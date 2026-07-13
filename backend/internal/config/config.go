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

	// IntegrationEncryptionKey はクリニック連携設定（LステップAPIキー等）をAES-256-GCMで暗号化するキー。
	// 32バイトのhex文字列（64文字）。本番では必須。
	IntegrationEncryptionKey string

	// S3 ファイルストレージ設定（shared_files用）
	S3SharedBucket string
	S3SharedRegion string

	// S3Endpoint はカスタム S3 互換エンドポイント（Cloudflare R2 等）。
	// 空文字（既定）の場合は AWS S3 のリージョナルエンドポイント・バーチャルホスト形式を維持する。
	S3Endpoint string

	// StorageType はファイルアップロード先を切り替える ("s3" または空文字=ローカル)。
	StorageType string
	// S3Bucket/S3Region はアップローダー用 S3 バケット（StorageType=s3 のとき必須）。
	// shared_files 用の S3SharedBucket/S3SharedRegion とは別設定。
	S3Bucket string
	S3Region string

	// TrustedProxyCIDR は release モードで rate-limit バイパス防止のため必須の ALB CIDR。
	TrustedProxyCIDR string

	// LogLevel はロガー初期化時のレベル切り替えに使用する ("debug" でデバッグレベル)。
	LogLevel string

	// CORSAllowedOrigin は CORS 許可オリジン（カンマ区切り）。空文字ならミドルウェア側で開発既定値を適用する。
	CORSAllowedOrigin string
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

		IntegrationEncryptionKey: os.Getenv("INTEGRATION_ENCRYPTION_KEY"),
		S3SharedBucket:           os.Getenv("S3_SHARED_BUCKET"),
		S3SharedRegion:           getEnv("S3_SHARED_REGION", "ap-northeast-1"),
		S3Endpoint:               os.Getenv("S3_ENDPOINT"),

		StorageType: os.Getenv("STORAGE_TYPE"),
		S3Bucket:    os.Getenv("S3_BUCKET"),
		S3Region:    os.Getenv("S3_REGION"),

		TrustedProxyCIDR: os.Getenv("TRUSTED_PROXY_CIDR"),

		LogLevel: os.Getenv("LOG_LEVEL"),

		CORSAllowedOrigin: os.Getenv("CORS_ALLOWED_ORIGIN"),
	}
}

// Validate は起動設定を検証する。
// dev デフォルト値チェックは全モードで実行する（staging 認証バイパス防止）。
func (c *Config) Validate() error {
	if c.JWTSecret == "dev-secret-change-me" {
		return fmt.Errorf("dev default JWT_SECRET is prohibited; set a secure value via the JWT_SECRET environment variable")
	}
	if c.StorageType == "s3" && (c.S3Bucket == "" || c.S3Region == "") {
		return fmt.Errorf("S3_BUCKET and S3_REGION are required when STORAGE_TYPE=s3")
	}
	if c.StorageType == "s3" && c.S3SharedBucket == "" {
		return fmt.Errorf("S3_SHARED_BUCKET is required when STORAGE_TYPE=s3")
	}
	if c.GinMode != "release" {
		return nil
	}
	if c.TrustedProxyCIDR == "" {
		return fmt.Errorf("TRUSTED_PROXY_CIDR is required in production (rate-limit bypass risk)")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET must be explicitly set in release mode")
	}
	if c.DBPass == "" || c.DBPass == "ekarte_password" {
		return fmt.Errorf("DB_PASSWORD must be explicitly set in release mode")
	}
	if c.IntegrationEncryptionKey == "" {
		return fmt.Errorf("INTEGRATION_ENCRYPTION_KEY must be explicitly set in release mode")
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
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName, c.DBSSLMode, JapanTimeZone,
	)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
