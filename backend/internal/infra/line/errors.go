package line

import "errors"

// ErrInvalidRecipient は受信者のLINE User IDが不正またはbot未追加の場合に返す。
// 処理継続可能なエラーとして扱う（上位層でスキップ判定する）。
var (
	ErrInvalidRecipient = errors.New("line: invalid recipient — user has not added the bot or blocked it")
	ErrRateLimit        = errors.New("line: messaging API rate limit exceeded")
)
