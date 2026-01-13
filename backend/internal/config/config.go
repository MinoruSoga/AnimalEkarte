package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port     string
	DBHost   string
	DBPort   string
	DBUser   string
	DBPass   string
	DBName   string
	GinMode  string
}

func Load() *Config {
	return &Config{
		Port:    getEnv("PORT", "8080"),
		DBHost:  getEnv("DB_HOST", "localhost"),
		DBPort:  getEnv("DB_PORT", "5432"),
		DBUser:  getEnv("DB_USER", "ekarte_user"),
		DBPass:  getEnv("DB_PASSWORD", "ekarte_password"),
		DBName:  getEnv("DB_NAME", "ekarte_db"),
		GinMode: getEnv("GIN_MODE", "debug"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName,
	)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
