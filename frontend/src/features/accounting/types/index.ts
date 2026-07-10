/**
 * Accounting feature types (UI-facing: camelCase, string IDs)
 * Backend types: {@link Billing}, {@link BillingItem}, {@link Payment} from models.ts
 */
import {
  BillingStatusWaiting,
  BillingStatusCompleted,
  BillingStatusCancelled,
  BillingStatusPending,
  PaymentMethodCash,
  PaymentMethodCreditCard,
  PaymentMethodElectronicMoney,
  PaymentMethodBankTransfer,
  ItemCategoryExamination,
  ItemCategoryTest,
  ItemCategoryProcedure,
  ItemCategorySurgery,
  ItemCategoryMedicine,
  ItemCategoryFood,
  ItemCategoryGoods,
  ItemCategoryOther,
  ItemCategoryTrimming,
  ItemCategoryVaccine,
  ItemCategoryHotel,
  ItemCategoryTraining,
} from "@/types/generated/models";
import type { TaxType } from "@/types/generated/models";

/** @see {@link import("@/types/generated/models").BillingStatus} */
export type AccountingStatus =
  | typeof BillingStatusWaiting
  | typeof BillingStatusCompleted
  | typeof BillingStatusCancelled
  | typeof BillingStatusPending;

/** @see {@link import("@/types/generated/models").PaymentMethod} */
export type PaymentMethod =
  | typeof PaymentMethodCash
  | typeof PaymentMethodCreditCard
  | typeof PaymentMethodElectronicMoney
  | typeof PaymentMethodBankTransfer;

/** @see {@link import("@/types/generated/models").ItemCategory} */
export type ItemCategory =
  | typeof ItemCategoryExamination
  | typeof ItemCategoryTest
  | typeof ItemCategoryProcedure
  | typeof ItemCategorySurgery
  | typeof ItemCategoryMedicine
  | typeof ItemCategoryFood
  | typeof ItemCategoryGoods
  | typeof ItemCategoryOther
  | typeof ItemCategoryTrimming
  | typeof ItemCategoryVaccine
  | typeof ItemCategoryHotel
  | typeof ItemCategoryTraining;

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
