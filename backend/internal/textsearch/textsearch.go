// Package textsearch provides shared normalization and escaping for
// PostgreSQL-backed human-name searches.
package textsearch

import "strings"

// KanaSourceChars and KanaTargetChars map the katakana range used by
// PostgreSQL translate() to its hiragana equivalent.
// SpaceSourceChars and SpaceTargetChars fold ideographic space (U+3000) to
// ASCII space on stored columns. Do not mutate the kana tables in place;
// compose via KanaAndSpaceSourceChars / KanaAndSpaceTargetChars at SQL call sites.
const (
	KanaSourceChars         = "ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ"
	KanaTargetChars         = "ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ"
	SpaceSourceChars        = "\u3000"
	SpaceTargetChars        = " "
	KanaAndSpaceSourceChars = KanaSourceChars + SpaceSourceChars
	KanaAndSpaceTargetChars = KanaTargetChars + SpaceTargetChars
)

// NormalizeQuerySpaces collapses full-width / extra whitespace so owner/pet search
// matches half-width space queries (BUG-008). Query-side only; stored columns
// must use SpaceSourceChars / KanaAndSpaceSourceChars with PostgreSQL translate().
func NormalizeQuerySpaces(value string) string {
	value = strings.ReplaceAll(value, "\u3000", " ")
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

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
