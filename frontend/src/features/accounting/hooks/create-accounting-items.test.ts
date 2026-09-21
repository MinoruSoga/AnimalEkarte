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
  it("商品マスタ由来の明細では merchandise_item_id を送る", async () => {
    const createItem = vi.fn().mockResolvedValue({});

    await createAccountingItemsSequentially(
      42,
      [{ ...ITEMS[0], merchandiseItemId: "77" }],
      createItem,
    );

    expect(createItem).toHaveBeenCalledWith(expect.objectContaining({ merchandise_item_id: 77 }));
  });

  it("ワクチン接種由来の明細では vaccination_id を数値で送る", async () => {
    const createItem = vi.fn().mockResolvedValue({});

    await createAccountingItemsSequentially(42, [{ ...ITEMS[1], vaccinationId: "88" }], createItem);

    expect(createItem).toHaveBeenCalledWith(expect.objectContaining({ vaccination_id: 88 }));
  });

  it("検査由来の明細では exam_id を数値で送る", async () => {
    const createItem = vi.fn().mockResolvedValue({});

    await createAccountingItemsSequentially(42, [{ ...ITEMS[1], examId: "55" }], createItem);

    expect(createItem).toHaveBeenCalledWith(expect.objectContaining({ exam_id: 55 }));
  });

  it("手入力otherの理由を other_reason として送る", async () => {
    const createItem = vi.fn().mockResolvedValue({});

    await createAccountingItemsSequentially(
      42,
      [{ ...ITEMS[0], otherReason: "締め時に確認する分類" }],
      createItem,
    );

    expect(createItem).toHaveBeenCalledWith(
      expect.objectContaining({ other_reason: "締め時に確認する分類" }),
    );
  });

  it("other以外または手入力以外の明細では other_reason を送らない", async () => {
    const createItem = vi.fn().mockResolvedValue({});

    await createAccountingItemsSequentially(
      42,
      [
        { ...ITEMS[0], category: "test", otherReason: "送信しない理由" },
        { ...ITEMS[1], category: "other", otherReason: "送信しない理由" },
      ],
      createItem,
    );

    expect(createItem).toHaveBeenCalledTimes(2);
    expect(createItem.mock.calls[0]?.[0]).not.toHaveProperty("other_reason");
    expect(createItem.mock.calls[1]?.[0]).not.toHaveProperty("other_reason");
  });

  it("明細作成が失敗した時点で後続POSTを止め、会計の部分失敗を限定する", async () => {
    const createItem = vi
      .fn()
      .mockResolvedValueOnce({})
      .mockRejectedValueOnce(new Error("item create failed"));

    await expect(
      createAccountingItemsSequentially(42, [...ITEMS, { ...ITEMS[0], id: "item-3" }], createItem),
    ).rejects.toThrow("item create failed");

    expect(createItem).toHaveBeenCalledTimes(2);
    expect(createItem.mock.calls.map(([request]) => request.name)).toEqual(["診察料", "内服薬"]);
  });

  it("治療明細では treatment_id を送り hospitalization/cage/checkup id を混在させない", async () => {
    const createItem = vi.fn().mockResolvedValue({});

    await createAccountingItemsSequentially(
      42,
      [{ ...ITEMS[1], treatmentId: "91", source: "medical_record" }],
      createItem,
    );

    const payload = createItem.mock.calls[0]?.[0] as Record<string, unknown>;
    expect(payload).toEqual(
      expect.objectContaining({
        treatment_id: 91,
      }),
    );
    expect(payload.merchandise_item_id).toBeUndefined();
    expect(payload.vaccination_id).toBeUndefined();
    expect(payload.exam_id).toBeUndefined();
    expect(payload).not.toHaveProperty("hospitalization_id");
    expect(payload).not.toHaveProperty("cage_id");
    expect(payload).not.toHaveProperty("checkup_id");
  });

  it("トリミング明細では course/option id のみを送り他マスタ id を混在させない", async () => {
    const createItem = vi.fn().mockResolvedValue({});

    await createAccountingItemsSequentially(
      42,
      [
        {
          ...ITEMS[0],
          name: "カット",
          trimmingCourseId: "15",
          trimmingOptionId: "16",
          source: "trimming",
        },
      ],
      createItem,
    );

    const payload = createItem.mock.calls[0]?.[0] as Record<string, unknown>;
    expect(payload).toEqual(
      expect.objectContaining({
        trimming_course_id: 15,
        trimming_option_id: 16,
      }),
    );
    expect(payload.merchandise_item_id).toBeUndefined();
    expect(payload.treatment_id).toBeUndefined();
    expect(payload.vaccination_id).toBeUndefined();
    expect(payload.exam_id).toBeUndefined();
    expect(payload).not.toHaveProperty("hospitalization_id");
    expect(payload).not.toHaveProperty("cage_id");
    expect(payload).not.toHaveProperty("checkup_id");
  });
});
