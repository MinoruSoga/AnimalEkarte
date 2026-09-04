import type { EstimateStatus } from "../types";

/** Create API（binding oneof=draft sent）と整合する許可 status — UI / hook の単一ソース */
export const CREATE_ALLOWED_STATUSES: readonly EstimateStatus[] = ["draft", "sent"];

/** Edit は全 status（4 値）を選択可 */
export const EDIT_STATUS_OPTIONS: { value: EstimateStatus; label: string }[] = [
  { value: "draft", label: "下書き" },
  { value: "sent", label: "送付済み" },
  { value: "approved", label: "承認済み" },
  { value: "rejected", label: "却下" },
];

/** Create は CREATE_ALLOWED_STATUSES 由来（draft / sent のみ） */
export const CREATE_STATUS_OPTIONS = EDIT_STATUS_OPTIONS.filter((opt) =>
  CREATE_ALLOWED_STATUSES.includes(opt.value),
);
