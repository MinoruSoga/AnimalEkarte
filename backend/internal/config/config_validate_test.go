package config

import (
	"strings"
	"testing"
)

func baseReleaseConfigForValidateTest() *Config {
	return &Config{
		GinMode:                  "release",
		JWTSecret:                "secure-jwt-secret",
		DBPass:                   "secure-db-password",
		SMTPPort:                 "587",
		IntegrationEncryptionKey: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
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
