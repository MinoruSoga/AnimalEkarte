import { describe, expect, it } from "vitest";

import type { TreatmentMasterItem } from "@/components/shared/TreatmentSearchDialog/TreatmentSearchDialog";

import {
  buildCopiedTreatmentPayload,
  buildMasterSelectionPayload,
  resolveItemTypeFromCategory,
  type CopyTreatmentMasters,
} from "../../lib/treatments-tab-model";
import type { Treatment } from "../../types";

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

// ─────────────────────────────────────────────────────────────
// EMR-219: 問診履歴コピー時の治療明細複写ペイロード
// ─────────────────────────────────────────────────────────────
describe("buildCopiedTreatmentPayload", () => {
  const baseSource: Treatment = {
    id: "500",
    medical_record_id: "101",
    item_type: "consultation",
    unit_price: 1500,
    quantity: 1,
    is_selected: true,
    status: "pending",
    content: "初診料",
    memo: "メモ",
    is_insurance: false,
    discount_rate: 10,
    discount_amount: 100,
    sort_order: 2,
    version: 1,
    created_at: "2026-02-01T00:00:00Z",
    updated_at: "2026-02-01T00:00:00Z",
  };

  const masters: CopyTreatmentMasters = {
    consultations: [{ id: "7", price: 2500 }],
    procedures: [{ id: "8", price: 4000 }],
    medicines: [{ id: "42", price: 120 }],
  };

  it("consultation 行はマスタ一致で master.id と現行価格へ再解決する（当時価格を持ち込まない）", () => {
    const payload = buildCopiedTreatmentPayload({
      source: { ...baseSource, consultation_id: "7" },
      masters,
      sortOrderOffset: 0,
    });

    expect(payload.consultation_id).toBe("7");
    expect(payload.unit_price).toBe(2500);
    expect(payload.item_type).toBe("consultation");
    expect(payload.sort_order).toBe(2);
  });

  it("procedure 行はマスタ一致で master.id と現行価格へ再解決する", () => {
    const payload = buildCopiedTreatmentPayload({
      source: {
        ...baseSource,
        item_type: "procedure",
        procedure_id: "8",
        consultation_id: undefined,
        unit_price: 999,
      },
      masters,
      sortOrderOffset: 0,
    });

    expect(payload.procedure_id).toBe("8");
    expect(payload.unit_price).toBe(4000);
  });

  it("medicine 行はマスタ一致で medicine_id を JSON number、unit_price を現行価格にする", () => {
    const payload = buildCopiedTreatmentPayload({
      source: {
        ...baseSource,
        item_type: "medicine",
        medicine_id: "42",
        consultation_id: undefined,
        unit_price: 80,
      },
      masters,
      sortOrderOffset: 0,
    });

    expect(payload.medicine_id).toBe(42);
    expect(typeof payload.medicine_id).toBe("number");
    expect(payload.unit_price).toBe(120);
  });

  it("マスタ不一致の consultation 行はソースの unit_price と参照を維持する", () => {
    const payload = buildCopiedTreatmentPayload({
      source: { ...baseSource, consultation_id: "999" },
      masters,
      sortOrderOffset: 0,
    });

    expect(payload.consultation_id).toBe("999");
    expect(payload.unit_price).toBe(1500);
  });

  it("マスタ不一致の medicine 行はソース medicine_id を number 化して送る", () => {
    const payload = buildCopiedTreatmentPayload({
      source: {
        ...baseSource,
        item_type: "medicine",
        medicine_id: "55",
        consultation_id: undefined,
      },
      masters,
      sortOrderOffset: 0,
    });

    expect(payload.medicine_id).toBe(55);
    expect(payload.unit_price).toBe(1500);
  });

  it("other（無型）行は単価・参照を含めソース値をそのまま保持する", () => {
    const payload = buildCopiedTreatmentPayload({
      source: {
        ...baseSource,
        item_type: "other",
        consultation_id: undefined,
        procedure_id: undefined,
        medicine_id: undefined,
        content: "検査キット",
        unit_price: 3200,
      },
      masters,
      sortOrderOffset: 0,
    });

    expect(payload.item_type).toBe("other");
    expect(payload.unit_price).toBe(3200);
    expect(payload.consultation_id).toBeUndefined();
    expect(payload.procedure_id).toBeUndefined();
    expect(payload.medicine_id).toBeUndefined();
  });

  it("content/memo/quantity/is_selected/is_insurance/discount/status をソースから保持する", () => {
    const payload = buildCopiedTreatmentPayload({
      source: {
        ...baseSource,
        consultation_id: "7",
        content: "再診料",
        memo: "要経過観察",
        quantity: 2.5,
        is_selected: false,
        is_insurance: true,
        discount_rate: 5,
        discount_amount: 250,
        status: "completed",
      },
      masters,
      sortOrderOffset: 0,
    });

    expect(payload.content).toBe("再診料");
    expect(payload.memo).toBe("要経過観察");
    expect(payload.quantity).toBe(2.5);
    expect(payload.is_selected).toBe(false);
    expect(payload.is_insurance).toBe(true);
    expect(payload.discount_rate).toBe(5);
    expect(payload.discount_amount).toBe(250);
    expect(payload.status).toBe("completed");
  });

  it("sort_order は複写先の現在最大 +1（offset）をソース値へ加算する", () => {
    const payload = buildCopiedTreatmentPayload({
      source: { ...baseSource, sort_order: 0 },
      masters,
      sortOrderOffset: 5,
    });

    expect(payload.sort_order).toBe(5);
  });

  it("dose_param_snapshot の dose_deviation_reason をペイロードへ引き継ぐ", () => {
    const payload = buildCopiedTreatmentPayload({
      source: {
        ...baseSource,
        item_type: "medicine",
        medicine_id: "42",
        consultation_id: undefined,
        dose_param_snapshot: { dose_deviation_reason: "臨床判断による増量" },
      },
      masters,
      sortOrderOffset: 0,
    });

    expect(payload.dose_deviation_reason).toBe("臨床判断による増量");
  });
});
