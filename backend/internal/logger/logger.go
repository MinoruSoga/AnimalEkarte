package logger

import (
	"io"
	"log/slog"
	"os"
)

// Logger はアプリケーション全体で使用するロガー
var defaultLogger *slog.Logger

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
	if defaultLogger == nil {
		Init(DefaultConfig())
	}
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
