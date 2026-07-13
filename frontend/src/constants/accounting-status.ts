import type { BillingStatus } from "@/types/generated/models";

// FE5-27: 履歴側(OwnerAccountingHistoryParts)の4値区別マップを正本化。
// 一覧側(AccountingListTable)は従来 waiting/pending をともに「会計待ち」に潰していたが、
// PO決定P-3により保留を区別表示する。
export const ACCOUNTING_STATUS_LABELS: Record<BillingStatus, string> = {
  waiting: "未精算",
  pending: "保留",
  completed: "精算済",
  cancelled: "取消",
};
