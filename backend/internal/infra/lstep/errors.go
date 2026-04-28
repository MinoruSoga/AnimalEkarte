package lstep

import "errors"

// ErrUserNotFound はLINE User IDがLステップに未登録の場合に返す。
// 処理継続可能なエラーとして扱う（上位層でスキップ判定する）。
var (
	ErrUserNotFound = errors.New("lstep: line user not registered in lstep")
	ErrRateLimit    = errors.New("lstep: rate limit exceeded")
)
