/**
 * カタカナをひらがなに変換する。ひらがな・漢字・ASCII は変換しない。
 * U+30A1 (ァ) 〜 U+30F6 (ヶ) → U+3041 (ぁ) 〜 U+3096 (ゖ)
 * Go 側の NormalizeKana (repository/kana_normalize.go) と同等の変換。
 */
export function normalizeKana(s: string): string {
  return s.replace(/[ァ-ヶ]/g, (ch) => String.fromCharCode(ch.charCodeAt(0) - 0x60));
}

/**
 * ひらがな・カタカナを区別しない部分一致比較。
 * 両辺を normalizeKana + toLowerCase した上で includes する。
 * クライアントサイドの検索フィルタで共通使用する。
 */
export function normalizedIncludes(text: string, term: string): boolean {
  return normalizeKana(text).toLowerCase().includes(normalizeKana(term).toLowerCase());
}
