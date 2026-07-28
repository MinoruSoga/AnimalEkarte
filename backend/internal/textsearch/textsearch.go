// Package textsearch provides shared normalization and escaping for
// PostgreSQL-backed human-name searches.
package textsearch

import "strings"

// KanaSourceChars and KanaTargetChars map the katakana range used by
// PostgreSQL translate() to its hiragana equivalent.
const (
	KanaSourceChars = "ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ"
	KanaTargetChars = "ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ"
)

// NormalizeKana converts katakana to hiragana while preserving every other
// rune. Search inputs use the same representation as translated database
// columns.
func NormalizeKana(value string) string {
	var normalized strings.Builder
	normalized.Grow(len(value))
	for _, character := range value {
		if character >= 'ァ' && character <= 'ヶ' {
			character -= 0x60
		}
		normalized.WriteRune(character)
	}
	return normalized.String()
}

// EscapeLike escapes PostgreSQL LIKE wildcards for an expression using
// ESCAPE '\'.
func EscapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	return strings.ReplaceAll(value, `_`, `\_`)
}
