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

func TestConfigValidate_ReleaseAcceptsIntegrationEncryptionKey(t *testing.T) {
	t.Setenv("LIFF_MOCK", "")
	cfg := baseReleaseConfigForValidateTest()

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidate_DebugDoesNotRequireIntegrationEncryptionKey(t *testing.T) {
	cfg := baseReleaseConfigForValidateTest()
	cfg.GinMode = "debug"
	cfg.IntegrationEncryptionKey = ""

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
