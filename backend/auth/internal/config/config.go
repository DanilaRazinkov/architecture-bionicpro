package config

import (
	"os"
	"strconv"
)

type Config struct {
	Keycloak KeycloakConfig
	Redis    RedisConfig
}

type KeycloakConfig struct {
	URL      string
	ClientID string
	Secret   string
	Realm    string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func Load() *Config {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	
	return &Config{
		Keycloak: KeycloakConfig{
			URL:      getOrDefault("KEYCLOAK_URL", "http://localhost:8080"),
			ClientID: getOrDefault("KEYCLOAK_CLIENT_ID", "backend-auth"),
			Secret:   getOrDefault("KEYCLOAK_SECRET", "oNwoLQdvJAvRcL89SydqCWCe5ry1jMgq"),
			Realm:    getOrDefault("KEYCLOAK_REALM", "reports-realm"),
		},
		Redis: RedisConfig{
			Host:     getOrDefault("REDIS_HOST", "localhost"),
			Port:     getOrDefault("REDIS_PORT", "6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       db,
		},
	}
}

func getOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}