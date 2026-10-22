import { describe, expect, it } from "vitest";

import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import type { Medicine } from "@/types";

import {
  MEDICINE_STRENGTH_REQUIRED_MESSAGE,
  medicineToFormData,
  type MedicineFormData,
} from "../lib/medicine-side-panel-model";
import {
  buildMedicineCreateRequest,
  buildMedicineUpdateRequest,
  groupFilteredMedicines,
  resolveMedicineDrag,
  validateMedicineForm,
} from "./medicine-settings-model";

function makeCategory(overrides: Partial<Medicine> = {}): Medicine {
  return makeMedicine({
    id: "10",
    name: "抗生剤",
    price: 0,
    dosageForm: undefined,
    medicineUnit: undefined,
    ...overrides,
  });
}

function makeMedicine(overrides: Partial<Medicine> = {}): Medicine {
  return {
    id: "1",
    name: "抗生剤A",
    parentId: undefined,
    dosageForm: "tablet",
    medicineUnit: "per_tablet",
    price: 1000,
    defaultQuantity: 1,
    inventoryId: undefined,
    description: "",
    isActive: true,
    sortOrder: 0,
    taxType: "excluded",
    taxRate: 0.1,
    isNonInsurance: false,
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function makeFormData(overrides: Partial<MedicineFormData> = {}): MedicineFormData {
  return {
    name: "抗生剤A",
    parentId: "",
    dosageForm: "tablet",
    medicineUnit: "per_tablet",
    price: 1000,
    description: "",
    isActive: true,
    taxType: "excluded",
    taxRate: 0.1,
    isNonInsurance: false,
    calculationType: "none",
    strength: "",
    frequencyPerDay: "",
    defaultDurationDays: "",
    ...overrides,
  };
}

describe("groupFilteredMedicines", () => {
  it("groups items under their parent category header", () => {
    const category = makeCategory();
    const child = makeMedicine({ id: "11", name: "アモキシシリン", parentId: "10" });
    const medicinesById = new Map([
      [category.id, category],
      [child.id, child],
    ]);

    const result = groupFilteredMedicines({
      orderedMedicines: [category, child],
      activeFilters: [],
      searchTerm: "",
      medicinesById,
    });

    expect(result.groupedMedicines.get("10")).toEqual({ header: category, items: [child] });
    expect(result.ungroupedMedicines).toEqual([]);
    expect(result.totalCount).toBe(2);
  });

  it("creates an empty group entry for a category with no children yet", () => {
    const category = makeCategory();
    const medicinesById = new Map([[category.id, category]]);

    const result = groupFilteredMedicines({
      orderedMedicines: [category],
      activeFilters: [],
      searchTerm: "",
      medicinesById,
    });

    expect(result.groupedMedicines.get("10")).toEqual({ header: category, items: [] });
  });

  it("treats a parentless medicine with dosage form as ungrouped even when price is 0 (BUG-006)", () => {
    const medicine = makeMedicine({ id: "21", price: 0, parentId: undefined });
    const medicinesById = new Map([[medicine.id, medicine]]);

    const result = groupFilteredMedicines({
      orderedMedicines: [medicine],
      activeFilters: [],
      searchTerm: "",
      medicinesById,
    });

    expect(result.groupedMedicines.size).toBe(0);
    expect(result.ungroupedMedicines).toEqual([medicine]);
  });

  it("treats a priced, parentless medicine as ungrouped", () => {
    const medicine = makeMedicine({ id: "20", price: 500 });
    const medicinesById = new Map([[medicine.id, medicine]]);

    const result = groupFilteredMedicines({
      orderedMedicines: [medicine],
      activeFilters: [],
      searchTerm: "",
      medicinesById,
    });

    expect(result.groupedMedicines.size).toBe(0);
    expect(result.ungroupedMedicines).toEqual([medicine]);
  });

  it("falls back to ungrouped when the parent category cannot be found", () => {
    const orphan = makeMedicine({ id: "30", parentId: "missing-parent" });
    const medicinesById = new Map<string, Medicine>();

    const result = groupFilteredMedicines({
      orderedMedicines: [orphan],
      activeFilters: [],
      searchTerm: "",
      medicinesById,
    });

    expect(result.groupedMedicines.size).toBe(0);
    expect(result.ungroupedMedicines).toEqual([orphan]);
  });

  it("applies the status filter (is active)", () => {
    const active = makeMedicine({ id: "1", isActive: true, price: 100 });
    const inactive = makeMedicine({ id: "2", isActive: false, price: 100 });
    const medicinesById = new Map([
      [active.id, active],
      [inactive.id, inactive],
    ]);
    const activeFilters: ActiveFilter[] = [
      { key: "status", condition: "is", value: "active", displayValue: "有効" },
    ];

    const result = groupFilteredMedicines({
      orderedMedicines: [active, inactive],
      activeFilters,
      searchTerm: "",
      medicinesById,
    });

    expect(result.ungroupedMedicines).toEqual([active]);
    expect(result.totalCount).toBe(1);
  });

  it("applies the status filter (is not active)", () => {
    const active = makeMedicine({ id: "1", isActive: true, price: 100 });
    const inactive = makeMedicine({ id: "2", isActive: false, price: 100 });
    const medicinesById = new Map([
      [active.id, active],
      [inactive.id, inactive],
    ]);
    const activeFilters: ActiveFilter[] = [
      { key: "status", condition: "is-not", value: "active", displayValue: "無効" },
    ];

    const result = groupFilteredMedicines({
      orderedMedicines: [active, inactive],
      activeFilters,
      searchTerm: "",
      medicinesById,
    });

    expect(result.ungroupedMedicines).toEqual([inactive]);
  });

  it("filters by search term using kana-normalized matching", () => {
    const target = makeMedicine({ id: "1", name: "アモキシシリン", price: 100 });
    const other = makeMedicine({ id: "2", name: "パルボワクチン", price: 100 });
    const medicinesById = new Map([
      [target.id, target],
      [other.id, other],
    ]);

    const result = groupFilteredMedicines({
      orderedMedicines: [target, other],
      activeFilters: [],
      searchTerm: "あもきし",
      medicinesById,
    });

    expect(result.ungroupedMedicines).toEqual([target]);
    expect(result.totalCount).toBe(1);
  });

  it("returns empty groups for an empty medicines array", () => {
    const result = groupFilteredMedicines({
      orderedMedicines: [],
      activeFilters: [],
      searchTerm: "",
      medicinesById: new Map(),
    });

    expect(result.groupedMedicines.size).toBe(0);
    expect(result.ungroupedMedicines).toEqual([]);
    expect(result.totalCount).toBe(0);
  });
});

describe("resolveMedicineDrag", () => {
  it("returns none when the active item cannot be found", () => {
    const over = makeMedicine({ id: "2" });
    const map = new Map([[over.id, over]]);

    const result = resolveMedicineDrag({
      activeItemId: "missing",
      overItemId: "2",
      orderedMedicinesById: map,
    });

    expect(result).toEqual({ type: "none" });
  });

  it("returns none when the over item cannot be found", () => {
    const active = makeMedicine({ id: "1" });
    const map = new Map([[active.id, active]]);

    const result = resolveMedicineDrag({
      activeItemId: "1",
      overItemId: "missing",
      orderedMedicinesById: map,
    });

    expect(result).toEqual({ type: "none" });
  });

  it("returns same-category when both items share the same parent", () => {
    const active = makeMedicine({ id: "1", parentId: "10" });
    const over = makeMedicine({ id: "2", parentId: "10" });
    const map = new Map([
      [active.id, active],
      [over.id, over],
    ]);

    const result = resolveMedicineDrag({
      activeItemId: "1",
      overItemId: "2",
      orderedMedicinesById: map,
    });

    expect(result).toEqual({ type: "same-category" });
  });

  it("returns same-category when both items are top-level (no parent)", () => {
    const active = makeMedicine({ id: "1", parentId: undefined });
    const over = makeMedicine({ id: "2", parentId: undefined });
    const map = new Map([
      [active.id, active],
      [over.id, over],
    ]);

    const result = resolveMedicineDrag({
      activeItemId: "1",
      overItemId: "2",
      orderedMedicinesById: map,
    });

    expect(result).toEqual({ type: "same-category" });
  });

  it("builds a move-category resolution with parent_id when dropped onto a categorized item", () => {
    const active = makeMedicine({ id: "1", parentId: "10" });
    const over = makeMedicine({ id: "2", parentId: "20" });
    const map = new Map([
      [active.id, active],
      [over.id, over],
    ]);

    const result = resolveMedicineDrag({
      activeItemId: "1",
      overItemId: "2",
      orderedMedicinesById: map,
    });

    expect(result).toEqual({
      type: "move-category",
      activeItemId: "1",
      overParentId: "20",
      updateRequest: { parent_id: 20 },
    });
  });

  it("builds a move-category resolution with clear_parent_id when dropped onto an uncategorized item", () => {
    const active = makeMedicine({ id: "1", parentId: "10" });
    const over = makeMedicine({ id: "2", parentId: undefined });
    const map = new Map([
      [active.id, active],
      [over.id, over],
    ]);

    const result = resolveMedicineDrag({
      activeItemId: "1",
      overItemId: "2",
      orderedMedicinesById: map,
    });

    expect(result).toEqual({
      type: "move-category",
      activeItemId: "1",
      overParentId: null,
      updateRequest: { clear_parent_id: true },
    });
  });
});

describe("validateMedicineForm (BUG-006)", () => {
  it("rejects an uncategorized medicine with no price and no dosage form", () => {
    expect(
      validateMedicineForm(makeFormData({ parentId: "", price: 0, dosageForm: "" }), false),
    ).toBe("薬剤として登録する場合は親カテゴリ、単価、または剤形のいずれかを入力してください");
  });

  it("accepts an uncategorized medicine when dosage form is set", () => {
    expect(
      validateMedicineForm(makeFormData({ parentId: "", price: 0, dosageForm: "tablet" }), false),
    ).toBeNull();
  });
});

describe("buildMedicineCreateRequest", () => {
  it("forces price to 0 for a category row regardless of form input", () => {
    const data = makeFormData({ price: 9999 });

    const request = buildMedicineCreateRequest(data, true);

    expect(request.price).toBe(0);
  });

  it("uses the form price for a non-category medicine", () => {
    const data = makeFormData({ price: 1500 });

    const request = buildMedicineCreateRequest(data, false);

    expect(request.price).toBe(1500);
  });

  it("omits parent_id when parentId is empty", () => {
    const data = makeFormData({ parentId: "" });

    const request = buildMedicineCreateRequest(data, false);

    expect(request).not.toHaveProperty("parent_id");
  });

  it("includes numeric parent_id when parentId is set", () => {
    const data = makeFormData({ parentId: "42" });

    const request = buildMedicineCreateRequest(data, false);

    expect(request.parent_id).toBe(42);
  });

  it("maps all remaining form fields to the request", () => {
    const data = makeFormData({
      name: "新薬",
      dosageForm: "injection",
      medicineUnit: "per_dose",
      description: "説明",
      isActive: false,
      taxType: "included",
      taxRate: 0.08,
      isNonInsurance: true,
    });

    const request = buildMedicineCreateRequest(data, false);

    expect(request).toMatchObject({
      name: "新薬",
      dosage_form: "injection",
      medicine_unit: "per_dose",
      description: "説明",
      is_active: false,
      tax_type: "included",
      tax_rate: "0.08",
      is_non_insurance: true,
    });
  });
});

describe("buildMedicineUpdateRequest", () => {
  it("forces price to 0 for a category row regardless of form input", () => {
    const data = makeFormData({ price: 9999 });

    const request = buildMedicineUpdateRequest({ data, isCategory: true, selectedMedicine: null });

    expect(request.price).toBe(0);
  });

  it("uses the form price for a non-category medicine", () => {
    const data = makeFormData({ price: 2000 });

    const request = buildMedicineUpdateRequest({ data, isCategory: false, selectedMedicine: null });

    expect(request.price).toBe(2000);
  });

  it("sets parent_id when the form specifies a parentId", () => {
    const data = makeFormData({ parentId: "42" });
    const selectedMedicine = makeMedicine({ id: "1", parentId: "10" });

    const request = buildMedicineUpdateRequest({ data, isCategory: false, selectedMedicine });

    expect(request.parent_id).toBe(42);
    expect(request).not.toHaveProperty("clear_parent_id");
  });

  it("sets clear_parent_id when parentId is cleared and the medicine previously had a parent", () => {
    const data = makeFormData({ parentId: "" });
    const selectedMedicine = makeMedicine({ id: "1", parentId: "10" });

    const request = buildMedicineUpdateRequest({ data, isCategory: false, selectedMedicine });

    expect(request.clear_parent_id).toBe(true);
    expect(request).not.toHaveProperty("parent_id");
  });

  it("does not set clear_parent_id when the medicine never had a parent", () => {
    const data = makeFormData({ parentId: "" });
    const selectedMedicine = makeMedicine({ id: "1", parentId: undefined });

    const request = buildMedicineUpdateRequest({ data, isCategory: false, selectedMedicine });

    expect(request).not.toHaveProperty("clear_parent_id");
    expect(request).not.toHaveProperty("parent_id");
  });

  it("does not set clear_parent_id when selectedMedicine is null", () => {
    const data = makeFormData({ parentId: "" });

    const request = buildMedicineUpdateRequest({ data, isCategory: false, selectedMedicine: null });

    expect(request).not.toHaveProperty("clear_parent_id");
    expect(request).not.toHaveProperty("parent_id");
  });
});

describe("calculation fields (#201)", () => {
  it("omits strength/frequency_per_day/default_duration_days when calculationType is none", () => {
    const data = makeFormData({
      calculationType: "none",
      strength: "50",
      frequencyPerDay: "2",
      defaultDurationDays: "7",
    });

    const request = buildMedicineCreateRequest(data, false);

    expect(request.calculation_type).toBe("none");
    expect(request).not.toHaveProperty("strength");
    expect(request).not.toHaveProperty("frequency_per_day");
    expect(request).not.toHaveProperty("default_duration_days");
  });

  it("includes parsed numeric fields when calculationType is per_weight", () => {
    const data = makeFormData({
      calculationType: "per_weight",
      strength: "50",
      frequencyPerDay: "2",
      defaultDurationDays: "7",
    });

    const request = buildMedicineCreateRequest(data, false);

    expect(request).toMatchObject({
      calculation_type: "per_weight",
      strength: "50",
      frequency_per_day: 2,
      default_duration_days: 7,
    });
  });

  it("omits blank optional numeric fields even when calculationType is per_weight", () => {
    const data = makeFormData({
      calculationType: "per_weight",
      strength: "",
      frequencyPerDay: "",
      defaultDurationDays: "",
    });

    const request = buildMedicineCreateRequest(data, false);

    expect(request).not.toHaveProperty("strength");
    expect(request).not.toHaveProperty("frequency_per_day");
    expect(request).not.toHaveProperty("default_duration_days");
  });

  it("applies the same calculation-field rules to the update request", () => {
    const data = makeFormData({ calculationType: "per_weight", strength: "12.5" });

    const request = buildMedicineUpdateRequest({ data, isCategory: false, selectedMedicine: null });

    expect(request).toMatchObject({ calculation_type: "per_weight", strength: "12.5" });
    expect(request).not.toHaveProperty("frequency_per_day");
  });
});

describe("EMR-447 medicine exact decimal wire", () => {
  it("sends tax_rate as a decimal string and keeps explicit 0", () => {
    expect(buildMedicineCreateRequest(makeFormData({ taxRate: 0.1 }), false).tax_rate).toBe("0.1");
    const update = buildMedicineUpdateRequest({
      data: makeFormData({ taxRate: 0 }),
      isCategory: false,
      selectedMedicine: null,
    });
    expect(update.tax_rate).toBe("0");
  });

  it("sends the strength input as typed without float conversion", () => {
    const request = buildMedicineCreateRequest(
      makeFormData({ calculationType: "per_weight", strength: " 0.0005 " }),
      false,
    );
    expect(request.strength).toBe("0.0005");
  });
});

// SPEC-QA-20261022-447-STRENGTH-CLEAR（EMR-447 Q4）の受入 fixture。②③⑤ と ①④ の判定を model で固定する。
describe("EMR-447 Q4 strength の必須・明示クリア", () => {
  it("①④ per_weight で含量が空（空白のみを含む）なら保存前エラーにする", () => {
    for (const strength of ["", "   "]) {
      const data = makeFormData({ calculationType: "per_weight", strength });
      expect(validateMedicineForm(data, false)).toBe(MEDICINE_STRENGTH_REQUIRED_MESSAGE);
    }
  });

  it("Codex P1: per_weight の含量は BE と同じ字句・桁・正値で保存前に拒否する（補正しない）", () => {
    const cases: [string, RegExp][] = [
      [".5", /符号・指数のない数値/],
      ["05", /符号・指数のない数値/],
      ["1e2", /符号・指数のない数値/],
      ["-1", /符号・指数のない数値/],
      ["1234567", /整数部は6桁まで/],
      ["12.34567", /小数第4位まで/],
      ["0", /0より大きい値/],
      ["0.0000", /0より大きい値/],
    ];
    for (const [strength, message] of cases) {
      const data = makeFormData({ calculationType: "per_weight", strength });
      expect(validateMedicineForm(data, false)).toMatch(message);
    }
  });

  it("per_weight で保存尺度内の含量（末尾0違いを含む）は受理し、none は検査しない", () => {
    for (const strength of ["12.5", "1.23000", "999999.9999", "0.0001"]) {
      const data = makeFormData({ calculationType: "per_weight", strength });
      expect(validateMedicineForm(data, false)).toBeNull();
    }
    for (const strength of ["", ".5"]) {
      const data = makeFormData({ calculationType: "none", strength });
      expect(validateMedicineForm(data, false)).toBeNull();
    }
  });

  it("② none への切替だけでは既存含量を消さず strength・clear_strength を送らない", () => {
    const request = buildMedicineUpdateRequest({
      data: makeFormData({ calculationType: "none", strength: "50" }),
      isCategory: false,
      selectedMedicine: makeMedicine({ strength: 50 }),
    });

    expect(request).not.toHaveProperty("strength");
    expect(request).not.toHaveProperty("clear_strength");
    expect(request.calculation_type).toBe("none");
  });

  it("③ 既存含量を空へ編集し none で保存すると clear_strength=true だけを送る", () => {
    const request = buildMedicineUpdateRequest({
      data: makeFormData({ calculationType: "none", strength: "" }),
      isCategory: false,
      selectedMedicine: makeMedicine({ strength: 50 }),
    });

    expect(request.clear_strength).toBe(true);
    expect(request).not.toHaveProperty("strength");
  });

  it("③ 再GET で strength が無い（NULL）薬剤はフォームで未設定（空欄）として表示する", () => {
    expect(medicineToFormData(makeMedicine({ strength: undefined })).strength).toBe("");
  });

  it("含量を持たない薬剤の none 保存では clear_strength を送らない", () => {
    const request = buildMedicineUpdateRequest({
      data: makeFormData({ calculationType: "none", strength: "" }),
      isCategory: false,
      selectedMedicine: makeMedicine({ strength: undefined }),
    });

    expect(request).not.toHaveProperty("clear_strength");
  });

  it("per_weight の有効な含量は strength として送り clear_strength を付けない", () => {
    const request = buildMedicineUpdateRequest({
      data: makeFormData({ calculationType: "per_weight", strength: "60" }),
      isCategory: false,
      selectedMedicine: makeMedicine({ strength: 50 }),
    });

    expect(request).toMatchObject({ calculation_type: "per_weight", strength: "60" });
    expect(request).not.toHaveProperty("clear_strength");
  });

  it("⑤ 新規作成の none・空欄は未指定のまま送り clear_strength を付けない", () => {
    const data = makeFormData({ calculationType: "none", strength: "" });
    const request = buildMedicineCreateRequest(data, false);

    expect(request).not.toHaveProperty("strength");
    expect(request).not.toHaveProperty("clear_strength");
  });
});
