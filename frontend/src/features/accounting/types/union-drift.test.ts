import { describe, it, expect } from "vitest";
import * as gen from "@/types/generated/models";
import type { AccountingStatus, PaymentMethod, ItemCategory } from "./index";

// FE4-1: tygo 生成定数は typeof で string に退化するため、
// 手書き literal union の値集合が生成定数のランタイム値と完全一致することを機械固定する。
describe("accounting union drift", () => {
  it("AccountingStatus の値集合が BillingStatus* 生成定数と一致する", () => {
    const values: AccountingStatus[] = ["waiting", "pending", "completed", "cancelled"];
    expect(new Set<string>(values)).toEqual(
      new Set([
        gen.BillingStatusWaiting,
        gen.BillingStatusPending,
        gen.BillingStatusCompleted,
        gen.BillingStatusCancelled,
      ]),
    );
  });

  it("PaymentMethod の値集合が PaymentMethod* 生成定数と一致する", () => {
    const values: PaymentMethod[] = ["cash", "credit_card", "electronic_money", "bank_transfer"];
    expect(new Set<string>(values)).toEqual(
      new Set([
        gen.PaymentMethodCash,
        gen.PaymentMethodCreditCard,
        gen.PaymentMethodElectronicMoney,
        gen.PaymentMethodBankTransfer,
      ]),
    );
  });

  it("ItemCategory の値集合が ItemCategory* 生成定数と一致する", () => {
    const values: ItemCategory[] = [
      "examination",
      "test",
      "procedure",
      "surgery",
      "medicine",
      "food",
      "goods",
      "other",
      "trimming",
      "vaccine",
      "hotel",
      "training",
    ];
    expect(new Set<string>(values)).toEqual(
      new Set([
        gen.ItemCategoryExamination,
        gen.ItemCategoryTest,
        gen.ItemCategoryProcedure,
        gen.ItemCategorySurgery,
        gen.ItemCategoryMedicine,
        gen.ItemCategoryFood,
        gen.ItemCategoryGoods,
        gen.ItemCategoryOther,
        gen.ItemCategoryTrimming,
        gen.ItemCategoryVaccine,
        gen.ItemCategoryHotel,
        gen.ItemCategoryTraining,
      ]),
    );
  });
});
