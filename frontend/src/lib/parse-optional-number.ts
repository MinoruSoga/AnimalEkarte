/**
 * 空文字/未入力は undefined、それ以外は数値へ変換する（NaN は undefined 扱い）。
 * PATCH/PUT ボディで optional な number フィールドを「未入力=送信しない」にする際の共通ヘルパー。
 * ts-review-201 MEDIUM: medicine-settings-model.ts と medicine-dose-params-editor-model.ts に
 * バイト単位で重複定義されていたため共通化。
 */
export function parseOptionalNumber(value: string): number | undefined {
  if (!value.trim()) return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}
