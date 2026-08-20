/**
 * Accounting feature types (UI-facing: camelCase, string IDs)
 * Backend types: {@link Billing}, {@link BillingItem}, {@link Payment} from models.ts
 */
// FE6-3: tygo enum_style: "union"（FE6-1/FE6-2）により生成定数が真の literal union に
// なったため、手書き union を生成型からの re-export へ移行した。drift テストは不要になり
// union-drift.test.ts から削除済み。AccountingStatus は生成側の BillingStatus に対応する
// （FE4-1 当時から名称が分岐していたため as で明示的にリネームして re-export する）。
import type {
  TaxType,
  BillingStatus as AccountingStatus,
  PaymentMethod,
  ItemCategory,
} from "@/types/generated/models";
export type { AccountingStatus, PaymentMethod, ItemCategory };

export interface AddAccountingItemInput {
  name: string;
  price: string;
  category: string;
  otherReason?: string;
  taxRate?: number;
  merchandiseItemId?: string;
}

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
  otherReason?: string;
  merchandiseItemId?: string;
  vaccinationId?: string;
  examId?: string;
  treatmentId?: string;
  /** treatment 由来の親カルテ（未請求候補などで付与） */
  medicalRecordId?: string;
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
  /** BUG-007: 未収残高（waiting 全額 or クレジット訂正後の patient_due−支払額） */
  outstandingAmount?: number;
  memo?: string;
}
