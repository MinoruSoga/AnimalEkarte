/**
 * Accounting feature types (UI-facing: camelCase, string IDs)
 * Backend types: {@link Billing}, {@link BillingItem}, {@link Payment} from models.ts
 */
import type { TaxType } from "@/types/generated/models";

/**
 * 手書き literal union（FE4-1）。
 * tygo 生成定数は `export const X: BillingStatus = "waiting"` 形式（BillingStatus = string）のため
 * `typeof 生成定数` は string に退化する（型安全性が失われる）。生成側は編集禁止のため、
 * 手書き literal union + ランタイム値集合の drift テスト（./union-drift.test.ts）で
 * 生成定数の値集合との一致を機械固定する。
 * @see {@link import("@/types/generated/models").BillingStatus}
 */
export type AccountingStatus = "waiting" | "pending" | "completed" | "cancelled";

/** 手書き literal union（FE4-1・上記 AccountingStatus と同じ理由）。@see {@link import("@/types/generated/models").PaymentMethod} */
export type PaymentMethod = "cash" | "credit_card" | "electronic_money" | "bank_transfer";

/** 手書き literal union（FE4-1・上記 AccountingStatus と同じ理由）。@see {@link import("@/types/generated/models").ItemCategory} */
export type ItemCategory =
  | "examination"
  | "test"
  | "procedure"
  | "surgery"
  | "medicine"
  | "food"
  | "goods"
  | "other"
  | "trimming"
  | "vaccine"
  | "hotel"
  | "training";

/** @see {@link import("@/types/generated/models").BillingItem} */
export interface AccountingItem {
  id: string;
  code?: string;
  category: ItemCategory;
  name: string;
  unitPrice: number;
  quantity: number;
  discountRate: number;   // #85: 項目別割引率(%)。入力補助、実控除は discountAmount
  discountAmount: number; // #85: 項目別割引額(円)。実際の控除額
  taxType: TaxType;
  taxRate: number;
  taxAmount: number; // BE が計算して返す
  subtotal: number;  // (unit_price × quantity − 割引額)（税抜・割引後 #85）
  isInsuranceApplicable: boolean;
  source: "medical_record" | "manual" | "hospitalization" | "trimming";
  treatmentId?: string;
  appointmentId?: string;
  trimmingCourseId?: string;
  trimmingOptionId?: string;
}

/** @see {@link import("@/types/generated/models").Payment} */
export interface PaymentInfo {
  subtotal: number; // 税抜小計
  taxTotal: number; // 消費税合計
  totalAmount: number; // 税込合計
  insuranceName?: string; // 保険会社名
  insuranceRatio?: number; // 負担割合 (0.5, 0.7 etc)
  insuranceAmount: number; // 保険負担額（マイナスのみ）
  discountAmount: number; // 値引き（マイナスのみ）
  billingAmount: number; // 請求金額 (total - insurance - discount)
  receivedAmount: number; // 預り金
  changeAmount: number; // お釣り
  method: PaymentMethod;
  paidByName?: string; // 支払処理スタッフ名
}

/** @see {@link import("@/types/generated/models").PaymentSplit} */
export interface PaymentSplitInfo {
  id: string;
  method: PaymentMethod;
  paymentMethodId?: string;
  amount: number;
  receivedAmount: number;
  changeAmount: number;
  paidByName?: string;
}

/** @see {@link import("@/types/generated/models").Billing} */
export interface Accounting {
  id: string;
  clinicId: string;
  medicalRecordId?: string;
  ownerId: string;
  ownerName: string;
  petId: string;
  petName: string;
  petSpecies?: string;
  status: AccountingStatus;
  scheduledDate: string; // 会計予定日（診療日）
  completedAt?: string; // 会計完了日時
  items: AccountingItem[];
  payment?: PaymentInfo;
  paymentSplits?: PaymentSplitInfo[];
  totalRefundedAmount: number; // 返金合計（0の場合はバッジ非表示）
  memo?: string;
}
