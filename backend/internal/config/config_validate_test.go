package config

import (
	"strings"
	"testing"
)

func baseReleaseConfigForValidateTest() *Config {
	return &Config{
		GinMode:                  "release",
		JWTSecret:                "0123456789abcdef0123456789abcdef",
		DBHost:                   "db.example.com",
		DBPort:                   "5432",
		DBUser:                   "animalekarte_app",
		DBPass:                   "secure-db-password",
		DBName:                   "animalekarte",
		DBSSLMode:                "verify-full",
		DBSSLRootCert:            "system",
		DBMaxOpenConns:           10,
		DBMaxIdleConns:           5,
		SMTPPort:                 "587",
		SMTPHost:                 "smtp.example.com",
		SMTPUser:                 "smtp-user",
		SMTPPass:                 "smtp-password",
		SMTPFrom:                 "sender@example.com",
		FrontendURL:              "https://example.com",
		IntegrationEncryptionKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		TrustedProxyCIDR:         "10.0.0.0/8",
		CORSAllowedOrigin:        "https://example.com",
		StorageType:              "s3",
		S3Bucket:                 "upload-bucket",
		S3Region:                 "ap-northeast-1",
		S3SharedBucket:           "shared-bucket",
		AWSAccessKeyID:           "test-access-key",
		AWSSecretAccessKey:       "test-secret-key",
	}
}

