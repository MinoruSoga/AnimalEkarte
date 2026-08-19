import { describe, expect, it } from "vitest";

import type { LabDeviceItemMaster } from "../api/lab-device-item-masters";
import type { ExaminationTypeMaster } from "../api/exam-types-master";
import {
  LAB_DEVICE_UNMAPPED_FIELD,
  buildExamFieldOptions,
  buildLabDeviceItemMasterUpdateRequest,
  examFieldOptionsForItem,
  examFieldSelectValue,
  groupLabDeviceItemMasters,
  labDeviceSourceLabel,
  labDeviceValueShapeLabel,
  parseExamFieldSelectValue,
  validateLabDeviceItemMasterDraft,
} from "./lab-device-item-master-settings-model";

function item(overrides: Partial<LabDeviceItemMaster> = {}): LabDeviceItemMaster {
  return {
    id: "1",
    sourceType: "fuji_nx600",
    deviceItemCode: "Na-P",
    displayName: "Na",
    unit: "mEq/l",
    valueShape: "numeric",
    examTypeFieldId: null,
    sortOrder: 10,
    isActive: true,
    ...overrides,
  };
}

describe("lab-device-item-master-settings-model", () => {
  it("機器ラベルと値の形を表示用に変換する", () => {
    expect(labDeviceSourceLabel("fuji_nx600")).toBe("NX600");
    expect(labDeviceSourceLabel("unknown")).toBe("unknown");
    expect(labDeviceValueShapeLabel("qual_and_num")).toBe("定性+数値");
  });

  it("検査種別フィールドを「種別 / 項目」の選択肢にする", () => {
    const examTypes: ExaminationTypeMaster[] = [
      {
        id: "10",
        name: "血液化学",
        price: 0,
        isActive: true,
        description: "",
        sortOrder: 1,
        isNonInsurance: false,
        createdAt: "",
        updatedAt: "",
        items: [
          {
            id: "21",
            examTypeId: "10",
            name: "Na",
            inspectionValue: "",
            normalValue: "",
            unit: "mEq/l",
            sortOrder: 1,
            createdAt: "",
            updatedAt: "",
            referenceRanges: [],
          },
        ],
      },
    ];
    expect(buildExamFieldOptions(examTypes)).toEqual([
      { id: "21", label: "血液化学 / Na" },
    ]);
  });

  it("欠落している載せる先を選択肢へ残す", () => {
    expect(examFieldOptionsForItem([], "99")).toEqual([
      { id: "99", label: "欠落フィールド (99)" },
    ]);
  });

  it("未設定の載せる先は sentinel と null を往復する", () => {
    expect(examFieldSelectValue(null)).toBe(LAB_DEVICE_UNMAPPED_FIELD);
    expect(parseExamFieldSelectValue(LAB_DEVICE_UNMAPPED_FIELD)).toBeNull();
    expect(parseExamFieldSelectValue("21")).toBe("21");
  });

  it("表示名空と不正な載せる先を拒否する", () => {
    expect(validateLabDeviceItemMasterDraft({ displayName: "  ", examTypeFieldId: null })).toBe(
      "表示名は必須です",
    );
    expect(validateLabDeviceItemMasterDraft({ displayName: "Na", examTypeFieldId: "abc" })).toBe(
      "載せる先が不正です",
    );
    expect(validateLabDeviceItemMasterDraft({ displayName: "Na", examTypeFieldId: "21" })).toBeNull();
  });

  it("更新リクエストはコードを送らず display_name / field / is_active だけを送る", () => {
    expect(
      buildLabDeviceItemMasterUpdateRequest({
        displayName: " 尿糖 ",
        unit: "mg/dL",
        examTypeFieldId: "21",
        isActive: false,
      }),
    ).toEqual({
      display_name: "尿糖",
      unit: "mg/dL",
      exam_type_field_id: 21,
      is_active: false,
    });
    expect(
      buildLabDeviceItemMasterUpdateRequest({
        displayName: "Na",
        unit: "mEq/l",
        examTypeFieldId: null,
        isActive: true,
      }).exam_type_field_id,
    ).toBeNull();
  });

  it("NX600 → AU10V → 尿の順でグループ化する", () => {
    const groups = groupLabDeviceItemMasters([
      item({ id: "3", sourceType: "arkray_pu4010", deviceItemCode: "GLU", displayName: "尿糖" }),
      item({ id: "2", sourceType: "fuji_au10v", deviceItemCode: "vf-SAA", displayName: "vf-SAA" }),
      item({ id: "1", sourceType: "fuji_nx600", deviceItemCode: "Na-P", displayName: "Na" }),
    ]);
    expect(groups.map((group) => group.sourceType)).toEqual([
      "fuji_nx600",
      "fuji_au10v",
      "arkray_pu4010",
    ]);
    expect(groups[0]?.label).toBe("NX600");
  });
});
