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
