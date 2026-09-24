package dbconn

import (
	"context"
	"errors"
	"log/slog"
	"time"

	gormlogger "gorm.io/gorm/logger"

	"gorm.io/gorm"
)

// slowQueryLogger implements gorm's logger.Interface with a single purpose:
// emit a Warn log for statements slower than the configured threshold. It is
// installed so that container-internal latency can be attributed to concrete
// queries during performance investigations.
//
// ParamsFilter strips bind values, so the fc() callback inside Trace returns
// the parameterized SQL template ("?"/$N placeholders) rather than inlined
// values — raw values may contain PII and are never written to logs.
type slowQueryLogger struct {
	threshold time.Duration
}

func newSlowQueryLogger(threshold time.Duration) gormlogger.Interface {
	return slowQueryLogger{threshold: threshold}
}

func (l slowQueryLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

// Info/Warn/Error are intentionally silent: gorm's default logger would emit
// statement text with inlined values and error details that can echo row
// contents (e.g. unique-violation detail). Errors propagate to callers, which
// apply the project's own sanitization policy when logging them.
func (slowQueryLogger) Info(context.Context, string, ...interface{})  {}
func (slowQueryLogger) Warn(context.Context, string, ...interface{})  {}
func (slowQueryLogger) Error(context.Context, string, ...interface{}) {}

// ParamsFilter drops bind parameters so Trace's fc() returns the SQL template
// with placeholders instead of inlined values.
func (slowQueryLogger) ParamsFilter(
	_ context.Context,
	sql string,
	_ ...interface{},
) (string, []interface{}) {
	return sql, nil
}

func (l slowQueryLogger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (string, int64),
	err error,
) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	elapsed := time.Since(begin)
	if err != nil {
		sql, _ := fc()
		slog.WarnContext(
			ctx,
			"db query error",
			"duration_ms", elapsed.Milliseconds(),
			"sql", sql,
		)
		return
	}
	if l.threshold <= 0 || elapsed <= l.threshold {
		return
	}
	sql, rows := fc()
	slog.WarnContext(
		ctx,
		"slow db query",
		"duration_ms", elapsed.Milliseconds(),
		"rows", rows,
		"sql", sql,
	)
}
