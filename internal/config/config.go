package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName        string
	Environment    string
	Port           string
	APIPrefix      string
	AllowedOrigins []string
	Database       DatabaseConfig
	JWT            JWTConfig
	Cloudinary     CloudinaryConfig
}

type DatabaseConfig struct {
	Enabled  bool
	DSN      string
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	Params   string
}

type JWTConfig struct {
	Secret         string
	ExpiresInHours int
}

type CloudinaryConfig struct {
	CloudName    string
	APIKey       string
	APISecret    string
	UploadFolder string
}

func (c *Config) GinMode() string {
	switch strings.ToLower(strings.TrimSpace(c.Environment)) {
	case "production":
		return "release"
	case "test":
		return "test"
	default:
		return "debug"
	}
}

func (c *Config) IsProduction() bool {
	return c.GinMode() == "release"
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppName:        getEnv("APP_NAME", "Absensi CN API"),
		Environment:    getEnv("APP_ENV", "development"),
		Port:           getEnv("APP_PORT", "8080"),
		APIPrefix:      getEnv("API_PREFIX", "/api/v1"),
		AllowedOrigins: splitCSV(getEnv("APP_ALLOWED_ORIGINS", "http://localhost:3000")),
		Database: DatabaseConfig{
			Enabled:  getEnvBool("DB_ENABLED", firstEnv("DB_DSN", "DATABASE_URL", "MYSQL_URL") != ""),
			DSN:      firstEnv("DB_DSN", "DATABASE_URL", "MYSQL_URL"),
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "3306"),
			Name:     getEnv("DB_NAME", "absensi_cn"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Params:   getEnv("DB_PARAMS", "charset=utf8mb4&parseTime=True&loc=Asia%2FJakarta"),
		},
		JWT: JWTConfig{
			Secret:         getEnv("JWT_SECRET", "development-secret"),
			ExpiresInHours: getEnvInt("JWT_EXPIRES_IN_HOURS", 24),
		},
		Cloudinary: CloudinaryConfig{
			CloudName:    getEnv("CLOUDINARY_CLOUD_NAME", ""),
			APIKey:       getEnv("CLOUDINARY_API_KEY", ""),
			APISecret:    getEnv("CLOUDINARY_API_SECRET", ""),
			UploadFolder: getEnv("CLOUDINARY_UPLOAD_FOLDER", "absensi-cn"),
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, errors.New("JWT_SECRET cannot be empty")
	}

	if cfg.APIPrefix == "" || cfg.APIPrefix[0] != '/' {
		return nil, fmt.Errorf("API_PREFIX must start with '/', got %q", cfg.APIPrefix)
	}

	if err := cfg.Cloudinary.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c CloudinaryConfig) Enabled() bool {
	return c.CloudName != "" && c.APIKey != "" && c.APISecret != ""
}

func (c CloudinaryConfig) Validate() error {
	values := []string{c.CloudName, c.APIKey, c.APISecret}
	filled := 0
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			filled++
		}
	}
	if filled > 0 && filled != len(values) {
		return errors.New("cloudinary config requires CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, and CLOUDINARY_API_SECRET together")
	}
	return nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}
	return ""
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func splitCSV(value string) []string {
	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))

	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
