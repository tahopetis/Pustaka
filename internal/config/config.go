package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Database   DatabaseConfig   `mapstructure:"database"`
	Neo4j      Neo4jConfig      `mapstructure:"neo4j"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Server     ServerConfig     `mapstructure:"server"`
	CORS       CORSConfig       `mapstructure:"cors"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Security   SecurityConfig   `mapstructure:"security"`
	Admin      AdminConfig      `mapstructure:"admin"`
	Env        string           `mapstructure:"environment"`
}

type DatabaseConfig struct {
	URL            string        `mapstructure:"url"`
	MaxOpenConns   int           `mapstructure:"max_open_conns"`
	MaxIdleConns   int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type Neo4jConfig struct {
	URI      string `mapstructure:"uri"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	MaxPool  int    `mapstructure:"max_pool"`
}

type RedisConfig struct {
	URL      string `mapstructure:"url"`
	Password string `mapstructure:"password"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret           string        `mapstructure:"secret"`
	AccessTokenTTL   time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `mapstructure:"refresh_token_ttl"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

type RateLimitConfig struct {
	RequestsPerMinute   int `mapstructure:"requests_per_minute"`
	AuthRequestsPerMinute int `mapstructure:"auth_requests_per_minute"`
}

type SecurityConfig struct {
	MaxUploadSize    string `mapstructure:"max_upload_size"`
	PasswordMinLen   int    `mapstructure:"password_min_length"`
	AccountLockout   int    `mapstructure:"account_lockout_attempts"`
	SessionTimeout   time.Duration `mapstructure:"session_timeout"`
}

type AdminConfig struct {
	Username string `mapstructure:"username"`
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file not found is not an error
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	// Set default values
	setDefaults()

	// Load config from environment variables
	viper.AutomaticEnv()

	// Bind environment variables to config keys
	viper.BindEnv("database.url", "DATABASE_URL", "PUSTAKA_DATABASE_URL")
	viper.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS", "PUSTAKA_DATABASE_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS", "PUSTAKA_DATABASE_MAX_IDLE_CONNS")
	viper.BindEnv("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME", "PUSTAKA_DATABASE_CONN_MAX_LIFETIME")

	viper.BindEnv("neo4j.uri", "NEO4J_URL", "PUSTAKA_NEO4J_URI")
	viper.BindEnv("neo4j.username", "NEO4J_USERNAME", "PUSTAKA_NEO4J_USERNAME")
	viper.BindEnv("neo4j.password", "NEO4J_PASSWORD", "PUSTAKA_NEO4J_PASSWORD")
	viper.BindEnv("neo4j.database", "NEO4J_DATABASE", "PUSTAKA_NEO4J_DATABASE")
	viper.BindEnv("neo4j.max_pool", "NEO4J_MAX_POOL", "PUSTAKA_NEO4J_MAX_POOL")

	viper.BindEnv("redis.url", "REDIS_URL", "PUSTAKA_REDIS_URL")
	viper.BindEnv("redis.password", "REDIS_PASSWORD", "PUSTAKA_REDIS_PASSWORD")
	viper.BindEnv("redis.pool_size", "REDIS_POOL_SIZE", "PUSTAKA_REDIS_POOL_SIZE")

	viper.BindEnv("jwt.secret", "JWT_SECRET", "PUSTAKA_JWT_SECRET")
	viper.BindEnv("jwt.access_token_ttl", "JWT_ACCESS_TOKEN_TTL", "PUSTAKA_JWT_ACCESS_TOKEN_TTL")
	viper.BindEnv("jwt.refresh_token_ttl", "JWT_REFRESH_TOKEN_TTL", "PUSTAKA_JWT_REFRESH_TOKEN_TTL")

	viper.BindEnv("server.host", "SERVER_HOST", "PUSTAKA_SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT", "PUSTAKA_SERVER_PORT")
	viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT", "PUSTAKA_SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT", "PUSTAKA_SERVER_WRITE_TIMEOUT")
	viper.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT", "PUSTAKA_SERVER_IDLE_TIMEOUT")

	viper.BindEnv("cors.allowed_origins", "CORS_ALLOWED_ORIGINS", "PUSTAKA_CORS_ALLOWED_ORIGINS")
	viper.BindEnv("cors.allowed_methods", "CORS_ALLOWED_METHODS", "PUSTAKA_CORS_ALLOWED_METHODS")
	viper.BindEnv("cors.allowed_headers", "CORS_ALLOWED_HEADERS", "PUSTAKA_CORS_ALLOWED_HEADERS")

	viper.BindEnv("logging.level", "LOG_LEVEL", "PUSTAKA_LOGGING_LEVEL")
	viper.BindEnv("logging.format", "LOG_FORMAT", "PUSTAKA_LOGGING_FORMAT")

	viper.BindEnv("admin.username", "ADMIN_USERNAME", "PUSTAKA_ADMIN_USERNAME")
	viper.BindEnv("admin.email", "ADMIN_EMAIL", "PUSTAKA_ADMIN_EMAIL")
	viper.BindEnv("admin.password", "ADMIN_PASSWORD", "PUSTAKA_ADMIN_PASSWORD")

	viper.BindEnv("environment", "ENVIRONMENT", "PUSTAKA_ENVIRONMENT")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate config
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Database defaults
	viper.SetDefault("database.url", "postgres://pustaka:password@localhost:5432/pustaka?sslmode=disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 25)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	// Neo4j defaults
	viper.SetDefault("neo4j.uri", "bolt://localhost:7687")
	viper.SetDefault("neo4j.username", "neo4j")
	viper.SetDefault("neo4j.password", "password")
	viper.SetDefault("neo4j.database", "neo4j")
	viper.SetDefault("neo4j.max_pool", 50)

	// Redis defaults
	viper.SetDefault("redis.url", "redis://localhost:6379")
	viper.SetDefault("redis.pool_size", 10)

	// JWT defaults
	viper.SetDefault("jwt.secret", "change-this-secret-key")
	viper.SetDefault("jwt.access_token_ttl", "24h")
	viper.SetDefault("jwt.refresh_token_ttl", "168h")

	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/metrics")

	// Rate limiting defaults
	viper.SetDefault("rate_limit.requests_per_minute", 1000)
	viper.SetDefault("rate_limit.auth_requests_per_minute", 60)

	// Security defaults
	viper.SetDefault("security.max_upload_size", "10MB")
	viper.SetDefault("security.password_min_length", 8)
	viper.SetDefault("security.account_lockout", 5)
	viper.SetDefault("security.session_timeout", "24h")

	// Admin defaults
	viper.SetDefault("admin.username", "admin")
	viper.SetDefault("admin.email", "admin@pustaka.dev")
	viper.SetDefault("admin.password", "Admin@123")

	// Environment defaults
	viper.SetDefault("environment", "development")
}

func validate(config *Config) error {
	if config.JWT.Secret == "change-this-secret-key" && config.Env == "production" {
		return fmt.Errorf("JWT secret must be changed in production")
	}

	if config.Database.URL == "" {
		return fmt.Errorf("database URL is required")
	}

	if config.Neo4j.URI == "" {
		return fmt.Errorf("neo4j URI is required")
	}

	if config.Server.Port <= 0 {
		return fmt.Errorf("server port must be positive")
	}

	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}