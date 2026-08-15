package config

import (
	"testing"
	"time"
)

func TestConfigureTimeZoneSetsTimeLocalToJapan(t *testing.T) {
	original := time.Local
	t.Cleanup(func() {
		time.Local = original
	})

	if err := ConfigureTimeZone(); err != nil {
		t.Fatalf("ConfigureTimeZone() error = %v", err)
	}

	if got := time.Local.String(); got != JapanTimeZone {
		t.Fatalf("time.Local = %q, want %q", got, JapanTimeZone)
	}
}

func TestDSNIncludesJapanTimeZone(t *testing.T) {
	cfg := &Config{
		DBHost:    "db",
		DBPort:    "5432",
		DBUser:    "user",
		DBPass:    "pass",
		DBName:    "ekarte",
		DBSSLMode: "disable",
	}

	want := "host=db port=5432 user=user password=pass dbname=ekarte sslmode=disable TimeZone=Asia/Tokyo"
	if got := cfg.DSN(); got != want {
		t.Fatalf("DSN() = %q, want %q", got, want)
	}
}

func TestDSNIncludesConfiguredSSLRootCert(t *testing.T) {
	cfg := &Config{
		DBHost:        "db.example.com",
		DBPort:        "5432",
		DBUser:        "user",
		DBPass:        "pass",
		DBName:        "ekarte",
		DBSSLMode:     "verify-full",
		DBSSLRootCert: "system",
	}

	want := "host=db.example.com port=5432 user=user password=pass dbname=ekarte sslmode=verify-full TimeZone=Asia/Tokyo sslrootcert=system"
	if got := cfg.DSN(); got != want {
		t.Fatalf("DSN() = %q, want %q", got, want)
	}
}
