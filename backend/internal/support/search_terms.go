package support

import (
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// search_terms.go — chat-context.ts の検索語抽出・根拠ゲートの Go 移植。
// 辞書データ（stopwords/suffixes/synonym_groups）は search_dict.go に
// 人間実行 codegen で生成される（FE search-dict.json が正本）。

// EMR-438: JS \s（chat-context.ts の分割契約）に合わせ Unicode 空白（U+3000 全角
// スペース等）も分割対象にする。Go の \s は ASCII のみのため、全角スペース入りの
// クエリが keyword-only モードで 1 語句扱いになり誤応答する問題への対応。
var searchTokenSplitRe = regexp.MustCompile(`[\s\p{Zs}\x{2028}\x{2029}\x{FEFF}、。，．！？!?,・「」『』（）()【】.:;のをにがはでとへもや]+`)

var searchSuffixRe = func() *regexp.Regexp {
	quoted := make([]string, len(searchDictSuffixes))
	for i, sfx := range searchDictSuffixes {
		quoted[i] = regexp.QuoteMeta(sfx)
	}
	return regexp.MustCompile(`(` + strings.Join(quoted, "|") + `)$`)
}()

// foldVNRunes は結合文字（声調・母音記号）を剥がし đ/Đ を d へ畳む。
// ラテン文字比較専用（JA 文字へ適用すると濁点が落ちる）。
func foldVNRunes(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if r == 'đ' || r == 'Đ' {
			b.WriteByte('d')
			continue
		}
		if unicode.Is(unicode.M, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func nfcLower(s string) string {
	return strings.ToLower(norm.NFC.String(s))
}

// stripSearchSuffixes は文末の活用形・丁寧語・形式名詞を収束まで剥がす。
func stripSearchSuffixes(term string) string {
	for {
		next := searchSuffixRe.ReplaceAllString(term, "")
		if next == term {
			return term
		}
		term = next
	}
}

// segmentJaTerm は分かち書き辞書の最長一致で結合語を切り出す。
// 辞書語を先頭に含まないトークンはそのまま返す。
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}

func segmentJaTerm(term string) []string {
	// JS の length は UTF-16 コードユニット数 — 補助文字は 2 と数える。
	if utf16Len(term) < 2 {
		return []string{term}
	}
	if _, ok := searchDictSegmentSet[term]; ok {
		return []string{term}
	}
	var parts []string
	rest := term
	for utf16Len(rest) >= 2 {
		matched := ""
		for w := range searchDictSegmentSet {
			if strings.HasPrefix(rest, w) && utf16Len(w) > utf16Len(matched) {
				matched = w
			}
		}
		if matched == "" {
			break
		}
		parts = append(parts, matched)
		rest = rest[len(matched):]
	}
	if rest != "" {
		parts = append(parts, rest)
	}
	// JS 実装と同じく、末尾 1 コードユニットの断片は直前パートへマージする。
	if len(parts) > 1 && utf16Len(parts[len(parts)-1]) == 1 {
		tail := parts[len(parts)-1]
		parts = parts[:len(parts)-1]
		parts[len(parts)-1] += tail
	}
	return parts
}

// expandJaTerm は同義語グループへ展開する。base term を先頭に含む。
func expandJaTerm(term string) []string {
	for _, group := range searchDictSynonymGroups {
		for _, w := range group {
			if w == term {
				out := []string{term}
				for _, v := range group {
					if v != term {
						out = append(out, v)
					}
				}
				return out
			}
		}
	}
	return []string{term}
}

// extractSearchTermGroups は質問から検索キーワードの同義語グループを抽出する
// （chat-context.ts extractSearchTermGroups と同一挙動）。
func extractSearchTermGroups(message string) [][]string {
	seen := map[string]struct{}{}
	var groups [][]string
	for _, raw := range searchTokenSplitRe.Split(message, -1) {
		token := norm.NFC.String(raw)
		for _, seg := range segmentJaTerm(stripSearchSuffixes(token)) {
			term := strings.ToLower(seg)
			// EMR-441: frontend 契約 (JS length = UTF-16 単位) に合わせ utf16Len で判定。
			// rune 数で棄却するとサロゲートペア1文字 (𠮷 等) が落ちて検索グループが空に
			// なり「該当なし」誤応答になる。
			if utf16Len(term) < 2 {
				continue
			}
			if _, stop := searchDictStopwords[term]; stop {
				continue
			}
			variants := expandJaTerm(term)
			sorted := append([]string(nil), variants...)
			sort.Strings(sorted)
			key := strings.Join(sorted, "|")
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			groups = append(groups, variants)
		}
	}
	return groups
}

// groupHasTextEvidence は haystack（NFC+小文字済み正規形と VN 畳み込み形）に
// 変種のどれかが文字列として存在するか（chat-context.ts groupHasEvidence と同一）。
func groupHasTextEvidence(normHaystack, foldHaystack string, variants []string) bool {
	for _, v := range variants {
		nv := nfcLower(v)
		if strings.Contains(normHaystack, nv) {
			return true
		}
		hasLatin := false
		for _, r := range nv {
			if unicode.Is(unicode.Latin, r) {
				hasLatin = true
				break
			}
		}
		if !hasLatin {
			continue
		}
		if strings.Contains(foldHaystack, foldVNRunes(nv)) {
			return true
		}
	}
	return false
}

// articleSearchHaystack は記事の根拠比較テキストを作る
// （searchText 相当 = title + section + markdown 記号を粗く除去した本文 + keywords）。
var markdownStripRe = regexp.MustCompile(`[#*` + "`" + `>\[\]()\-_!]`)

func articleSearchHaystack(title, section, body string, keywords, keywordsVI []string) (normHaystack, foldHaystack string) {
	base := title + " " + section + " " + markdownStripRe.ReplaceAllString(body, " ") +
		" " + strings.Join(keywords, " ") + " " + strings.Join(keywordsVI, " ")
	normHaystack = nfcLower(base)
	return normHaystack, foldVNRunes(normHaystack)
}
