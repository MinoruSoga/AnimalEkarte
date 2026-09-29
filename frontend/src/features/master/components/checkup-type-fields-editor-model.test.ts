import { describe, expect, it } from "vitest";

import type { CheckupTypeFieldRow } from "@/hooks/use-checkup-fields";
import {
  buildCheckupFieldCreateRequest,
  buildCheckupFieldUpdateRequest,
  checkupFieldToDraft,
  emptyCheckupFieldDraft,
  isCheckupSelectFieldType,
  validateCheckupFieldDraft,
  type CheckupFieldDraft,
} from "./checkup-type-fields-editor-model";

// EMR-225: 健診フィールド定義エディタの純粋ロジック固定（BE 検証と対）。

const baseDraft = (overrides: Partial<CheckupFieldDraft> = {}): CheckupFieldDraft => ({
  name: "体重",
  fieldType: "number",
  unit: "kg",
  minValue: "",
  maxValue: "",
  options: [],
  ...overrides,
});

const numberRow: CheckupTypeFieldRow = {
  id: 31,
  checkupTypeId: 7,
  name: "体重",
  fieldType: "number",
  unit: "kg",
  minValue: 1,
  maxValue: 80,
  options: [],
  isProvisional: false,
  sortOrder: 1,
};

const selectRow: CheckupTypeFieldRow = {
  id: 32,
  checkupTypeId: 7,
  name: "総合評価",
  fieldType: "single_select",
  unit: "",
  options: [
    { value: "a", label: "良好" },
    { value: "b", label: "要注意" },
  ],
  isProvisional: false,
  sortOrder: 2,
};

describe("isCheckupSelectFieldType", () => {
  it("matches the backend select-type set", () => {
    expect(isCheckupSelectFieldType("single_select")).toBe(true);
    expect(isCheckupSelectFieldType("multi_select")).toBe(true);
    expect(isCheckupSelectFieldType("checklist")).toBe(true);
    expect(isCheckupSelectFieldType("number")).toBe(false);
    expect(isCheckupSelectFieldType("boolean")).toBe(false);
    expect(isCheckupSelectFieldType("text")).toBe(false);
  });
});

describe("validateCheckupFieldDraft", () => {
  it("rejects a blank name", () => {
    expect(validateCheckupFieldDraft(baseDraft({ name: "  " }))).toBe("項目名を入力してください");
  });

  it("rejects non-finite and reversed numeric bounds", () => {
    expect(validateCheckupFieldDraft(baseDraft({ minValue: "abc" }))).toBe(
      "基準値には有限の数値を入力してください",
    );
    expect(validateCheckupFieldDraft(baseDraft({ minValue: "Infinity" }))).toBe(
      "基準値には有限の数値を入力してください",
    );
    expect(validateCheckupFieldDraft(baseDraft({ minValue: "10", maxValue: "5" }))).toBe(
      "基準値の下限は上限以下にしてください",
    );
  });

  it("skips bound validation for non-number types", () => {
    expect(validateCheckupFieldDraft(baseDraft({ fieldType: "text", minValue: "zzz" }))).toBeNull();
  });

  it("requires non-empty options for select types", () => {
    for (const fieldType of ["single_select", "multi_select", "checklist"] as const) {
      expect(validateCheckupFieldDraft(baseDraft({ fieldType, options: [] }))).toBe(
        "選択式の項目には選択肢を1件以上登録してください",
      );
    }
  });

  it("rejects empty option fields and duplicate values", () => {
    expect(
      validateCheckupFieldDraft(
        baseDraft({ fieldType: "single_select", options: [{ value: " ", label: "A" }] }),
      ),
    ).toBe("選択肢の値と表示名を入力してください");
    expect(
      validateCheckupFieldDraft(
        baseDraft({ fieldType: "single_select", options: [{ value: "a", label: "  " }] }),
      ),
    ).toBe("選択肢の値と表示名を入力してください");
    expect(
      validateCheckupFieldDraft(
        baseDraft({
          fieldType: "multi_select",
          options: [
            { value: "a", label: "A" },
            { value: "a", label: "B" },
          ],
        }),
      ),
    ).toBe("選択肢の値は一意にしてください");
  });

  it("accepts a valid draft of every field type", () => {
    expect(validateCheckupFieldDraft(emptyCheckupFieldDraft())).not.toBeNull();
    expect(validateCheckupFieldDraft(baseDraft({ fieldType: "boolean", name: "無" }))).toBeNull();
    expect(
      validateCheckupFieldDraft(
        baseDraft({
          fieldType: "checklist",
          options: [{ value: "scale", label: "スケーリング" }],
        }),
      ),
    ).toBeNull();
  });
});

describe("checkupFieldToDraft", () => {
  it("maps stored rows into editable drafts", () => {
    expect(checkupFieldToDraft(numberRow)).toEqual({
      name: "体重",
      fieldType: "number",
      unit: "kg",
      minValue: "1",
      maxValue: "80",
      options: [],
    });
    expect(checkupFieldToDraft(selectRow).options).toEqual([
      { value: "a", label: "良好" },
      { value: "b", label: "要注意" },
    ]);
  });
});

describe("buildCheckupFieldCreateRequest", () => {
  it("sends unit/bounds only for number and options only for select types", () => {
    expect(buildCheckupFieldCreateRequest(baseDraft({ minValue: "1", maxValue: "80" }), 4)).toEqual(
      {
        name: "体重",
        field_type: "number",
        unit: "kg",
        min_value: 1,
        max_value: 80,
        sort_order: 4,
      },
    );

    expect(
      buildCheckupFieldCreateRequest(
        baseDraft({
          fieldType: "single_select",
          unit: "残置しない",
          minValue: "9",
          options: [{ value: " a ", label: " 良好 " }],
        }),
      ),
    ).toEqual({
      name: "体重",
      field_type: "single_select",
      unit: "残置しない",
      options: [{ value: "a", label: "良好" }],
    });
  });

  it("omits sort_order when not given", () => {
    expect(buildCheckupFieldCreateRequest(baseDraft())).not.toHaveProperty("sort_order");
  });
});

describe("buildCheckupFieldUpdateRequest", () => {
  it("always sends name/field_type/unit and uses clear flags for emptied bounds", () => {
    const req = buildCheckupFieldUpdateRequest(
      baseDraft({ name: "体重（朝）", minValue: "", maxValue: "90" }),
      numberRow,
    );
    expect(req).toEqual({
      name: "体重（朝）",
      field_type: "number",
      unit: "kg",
      clear_min_value: true,
      max_value: 90,
    });
  });

  it("sends explicit empty options when switching select → non-select", () => {
    const req = buildCheckupFieldUpdateRequest(
      baseDraft({ fieldType: "boolean", options: [] }),
      selectRow,
    );
    expect(req.field_type).toBe("boolean");
    expect(req.options).toEqual([]);
  });

  it("omits options entirely when a non-select field stays non-select", () => {
    const req = buildCheckupFieldUpdateRequest(baseDraft(), numberRow);
    expect(req).not.toHaveProperty("options");
  });

  it("sends normalized options for select types", () => {
    const req = buildCheckupFieldUpdateRequest(
      baseDraft({
        fieldType: "checklist",
        options: [{ value: " a ", label: " A " }],
      }),
      selectRow,
    );
    expect(req.options).toEqual([{ value: "a", label: "A" }]);
  });
});
