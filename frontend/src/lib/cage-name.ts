/**
 * EMR-306: ケージ ID → ケージ名の解決（連結中バッジ・確認ダイアログ等の共通ヘルパー）。
 * 一覧に該当 ID が存在しない場合は `ID <id>` にフォールバックする。
 * master feature の Cage（id: string）と汎用マスタの MasterItem（id: string|number）
 * の両方に使えるよう構造型で受け、比較は String 正規化で行う。
 */
export function resolveCageNameById(
  cages: readonly { id: string | number; name: string }[],
  cageId: string,
): string {
  return cages.find((c) => String(c.id) === cageId)?.name ?? `ID ${cageId}`;
}
