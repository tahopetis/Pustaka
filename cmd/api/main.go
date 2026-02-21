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

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pustaka/pustaka/internal/amortization"
	"github.com/pustaka/pustaka/internal/api"
	"github.com/pustaka/pustaka/internal/api/handlers"
	"github.com/pustaka/pustaka/internal/api/middleware"
	"github.com/pustaka/pustaka/internal/auth"
	"github.com/pustaka/pustaka/internal/ci"
	"github.com/pustaka/pustaka/internal/config"
	"github.com/pustaka/pustaka/internal/database"
	"github.com/pustaka/pustaka/internal/ea"
	"github.com/pustaka/pustaka/internal/repository"
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
		// Check if the error is because the users table doesn't exist (fresh database)
		if err.Error() == "ERROR: relation \"users\" does not exist (SQLSTATE 42P01)" ||
		   err.Error() == "relation \"users\" does not exist" {
			logger.Info().Msg("Users table doesn't exist yet, skipping admin user initialization")
			return nil
		}
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

	// Initialize databases first (needed for admin user creation)
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

	// Initialize minimal services needed for admin user creation
	passwordService := auth.NewPasswordService()
	rbacService := auth.NewRBACService(postgresDB.Pool)

	// Try to initialize admin user before running migrations (for existing databases)
	// This will skip if tables don't exist yet (fresh database)
	logger.Info().Msg("Checking for existing admin user...")
	if err := initializeAdminUser(postgresDB.Pool, rbacService, passwordService, cfg.Admin, logger); err != nil {
		logger.Warn().Err(err).Msg("Failed to check/admin user, will retry after migrations")
	}

	// Run database migrations
	// For fresh databases, this creates all tables including users
	// For existing databases, this applies any pending migrations
	logger.Info().Msg("Running database migrations...")
	migrationsPath := "cmd/migrations"
	if err := database.RunMigrations(cfg.Database.URL, migrationsPath, logger); err != nil {
		logger.Fatal().Err(err).Msg("Failed to run database migrations")
	}
	logger.Info().Msg("Database migrations completed successfully")

	// Initialize admin user after migrations (ensures admin user exists)
	// This is required because migration 009 references the admin user
	logger.Info().Msg("Initializing admin user...")
	if err := initializeAdminUser(postgresDB.Pool, rbacService, passwordService, cfg.Admin, logger); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize admin user")
	}
	logger.Info().Msg("Admin user initialized successfully")

	// Initialize remaining databases
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

	// Create CI services
	ciRepo := ci.NewRepository(postgresDB.Pool, logger)
	neo4jService := ci.NewNeo4jService(neo4jDB.Driver, logger)

	// Create audit service
	auditRepo := ci.NewAuditLogRepository(postgresDB.Pool, logger)
	auditService := ci.NewAuditService(auditRepo, logger)

	ciService := ci.NewService(ciRepo, neo4jService, redisDB.Client, auditService, logger)

	// Create relationship type services
	relationshipTypeRepo := ci.NewRelationshipTypeRepository(postgresDB.Pool)
	relationshipTypeService := ci.NewRelationshipTypeService(relationshipTypeRepo, ciRepo, neo4jService, redisDB.Client, auditService, logger, postgresDB.Pool)

	// Create lifecycle status services
	lifecycleStatusRepo := ci.NewLifecycleStatusRepository(postgresDB.Pool)
	lifecycleStatusService := ci.NewLifecycleStatusService(lifecycleStatusRepo, ciRepo, redisDB.Client, auditService, logger)

	// Create amortization services
	amortizationRepo := amortization.NewRepository(postgresDB.Pool, logger)
	amortizationCache := amortization.NewCacheRepository(redisDB.Client, logger)
	amortizationCalculator := amortization.NewCalculator(logger)
	amortizationScheduler := amortization.NewScheduler(amortizationRepo, amortizationCache, logger)

	// Create EA services
	eaRepo := ea.NewRepository(postgresDB.Pool)
	eaService := ea.NewService(ciService, ciRepo, eaRepo, neo4jService, redisDB.Client, auditService, logger)
	eaImportService := ea.NewImportService(eaRepo, ciRepo, lifecycleStatusRepo, logger)

	// Create EA data quality repository
	qualityRepo := repository.NewEADataQualityRepository(postgresDB.Pool, logger)

	// Create adapters for CI service interfaces
	ciAdapter := &amortization.CIServiceAdapter{
		Service: ciService,
	}

	lifecycleAdapter := &amortization.LifecycleServiceAdapter{
		Service: lifecycleStatusService,
	}

	// Create simple stub implementations for missing interfaces
	amortizationEventPublisher := amortization.NewEventPublisher(logger)
	amortizationAuditLogger := amortization.NewAuditLogger(auditService, logger)

	// Create amortization service with proper interface implementations
	amortizationService := amortization.NewAmortizationService(
		amortizationRepo,
		amortizationCache,
		amortizationCalculator,
		amortizationScheduler,
		amortizationEventPublisher,
		amortizationAuditLogger,
		ciAdapter,
		lifecycleAdapter,
		logger,
	)

	// Initialize handlers
	baseHandler := api.NewHandler(logger)
	authHandler := handlers.NewAuthHandler(jwtService, passwordService, rbacService, logger)
	userHandler := handlers.NewUserHandler(rbacService, passwordService, logger)
	ciHandlers := api.NewCIHandlers(baseHandler, ciService)
	ciTypeHandlers := api.NewCITypeHandlers(baseHandler, ciService)
	relationshipHandlers := api.NewRelationshipHandlers(baseHandler, ciService)
	relationshipTypeHandlers := handlers.NewRelationshipTypeHandler(relationshipTypeService, rbacService, logger)
	lifecycleStatusHandlers := handlers.NewLifecycleStatusHandler(lifecycleStatusService, rbacService, logger)
	auditHandlers := api.NewAuditHandlers(baseHandler, auditService)
	amortizationHandlers := handlers.NewAmortizationHandler(amortizationService, logger)
	eaHandlers := handlers.NewEAHandlers(eaService, logger)
	importHandlers := handlers.NewImportHandlers(eaImportService, logger)
	dataQualityHandlers := handlers.NewDataQualityHandlers(qualityRepo, logger)

	// Get base URL for QR codes (from server config or default)
	baseURL := cfg.Server.BaseURL
	if baseURL == "" {
		baseURL = "http://" + cfg.Server.Host + fmt.Sprintf(":%d", cfg.Server.Port)
	}

	// Initialize QR handlers
	qrHandlers := api.NewQRHandlers(baseHandler, ciService, logger, baseURL)

	// Setup router
	router := setupRouter(cfg, logger, authHandler, userHandler, ciHandlers, ciTypeHandlers, relationshipHandlers, relationshipTypeHandlers, lifecycleStatusHandlers, auditHandlers, amortizationHandlers, eaHandlers, importHandlers, dataQualityHandlers, qrHandlers, jwtService, rbacService)

	// Start amortization scheduler in background
	go func() {
		logger.Info().Msg("Starting amortization scheduler")
		if err := amortizationScheduler.ScheduleDailyRun(context.Background()); err != nil {
			logger.Error().Err(err).Msg("Failed to start amortization scheduler")
		}
	}()

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

	// Shutdown amortization scheduler
	if err := amortizationScheduler.UnscheduleDailyRun(); err != nil {
		logger.Error().Err(err).Msg("Failed to stop amortization scheduler")
	}

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
	relationshipTypeHandlers *handlers.RelationshipTypeHandler,
	lifecycleStatusHandlers *handlers.LifecycleStatusHandler,
	auditHandlers *api.AuditHandlers,
	amortizationHandlers *handlers.AmortizationHandler,
	eaHandlers *handlers.EAHandlers,
	importHandlers *handlers.ImportHandlers,
	dataQualityHandlers *handlers.DataQualityHandlers,
	qrHandlers *api.QRHandlers,
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

			// Activity tracking middleware
			r.Use(middleware.ActivityTracker(rbacService))

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
					r.Post("/bulk-sources", relationshipHandlers.CreateRelationshipsFromSources)
					r.Post("/bulk-matrix", relationshipHandlers.CreateRelationshipsMatrix)
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

			// Relationship Type routes
			r.Route("/relationship-types", func(r chi.Router) {
				r.Use(middleware.RBAC("relationship_type:read"))
				r.Get("/", relationshipTypeHandlers.ListRelationshipTypes)
				r.Get("/active", relationshipTypeHandlers.GetActiveRelationshipTypes)
				r.Get("/usage", relationshipTypeHandlers.GetRelationshipTypeUsage)
				r.Get("/statistics", relationshipTypeHandlers.GetRelationshipTypeStatistics)
				r.Get("/categories", relationshipTypeHandlers.GetRelationshipTypeCategories)
				r.Post("/validate", relationshipTypeHandlers.ValidateRelationship)
				r.Get("/{id}", relationshipTypeHandlers.GetRelationshipType)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("relationship_type:create"))
					r.Post("/", relationshipTypeHandlers.CreateRelationshipType)
				})
				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("relationship_type:update"))
					r.Put("/{id}", relationshipTypeHandlers.UpdateRelationshipType)
				})
				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("relationship_type:delete"))
					r.Delete("/{id}", relationshipTypeHandlers.DeleteRelationshipType)
				})
			})

			// Lifecycle Status routes
			r.Route("/lifecycle-status", func(r chi.Router) {
				r.Use(middleware.RBAC("lifecycle_status:read"))
				r.Get("/", lifecycleStatusHandlers.ListLifecycleStatuses)
				r.Get("/active", lifecycleStatusHandlers.GetActiveLifecycleStatuses)
				r.Get("/usage", lifecycleStatusHandlers.GetLifecycleStatusUsage)
				r.Get("/distribution", lifecycleStatusHandlers.GetCIStatusDistribution)
				r.Get("/{id}", lifecycleStatusHandlers.GetLifecycleStatus)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("lifecycle_status:create"))
					r.Post("/", lifecycleStatusHandlers.CreateLifecycleStatus)
				})
				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("lifecycle_status:update"))
					r.Put("/{id}", lifecycleStatusHandlers.UpdateLifecycleStatus)
				})
				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("lifecycle_status:delete"))
					r.Delete("/{id}", lifecycleStatusHandlers.DeleteLifecycleStatus)
				})
			})

			// Amortization routes
			r.Route("/amortization", func(r chi.Router) {
				// Read access for most operations
				r.Use(middleware.RBAC("amortization:read"))

				// Configuration Items with amortization
				r.Get("/configuration-items", amortizationHandlers.ListAmortizableCIs)
				r.Get("/configuration-items/{ciId}", amortizationHandlers.GetAmortizationDetails)

				// Ledger management
				r.Get("/ledger", amortizationHandlers.GetLedgerEntries)
				r.Get("/ledger/{entryId}", amortizationHandlers.GetLedgerEntry)

				// Amortization runs
				r.Get("/runs", amortizationHandlers.ListAmortizationRuns)
				r.Get("/runs/{runId}", amortizationHandlers.GetAmortizationRun)

				// Reports and summaries
				r.Get("/summaries", amortizationHandlers.GetAmortizationSummaries)
				r.Get("/reports/depreciation-schedule", amortizationHandlers.GenerateDepreciationSchedule)
				r.Get("/settings", amortizationHandlers.GetAmortizationSettings)

				// Routes requiring additional permissions
				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("amortization:configure"))
					r.Put("/configuration-items/{ciId}", amortizationHandlers.UpdateAmortizationConfig)
					r.Post("/adjustments", amortizationHandlers.CreateAdjustment)
					r.Put("/settings", amortizationHandlers.UpdateAmortizationSettings)

					// Restructuring routes (require configure permission)
					r.Post("/restructuring/preview", amortizationHandlers.PreviewRestructuring)
					r.Post("/restructuring", amortizationHandlers.ExecuteRestructuring)
				})

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("amortization:run"))
					r.Post("/runs", amortizationHandlers.TriggerManualRun)
				})
			})

			// Graph routes
			r.Route("/graph", func(r chi.Router) {
				r.Use(middleware.RBAC("ci:read"))
				r.Get("/", ciHandlers.GetGraphData)
				r.Get("/explore", ciHandlers.ExploreGraph)
			})

			// QR Code routes
			r.Route("/qr", func(r chi.Router) {
				r.Use(middleware.RBAC("ci:read"))
				r.Get("/ci/{id}", qrHandlers.GetCIQRCode)
				r.Get("/ci/{id}/image", qrHandlers.GetCIQRCodeImage)
			})

			// CI relationship routes
			r.Route("/ci/{id}/relationships", func(r chi.Router) {
				r.Use(middleware.RBAC("relationship:read"))
				r.Get("/", ciHandlers.GetCIRelationships)
			})

			// Audit routes
			r.Route("/audit", func(r chi.Router) {
				r.Use(middleware.RBAC("audit:read"))
				r.Get("/", auditHandlers.ListAuditLogsFrontend) // For frontend compatibility
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
				r.Get("/ci-growth", ciHandlers.GetCIGrowth)
				r.Get("/relationship-types/usage", relationshipTypeHandlers.GetRelationshipTypeUsage)
			})

			// Current user profile
			r.Get("/me", authHandler.GetCurrentUser)

			// EA Entity routes
			r.Route("/ea/entities", func(r chi.Router) {
				r.Use(middleware.RBAC("ea:read"))
				r.Get("/", eaHandlers.ListEAEntities)
				r.Get("/{id}", eaHandlers.GetEAEntity)
				r.Get("/{id}/validate", eaHandlers.ValidateEAEntity)
				r.Get("/{id}/audit", eaHandlers.GetEAEntityAuditLogs)

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("ea:create"))
					r.Post("/", eaHandlers.CreateEAEntity)
				})

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("ea:update"))
					r.Put("/{id}", eaHandlers.UpdateEAEntity)
				})

				r.Group(func(r chi.Router) {
					r.Use(middleware.RBAC("ea:delete"))
					r.Delete("/{id}", eaHandlers.DeleteEAEntity)
				})
			})

			// EA Import routes
			r.Route("/ea/import", func(r chi.Router) {
				r.Use(middleware.RBAC("ea:create"))

				// Template generation
				r.Post("/template", importHandlers.GenerateImportTemplate)

				// Import validation and execution
				r.Post("/validate", importHandlers.ValidateImport)
				r.Post("/execute", importHandlers.ExecuteImport)

				// Error export
				r.Post("/errors/download", importHandlers.DownloadErrorCSV)

				// Import status (for async imports)
				r.Get("/status/{batch_id}", importHandlers.GetImportStatus)
			})

			// EA Data Quality routes
			r.Route("/ea/data-quality", func(r chi.Router) {
				r.Use(middleware.RBAC("ea:read"))
				r.Get("/", dataQualityHandlers.GetDataQualityMetrics)
				r.Get("/stale", dataQualityHandlers.GetStaleEntities)
				r.Get("/errors", dataQualityHandlers.GetEntitiesWithErrors)
				r.Get("/lifecycle", dataQualityHandlers.GetLifecycleBreakdown)
			})

			// Dashboard routes
			r.Route("/dashboard", func(r chi.Router) {
				r.Use(middleware.RBAC("ci:read"))
				r.Get("/stats", ciHandlers.GetDashboardStats)
				r.Get("/health-score", ciHandlers.GetHealthScore)
				r.Get("/data-quality", ciHandlers.GetDataQualityMetrics)
				r.Get("/asset-aging", ciHandlers.GetAssetAgingMetrics)
				r.Get("/risk-metrics", ciHandlers.GetRiskMetrics)
			})
		})
	})

	// Public routes (no authentication required)
	r.Route("/public", func(r chi.Router) {
		r.Get("/ci/{id}", qrHandlers.GetPublicCI)
	})

	return r
}
