import { describe, expect, it, vi } from "vitest";

import type { AccountingItem } from "../types";
import { createAccountingItemsSequentially } from "./create-accounting-items";

const ITEMS = [
  {
    id: "item-1",
    category: "other",
    name: "診察料",
    unitPrice: 1000,
    quantity: 1,
    discountRate: 0,
    discountAmount: 0,
    taxType: "excluded",
    taxRate: 0.1,
    taxAmount: 100,
    subtotal: 1000,
    isInsuranceApplicable: false,
    source: "manual",
  },
  {
    id: "item-2",
    category: "medicine",
    name: "内服薬",
    unitPrice: 500,
    quantity: 2,
    discountRate: 0,
    discountAmount: 0,
    taxType: "included",
    taxRate: 0.1,
    taxAmount: 91,
    subtotal: 1000,
    isInsuranceApplicable: true,
    source: "medical_record",
    treatmentId: "10",
  },
] satisfies AccountingItem[];

describe("createAccountingItems", () => {
  it("明細作成が失敗した時点で後続POSTを止め、会計の部分失敗を限定する", async () => {
    const createItem = vi.fn()
      .mockResolvedValueOnce({})
      .mockRejectedValueOnce(new Error("item create failed"));

    await expect(
      createAccountingItemsSequentially(42, [...ITEMS, { ...ITEMS[0], id: "item-3" }], createItem),
    ).rejects.toThrow("item create failed");

    expect(createItem).toHaveBeenCalledTimes(2);
    expect(createItem.mock.calls.map(([request]) => request.name)).toEqual(["診察料", "内服薬"]);
  });
});
