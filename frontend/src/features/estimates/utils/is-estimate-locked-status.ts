import type { EstimateStatus } from "../types";

/** locked 見積の編集拒否時に Form / hook で共有するメッセージ */
export const ESTIMATE_LOCKED_EDIT_MESSAGE =
  "承認済みまたは却下済みの見積書は編集できません";

/**
 * Backend `isEstimateLocked`（approved / rejected）と同義。
 * draft / sent はロックしない（draft→approved への遷移は編集画面で許可）。
 */
export function isEstimateLockedStatus(status: EstimateStatus): boolean {
  return status === "approved" || status === "rejected";
}
