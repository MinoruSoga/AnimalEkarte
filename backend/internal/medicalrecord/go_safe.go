package medicalrecord

import "github.com/animal-ekarte/backend/internal/sharedkernel"

// goSafe は sharedkernel.GoSafe への既存呼び出し面互換 delegate（実装正本は sharedkernel・
// 共有カーネル昇格batch。④b までの service との複製は解消済み）。
func goSafe(name string, fn func()) {
	sharedkernel.GoSafe(name, fn)
}
