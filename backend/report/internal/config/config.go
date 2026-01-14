package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ClickHouse ClickHouseConfig
	S3         S3Config
	CDN        CDNConfig
}

type ClickHouseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
	UseSSL          bool
}

type CDNConfig struct {
	Host string
}

func Load() *Config {
	return &Config{
		ClickHouse: ClickHouseConfig{
			Host:     getEnv("CLICKHOUSE_HOST", "olap_db"),
			Port:     getEnvAsInt("CLICKHOUSE_PORT", 9000),
			Database: getEnv("CLICKHOUSE_DATABASE", "default"),
			Username: getEnv("CLICKHOUSE_USERNAME", "default"),
			Password: getEnv("CLICKHOUSE_PASSWORD", ""),
		},
		S3: S3Config{
			Endpoint:        getEnv("S3_ENDPOINT", "minio:9000"),
			AccessKeyID:     getEnv("S3_ACCESS_KEY_ID", "minio_user"),
			SecretAccessKey: getEnv("S3_SECRET_ACCESS_KEY", "minio_password"),
			BucketName:      getEnv("S3_BUCKET_NAME", "reports"),
			Region:          getEnv("S3_REGION", "us-east-1"),
			UseSSL:          getEnvAsBool("S3_USE_SSL", false),
		},
		CDN: CDNConfig{
			Host: getEnv("CDN_HOST", "localhost:8888"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		value = strings.ToLower(value)
		return value == "true" || value == "1" || value == "yes" || value == "on"
	}
	return defaultValue
}
