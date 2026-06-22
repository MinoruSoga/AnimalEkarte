/**
 * #158 飼主レポート: DM（ダイレクトメール）送付希望の表示ラベル。
 *
 * レガシー EMR（Figma 37:142）が飼主欄に持つ「DM 区分」に対応する。
 * 既存レコードは値を持たない（undefined/null）ため、「未設定」と「不要(false)」を
 * 明確に区別する。これにより未設定の飼主を「不要」と誤表示しない。
 */
export function formatDMPreference(pref: boolean | null | undefined): string {
  if (pref == null) return "未設定";
  return pref ? "必要" : "不要";
}
