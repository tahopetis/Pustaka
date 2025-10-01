package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/pustaka/pustaka/internal/api"
	"github.com/pustaka/pustaka/internal/api/handlers"
	"github.com/pustaka/pustaka/internal/api/middleware"
	"github.com/pustaka/pustaka/internal/auth"
	"github.com/pustaka/pustaka/internal/ci"
	"github.com/pustaka/pustaka/internal/config"
	"github.com/pustaka/pustaka/internal/database"
	pustakaLogger "github.com/pustaka/pustaka/pkg/logger"
)

func initializeAdminUser(pool *pgxpool.Pool, rbacService *auth.RBACService, passwordService *auth.PasswordService, adminConfig config.AdminConfig, logger *pustakaLogger.Logger) error {
	ctx := context.Background()

	// Hash the admin password
	hashedPassword, err := passwordService.HashPassword(adminConfig.Password)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Check if admin user already exists
	var userID uuid.UUID
	var existingEmail string
	err = pool.QueryRow(ctx, "SELECT id, email FROM users WHERE username = $1", adminConfig.Username).Scan(&userID, &existingEmail)

	if err == nil {
		// Admin user exists, update password and email to match config
		_, err = pool.Exec(ctx, `
			UPDATE users
			SET password_hash = $1, email = $2, updated_at = $3
			WHERE username = $4
		`, hashedPassword, adminConfig.Email, time.Now(), adminConfig.Username)

		if err != nil {
			return fmt.Errorf("failed to update admin user: %w", err)
		}

		logger.Info().Str("username", adminConfig.Username).Msg("Admin user updated successfully")
	} else if err.Error() == "no rows in result set" {
		// Admin user doesn't exist, create it
		userID = uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO users (id, username, email, password_hash, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, userID, adminConfig.Username, adminConfig.Email, hashedPassword, true, time.Now(), time.Now())

		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		logger.Info().
			Str("username", adminConfig.Username).
			Str("email", adminConfig.Email).
			Msg("Admin user created successfully")
	} else {
		return fmt.Errorf("failed to check admin user existence: %w", err)
	}

	// Get admin role
	var adminRoleID uuid.UUID
	err = pool.QueryRow(ctx, "SELECT id FROM roles WHERE name = 'admin'").Scan(&adminRoleID)
	if err != nil {
		return fmt.Errorf("failed to get admin role: %w", err)
	}

	// Assign admin role to user (ensure role is assigned)
	_, err = pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, role_id) DO UPDATE SET created_at = $3
	`, userID, adminRoleID, time.Now())

	if err != nil {
		return fmt.Errorf("failed to assign admin role: %w", err)
	}

	logger.Info().
		Str("username", adminConfig.Username).
		Msg("Admin role assigned successfully")

	return nil
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := pustakaLogger.New(pustakaLogger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})

	logger.Info().
		Str("version", "1.0.0").
		Str("environment", cfg.Env).
		Msg("Starting Pustaka API Server")

	// Initialize databases
	postgresDB, err := database.NewPostgresDB(
		cfg.Database.URL,
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
		cfg.Database.ConnMaxLifetime,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer postgresDB.Close()

	neo4jDB, err := database.NewNeo4jDB(
		cfg.Neo4j.URI,
		cfg.Neo4j.Username,
		cfg.Neo4j.Password,
		cfg.Neo4j.Database,
		cfg.Neo4j.MaxPool,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Neo4j")
	}
	defer neo4jDB.Close()

	// Initialize Neo4j indexes
	if err := neo4jDB.InitializeIndexes(context.Background()); err != nil {
		logger.Error().Err(err).Msg("Failed to initialize Neo4j indexes")
	}

	redisDB, err := database.NewRedisDB(
		cfg.Redis.URL,
		cfg.Redis.Password,
		cfg.Redis.PoolSize,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisDB.Close()

	// Initialize services
	jwtService := auth.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		"pustaka",
	)

	passwordService := auth.NewPasswordService()

	rbacService := auth.NewRBACService(postgresDB.Pool)

	// Create CI services
	ciRepo := ci.NewRepository(postgresDB.Pool, logger)
	neo4jService := ci.NewNeo4jService(neo4jDB.Driver, logger)
	ciService := ci.NewService(ciRepo, neo4jService, redisDB.Client, logger)

	// Create audit service
	auditRepo := ci.NewAuditLogRepository(postgresDB.Pool, logger)
	auditService := ci.NewAuditService(auditRepo, logger)

	// Initialize admin user
	if err := initializeAdminUser(postgresDB.Pool, rbacService, passwordService, cfg.Admin, logger); err != nil {
		logger.Error().Err(err).Msg("Failed to initialize admin user")
	}

	// Initialize handlers
	baseHandler := api.NewHandler(logger)
	authHandler := handlers.NewAuthHandler(jwtService, passwordService, rbacService, logger)
	userHandler := handlers.NewUserHandler(rbacService, passwordService, logger)
	ciHandlers := api.NewCIHandlers(baseHandler, ciService)
	ciTypeHandlers := api.NewCITypeHandlers(baseHandler, ciService)
	relationshipHandlers := api.NewRelationshipHandlers(baseHandler, ciService)
	auditHandlers := api.NewAuditHandlers(baseHandler, auditService)

	// Setup router
	router := setupRouter(cfg, logger, authHandler, userHandler, ciHandlers, ciTypeHandlers, relationshipHandlers, auditHandlers, jwtService, rbacService)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info().
			Str("address", server.Addr).
			Msg("HTTP server starting")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down HTTP server...")

	// Shutdown server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to shutdown HTTP server gracefully")
	}

	logger.Info().Msg("HTTP server stopped")
}

func setupRouter(
	cfg *config.Config,
	logger *pustakaLogger.Logger,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	ciHandlers *api.CIHandlers,
	ciTypeHandlers *api.CITypeHandlers,
	relationshipHandlers *api.RelationshipHandlers,
	auditHandlers *api.AuditHandlers,
	jwtService *auth.JWTService,
	rbacService *auth.RBACService,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))
	r.Use(chiMiddleware.AllowContentType("application/json"))
	r.Use(chiMiddleware.CleanPath)

	// Custom middleware
	r.Use(middleware.Logger)
	r.Use(middleware.CORS(
		cfg.CORS.AllowedOrigins,
		cfg.CORS.AllowedMethods,
		cfg.CORS.AllowedHeaders,
	))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s"}`, time.Now().UTC().Format(time.RFC3339))
	})

	// Metrics endpoint (if enabled)
	if cfg.Metrics.Enabled {
		// TODO: Setup Prometheus metrics
		// r.Handle(cfg.Metrics.Path, promhttp.Handler())
	}

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			// JWT authentication middleware
			r.Use(middleware.JWTAuth(jwtService))

			// Audit logging middleware
			r.Use(middleware.AuditLogging(rbacService, logger))

			// User routes
			r.Route("/users", func(r chi.Router) {
				r.Use(middleware.RBAC("user:read"))
				r.Get("/", userHandler.ListUsers)
				r.Get("/{id}", userHandler.GetUser)

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("user:create"))
					r.Post("/", userHandler.CreateUser)
				})

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("user:update"))
					r.Put("/{id}", userHandler.UpdateUser)
				})

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("user:delete"))
					r.Delete("/{id}", userHandler.DeleteUser)
				})
			})

			// CI Type routes
			r.Route("/ci-types", func(r chi.Router) {
				r.Use(middleware.RBAC("ci_type:read"))
				r.Get("/", ciTypeHandlers.ListCITypes)
				r.Get("/{id}", ciTypeHandlers.GetCIType)

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("ci_type:create"))
					r.Post("/", ciTypeHandlers.CreateCIType)
				})

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("ci_type:update"))
					r.Put("/{id}", ciTypeHandlers.UpdateCIType)
				})

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("ci_type:delete"))
					r.Delete("/{id}", ciTypeHandlers.DeleteCIType)
				})
			})

			// Configuration Item routes
			r.Route("/ci", func(r chi.Router) {
				r.Use(middleware.RBAC("ci:read"))
				r.Get("/", ciHandlers.ListCIs)
				r.Get("/{id}", ciHandlers.GetCI)

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("ci:create"))
					r.Post("/", ciHandlers.CreateCI)
				})

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("ci:update"))
					r.Put("/{id}", ciHandlers.UpdateCI)
				})

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("ci:delete"))
					r.Delete("/{id}", ciHandlers.DeleteCI)
				})
			})

			// Relationship routes
			r.Route("/relationships", func(r chi.Router) {
				r.Use(middleware.RBAC("relationship:read"))
				r.Get("/", relationshipHandlers.ListRelationships)
				r.Get("/{id}", relationshipHandlers.GetRelationship)

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("relationship:create"))
					r.Post("/", relationshipHandlers.CreateRelationship)
				})

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("relationship:update"))
					r.Put("/{id}", relationshipHandlers.UpdateRelationship)
				})

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("relationship:delete"))
					r.Delete("/{id}", relationshipHandlers.DeleteRelationship)
				})
			})

			// Graph routes
			r.Route("/graph", func(r chi.Router) {
				r.Use(middleware.RBAC("ci:read"))
				r.Get("/", ciHandlers.GetGraphData)
				r.Get("/explore", ciHandlers.ExploreGraph)
			})

			// CI relationship routes
			r.Route("/ci/{id}/relationships", func(r chi.Router) {
				r.Use(middleware.RBAC("relationship:read"))
				r.Get("/", ciHandlers.GetCIRelationships)
			})

			// Audit routes
			r.Route("/audit", func(r chi.Router) {
				r.Use(middleware.RBAC("audit:read"))
				r.Get("/logs", auditHandlers.ListAuditLogs)
				r.Get("/logs/{id}", auditHandlers.GetAuditLog)
				r.Get("/stats", auditHandlers.GetAuditStats)
				r.Get("/export", auditHandlers.ExportAuditLogs)

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("audit:delete"))
					r.Delete("/logs/{id}", auditHandlers.DeleteAuditLog)
					r.Delete("/cleanup", auditHandlers.CleanupOldAuditLogs)
				})
			})

			// Analytics routes
			r.Route("/analytics", func(r chi.Router) {
				r.Use(middleware.RBAC("ci:read"))
				r.Get("/ci-types/usage", ciTypeHandlers.GetCITypesByUsage)
				r.Get("/cycles", relationshipHandlers.FindCycles)
				r.Get("/most-connected", relationshipHandlers.GetMostConnectedCIs)
			})

			// Current user profile
			r.Get("/me", authHandler.GetCurrentUser)

			// Dashboard routes
			r.Route("/dashboard", func(r chi.Router) {
				r.Use(middleware.RBAC("ci:read"))
				r.Get("/stats", ciHandlers.GetDashboardStats)
			})
		})
	})

	return r
}