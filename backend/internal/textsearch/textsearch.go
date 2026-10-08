// Package textsearch provides shared normalization and escaping for
// PostgreSQL-backed human-name searches.
package textsearch

import (
	"strings"
	"unicode"
)

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

// FoldedExpr renders the column-side fold expression as a SQL fragment with the
// character sets embedded as literals: translate(<column>, '<src>', '<dst>').
//
// Search predicates MUST use this literal-embedded form. Passing the charset as
// bind parameters (translate(col, ?, ?)) is semantically identical but prevents
// PostgreSQL from matching the predicate against the GIN trigram expression
// indexes added by migrations/015_search_fold_indexes.sql — expression indexes
// bind by literal equality, so any change to the fold charsets must update both
// this helper and that migration together.
func FoldedExpr(column string) string {
	return "translate(" + column + ", '" + KanaAndSpaceSourceChars + "', '" + KanaAndSpaceTargetChars + "')"
}

// SpaceStripRegexp removes every ASCII and ideographic space from already-folded
// text. Used for order-preserving "姓 名" matching; the same literal appears in
// the idx_owners_name_compact_fold_trgm index expression.
const SpaceStripRegexp = "[[:space:]　]+"

// NormalizeQuerySpaces collapses full-width / extra whitespace so owner/pet search
// matches half-width space queries (BUG-008). Query-side only; stored columns
// must use SpaceSourceChars / KanaAndSpaceSourceChars with PostgreSQL translate().
func NormalizeQuerySpaces(value string) string {
	value = strings.ReplaceAll(value, "\u3000", " ")
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

// CompactSearchText removes the spaces SpaceStripRegexp removes on PostgreSQL 18
// with libc en_US.utf8, preserving character order. Query-side companion of:
//
//	regexp_replace(FoldedExpr(column), SpaceStripRegexp, '', 'g')
//
// That POSIX class includes ASCII whitespace, U+3000, and the other spaces it
// matches in this locale. It leaves U+00A0, U+1680, U+2007, and U+202F in place.
// Those runes stay on the stored side too, so the query keeps them. Removing
// them in SQL needs a new migration and idx_owners_name_compact_fold_trgm
// together. Do not edit migration 015. Callers still kana-normalize and
// EscapeLike the result. Do not put the strip charset in a SQL bind parameter.
func CompactSearchText(value string) string {
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if compactSearchSpace(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// compactSearchSpace matches [[:space:]] on postgres:18-alpine en_US.utf8.
// unicode.IsSpace also treats U+00A0, U+1680, U+2007, and U+202F as space.
// The POSIX class on this locale does not.
func compactSearchSpace(r rune) bool {
	switch r {
	case '\u00A0', '\u1680', '\u2007', '\u202F':
		return false
	default:
		return unicode.IsSpace(r)
	}
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
