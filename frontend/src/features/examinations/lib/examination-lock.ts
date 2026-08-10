/**
 * BUG-033 / S02 examination edit locks (server-persisted status only).
 *
 * - confirmed: full form lock
 * - completed without revision history: first-pass completion seal (results + delete)
 * - completed with revision history: post-unconfirm working copy (editable)
 */

export type ExaminationLockStatus =
  | "依頼中"
  | "検査中"
  | "結果入力済み"
  | "完了"
  | "確定";

export function isPersistedConfirmedStatus(
  status: string | undefined | null,
): boolean {
  return status === "確定";
}

/** First-pass「完了」seal — no official revision history yet. */
export function isPersistedCompletedSeal(
  status: string | undefined | null,
  currentRevisionVersion: number | undefined | null,
): boolean {
  return status === "完了" && currentRevisionVersion == null;
}

export function isPersistedResultsLocked(
  status: string | undefined | null,
  currentRevisionVersion: number | undefined | null,
): boolean {
  return (
    isPersistedConfirmedStatus(status) ||
    isPersistedCompletedSeal(status, currentRevisionVersion)
  );
}
