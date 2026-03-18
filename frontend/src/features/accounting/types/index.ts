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
  ItemCategoryExamination,
  ItemCategoryTest,
  ItemCategoryProcedure,
  ItemCategorySurgery,
  ItemCategoryMedicine,
  ItemCategoryFood,
  ItemCategoryGoods,
  ItemCategoryOther,
} from "@/types/generated/models";

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
  | typeof PaymentMethodElectronicMoney;

/** @see {@link import("@/types/generated/models").ItemCategory} */
export type ItemCategory =
  | typeof ItemCategoryExamination
  | typeof ItemCategoryTest
  | typeof ItemCategoryProcedure
  | typeof ItemCategorySurgery
  | typeof ItemCategoryMedicine
  | typeof ItemCategoryFood
  | typeof ItemCategoryGoods
  | typeof ItemCategoryOther;

/** @see {@link import("@/types/generated/models").BillingItem} */
export interface AccountingItem {
  id: string;
  code?: string;
  category: ItemCategory;
  name: string;
  unitPrice: number;
  quantity: number;
  taxRate: 0.1 | 0.08; // 10% or 8%
  isInsuranceApplicable: boolean;
  source: "medical_record" | "manual"; // カルテ連携か手動追加か
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
}

/** @see {@link import("@/types/generated/models").Billing} */
export interface Accounting {
  id: string;
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
  memo?: string;
}
