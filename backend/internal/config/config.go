package config

import (
	"fmt"
	"os"
)

// BcryptCost は bcrypt ハッシュ生成コストの標準値。
// DefaultCost(10) は 2013 年基準で現代では不十分なため 12 を使用する。
const BcryptCost = 12

type Config struct {
	Port      string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string
	GinMode   string

	JWTSecret string

	// SMTP設定（空文字=無効）。LINE アクセストークン・通知先メールはクリニックごとに DB で管理する。
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	// FrontendURL はパスワードリセットメール等に含めるフロントエンドのベースURL。
	FrontendURL string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "ekarte_user"),
		DBPass:    getEnv("DB_PASSWORD", "ekarte_password"),
		DBName:    getEnv("DB_NAME", "ekarte_db"),
		DBSSLMode: getEnv("DB_SSL_MODE", "disable"),
		GinMode:   getEnv("GIN_MODE", "debug"),

		JWTSecret: getEnv("JWT_SECRET", "dev-secret-change-me"),

		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: getEnv("SMTP_PORT", "587"),
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
		SMTPFrom: os.Getenv("SMTP_FROM"),

		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
	}
}

// Validate は本番環境（GIN_MODE=release）での必須設定を検証する。
// 未設定の場合はエラーを返す。
func (c *Config) Validate() error {
	if c.GinMode != "release" {
		return nil
	}
	if c.JWTSecret == "" || c.JWTSecret == "dev-secret-change-me" {
		return fmt.Errorf("JWT_SECRET must be explicitly set in release mode")
	}
	if c.DBPass == "" || c.DBPass == "ekarte_password" {
		return fmt.Errorf("DB_PASSWORD must be explicitly set in release mode")
	}
	if c.SMTPHost != "" && c.SMTPPort != "465" && c.SMTPPort != "587" {
		return fmt.Errorf("SMTP_PORT must be 465 (TLS) or 587 (STARTTLS) in release mode, got %s", c.SMTPPort)
	}
	if os.Getenv("LIFF_MOCK") == "true" {
		return fmt.Errorf("LIFF_MOCK must not be set in release mode")
	}
	return nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
