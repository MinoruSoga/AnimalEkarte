/** The version belongs to the editing snapshot, never a later refetch. */
export const MASTER_EDIT_VERSION_ERROR = "編集開始時の版を取得できません。再読込してください";

export function isMasterEditVersion(value: unknown): value is number {
  return typeof value === "number" && Number.isSafeInteger(value) && value > 0;
}

export function requireMasterEditVersion(value: unknown): number {
  if (!isMasterEditVersion(value)) throw new Error(MASTER_EDIT_VERSION_ERROR);
  return value;
}
