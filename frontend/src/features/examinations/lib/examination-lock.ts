/**
 * BUG-033 / S02 examination edit locks (server-persisted status only).
 *
 * - confirmed: full form lock
 * - completed without revision history: first-pass completion seal (results + delete)
 * - completed with revision history: post-unconfirm working copy (editable)
 *
 * Accepts JA labels (form/transform) and EN API enums so a missed transform
 * cannot silently unlock a sealed exam.
 */

export type ExaminationLockStatus =
  | "依頼中"
  | "検査中"
  | "結果入力済み"
  | "完了"
  | "確定";

/** Normalize FE JA labels and BE EN enums used by lock checks. */
export function normalizeExaminationLockStatus(
  status: string | undefined | null,
): ExaminationLockStatus | string | undefined | null {
  if (status == null || status === "") return status;
  switch (status) {
    case "pending":
      return "依頼中";
    case "in_progress":
      return "検査中";
    case "result_entered":
      return "結果入力済み";
    case "completed":
    case "完了":
      return "完了";
    case "confirmed":
    case "確定":
      return "確定";
    default:
      return status;
  }
}

export function isPersistedConfirmedStatus(
  status: string | undefined | null,
): boolean {
  return normalizeExaminationLockStatus(status) === "確定";
}

/** First-pass「完了」seal — no official revision history yet. */
export function isPersistedCompletedSeal(
  status: string | undefined | null,
  currentRevisionVersion: number | undefined | null,
): boolean {
  return (
    normalizeExaminationLockStatus(status) === "完了" &&
    currentRevisionVersion == null
  );
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
