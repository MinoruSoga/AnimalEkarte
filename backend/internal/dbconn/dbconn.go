// Package dbconn centralizes runtime database connection setup and the
// DSN/environment safety helpers shared by database command tools.
package dbconn

import (
	"fmt"
	"os"

	"github.com/animal-ekarte/backend/internal/config"
)

// JapanTimeZone re-exports config.JapanTimeZone for callers that only import
// dbconn for DSN construction.
const JapanTimeZone = config.JapanTimeZone

// ConnParams holds the connection parameters shared by every DB target this
// tool set talks to. It intentionally excludes the database name: callers
// pick the target database per-connection via DSN(dbname), since several
// tools reuse the same host/user/password against multiple database names.
type ConnParams struct {
	Host        string
	Port        string
	User        string
	Password    string
	SSLMode     string
	SSLRootCert string
}

// FromEnv reads DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_SSL_MODE/DB_SSL_ROOT_CERT, applying the
// same defaults (port 5432, sslmode disable) all 4 cmd tools already used.
// DB_HOST/DB_USER/DB_PASSWORD are required; DB_NAME is deliberately NOT read
// here — each tool has different requiredness/defaults for it, so name
// selection stays with the caller.
func FromEnv() (ConnParams, error) {
	c := ConnParams{
		Host:        os.Getenv("DB_HOST"),
		Port:        EnvOr("DB_PORT", "5432"),
		User:        os.Getenv("DB_USER"),
		Password:    os.Getenv("DB_PASSWORD"),
		SSLMode:     EnvOr("DB_SSL_MODE", "disable"),
		SSLRootCert: os.Getenv("DB_SSL_ROOT_CERT"),
	}
	if c.Host == "" || c.User == "" || c.Password == "" {
		return ConnParams{}, fmt.Errorf("missing required DB env vars (DB_HOST, DB_USER, DB_PASSWORD)")
	}
	return c, nil
}

// DSN builds a libpq-style connection string for dbname, with the fixed
// TimeZone=Asia/Tokyo parameter every existing call site already appended.
func (c ConnParams) DSN(dbname string) string {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host, c.Port, c.User, c.Password, dbname, c.SSLMode, config.JapanTimeZone,
	)
	if c.SSLRootCert != "" {
		dsn += " sslrootcert=" + c.SSLRootCert
	}
	return dsn
}

// localHosts is the superset guard (5 entries, including IPv6 loopback forms)
// shared by the one-shot DB tools. The 3-entry guards were a strict subset
// with no observed case where the extra 2 entries caused a false accept, so
// this is the single source of truth going forward (BE-refactor.md G10-5).
var localHosts = map[string]bool{
	"db":        true,
	"localhost": true,
	"127.0.0.1": true,
	"::1":       true,
	"[::1]":     true,
}

// IsLocalHost reports whether host is a known-local DB host. Every write-path
// cmd tool in this module refuses to run against a host that fails this
// check, to prevent an accidental non-local (staging/production) target.
func IsLocalHost(host string) bool {
	return localHosts[host]
}

// EnvOr returns os.Getenv(key) if non-empty, otherwise fallback.
func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
