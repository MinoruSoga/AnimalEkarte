package commonbasic

import (
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// unresolved は R-1 の失敗を固定の 422（CodeCommonBasicUnresolved）で返す。
// detail にはローカル ID など患者データを含まない識別子だけを入れる（サーバーログ用で応答には出ない）。
func unresolved(format string, args ...any) error {
	return apperrors.WrapCommonBasicUnresolved(fmt.Sprintf(format, args...))
}
