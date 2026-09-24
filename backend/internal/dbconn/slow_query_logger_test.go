package dbconn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"gorm.io/gorm"
)

func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() {
		slog.SetDefault(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))
	})
	return &buf
}

func TestSlowQueryLogger_BelowThresholdSilent(t *testing.T) {
	buf := captureSlog(t)
	l := newSlowQueryLogger(500 * time.Millisecond)
	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT * FROM pets WHERE id = ?", 1
	}, nil)
	if buf.Len() != 0 {
		t.Fatalf("expected no log output, got %q", buf.String())
	}
}

func TestSlowQueryLogger_AboveThresholdWarns(t *testing.T) {
	buf := captureSlog(t)
	l := newSlowQueryLogger(time.Millisecond)
	l.Trace(context.Background(), time.Now().Add(-time.Second), func() (string, int64) {
		return "SELECT * FROM pets WHERE name = ?", 7
	}, nil)

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("expected JSON log line, got %q", buf.String())
	}
	if entry["msg"] != "slow db query" {
		t.Errorf("msg = %v, want %q", entry["msg"], "slow db query")
	}
	if entry["sql"] != "SELECT * FROM pets WHERE name = ?" {
		t.Errorf("sql = %v, want parameterized template", entry["sql"])
	}
	if entry["rows"] != float64(7) {
		t.Errorf("rows = %v, want 7", entry["rows"])
	}
}

func TestSlowQueryLogger_RecordNotFoundSilent(t *testing.T) {
	buf := captureSlog(t)
	l := newSlowQueryLogger(time.Millisecond)
	l.Trace(context.Background(), time.Now().Add(-time.Hour), func() (string, int64) {
		return "SELECT 1", 0
	}, gorm.ErrRecordNotFound)
	if buf.Len() != 0 {
		t.Fatalf("expected no log output, got %q", buf.String())
	}
}

func TestSlowQueryLogger_ErrorLogsTemplateNotErrText(t *testing.T) {
	buf := captureSlog(t)
	l := newSlowQueryLogger(time.Hour)
	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "INSERT INTO owners (email) VALUES (?)", 0
	}, errors.New("duplicate key value violates unique constraint (email=secret@example.test)"))

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &entry); err != nil {
		t.Fatalf("expected JSON log line, got %q", buf.String())
	}
	if entry["msg"] != "db query error" {
		t.Errorf("msg = %v, want %q", entry["msg"], "db query error")
	}
	if bytes.Contains(buf.Bytes(), []byte("secret@example.test")) {
		t.Errorf("log leaked error detail: %q", buf.String())
	}
}

func TestSlowQueryLogger_ParamsFilterStripsValues(t *testing.T) {
	l := slowQueryLogger{threshold: time.Millisecond}
	sql, vars := l.ParamsFilter(
		context.Background(),
		"SELECT * FROM owners WHERE email = ?",
		"user@example.test",
	)
	if sql != "SELECT * FROM owners WHERE email = ?" {
		t.Errorf("sql = %q, want unchanged template", sql)
	}
	if vars != nil {
		t.Errorf("vars = %v, want nil", vars)
	}
}

func TestSlowQueryLogger_ZeroThresholdSilent(t *testing.T) {
	buf := captureSlog(t)
	l := newSlowQueryLogger(0)
	l.Trace(context.Background(), time.Now().Add(-time.Hour), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)
	if buf.Len() != 0 {
		t.Fatalf("expected no log output, got %q", buf.String())
	}
}
