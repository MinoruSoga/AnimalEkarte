import { describe, it, expect } from "vitest";
import { transformCashRegisterClose } from "./cash-register";
import type { CashRegisterClose as BackendCashRegisterClose } from "@/types/generated/models";

const minimalClose: BackendCashRegisterClose = {
  id: 1,
  clinic_id: 1,
  close_date: "2026-07-10",
  period: "am",
  theoretical_cash: 10000,
  actual_cash: 10000,
  cash_difference: 0,
  category_breakdown: {
    categories: {
      examination: { cash: 3000 },
      medicine: { cash: 2000 },
    },
  },
  memo: "",
  closed_at: "2026-07-10T09:00:00Z",
  created_at: "2026-07-10T09:00:00Z",
  updated_at: "2026-07-10T09:00:00Z",
};

describe("transformCashRegisterClose", () => {
  it("categoryBreakdown を型上 unknown として保持しつつ実行時値をそのまま返す", () => {
    const result = transformCashRegisterClose(minimalClose);

    // 実行時の中身は従来どおり category_breakdown の生オブジェクト
    expect(result.categoryBreakdown).toEqual(minimalClose.category_breakdown);
  });
});
