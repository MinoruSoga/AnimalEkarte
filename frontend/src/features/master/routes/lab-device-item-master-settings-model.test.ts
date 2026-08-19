import { describe, expect, it } from "vitest";

import type { LabDevice } from "../api/lab-devices";
import type { LabDeviceItemMaster } from "../api/lab-device-item-masters";
import type { ExaminationTypeMaster } from "../api/exam-types-master";
import {
  LAB_DEVICE_EXAM_MIXED,
  LAB_DEVICE_EXAM_SELECT_UNSET,
  LAB_DEVICE_EXAM_UNSET,
  LAB_DEVICE_UNMAPPED_FIELD,
  availableLabDeviceSourceTypes,
  buildExamFieldOptions,
  buildLabDeviceCreateRequest,
  buildLabDeviceItemMasterUpdateRequest,
  buildLabDeviceUpdateRequest,
  collectDirtyLabDeviceUpdates,
  countDraftsUnmappedByExamChange,
  examFieldOptionsForExamType,
  examFieldOptionsForItem,
  examFieldSelectValue,
  examTypeSelectOptions,
  examTypeSelectValue,
  groupLabDeviceItemMasters,
  itemToLabDeviceDraft,
  itemsForLabDevice,
  labDeviceExamLabel,
  labDeviceExamTypeId,
  labDeviceFieldName,
  labDeviceFieldUnit,
  labDeviceItemSelectOptions,
  labDeviceSourceLabel,
  labDeviceUnitMismatch,
  labDeviceToFormData,
  labDeviceValueShapeLabel,
  parseExamFieldSelectValue,
  parseExamTypeSelectValue,
  parseLabDeviceSourceQuery,
  restrictDraftsToExamType,
  toLabDeviceRows,
  validateLabDeviceDraft,
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

function device(overrides: Partial<LabDevice> = {}): LabDevice {
  return {
    id: "1",
    sourceType: "fuji_nx600",
    name: "院内NX",
    examTypeId: null,
    isActive: true,
    sortOrder: 10,
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

  it("対応済み項目から検査名を出し、未設定と複数を区別する", () => {
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
      {
        id: "11",
        name: "尿検査",
        price: 0,
        isActive: true,
        description: "",
        sortOrder: 2,
        isNonInsurance: false,
        createdAt: "",
        updatedAt: "",
        items: [
          {
            id: "31",
            examTypeId: "11",
            name: "PRO",
            inspectionValue: "",
            normalValue: "",
            unit: "mg/dL",
            sortOrder: 1,
            createdAt: "",
            updatedAt: "",
            referenceRanges: [],
          },
        ],
      },
    ];
    expect(labDeviceExamLabel([item()], examTypes)).toBe(LAB_DEVICE_EXAM_UNSET);
    expect(labDeviceExamLabel([item({ examTypeFieldId: "21" })], examTypes)).toBe("血液化学");
    expect(
      labDeviceExamLabel(
        [item({ examTypeFieldId: "21" }), item({ id: "2", examTypeFieldId: "31" })],
        examTypes,
      ),
    ).toBe(LAB_DEVICE_EXAM_MIXED);
    expect(labDeviceExamTypeId([item({ examTypeFieldId: "21" })], examTypes)).toBe("10");
    expect(labDeviceExamTypeId([item()], examTypes)).toBeNull();
    expect(labDeviceFieldName("21", examTypes)).toBe("Na");
    expect(labDeviceFieldName(null, examTypes)).toBe("");
    expect(
      examFieldOptionsForExamType(examTypes, "10"),
    ).toEqual([{ id: "21", label: "Na" }]);
    expect(examFieldOptionsForExamType(examTypes, null)).toEqual([]);
    expect(
      restrictDraftsToExamType(
        [
          { id: "1", examTypeFieldId: "21", isActive: true },
          { id: "2", examTypeFieldId: "31", isActive: true },
        ],
        "10",
        examTypes,
      ),
    ).toEqual([
      { id: "1", examTypeFieldId: "21", isActive: true },
      { id: "2", examTypeFieldId: null, isActive: true },
    ]);
  });

  it("欠落している検査項目を選択肢へ残す", () => {
    expect(examFieldOptionsForItem([], "99")).toEqual([
      { id: "99", label: "欠落フィールド (99)" },
    ]);
  });

  it("未選択の検査項目は空文字と null を往復する", () => {
    expect(examFieldSelectValue(null)).toBe("");
    expect(parseExamFieldSelectValue("")).toBeNull();
    expect(parseExamFieldSelectValue(LAB_DEVICE_UNMAPPED_FIELD)).toBeNull();
    expect(parseExamFieldSelectValue("21")).toBe("21");
  });

  it("不正な検査項目を拒否する", () => {
    expect(validateLabDeviceItemMasterDraft({ examTypeFieldId: "abc" })).toBe(
      "検査項目が不正です",
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

  it("一覧はDBの機器だけを出し、検査名は機器の exam_type_id を正とする", () => {
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
    const rows = toLabDeviceRows(
      [
        device({
          id: "3",
          sourceType: "fuji_au10v",
          name: "院内AU",
          examTypeId: "10",
          sortOrder: 20,
        }),
        device({ id: "1", name: "院内NX", examTypeId: null, sortOrder: 10 }),
      ],
      [
        item({ id: "1", sourceType: "fuji_nx600", examTypeFieldId: null }),
        item({ id: "2", sourceType: "fuji_nx600", examTypeFieldId: "21" }),
      ],
      examTypes,
    );
    expect(rows.map((row) => row.name)).toEqual(["院内NX", "院内AU"]);
    expect(rows[0]).toMatchObject({
      examLabel: "血液化学",
      itemCount: 2,
      unmappedCount: 1,
    });
    expect(rows[1]).toMatchObject({ examLabel: "血液化学", itemCount: 0 });
    expect(toLabDeviceRows([], [item()])).toEqual([]);
  });

  it("未使用プロトコルと機器フォームの保存リクエストを組み立てる", () => {
    expect(availableLabDeviceSourceTypes([device()])).toEqual(["fuji_au10v", "arkray_pu4010"]);
    expect(availableLabDeviceSourceTypes([
      device(),
      device({ id: "2", sourceType: "fuji_au10v" }),
      device({ id: "3", sourceType: "arkray_pu4010" }),
    ])).toEqual([]);
    expect(labDeviceToFormData(null, ["fuji_au10v"])).toEqual({
      name: "AU10V",
      sourceType: "fuji_au10v",
      examTypeId: null,
      isActive: true,
      sortOrder: 20,
    });
    expect(examTypeSelectValue(null)).toBe(LAB_DEVICE_EXAM_SELECT_UNSET);
    expect(parseExamTypeSelectValue(LAB_DEVICE_EXAM_SELECT_UNSET)).toBeNull();
    expect(examTypeSelectOptions([{
      id: "10",
      name: "血液化学",
      price: 0,
      isActive: true,
      description: "",
      sortOrder: 1,
      isNonInsurance: false,
      createdAt: "",
      updatedAt: "",
      items: [],
    }])).toEqual([
      { value: LAB_DEVICE_EXAM_SELECT_UNSET, label: LAB_DEVICE_EXAM_UNSET },
      { value: "10", label: "血液化学" },
    ]);
    expect(validateLabDeviceDraft({
      name: "",
      sourceType: "fuji_nx600",
      examTypeId: null,
      requireSourceType: true,
    })).toBe("機器名は必須です");
    expect(validateLabDeviceDraft({
      name: "院内NX",
      sourceType: "",
      examTypeId: null,
      requireSourceType: true,
    })).toBe("プロトコルを選んでください");
    expect(validateLabDeviceDraft({
      name: "院内NX",
      sourceType: "fuji_nx600",
      examTypeId: "10",
      requireSourceType: true,
    })).toBeNull();
    expect(buildLabDeviceCreateRequest({
      name: " 院内NX ",
      sourceType: "fuji_nx600",
      examTypeId: "10",
      isActive: true,
      sortOrder: 10,
    })).toEqual({
      name: "院内NX",
      source_type: "fuji_nx600",
      exam_type_id: 10,
      is_active: true,
      sort_order: 10,
    });
    expect(buildLabDeviceUpdateRequest({
      name: "院内NX",
      sourceType: "fuji_nx600",
      examTypeId: null,
      isActive: false,
      sortOrder: 15,
    })).toEqual({
      name: "院内NX",
      exam_type_id: null,
      is_active: false,
      sort_order: 15,
    });
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
    ).toBe("K-P: 検査項目が不正です");
  });

  it("一覧の検査ラベルは項目対応（実挙動）を機器設定より優先する", () => {
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
          { id: "21", examTypeId: "10", name: "Na", inspectionValue: "", normalValue: "", unit: "mEq/l", sortOrder: 1, createdAt: "", updatedAt: "", referenceRanges: [] },
        ],
      },
      {
        id: "11",
        name: "尿検査",
        price: 0,
        isActive: true,
        description: "",
        sortOrder: 2,
        isNonInsurance: false,
        createdAt: "",
        updatedAt: "",
        items: [
          { id: "31", examTypeId: "11", name: "尿糖", inspectionValue: "", normalValue: "", unit: "mg/dL", sortOrder: 1, createdAt: "", updatedAt: "", referenceRanges: [] },
        ],
      },
    ];
    // 機器には尿検査(11)が設定されているが、項目は血液化学(10)の field に張られている
    const rows = toLabDeviceRows(
      [device({ id: "1", examTypeId: "11" })],
      [item({ id: "1", examTypeFieldId: "21" })],
      examTypes,
    );
    expect(rows[0]!.examLabel).toBe("血液化学");
  });

  it("項目 select は機器の検査で絞り、欠落フィールドと未設定を落とさない", () => {
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
          { id: "21", examTypeId: "10", name: "Na", inspectionValue: "", normalValue: "", unit: "mEq/l", sortOrder: 1, createdAt: "", updatedAt: "", referenceRanges: [] },
          { id: "22", examTypeId: "10", name: "K", inspectionValue: "", normalValue: "", unit: "mEq/l", sortOrder: 2, createdAt: "", updatedAt: "", referenceRanges: [] },
        ],
      },
    ];
    expect(labDeviceItemSelectOptions(examTypes, "10", null)).toEqual([
      { value: LAB_DEVICE_UNMAPPED_FIELD, label: LAB_DEVICE_EXAM_UNSET },
      { value: "21", label: "Na" },
      { value: "22", label: "K" },
    ]);
    // 絞り込み外の現在値は欠落フィールドとして残す
    expect(labDeviceItemSelectOptions(examTypes, "10", "99").map((option) => option.value)).toEqual([
      LAB_DEVICE_UNMAPPED_FIELD,
      "21",
      "22",
      "99",
    ]);
    // 機器の検査が未設定なら全検査の「種別 / 項目」から選べる
    expect(labDeviceItemSelectOptions(examTypes, null, null)[1]).toEqual({
      value: "21",
      label: "血液化学 / Na",
    });
  });

  it("検査変更で外れる項目対応の件数を数える", () => {
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
          { id: "21", examTypeId: "10", name: "Na", inspectionValue: "", normalValue: "", unit: "mEq/l", sortOrder: 1, createdAt: "", updatedAt: "", referenceRanges: [] },
        ],
      },
      {
        id: "11",
        name: "尿検査",
        price: 0,
        isActive: true,
        description: "",
        sortOrder: 2,
        isNonInsurance: false,
        createdAt: "",
        updatedAt: "",
        items: [
          { id: "31", examTypeId: "11", name: "尿糖", inspectionValue: "", normalValue: "", unit: "mg/dL", sortOrder: 1, createdAt: "", updatedAt: "", referenceRanges: [] },
        ],
      },
    ];
    const drafts = [
      { id: "1", examTypeFieldId: "21", isActive: true },
      { id: "2", examTypeFieldId: null, isActive: true },
    ];
    expect(countDraftsUnmappedByExamChange(drafts, "11", examTypes)).toBe(1);
    expect(countDraftsUnmappedByExamChange(drafts, "10", examTypes)).toBe(0);
    expect(countDraftsUnmappedByExamChange(drafts, null, examTypes)).toBe(0);
  });

  it("単位不一致は両方に値があって食い違うときだけ警告する", () => {
    expect(labDeviceUnitMismatch("mg/dL", "mmol/L")).toBe(true);
    expect(labDeviceUnitMismatch("mg/dL", "mg/dl")).toBe(false);
    expect(labDeviceUnitMismatch("", "mmol/L")).toBe(false);
    expect(labDeviceUnitMismatch("mg/dL", "")).toBe(false);
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
          { id: "21", examTypeId: "10", name: "Na", inspectionValue: "", normalValue: "", unit: "mmol/L", sortOrder: 1, createdAt: "", updatedAt: "", referenceRanges: [] },
        ],
      },
    ];
    expect(labDeviceFieldUnit("21", examTypes)).toBe("mmol/L");
    expect(labDeviceFieldUnit(null, examTypes)).toBe("");
  });
});
