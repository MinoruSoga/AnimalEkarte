package repository

import "strings"

// ひらがな⇔カタカナ正規化用の文字マッピング。
// PostgreSQL translate(s, kanaSourceChars, kanaTargetChars) で DB 列のカタカナをひらがなに正規化する。
// 86 文字対応: U+30A1 (ァ) 〜 U+30F6 (ヶ) → U+3041 (ぁ) 〜 U+3096 (ゖ)
const (
	kanaSourceChars = "ァアィイゥウェエォオカガキギクグケゲコゴサザシジスズセゼソゾタダチヂッツヅテデトドナニヌネノハバパヒビピフブプヘベペホボポマミムメモャヤュユョヨラリルレロヮワヰヱヲンヴヵヶ"
	kanaTargetChars = "ぁあぃいぅうぇえぉおかがきぎくぐけげこごさざしじすずせぜそぞただちぢっつづてでとどなにぬねのはばぱひびぴふぶぷへべぺほぼぽまみむめもゃやゅゆょよらりるれろゎわゐゑをんゔゕゖ"
)

// NormalizeKana はカタカナをひらがなに変換する。ひらがな・漢字・ASCII は変換しない。
// 検索語を正規化し、DB 列の translate() 正規化値と統一比較するために使用する。
// U+30A1 (ァ) 〜 U+30F6 (ヶ) はひらがなブロック (U+3041〜U+3096) と
// オフセット U+0060 (= 96) だけ離れているため、単純な減算で変換できる。
func NormalizeKana(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= 'ァ' && r <= 'ヶ' { // U+30A1 〜 U+30F6
			r -= 0x60
		}
		b.WriteRune(r)
	}
	return b.String()
}