func TestConfigLoad(t *testing.T) {
	t.Setenv("PORT", "9999")
	t.Setenv("DB_HOST", "myhost")
	t.Setenv("DB_PORT", "1234")
	t.Setenv("DB_USER", "myuser")
	t.Setenv("DB_PASSWORD", "mypass")
	t.Setenv("DB_NAME", "mydb")
	t.Setenv("DB_SSL_MODE", "require")
	t.Setenv("DB_SSL_ROOT_CERT", "system")
	t.Setenv("GIN_MODE", "release")
	t.Setenv("JWT_SECRET", "mysecret")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_USER", "smtpuser")
	t.Setenv("SMTP_PASS", "smtppass")
	t.Setenv("SMTP_FROM", "from@example.com")
	t.Setenv("FRONTEND_URL", "http://frontend.local")
	t.Setenv("INTEGRATION_ENCRYPTION_KEY", "key")
	t.Setenv("S3_SHARED_BUCKET", "bucket")
	t.Setenv("S3_SHARED_REGION", "us-east-1")
	t.Setenv("S3_ENDPOINT", "https://example.r2.cloudflarestorage.com")
	t.Setenv("STORAGE_TYPE", "s3")
	t.Setenv("S3_BUCKET", "upload-bucket")
	t.Setenv("S3_REGION", "ap-northeast-1")
	t.Setenv("S3_PUBLIC_BASE_URL", "https://images.example.com")
	t.Setenv("AWS_ACCESS_KEY_ID", "access-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "secret-key")
	t.Setenv("TRUSTED_PROXY_CIDR", "10.0.0.0/8")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("CORS_ALLOWED_ORIGIN", "https://example.com")

	cfg := Load()

	if cfg.Port != "9999" {
		t.Errorf("Port = %s, want 9999", cfg.Port)
	}
	if cfg.DBHost != "myhost" {
		t.Errorf("DBHost = %s, want myhost", cfg.DBHost)
	}
	if cfg.DBPort != "1234" {
		t.Errorf("DBPort = %s, want 1234", cfg.DBPort)
	}
	if cfg.DBUser != "myuser" {
		t.Errorf("DBUser = %s, want myuser", cfg.DBUser)
	}
	if cfg.DBPass != "mypass" {
		t.Errorf("DBPass = %s, want mypass", cfg.DBPass)
	}
	if cfg.DBName != "mydb" {
		t.Errorf("DBName = %s, want mydb", cfg.DBName)
	}
	if cfg.DBSSLMode != "require" {
		t.Errorf("DBSSLMode = %s, want require", cfg.DBSSLMode)
	}
	if cfg.DBSSLRootCert != "system" {
		t.Errorf("DBSSLRootCert = %s, want system", cfg.DBSSLRootCert)
	}
	if cfg.GinMode != "release" {
		t.Errorf("GinMode = %s, want release", cfg.GinMode)
	}
	if cfg.JWTSecret != "mysecret" {
		t.Errorf("JWTSecret = %s, want mysecret", cfg.JWTSecret)
	}
	if cfg.SMTPHost != "smtp.example.com" {
		t.Errorf("SMTPHost = %s, want smtp.example.com", cfg.SMTPHost)
	}
	if cfg.SMTPPort != "465" {
		t.Errorf("SMTPPort = %s, want 465", cfg.SMTPPort)
	}
	if cfg.SMTPUser != "smtpuser" {
		t.Errorf("SMTPUser = %s, want smtpuser", cfg.SMTPUser)
	}
	if cfg.SMTPPass != "smtppass" {
		t.Errorf("SMTPPass = %s, want smtppass", cfg.SMTPPass)
	}
	if cfg.SMTPFrom != "from@example.com" {
		t.Errorf("SMTPFrom = %s, want from@example.com", cfg.SMTPFrom)
	}
	if cfg.FrontendURL != "http://frontend.local" {
		t.Errorf("FrontendURL = %s, want http://frontend.local", cfg.FrontendURL)
	}
	if cfg.IntegrationEncryptionKey != "key" {
		t.Errorf("IntegrationEncryptionKey = %s, want key", cfg.IntegrationEncryptionKey)
	}
	if cfg.S3SharedBucket != "bucket" {
		t.Errorf("S3SharedBucket = %s, want bucket", cfg.S3SharedBucket)
	}
	if cfg.S3SharedRegion != "us-east-1" {
		t.Errorf("S3SharedRegion = %s, want us-east-1", cfg.S3SharedRegion)
	}
	if cfg.S3Endpoint != "https://example.r2.cloudflarestorage.com" {
		t.Errorf("S3Endpoint = %s, want https://example.r2.cloudflarestorage.com", cfg.S3Endpoint)
	}
	if cfg.StorageType != "s3" {
		t.Errorf("StorageType = %s, want s3", cfg.StorageType)
	}
	if cfg.S3Bucket != "upload-bucket" {
		t.Errorf("S3Bucket = %s, want upload-bucket", cfg.S3Bucket)
	}
	if cfg.S3Region != "ap-northeast-1" {
		t.Errorf("S3Region = %s, want ap-northeast-1", cfg.S3Region)
	}
	if cfg.S3PublicBaseURL != "https://images.example.com" {
		t.Errorf("S3PublicBaseURL = %s, want https://images.example.com", cfg.S3PublicBaseURL)
	}
	if cfg.AWSAccessKeyID != "access-key" {
		t.Errorf("AWSAccessKeyID was not loaded")
	}
	if cfg.AWSSecretAccessKey != "secret-key" {
		t.Errorf("AWSSecretAccessKey was not loaded")
	}
	if cfg.TrustedProxyCIDR != "10.0.0.0/8" {
		t.Errorf("TrustedProxyCIDR = %s, want 10.0.0.0/8", cfg.TrustedProxyCIDR)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %s, want debug", cfg.LogLevel)
	}
	if cfg.CORSAllowedOrigin != "https://example.com" {
		t.Errorf("CORSAllowedOrigin = %s, want https://example.com", cfg.CORSAllowedOrigin)
	}
}

func TestConfigLoad_S3EndpointDefaultsEmpty(t *testing.T) {
	t.Setenv("S3_ENDPOINT", "") // CI/デプロイ環境で S3_ENDPOINT が既に設定されていても隔離する

	cfg := Load()

	if cfg.S3Endpoint != "" {
		t.Errorf("S3Endpoint = %q, want empty (AWS S3 既定挙動を維持)", cfg.S3Endpoint)
	}
}

func TestConfigLoad_ReleaseDoesNotInheritDevelopmentDatabaseDefaults(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")

	cfg := Load()

	if cfg.DBHost != "" || cfg.DBPort != "" || cfg.DBUser != "" || cfg.DBPass != "" || cfg.DBName != "" {
		t.Fatalf(
			"release database target inherited a default: host=%q port=%q user=%q password_set=%t name=%q",
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBUser,
			cfg.DBPass != "",
			cfg.DBName,
		)
	}
}

func TestConfigLoad_S3PublicBaseURLDefaultsEmpty(t *testing.T) {
	t.Setenv("S3_PUBLIC_BASE_URL", "") // 未設定時は推測公開ドメインを捏造しない（P2-5: 値投入は USER 運用）

	cfg := Load()

	if cfg.S3PublicBaseURL != "" {
		t.Errorf("S3PublicBaseURL = %q, want empty (未設定時は公開 base を持たない)", cfg.S3PublicBaseURL)
	}
}

func TestConfigValidate_DevDefaultJWTSecretProhibited(t *testing.T) {
	cfg := &Config{
		JWTSecret: "dev-secret-change-me",
	}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "dev default JWT_SECRET is prohibited") {
		t.Fatalf("expected error for dev default JWT_SECRET, got %v", err)
	}
}

