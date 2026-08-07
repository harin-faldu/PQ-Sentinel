package config

import (
	"os"
)

type Config struct {
	Port         string
	DBDriver     string // "mysql", "sqlite", "memory"
	DBSource     string
	AllowOrigins string
}

func LoadConfig() *Config {
	port := getEnv("PORT", "8080")
	dbDriver := getEnv("DB_DRIVER", "mysql")
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASSWORD", "harin")
	dbName := getEnv("DB_NAME", "qv_cbom")

	// Default MySQL DSN
	dbSource := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?parseTime=true"
	if customDSN := os.Getenv("DB_SOURCE"); customDSN != "" {
		dbSource = customDSN
	}

	return &Config{
		Port:         port,
		DBDriver:     dbDriver,
		DBSource:     dbSource,
		AllowOrigins: getEnv("ALLOW_ORIGINS", "*"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
