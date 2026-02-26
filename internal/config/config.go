// Package config provides application configuration loading and validation.
package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration loaded from env/files and used across the service.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	AWS      AWSConfig
	Upload   UploadConfig
	SMTP     SMTPConfig
}

// ServerConfig contains HTTP server settings such as port and GinMode.
type ServerConfig struct {
	Port    string
	GinMode string
}

// DatabaseConfig contains database connection settings such as host, port,
// credentials, database name
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWTConfig contains settings used to issue and validate JWTs.
type JWTConfig struct {

	// Secret is the signing key used to sign and verify tokens.
	// For HS* algorithms this is a shared secret; keep it private and rotate when needed.
	Secret string

	// ExpiresIn is the access token time-to-live (TTL).
	ExpiresIn time.Duration

	// RefreshTokenExpires is the refresh token time-to-live (TTL).
	RefreshTokenExpires time.Duration
}

// AWSConfig contains AWS-related settings used by the application, including
// region, credentials (if not using IAM roles), and service-specific options.
type AWSConfig struct {
	// Region is the AWS region (e.g., "ap-south-1") used for service clients.
	Region string

	// AccessKeyID is the AWS access key ID.
	// Prefer IAM roles / workload identity in production instead of static keys.
	AccessKeyID string

	// SecretAccessKey is the AWS secret access key.
	// Do not log this value; load it from a secure secret store or environment variable.
	SecretAccessKey string

	// S3Bucket is the S3 bucket name used by the application.
	S3Bucket string

	// S3Endpoint is an optional custom endpoint (useful for S3-compatible storage or local testing).
	// Leave empty to use AWS default endpoints.
	S3Endpoint string

	// EventQueueName is the name of the queue used for event processing (e.g., SQS queue name).
	EventQueueName string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// UploadConfig contains settings for file uploads, including storage location,
// provider selection, and size limits.
type UploadConfig struct {
	Path        string
	MaxFileSize int64

	// UploadProvider  can be s3 or local
	UploadProvider string
}

// Load loads the application configuration from environment variables and/or
// configuration files and returns a validated Config.
func Load() (*Config, error) {
	_ = godotenv.Load()

	jwtExpiresIn, _ := time.ParseDuration(getEnv("JWT_EXPIRES_IN", "24h"))
	refreshTokenExpires, _ := time.ParseDuration(getEnv("REFRESH_TOKEN_EXPIRES_IN", "720h"))
	maxUploadSize, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "10485760"), 10, 64)
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "1025"))

	return &Config{
		Server: ServerConfig{
			Port:    getEnv("PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5434"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "ecommerce"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:              getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
			ExpiresIn:           jwtExpiresIn,
			RefreshTokenExpires: refreshTokenExpires,
		},
		AWS: AWSConfig{
			Region:          getEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", "test"),
			SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", "test"),
			S3Bucket:        getEnv("AWS_S3_BUCKET", "ecommerce-uploads"),
			S3Endpoint:      getEnv("AWS_S3_ENDPOINT", "http://localhost:4566"),
			EventQueueName:  getEnv("AWS_EVENT_QUEUE_NAME", "ecommerce-events"),
		},
		Upload: UploadConfig{
			Path:           getEnv("UPLOAD_PATH", "./uploads"),
			MaxFileSize:    maxUploadSize,
			UploadProvider: getEnv("UPLOAD_PROVIDER", "local"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "localhost"),
			Port:     smtpPort,
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@shop.com"),
		},
	}, nil

}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
