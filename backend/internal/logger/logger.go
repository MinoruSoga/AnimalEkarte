package logger

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

// Logger はアプリケーション全体で使用するロガー
var defaultLogger *slog.Logger

// fallbackOnce は Default() が Init() 未実行時に行うフォールバック初期化を保護する。
// 全エントリポイントは main で Init 済みのため実害は理論値だが、
// 非同期 read/write の -race 検出クラスを残さない（BE-refactor.md B-4）。
var fallbackOnce sync.Once

// Config はロガーの設定
type Config struct {
	Level  slog.Level
	Format string // "json" or "text"
	Output io.Writer
}

// DefaultConfig はデフォルト設定を返す
func DefaultConfig() Config {
	return Config{
		Level:  slog.LevelInfo,
		Format: "json",
		Output: os.Stdout,
	}
}

// Init はロガーを初期化する
func Init(cfg Config) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.Level == slog.LevelDebug,
	}

	if cfg.Format == "text" {
		handler = slog.NewTextHandler(cfg.Output, opts)
	} else {
		handler = slog.NewJSONHandler(cfg.Output, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Default はデフォルトロガーを返す
func Default() *slog.Logger {
	// defaultLogger への読み書きは sync.Once の外側で行わない
	// （Once.Do() の外の nil チェックは write と競合し -race を誘発する）。
	fallbackOnce.Do(func() {
		if defaultLogger == nil {
			Init(DefaultConfig())
		}
	})
	return defaultLogger
}

// Info はInfoレベルのログを出力する
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

// Warn はWarnレベルのログを出力する
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

// Error はErrorレベルのログを出力する
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}
