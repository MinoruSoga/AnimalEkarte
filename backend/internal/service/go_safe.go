package service

import "github.com/animal-ekarte/backend/internal/sharedkernel"

// goSafe は sharedkernel.GoSafe への既存呼び出し面互換 delegate（panic 回復付き goroutine 起動。
// 実装正本は sharedkernel・共有カーネル昇格batch）。
func goSafe(name string, fn func()) {
	sharedkernel.GoSafe(name, fn)
}
