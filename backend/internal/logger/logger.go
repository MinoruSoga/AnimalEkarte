package logger

import (
	"context"
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

// WithContext はコンテキストからロガーを取得する（将来のリクエストID等の拡張用）
func WithContext(ctx context.Context) *slog.Logger {
	// 将来的にリクエストIDなどをコンテキストから取得して付与できる
	return Default()
}

// Info はInfoレベルのログを出力する
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

// Debug はDebugレベルのログを出力する
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

// Warn はWarnレベルのログを出力する
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

// Error はErrorレベルのログを出力する
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

// InfoContext はコンテキスト付きInfoログを出力する
func InfoContext(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).InfoContext(ctx, msg, args...)
}

// ErrorContext はコンテキスト付きErrorログを出力する
func ErrorContext(ctx context.Context, msg string, args ...any) {
	WithContext(ctx).ErrorContext(ctx, msg, args...)
}

// With は追加属性を持つロガーを返す
func With(args ...any) *slog.Logger {
	return Default().With(args...)
}
