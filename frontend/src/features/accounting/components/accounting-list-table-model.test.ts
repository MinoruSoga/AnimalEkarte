import { describe, expect, it } from "vitest";
import type { Accounting } from "../types";
import { calculateAccountingTotal } from "../lib/accounting-list-table-model";

describe("calculateAccountingTotal recorded negatives", () => {
  it("入金が無い明細の負額を 0 にクランプしない", () => {
    const accounting = {
      id: "1",
      clinicId: "1",
      ownerId: "1",
      ownerName: "a",
      petId: "1",
      petName: "b",
      status: "waiting",
      scheduledDate: "2026-08-24",
      items: [
        {
          id: "i",
          category: "examination",
          name: "赤伝",
          unitPrice: -3000,
          quantity: 1,
          discountRate: 0,
          discountAmount: 0,
          taxType: "excluded",
          taxRate: 0.1,
          taxAmount: -300,
          subtotal: -3000,
          isInsuranceApplicable: false,
          source: "manual",
        },
      ],
      totalRefundedAmount: 0,
    } as Accounting;
    expect(calculateAccountingTotal(accounting)).toBe(-3300);
  });

  it("一覧で items が空なら billings.total_amount を使う", () => {
    const accounting = {
      id: "1",
      clinicId: "1",
      ownerId: "1",
      ownerName: "a",
      petId: "1",
      petName: "b",
      status: "waiting",
      scheduledDate: "2026-08-24",
      items: [],
      totalAmount: 4400,
      totalRefundedAmount: 0,
    } as Accounting;
    expect(calculateAccountingTotal(accounting)).toBe(4400);
  });
});
