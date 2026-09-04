import { describe, expect, it } from "vitest";

import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

import {
  buildMasterSelectionPayload,
  resolveItemTypeFromCategory,
} from "../../lib/treatments-tab-model";

describe("resolveItemTypeFromCategory", () => {
  it.each([
    ["薬剤", "medicine"],
    ["処置", "procedure"],
    ["診察", "consultation"],
    ["検査", "other"],
    ["予防", "other"],
  ])("%s -> %s", (category, want) => {
    expect(resolveItemTypeFromCategory(category)).toBe(want);
  });
});

describe("buildMasterSelectionPayload", () => {
  // ts-review-201 CRITICAL 回帰テスト: medicine_id は BE の *uint64 (JSON number) 期待に
  // 合わせて必ず number へ変換されていること。string のまま送ると全ての薬剤選択が 400 になる。
  it("薬剤選択時、medicine_id は文字列ではなく number に変換される", () => {
    const item: TreatmentMasterItem = {
      id: "42",
      name: "アモキシシリン",
      unitPrice: 100,
      category: "薬剤",
      medicineId: "42",
    };

    const payload = buildMasterSelectionPayload({ item, quantity: 2, sortOrder: 0 });

    expect(payload.medicine_id).toBe(42);
    expect(typeof payload.medicine_id).toBe("number");
    expect(payload.item_type).toBe("medicine");
    expect(payload.quantity).toBe(2);
  });

  it("薬剤以外の選択では medicine_id は undefined になる", () => {
    const item: TreatmentMasterItem = {
      id: "7",
      name: "一般診察",
      unitPrice: 3000,
      category: "診察",
    };

    const payload = buildMasterSelectionPayload({ item, quantity: 1, sortOrder: 3 });

    expect(payload.medicine_id).toBeUndefined();
    expect(payload.item_type).toBe("consultation");
    expect(payload.sort_order).toBe(3);
  });

  it("medicineId が空文字の場合も medicine_id は undefined（0 を誤送信しない）", () => {
    const item: TreatmentMasterItem = {
      id: "1",
      name: "テスト",
      unitPrice: 0,
      category: "薬剤",
      medicineId: "",
    };

    const payload = buildMasterSelectionPayload({ item, quantity: 1, sortOrder: 0 });

    expect(payload.medicine_id).toBeUndefined();
  });
});
