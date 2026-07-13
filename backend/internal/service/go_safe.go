package service

import (
	"log/slog"
	"runtime/debug"
)

// goSafe は panic を回復してログする goroutine 起動ヘルパー。
// バックグラウンド goroutine の panic はプロセスを落とすため、直接 go fn() を書かない。
func goSafe(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("background goroutine panicked", "name", name, "panic", r, "stack", string(debug.Stack()))
			}
		}()
		fn()
	}()
}
