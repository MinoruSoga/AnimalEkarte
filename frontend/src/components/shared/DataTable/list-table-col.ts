/**
 * Shared list DataTable column width tokens (Lane B common table).
 * Same family as MASTER_TABLE_COL: status labels like 「ステータス」 need ≥100px
 * plus whitespace-nowrap so full-width Japanese does not wrap one character per line (BUG-020).
 */
export const LIST_TABLE_COL = {
  /** Status header/body cells on medical-records / examinations / accounting lists */
  status: "w-[100px] min-w-[100px] whitespace-nowrap",
} as const;