func TestConfigValidate_ReleaseRequiresJWTSecret(t *testing.T) {
	cfg := baseReleaseConfigForValidateTest()
	cfg.JWTSecret = ""
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "JWT_SECRET must be explicitly set") {
		t.Fatalf("expected error for empty JWT_SECRET in release mode, got %v", err)
	}
}

func TestConfigValidate_ReleaseRejectsShortJWTSecret(t *testing.T) {
	cfg := baseReleaseConfigForValidateTest()
	cfg.JWTSecret = strings.Repeat("a", 31)

	err := cfg.Validate()

	if err == nil || !strings.Contains(err.Error(), "at least 32 bytes") {
		t.Fatalf("expected short JWT_SECRET error, got %v", err)
	}
}

func TestConfigValidate_ReleaseProhibitsDevDBPassword(t *testing.T) {
	cfg := baseReleaseConfigForValidateTest()
	cfg.DBPass = "ekarte_password"
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "DB_PASSWORD must be explicitly set") {
		t.Fatalf("expected error for default DB_PASSWORD in release mode, got %v", err)
	}

	cfg.DBPass = ""
	err = cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "DB_PASSWORD must be explicitly set") {
		t.Fatalf("expected error for empty DB_PASSWORD in release mode, got %v", err)
	}
}

func TestConfigValidate_ReleaseRequiresExplicitDatabaseTarget(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*Config)
		wantErrMsg string
	}{
		{
			name:       "host is required",
			mutate:     func(cfg *Config) { cfg.DBHost = "" },
			wantErrMsg: "DB_HOST",
		},
		{
			name:       "development host is prohibited",
			mutate:     func(cfg *Config) { cfg.DBHost = "localhost" },
			wantErrMsg: "DB_HOST",
		},
		{
			name:       "loopback host is prohibited",
			mutate:     func(cfg *Config) { cfg.DBHost = "127.0.0.1" },
			wantErrMsg: "DB_HOST",
		},
		{
			name:       "port is required",
			mutate:     func(cfg *Config) { cfg.DBPort = "" },
			wantErrMsg: "DB_PORT",
		},
		{
			name:       "port must be numeric",
			mutate:     func(cfg *Config) { cfg.DBPort = "postgres" },
			wantErrMsg: "DB_PORT",
		},
		{
			name:       "port must be in range",
			mutate:     func(cfg *Config) { cfg.DBPort = "65536" },
			wantErrMsg: "DB_PORT",
		},
		{
			name:       "user is required",
			mutate:     func(cfg *Config) { cfg.DBUser = "" },
			wantErrMsg: "DB_USER",
		},
		{
			name:       "development user is prohibited",
			mutate:     func(cfg *Config) { cfg.DBUser = "ekarte_user" },
			wantErrMsg: "DB_USER",
		},
		{
			name:       "name is required",
			mutate:     func(cfg *Config) { cfg.DBName = "" },
			wantErrMsg: "DB_NAME",
		},
		{
			name:       "development name is prohibited",
			mutate:     func(cfg *Config) { cfg.DBName = "ekarte_db" },
			wantErrMsg: "DB_NAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseReleaseConfigForValidateTest()
			tt.mutate(cfg)

			err := cfg.Validate()

			if err == nil || !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Fatalf("Validate() error = %v, want %s error", err, tt.wantErrMsg)
			}
		})
	}
}

