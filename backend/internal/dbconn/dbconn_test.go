package dbconn

import (
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/config"
)

func TestFromEnv_RequiresHostUserPassword(t *testing.T) {
	for _, unset := range []string{"DB_HOST", "DB_USER", "DB_PASSWORD"} {
		t.Run("missing "+unset, func(t *testing.T) {
			t.Setenv("DB_HOST", "localhost")
			t.Setenv("DB_USER", "u")
			t.Setenv("DB_PASSWORD", "p")
			t.Setenv(unset, "")

			if _, err := FromEnv(); err == nil {
				t.Fatalf("FromEnv() with %s unset must return an error", unset)
			}
		})
	}
}

func TestFromEnv_AppliesDefaults(t *testing.T) {
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_SSL_MODE", "")
	t.Setenv("DB_SSL_ROOT_CERT", "")

	got, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv() unexpected error: %v", err)
	}
	if got.Port != "5432" {
		t.Errorf("FromEnv().Port = %q, want default 5432", got.Port)
	}
	if got.SSLMode != "disable" {
		t.Errorf("FromEnv().SSLMode = %q, want default disable", got.SSLMode)
	}
	if got.SSLRootCert != "" {
		t.Errorf("FromEnv().SSLRootCert = %q, want empty default", got.SSLRootCert)
	}
}

func TestFromEnv_HonoursOverrides(t *testing.T) {
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_USER", "u")
	t.Setenv("DB_PASSWORD", "p")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_SSL_MODE", "verify-full")
	t.Setenv("DB_SSL_ROOT_CERT", "system")

	got, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv() unexpected error: %v", err)
	}
	if got.Port != "5433" || got.SSLMode != "verify-full" || got.SSLRootCert != "system" {
		t.Fatalf("FromEnv() = %+v, want overrides applied", got)
	}
}

func TestConnParams_DSN(t *testing.T) {
	c := ConnParams{
		Host:        "db",
		Port:        "5432",
		User:        "u",
		Password:    "p",
		SSLMode:     "verify-full",
		SSLRootCert: "system",
	}
	got := c.DSN("mydb")
	for _, want := range []string{
		"host=db",
		"port=5432",
		"user=u",
		"password=p",
		"dbname=mydb",
		"sslmode=verify-full",
		"sslrootcert=system",
		"TimeZone=" + config.JapanTimeZone,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("DSN() = %q, want it to contain %q", got, want)
		}
	}
}

func TestIsLocalHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"db", true},
		{"localhost", true},
		{"127.0.0.1", true},
		{"::1", true},
		{"[::1]", true},
		{"prod-db.example.com", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsLocalHost(tt.host); got != tt.want {
			t.Errorf("IsLocalHost(%q) = %v, want %v", tt.host, got, tt.want)
		}
	}
}

func TestEnvOr(t *testing.T) {
	t.Run("returns env value when set", func(t *testing.T) {
		t.Setenv("DBCONN_TEST_KEY", "custom")
		if got := EnvOr("DBCONN_TEST_KEY", "fallback"); got != "custom" {
			t.Fatalf("EnvOr() = %q, want %q", got, "custom")
		}
	})

	t.Run("returns fallback when unset", func(t *testing.T) {
		t.Setenv("DBCONN_TEST_KEY_UNSET", "")
		if got := EnvOr("DBCONN_TEST_KEY_UNSET", "fallback"); got != "fallback" {
			t.Fatalf("EnvOr() = %q, want %q", got, "fallback")
		}
	})
}
