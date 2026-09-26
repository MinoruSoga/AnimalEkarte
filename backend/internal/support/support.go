package support

import "strconv"

// uintToString はエラーメッセージ用の uint64 → string 変換
func uintToString(v uint64) string {
	return strconv.FormatUint(v, 10)
}