func TestConfigValidate_ReleaseRequiresIntegrationEncryptionKey(t *testing.T) {
	t.Setenv("LIFF_MOCK", "")
	cfg := baseReleaseConfigForValidateTest()
	cfg.IntegrationEncryptionKey = ""

	err := cfg.Validate()

	if err == nil {
		t.Fatal("Validate() error = nil, want integration encryption key error")
	}
	if !strings.Contains(err.Error(), "INTEGRATION_ENCRYPTION_KEY") {
		t.Fatalf("Validate() error = %q, want INTEGRATION_ENCRYPTION_KEY", err.Error())
	}
}

func TestConfigValidate_ReleaseRestrictsSMTPPort(t *testing.T) {
	cfg := baseReleaseConfigForValidateTest()
	cfg.SMTPHost = "smtp.example.com"
	cfg.SMTPPort = "25"
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "SMTP_PORT must be 465 (TLS) or 587 (STARTTLS)") {
		t.Fatalf("expected error for invalid SMTP_PORT in release mode, got %v", err)
	}
}

func TestConfigValidate_ReleaseProhibitsLiffMock(t *testing.T) {
	t.Setenv("LIFF_MOCK", "true")
	cfg := baseReleaseConfigForValidateTest()
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "LIFF_MOCK must not be set") {
		t.Fatalf("expected error for LIFF_MOCK in release mode, got %v", err)
	}
}

