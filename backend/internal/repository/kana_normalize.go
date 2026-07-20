package repository

import "github.com/animal-ekarte/backend/internal/repository/repohelpers"

// BE9-2D ⑦: 実装は repohelpers へ hoist（medicalrecord の medical_record_repository も共有するため。
// BE8 §6 で ltv 解錠手段として文書化済みの移設）。既存呼び出し面互換の delegate。

// NormalizeKana はカタカナ/ひらがな正規化（検索用）。
func NormalizeKana(s string) string { return repohelpers.NormalizeKana(s) }

func escapeLike(s string) string { return repohelpers.EscapeLike(s) }

var kanaSourceChars = repohelpers.KanaSourceChars
var kanaTargetChars = repohelpers.KanaTargetChars
