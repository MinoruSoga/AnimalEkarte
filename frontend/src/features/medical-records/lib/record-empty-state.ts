/**
 * EMR-412 K5: ロック中の空状態文言。ロック理由で分ける。
 * - 確定済: 「確定済のため変更できません」
 * - 編集権限なし（canSubmit false など）: 「編集権限がありません」
 * - 通常: null（呼び出し側が実ボタン名つきの通常文言を出す）
 */
export function getLockedEmptyStateMessage(input: {
  isFinalized: boolean;
  isLocked: boolean;
}): string | null {
  if (input.isFinalized) return "確定済のため変更できません";
  if (input.isLocked) return "編集権限がありません";
  return null;
}
