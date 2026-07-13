import type { PaymentMethod } from "@/types/generated/models";

/** 支払方法の正式ラベル（一覧・ダイアログ等の標準表示。FE5-7） */
export const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  cash: "現金",
  credit_card: "クレジットカード",
  electronic_money: "電子マネー",
  bank_transfer: "銀行振込",
};
