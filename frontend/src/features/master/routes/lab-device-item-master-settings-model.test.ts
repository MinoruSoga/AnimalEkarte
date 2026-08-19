import { describe, expect, it } from "vitest";

import type { LabDeviceItemMaster } from "../api/lab-device-item-masters";
import type { ExaminationTypeMaster } from "../api/exam-types-master";
import {
  LAB_DEVICE_UNMAPPED_FIELD,
  buildExamFieldOptions,
  buildLabDeviceItemMasterUpdateRequest,
  collectDirtyLabDeviceUpdates,
  examFieldOptionsForItem,
  examFieldSelectValue,
  groupLabDeviceItemMasters,
  itemToLabDeviceDraft,
  itemsForLabDevice,
  labDeviceSourceLabel,
  labDeviceValueShapeLabel,
  parseExamFieldSelectValue,
  parseLabDeviceSourceQuery,
  toLabDeviceRows,
  validateLabDeviceItemMasterDraft,
} from "./lab-device-item-master-settings-model";

function item(overrides: Partial<LabDeviceItemMaster> = {}): LabDeviceItemMaster {
  return {
    id: "1",
    sourceType: "fuji_nx600",
    deviceItemCode: "Na-P",
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

  it("受信画面の source クエリで該当機器を開く", () => {
    expect(parseLabDeviceSourceQuery("fuji_nx600")).toBe("fuji_nx600");
    expect(parseLabDeviceSourceQuery("  ")).toBeNull();
    expect(parseLabDeviceSourceQuery(null)).toBeNull();
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

  it("不正な載せる先を拒否する", () => {
    expect(validateLabDeviceItemMasterDraft({ examTypeFieldId: "abc" })).toBe(
      "載せる先が不正です",
    );
    expect(validateLabDeviceItemMasterDraft({ examTypeFieldId: null })).toBeNull();
    expect(validateLabDeviceItemMasterDraft({ examTypeFieldId: "21" })).toBeNull();
  });

  it("更新リクエストはコードを送らず field / is_active だけを送る", () => {
    expect(
      buildLabDeviceItemMasterUpdateRequest({
        unit: "mg/dL",
        examTypeFieldId: "21",
        isActive: false,
      }),
    ).toEqual({
      unit: "mg/dL",
      exam_type_field_id: 21,
      is_active: false,
    });
    expect(
      buildLabDeviceItemMasterUpdateRequest({
        unit: "mEq/l",
        examTypeFieldId: null,
        isActive: true,
      }).exam_type_field_id,
    ).toBeNull();
  });

  it("NX600 → AU10V → 尿の順でグループ化する", () => {
    const groups = groupLabDeviceItemMasters([
      item({ id: "3", sourceType: "arkray_pu4010", deviceItemCode: "GLU" }),
      item({ id: "2", sourceType: "fuji_au10v", deviceItemCode: "vf-SAA" }),
      item({ id: "1", sourceType: "fuji_nx600", deviceItemCode: "Na-P" }),
    ]);
    expect(groups.map((group) => group.sourceType)).toEqual([
      "fuji_nx600",
      "fuji_au10v",
      "arkray_pu4010",
    ]);
    expect(groups[0]?.label).toBe("NX600");
  });

  it("一覧は項目が空でも3機器を出し、未設定数を集計する", () => {
    const rows = toLabDeviceRows([
      item({ id: "1", sourceType: "fuji_nx600", examTypeFieldId: null, isActive: true }),
      item({ id: "2", sourceType: "fuji_nx600", examTypeFieldId: "21", isActive: false }),
    ]);
    expect(rows.map((row) => row.sourceType)).toEqual([
      "fuji_nx600",
      "fuji_au10v",
      "arkray_pu4010",
    ]);
    expect(rows[0]).toMatchObject({
      name: "NX600",
      itemCount: 2,
      unmappedCount: 1,
    });
    expect(rows[1]).toMatchObject({ name: "AU10V", itemCount: 0, unmappedCount: 0 });
    expect(rows[2]).toMatchObject({ name: "尿（PU-4010）", itemCount: 0 });
  });

  it("未知の機器は末尾に出す", () => {
    const rows = toLabDeviceRows([
      item({ id: "9", sourceType: "other_device", isActive: false }),
    ]);
    expect(rows.map((row) => row.sourceType)).toEqual([
      "fuji_nx600",
      "fuji_au10v",
      "arkray_pu4010",
      "other_device",
    ]);
  });

  it("機器の項目は sort_order で並べ、変更分だけ PATCH する", () => {
    const items = [
      item({ id: "2", deviceItemCode: "K-P", sortOrder: 20 }),
      item({ id: "1", deviceItemCode: "Na-P", sortOrder: 10, examTypeFieldId: null }),
    ];
    expect(itemsForLabDevice(items, "fuji_nx600").map((row) => row.id)).toEqual(["1", "2"]);
    const drafts = items.map(itemToLabDeviceDraft);
    drafts[0] = { ...drafts[0]!, examTypeFieldId: "21" };
    expect(collectDirtyLabDeviceUpdates(items, drafts)).toEqual({
      error: null,
      updates: [
        {
          id: "2",
          req: {
            unit: "mEq/l",
            exam_type_field_id: 21,
            is_active: true,
          },
        },
      ],
    });
    expect(
      collectDirtyLabDeviceUpdates(items, [{ ...itemToLabDeviceDraft(items[0]!), examTypeFieldId: "x" }]).error,
    ).toBe("K-P: 載せる先が不正です");
  });
});