func TestConfigValidate_ReleaseAcceptsValidConfig(t *testing.T) {
	t.Setenv("LIFF_MOCK", "")
	cfg := baseReleaseConfigForValidateTest()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidate_DebugModeBypassesReleaseChecks(t *testing.T) {
	cfg := &Config{
		GinMode:   "debug",
		JWTSecret: "secure-jwt-secret",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidate_ReleaseRequiresTrustedProxyCIDR(t *testing.T) {
	cfg := baseReleaseConfigForValidateTest()
	cfg.TrustedProxyCIDR = ""
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "TRUSTED_PROXY_CIDR is required in production") {
		t.Fatalf("expected error for empty TRUSTED_PROXY_CIDR in release mode, got %v", err)
	}
}

func TestConfigValidate_ReleaseRejectsInvalidTrustedProxyCIDR(t *testing.T) {
	cfg := baseReleaseConfigForValidateTest()
	cfg.TrustedProxyCIDR = "not-a-cidr"

	err := cfg.Validate()

	if err == nil || !strings.Contains(err.Error(), "TRUSTED_PROXY_CIDR must be a valid CIDR") {
		t.Fatalf("expected invalid TRUSTED_PROXY_CIDR error, got %v", err)
	}
}

func TestConfigValidate_ReleaseRejectsUntrustedProxyNetworks(t *testing.T) {
	for _, cidr := range []string{
		"0.0.0.0/0",
		"::/0",
		"203.0.113.0/24",
	} {
		t.Run(cidr, func(t *testing.T) {
			cfg := baseReleaseConfigForValidateTest()
			cfg.TrustedProxyCIDR = cidr

			err := cfg.Validate()

			if err == nil || !strings.Contains(err.Error(), "TRUSTED_PROXY_CIDR") {
				t.Fatalf("Validate() error = %v, want TRUSTED_PROXY_CIDR error", err)
			}
		})
	}
}

func TestConfigValidate_ReleaseRejectsUnsafeExternalEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*Config)
		wantErrMsg string
	}{
		{
			name: "CORS origin is required",
			mutate: func(cfg *Config) {
				cfg.CORSAllowedOrigin = ""
			},
			wantErrMsg: "CORS_ALLOWED_ORIGIN",
		},
		{
			name: "CORS wildcard is prohibited",
			mutate: func(cfg *Config) {
				cfg.CORSAllowedOrigin = "*"
			},
			wantErrMsg: "CORS_ALLOWED_ORIGIN",
		},
		{
			name: "CORS origin must use HTTPS",
			mutate: func(cfg *Config) {
				cfg.CORSAllowedOrigin = "http://example.com"
			},
			wantErrMsg: "CORS_ALLOWED_ORIGIN",
		},
		{
			name: "CORS origin must not use localhost",
			mutate: func(cfg *Config) {
				cfg.CORSAllowedOrigin = "https://localhost"
			},
			wantErrMsg: "CORS_ALLOWED_ORIGIN",
		},
		{
			name: "CORS origin must not contain a path",
			mutate: func(cfg *Config) {
				cfg.CORSAllowedOrigin = "https://example.com/"
			},
			wantErrMsg: "CORS_ALLOWED_ORIGIN",
		},
		{
			name: "frontend URL is required",
			mutate: func(cfg *Config) {
				cfg.FrontendURL = ""
			},
			wantErrMsg: "FRONTEND_URL",
		},
		{
			name: "frontend URL must use HTTPS",
			mutate: func(cfg *Config) {
				cfg.FrontendURL = "http://example.com"
			},
			wantErrMsg: "FRONTEND_URL",
		},
		{
			name: "frontend URL must not use loopback",
			mutate: func(cfg *Config) {
				cfg.FrontendURL = "https://127.0.0.1"
			},
			wantErrMsg: "FRONTEND_URL",
		},
		{
			name: "frontend URL must not contain user info",
			mutate: func(cfg *Config) {
				cfg.FrontendURL = "https://user:pass@example.com"
			},
			wantErrMsg: "FRONTEND_URL",
		},
		{
			name: "frontend URL must not contain query or fragment",
			mutate: func(cfg *Config) {
				cfg.FrontendURL = "https://example.com?redirect=unsafe#fragment"
			},
			wantErrMsg: "FRONTEND_URL",
		},
		{
			name: "database SSL mode must not be disabled",
			mutate: func(cfg *Config) {
				cfg.DBSSLMode = "disable"
			},
			wantErrMsg: "DB_SSL_MODE",
		},
		{
			name: "database SSL mode must be recognized",
			mutate: func(cfg *Config) {
				cfg.DBSSLMode = "prefer"
			},
			wantErrMsg: "DB_SSL_MODE",
		},
		{
			name: "database SSL mode must verify server identity",
			mutate: func(cfg *Config) {
				cfg.DBSSLMode = "require"
			},
			wantErrMsg: "DB_SSL_MODE",
		},
		{
			name: "database SSL root cert is required",
			mutate: func(cfg *Config) {
				cfg.DBSSLRootCert = ""
			},
			wantErrMsg: "DB_SSL_ROOT_CERT",
		},
		{
			name: "database SSL root cert must use system trust",
			mutate: func(cfg *Config) {
				cfg.DBSSLRootCert = "/tmp/untrusted.pem"
			},
			wantErrMsg: "DB_SSL_ROOT_CERT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LIFF_MOCK", "")
			cfg := baseReleaseConfigForValidateTest()
			tt.mutate(cfg)

			err := cfg.Validate()

			if err == nil || !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Fatalf("Validate() error = %v, want %s error", err, tt.wantErrMsg)
			}
		})
	}
}

