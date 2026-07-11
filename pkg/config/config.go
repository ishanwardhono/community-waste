package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppPort            string
	LogLevel           string
	DB                 DBConfig
	S3                 S3Config
	ShutdownTimeout    time.Duration
	AutoCancelInterval time.Duration
	AutoCancelMaxAge   time.Duration
}

type S3Config struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	Bucket         string
	UseSSL         bool
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

func Load() (*Config, error) {
	var missing []string
	env := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			missing = append(missing, key)
		}
		return v
	}

	cfg := &Config{
		AppPort:  env("APP_PORT"),
		LogLevel: env("LOG_LEVEL"),
		DB: DBConfig{
			Host:     env("DB_HOST"),
			Port:     env("DB_PORT"),
			User:     env("DB_USER"),
			Password: env("DB_PASSWORD"),
			Name:     env("DB_NAME"),
			SSLMode:  env("DB_SSLMODE"),
		},
		S3: S3Config{
			Endpoint:       env("S3_ENDPOINT"),
			PublicEndpoint: env("S3_PUBLIC_ENDPOINT"),
			AccessKey:      env("S3_ACCESS_KEY"),
			SecretKey:      env("S3_SECRET_KEY"),
			Bucket:         env("S3_BUCKET"),
		},
	}
	shutdownRaw := env("SHUTDOWN_TIMEOUT")
	s3SSLRaw := env("S3_USE_SSL")
	autoCancelIntervalRaw := env("AUTOCANCEL_INTERVAL")
	autoCancelMaxAgeRaw := env("AUTOCANCEL_MAX_AGE")

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing env vars: %s, provide an .env file (see .env.example)",
			strings.Join(missing, ", "))
	}

	shutdown, err := time.ParseDuration(shutdownRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
	}
	cfg.ShutdownTimeout = shutdown

	autoCancelInterval, err := time.ParseDuration(autoCancelIntervalRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid AUTOCANCEL_INTERVAL: %w", err)
	}
	cfg.AutoCancelInterval = autoCancelInterval

	autoCancelMaxAge, err := time.ParseDuration(autoCancelMaxAgeRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid AUTOCANCEL_MAX_AGE: %w", err)
	}
	cfg.AutoCancelMaxAge = autoCancelMaxAge

	s3SSL, err := strconv.ParseBool(s3SSLRaw)
	if err != nil {
		return nil, fmt.Errorf("invalid S3_USE_SSL: %w", err)
	}
	cfg.S3.UseSSL = s3SSL

	return cfg, nil
}
