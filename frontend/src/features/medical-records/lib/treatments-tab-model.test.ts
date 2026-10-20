import { describe, expect, it } from "vitest";

import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

import { buildMasterSelectionPayload } from "./treatments-tab-model";

/** 既定選択フィクスチャ。overrides でテスト対象フィールドだけを変える。 */
function masterItem(overrides: Partial<TreatmentMasterItem> = {}): TreatmentMasterItem {
  return {
    id: "1",
    name: "テスト項目",
    unitPrice: 1000,
    category: "検査",
    insuranceContract: "master",
    isNonInsurance: true,
    ...overrides,
  };
}

describe("buildMasterSelectionPayload (EMR-280)", () => {
  it.each([
    {
      category: "診察",
      id: "7",
      unitPrice: 3307,
      expected: { item_type: "consultation", consultation_id: 7 },
    },
    {
      category: "処置",
      id: "8",
      unitPrice: 3306,
      expected: { item_type: "procedure", procedure_id: 8 },
    },
  ])("$category 選択は参照 id を number で紐付ける", ({ category, id, unitPrice, expected }) => {
    const payload = buildMasterSelectionPayload({
      item: masterItem({ category, id, unitPrice }),
      quantity: 1,
      sortOrder: 0,
    });

    expect(payload).toMatchObject(expected);
    if (expected.item_type === "consultation") {
      expect(typeof payload.consultation_id).toBe("number");
      expect(payload.procedure_id).toBeUndefined();
    } else {
      expect(typeof payload.procedure_id).toBe("number");
      expect(payload.consultation_id).toBeUndefined();
    }
    expect(payload.medicine_id).toBeUndefined();
    expect(payload.unit_price).toBe(unitPrice);
  });

  it.each(["検査", "予防", "入院", "その他"])("%s の同一 id は参照 id を紐付けない", (category) => {
    const payload = buildMasterSelectionPayload({
      item: masterItem({ category, id: "7" }),
      quantity: 1,
      sortOrder: 0,
    });

    expect(payload.item_type).toBe("other");
    expect(payload.consultation_id).toBeUndefined();
    expect(payload.procedure_id).toBeUndefined();
    expect(payload.medicine_id).toBeUndefined();
  });

  it("診察/処置でも数値化できない id は参照を送信しない（0 を誤送信しない）", () => {
    const payload = buildMasterSelectionPayload({
      item: masterItem({ category: "診察", id: "not-a-number" }),
      quantity: 1,
      sortOrder: 0,
    });

    expect(payload.item_type).toBe("consultation");
    expect(payload.consultation_id).toBeUndefined();
  });

  it("薬剤選択の medicine_id number 化契約は維持される", () => {
    const payload = buildMasterSelectionPayload({
      item: masterItem({
        category: "薬剤",
        id: "42",
        unitPrice: 3303,
        medicineId: "42",
      }),
      quantity: 2,
      sortOrder: 0,
    });

    expect(payload.item_type).toBe("medicine");
    expect(payload.medicine_id).toBe(42);
    expect(typeof payload.medicine_id).toBe("number");
    expect(payload.consultation_id).toBeUndefined();
    expect(payload.procedure_id).toBeUndefined();
  });
});

describe("buildMasterSelectionPayload 品目別保険設定の受渡し (EMR-454)", () => {
  it.each([
    { category: "診察", isNonInsurance: true, expected: false },
    { category: "診察", isNonInsurance: false, expected: true },
    { category: "処置", isNonInsurance: true, expected: false },
    { category: "処置", isNonInsurance: false, expected: true },
    { category: "薬剤", isNonInsurance: true, expected: false },
    { category: "薬剤", isNonInsurance: false, expected: true },
  ])(
    "$category の master 明示値 is_non_insurance=$isNonInsurance は is_insurance=$expected へ反転写像される",
    ({ category, isNonInsurance, expected }) => {
      const payload = buildMasterSelectionPayload({
        item: masterItem({ category, id: "5", isNonInsurance }),
        quantity: 1,
        sortOrder: 0,
      });

      expect(payload.is_insurance).toBe(expected);
    },
  );

  it.each([null, undefined])(
    "unknown insurance %s is blocked before payload creation",
    (isNonInsurance) => {
      expect(() =>
        buildMasterSelectionPayload({
          item: masterItem({ isNonInsurance }),
          quantity: 1,
          sortOrder: 0,
        }),
      ).toThrow("保険設定が未設定");
    },
  );
});

it.each(
  [true, false].flatMap((insuranceFallback) =>
    [null, undefined].map((isNonInsurance) => ({ insuranceFallback, isNonInsurance })),
  ),
)(
  "adopted unknown $isNonInsurance rejects caller fallback $insuranceFallback",
  ({ insuranceFallback, isNonInsurance }) => {
    expect(() =>
      buildMasterSelectionPayload({
        item: masterItem({ isNonInsurance }),
        quantity: 1,
        sortOrder: 0,
        insuranceFallback,
      }),
    ).toThrow("保険設定が未設定");
  },
);

// 未採用では master 値を推測せず、caller の従来既定だけを維持する。
it.each(
  [undefined, true, false].flatMap((insuranceFallback) =>
    [undefined, null, true, false].map((isNonInsurance) => ({ insuranceFallback, isNonInsurance })),
  ),
)(
  "legacy ignores setting $isNonInsurance and preserves fallback $insuranceFallback",
  ({ insuranceFallback, isNonInsurance }) => {
    const payload = buildMasterSelectionPayload({
      item: masterItem({ insuranceContract: undefined, isNonInsurance }),
      quantity: 1,
      sortOrder: 0,
      insuranceFallback,
    });
    expect(payload.is_insurance).toBe(insuranceFallback ?? false);
  },
);