func TestConfigValidate_ReleaseRequiresSMTPDelivery(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*Config)
		wantErrMsg string
	}{
		{
			name:       "SMTP host is required",
			mutate:     func(cfg *Config) { cfg.SMTPHost = "" },
			wantErrMsg: "SMTP_HOST",
		},
		{
			name:       "SMTP user is required",
			mutate:     func(cfg *Config) { cfg.SMTPUser = "" },
			wantErrMsg: "SMTP_USER",
		},
		{
			name:       "SMTP password is required",
			mutate:     func(cfg *Config) { cfg.SMTPPass = "" },
			wantErrMsg: "SMTP_PASS",
		},
		{
			name:       "SMTP from address is required",
			mutate:     func(cfg *Config) { cfg.SMTPFrom = "" },
			wantErrMsg: "SMTP_FROM",
		},
		{
			name: "SMTP from address rejects CRLF",
			mutate: func(cfg *Config) {
				cfg.SMTPFrom = "sender@example.com\r\nBcc: attacker@example.com"
			},
			wantErrMsg: "SMTP_FROM",
		},
		{
			name:       "SMTP from address must be valid",
			mutate:     func(cfg *Config) { cfg.SMTPFrom = "not-an-email" },
			wantErrMsg: "SMTP_FROM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LIFF_MOCK", "")
			cfg := baseReleaseConfigForValidateTest()
			tt.mutate(cfg)

			err := cfg.Validate()

			if err == nil || !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Fatalf("Validate() error = %v, want %s error", err, tt.wantErrMsg)
			}
		})
	}
}

func TestConfigValidate_ReleaseRequiresDurableHTTPSStorage(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		mutate     func(*Config)
		wantErrMsg string
	}{
		{
			name:       "release requires S3 storage",
			config:     baseReleaseConfigForValidateTest(),
			mutate:     func(cfg *Config) { cfg.StorageType = "" },
			wantErrMsg: "STORAGE_TYPE",
		},
		{
			name:       "release rejects local storage",
			config:     baseReleaseConfigForValidateTest(),
			mutate:     func(cfg *Config) { cfg.StorageType = "local" },
			wantErrMsg: "STORAGE_TYPE",
		},
		{
			name:       "debug rejects unknown storage type",
			config:     &Config{GinMode: "debug", JWTSecret: "secure-jwt-secret"},
			mutate:     func(cfg *Config) { cfg.StorageType = "unknown" },
			wantErrMsg: "STORAGE_TYPE",
		},
		{
			name:   "S3 endpoint must use HTTPS",
			config: baseReleaseConfigForValidateTest(),
			mutate: func(cfg *Config) {
				cfg.S3Endpoint = "http://storage.example.com"
			},
			wantErrMsg: "S3_ENDPOINT",
		},
		{
			name:   "S3 public base URL must use HTTPS",
			config: baseReleaseConfigForValidateTest(),
			mutate: func(cfg *Config) {
				cfg.S3PublicBaseURL = "http://images.example.com"
			},
			wantErrMsg: "S3_PUBLIC_BASE_URL",
		},
		{
			name:   "S3 public base URL must not contain user info",
			config: baseReleaseConfigForValidateTest(),
			mutate: func(cfg *Config) {
				cfg.S3PublicBaseURL = "https://user:pass@images.example.com"
			},
			wantErrMsg: "S3_PUBLIC_BASE_URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LIFF_MOCK", "")
			tt.mutate(tt.config)

			err := tt.config.Validate()

			if err == nil || !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Fatalf("Validate() error = %v, want %s error", err, tt.wantErrMsg)
			}
		})
	}
}

func TestConfigValidate_ReleaseRequiresS3Credentials(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*Config)
		wantErrMsg string
	}{
		{
			name:       "access key is required",
			mutate:     func(cfg *Config) { cfg.AWSAccessKeyID = "" },
			wantErrMsg: "AWS_ACCESS_KEY_ID",
		},
		{
			name:       "secret key is required",
			mutate:     func(cfg *Config) { cfg.AWSSecretAccessKey = "" },
			wantErrMsg: "AWS_SECRET_ACCESS_KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseReleaseConfigForValidateTest()
			tt.mutate(cfg)

			err := cfg.Validate()

			if err == nil || !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Fatalf("Validate() error = %v, want %s error", err, tt.wantErrMsg)
			}
		})
	}
}

