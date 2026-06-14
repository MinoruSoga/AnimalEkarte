/**
 * カタカナをひらがなに変換する。ひらがな・漢字・ASCII は変換しない。
 * U+30A1 (ァ) 〜 U+30F6 (ヶ) → U+3041 (ぁ) 〜 U+3096 (ゖ)
 * Go 側の NormalizeKana (repository/kana_normalize.go) と同等の変換。
 */
export function normalizeKana(s: string): string {
  return s.replace(/[ァ-ヶ]/g, (ch) =>
    String.fromCharCode(ch.charCodeAt(0) - 0x60)
  );
}
