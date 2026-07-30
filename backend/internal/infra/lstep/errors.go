package lstep

import "errors"

// ErrUserNotFound はLINE User IDがLステップに未登録の場合に返す。
// 処理継続可能なエラーとして扱う（上位層でスキップ判定する）。
//
// ErrRateLimit はレート制限リトライを使い切った場合に返す。
//
// ErrWriteDisabled は deploy-level kill switch（LSTEP_WRITE_API_ENABLED）が
// 有効でないとき、write 系メソッドが HTTP を送らずに返す。
// 成功（nil）扱いしないこと。delivery fired / tag cache receipt 更新に進ませない。
var (
	ErrUserNotFound  = errors.New("lstep: line user not registered in lstep")
	ErrRateLimit     = errors.New("lstep: rate limit exceeded")
	ErrWriteDisabled = errors.New("lstep: write API disabled by deploy gate")
)
