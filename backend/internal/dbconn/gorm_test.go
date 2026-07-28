package dbconn

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/config"
)

type noConnectDriver struct{}

func (noConnectDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("unexpected database connection")
}

type noConnectConnector struct{}

func (noConnectConnector) Connect(context.Context) (driver.Conn, error) {
	return nil, errors.New("unexpected database connection")
}

func (noConnectConnector) Driver() driver.Driver {
	return noConnectDriver{}
}

type poolSettingsRecorder struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
}

func (r *poolSettingsRecorder) SetMaxOpenConns(value int) {
	r.maxOpenConns = value
}

func (r *poolSettingsRecorder) SetMaxIdleConns(value int) {
	r.maxIdleConns = value
}

func (r *poolSettingsRecorder) SetConnMaxLifetime(value time.Duration) {
	r.connMaxLifetime = value
}

func (r *poolSettingsRecorder) SetConnMaxIdleTime(value time.Duration) {
	r.connMaxIdleTime = value
}

func TestApplyPoolSettings_PreservesRuntimeDefaults(t *testing.T) {
	recorder := &poolSettingsRecorder{}
	cfg := &config.Config{
		DBMaxOpenConns: 37,
		DBMaxIdleConns: 11,
	}

	applyPoolSettings(recorder, cfg)

	if recorder.maxOpenConns != 37 {
		t.Errorf("SetMaxOpenConns() = %d, want 37", recorder.maxOpenConns)
	}
	if recorder.maxIdleConns != 11 {
		t.Errorf("SetMaxIdleConns() = %d, want 11", recorder.maxIdleConns)
	}
	if recorder.connMaxLifetime != 30*time.Minute {
		t.Errorf("SetConnMaxLifetime() = %s, want 30m", recorder.connMaxLifetime)
	}
	if recorder.connMaxIdleTime != 5*time.Minute {
		t.Errorf("SetConnMaxIdleTime() = %s, want 5m", recorder.connMaxIdleTime)
	}
}

func TestOpenGORMWith_UsesConfigDSNAndAppliesPoolSettings(t *testing.T) {
	cfg := &config.Config{
		DBHost:         "database.internal",
		DBPort:         "5433",
		DBUser:         "app",
		DBPass:         "secret",
		DBName:         "ekarte",
		DBSSLMode:      "verify-full",
		DBSSLRootCert:  "system",
		DBMaxOpenConns: 23,
		DBMaxIdleConns: 7,
	}
	sqlDB := sql.OpenDB(noConnectConnector{})
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close test sql.DB: %v", err)
		}
	})
	wantDB := &gorm.DB{Config: &gorm.Config{ConnPool: sqlDB}}
	var gotDSN string

	gotDB, err := openGORMWith(cfg, func(dsn string) (*gorm.DB, error) {
		gotDSN = dsn
		return wantDB, nil
	})

	if err != nil {
		t.Fatalf("openGORMWith() unexpected error: %v", err)
	}
	if gotDB != wantDB {
		t.Fatalf("openGORMWith() DB = %p, want %p", gotDB, wantDB)
	}
	wantDSN := "host=database.internal port=5433 user=app password=secret dbname=ekarte sslmode=verify-full TimeZone=Asia/Tokyo sslrootcert=system"
	if gotDSN != wantDSN {
		t.Fatalf("openGORMWith() DSN = %q, want %q", gotDSN, wantDSN)
	}
	if got := sqlDB.Stats().MaxOpenConnections; got != 23 {
		t.Errorf("MaxOpenConnections = %d, want 23", got)
	}
}

func TestOpenGORMWith_WrapsErrors(t *testing.T) {
	openErr := errors.New("open failed")
	tests := []struct {
		name        string
		open        gormOpener
		wantCause   error
		wantMessage string
	}{
		{
			name: "GORM接続開始失敗",
			open: func(string) (*gorm.DB, error) {
				return nil, openErr
			},
			wantCause:   openErr,
			wantMessage: "failed to open database connection",
		},
		{
			name: "sql.DB取得失敗",
			open: func(string) (*gorm.DB, error) {
				return &gorm.DB{Config: &gorm.Config{}}, nil
			},
			wantCause:   gorm.ErrInvalidDB,
			wantMessage: "get sql.DB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := openGORMWith(&config.Config{}, tt.open)

			if err == nil {
				t.Fatal("openGORMWith() error = nil, want error")
			}
			if db != nil {
				t.Fatalf("openGORMWith() DB = %p, want nil", db)
			}
			if !errors.Is(err, tt.wantCause) {
				t.Errorf("openGORMWith() error = %v, want cause %v", err, tt.wantCause)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Errorf("openGORMWith() error = %q, want it to contain %q", err, tt.wantMessage)
			}
		})
	}
}

func TestOpenGORM_InvalidSSLModePreservesWrappedError(t *testing.T) {
	cfg := &config.Config{
		DBHost:    "database.invalid",
		DBPort:    "5432",
		DBUser:    "app",
		DBPass:    "secret",
		DBName:    "ekarte",
		DBSSLMode: "not-a-real-sslmode",
	}

	db, err := OpenGORM(cfg)

	if err == nil {
		t.Fatal("OpenGORM() error = nil, want invalid sslmode error")
	}
	if db != nil {
		t.Fatalf("OpenGORM() DB = %p, want nil", db)
	}
	if !strings.Contains(err.Error(), "failed to open database connection") {
		t.Errorf("OpenGORM() error = %q, want wrapped connection error", err)
	}
}
