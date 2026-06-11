// Package config loads strongly-typed application configuration from the
// environment (12-factor). A single Config struct is constructed once at
// startup and injected down the dependency graph — no global state, no
// scattered os.Getenv calls in business code.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment is the deployment environment the process believes it is in.
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

// Config is the root configuration aggregate, composed of per-concern sub-configs.
type Config struct {
	Env      Environment
	LogLevel string
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	MQTT     MQTTConfig
	JWT      JWTConfig
	Ingest   IngestConfig
}

// IngestConfig tunes the telemetry pipeline.
type IngestConfig struct {
	Workers           int
	QueueSize         int
	HeartbeatInterval time.Duration
	OfflineAfter      time.Duration
}

// HTTPConfig configures the Gin HTTP server.
type HTTPConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
}

// Addr returns the host:port the server binds to.
func (h HTTPConfig) Addr() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

// PostgresConfig describes the system-of-record connection (used from Phase 3).
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN renders a libpq connection string.
func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode,
	)
}

// RedisConfig describes the cache / hot-path store (used from Phase 5).
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// MQTTConfig describes the broker connection (used from Phase 5).
type MQTTConfig struct {
	Host     string
	Port     int
	ClientID string
	Username string
	Password string
}

// JWTConfig holds token signing material (used from Phase 4).
type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

// Load reads configuration from the environment, applying sane development
// defaults. It fails fast with a descriptive error if a required value is
// malformed, so misconfiguration is caught at boot rather than at first use.
func Load() (*Config, error) {
	cfg := &Config{
		Env:      Environment(getEnv("APP_ENV", string(EnvDevelopment))),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		HTTP: HTTPConfig{
			Host:            getEnv("API_HOST", "0.0.0.0"),
			Port:            getEnvInt("API_PORT", 8080),
			ReadTimeout:     getEnvDuration("API_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvDuration("API_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getEnvDuration("API_SHUTDOWN_TIMEOUT", 10*time.Second),
			AllowedOrigins:  getEnvList("API_CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "iot"),
			Password: getEnv("POSTGRES_PASSWORD", "iot_dev_password"),
			DBName:   getEnv("POSTGRES_DB", "iot_dashboard"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		MQTT: MQTTConfig{
			Host:     getEnv("MQTT_HOST", "localhost"),
			Port:     getEnvInt("MQTT_PORT", 1883),
			ClientID: getEnv("MQTT_CLIENT_ID", "iot-backend"),
			Username: getEnv("MQTT_USERNAME", ""),
			Password: getEnv("MQTT_PASSWORD", ""),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "dev_access_secret_change_me"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "dev_refresh_secret_change_me"),
			AccessTTL:     getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    getEnvDuration("JWT_REFRESH_TTL", 168*time.Hour),
		},
		Ingest: IngestConfig{
			Workers:           getEnvInt("INGEST_WORKERS", 4),
			QueueSize:         getEnvInt("INGEST_QUEUE_SIZE", 1024),
			HeartbeatInterval: getEnvDuration("INGEST_HEARTBEAT_INTERVAL", 30*time.Second),
			OfflineAfter:      getEnvDuration("INGEST_OFFLINE_AFTER", 90*time.Second),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// IsProduction reports whether the process is running in production.
func (c *Config) IsProduction() bool { return c.Env == EnvProduction }

func (c *Config) validate() error {
	if c.HTTP.Port <= 0 || c.HTTP.Port > 65535 {
		return fmt.Errorf("config: invalid API_PORT %d", c.HTTP.Port)
	}
	if c.IsProduction() {
		if c.JWT.AccessSecret == "dev_access_secret_change_me" {
			return fmt.Errorf("config: JWT_ACCESS_SECRET must be set in production")
		}
	}
	return nil
}

// --- typed env helpers -------------------------------------------------------

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvList(key string, fallback []string) []string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				out = append(out, t)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return fallback
}