func TestConfigValidate_ReleaseRejectsInvalidDatabasePoolLimits(t *testing.T) {
	tests := []struct {
		name string
		open int
		idle int
	}{
		{name: "open must be positive", open: 0, idle: 0},
		{name: "idle must be positive", open: 10, idle: 0},
		{name: "idle must not exceed open", open: 5, idle: 10},
		{name: "open must not exceed Cloudflare safe limit", open: 11, idle: 5},
		{name: "idle must not exceed Cloudflare safe limit", open: 10, idle: 6},
		{name: "legacy defaults must not pass release validation", open: 50, idle: 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := baseReleaseConfigForValidateTest()
			cfg.DBMaxOpenConns = tt.open
			cfg.DBMaxIdleConns = tt.idle

			err := cfg.Validate()

			if err == nil || !strings.Contains(err.Error(), "DB_MAX_") {
				t.Fatalf("Validate() error = %v, want DB_MAX_* error", err)
			}
		})
	}
}

func TestConfigValidate_ReleaseFailsClosedForMissingOrInvalidDatabasePoolEnv(t *testing.T) {
	tests := []struct {
		name string
		open string
		idle string
	}{
		{name: "missing", open: "", idle: ""},
		{name: "invalid", open: "not-a-number", idle: "also-invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DB_MAX_OPEN_CONNS", tt.open)
			t.Setenv("DB_MAX_IDLE_CONNS", tt.idle)
			loaded := Load()
			cfg := baseReleaseConfigForValidateTest()
			cfg.DBMaxOpenConns = loaded.DBMaxOpenConns
			cfg.DBMaxIdleConns = loaded.DBMaxIdleConns

			err := cfg.Validate()

			if err == nil || !strings.Contains(err.Error(), "DB_MAX_") {
				t.Fatalf("Validate() error = %v, want DB_MAX_* error", err)
			}
		})
	}
}

func TestConfigValidate_StorageTypeS3RequiresBucketAndRegion(t *testing.T) {
	tests := []struct {
		name      string
		ginMode   string
		s3Bucket  string
		s3Region  string
		wantError bool
	}{
		{"debugモードでもBucket欠落はエラー", "debug", "", "ap-northeast-1", true},
		{"debugモードでもRegion欠落はエラー", "debug", "bucket", "", true},
		{"debugモードでBucket/Region両方あればOK", "debug", "bucket", "ap-northeast-1", false},
		{"releaseモードでもBucket/Region両方あればOK", "release", "bucket", "ap-northeast-1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LIFF_MOCK", "")
			var cfg *Config
			if tt.ginMode == "release" {
				cfg = baseReleaseConfigForValidateTest()
			} else {
				cfg = &Config{GinMode: tt.ginMode, JWTSecret: "secure-jwt-secret"}
			}
			cfg.StorageType = "s3"
			cfg.S3Bucket = tt.s3Bucket
			cfg.S3Region = tt.s3Region
			// S3SharedBucket は本テストの対象外（A-4 で別途検証）なのでOKケースでは満たしておく。
			cfg.S3SharedBucket = "shared-bucket"

			err := cfg.Validate()

			if tt.wantError && (err == nil || !strings.Contains(err.Error(), "S3_BUCKET and S3_REGION are required")) {
				t.Fatalf("expected S3_BUCKET/S3_REGION error, got %v", err)
			}
			if !tt.wantError && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestConfigValidate_StorageTypeEmptyUsesLocalDevelopmentStorage(t *testing.T) {
	cfg := &Config{GinMode: "debug", JWTSecret: "secure-jwt-secret", StorageType: ""}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

// TestConfigValidate_StorageTypeS3RequiresSharedBucket は STORAGE_TYPE=s3 かつ
// S3_SHARED_BUCKET 未設定の場合に Validate() が fail-fast することを検証する
// （main.go の S3_SHARED_BUCKET 参照は最初の共有ファイルアップロードまで検証されなかった）。
func TestConfigValidate_StorageTypeS3RequiresSharedBucket(t *testing.T) {
	cfg := &Config{GinMode: "debug", JWTSecret: "secure-jwt-secret"}
	cfg.StorageType = "s3"
	cfg.S3Bucket = "bucket"
	cfg.S3Region = "ap-northeast-1"
	cfg.S3SharedBucket = ""

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "S3_SHARED_BUCKET is required") {
		t.Fatalf("expected S3_SHARED_BUCKET error, got %v", err)
	}
}
