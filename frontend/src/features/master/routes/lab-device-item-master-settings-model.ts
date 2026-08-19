import type {
  LabDeviceItemMaster,
  UpdateLabDeviceItemMasterRequest,
} from "../api/lab-device-item-masters";
import type { ExaminationTypeMaster } from "../api/exam-types-master";

export const LAB_DEVICE_UNMAPPED_FIELD = "__unmapped__";

export const LAB_DEVICE_SOURCE_ORDER = [
  "fuji_nx600",
  "fuji_au10v",
  "arkray_pu4010",
] as const;

const LAB_DEVICE_SOURCE_LABELS: Record<string, string> = {
  fuji_nx600: "NX600",
  fuji_au10v: "AU10V",
  arkray_pu4010: "尿（PU-4010）",
};

const LAB_DEVICE_VALUE_SHAPE_LABELS: Record<string, string> = {
  numeric: "数値",
  inequality: "不等号",
  qual_and_num: "定性+数値",
  dash: "ダッシュ",
  text: "テキスト",
};

export type LabDeviceExamFieldOption = {
  id: string;
  label: string;
};

export type LabDeviceItemMasterGroup = {
  sourceType: string;
  label: string;
  items: LabDeviceItemMaster[];
};

export function labDeviceSourceLabel(sourceType: string): string {
  return LAB_DEVICE_SOURCE_LABELS[sourceType] ?? sourceType;
}

export function labDeviceValueShapeLabel(valueShape: string): string {
  return LAB_DEVICE_VALUE_SHAPE_LABELS[valueShape] ?? valueShape;
}

export function buildExamFieldOptions(
  examTypes: ExaminationTypeMaster[] | undefined,
): LabDeviceExamFieldOption[] {
  return (examTypes ?? []).flatMap((examType) =>
    (examType.items ?? []).map((field) => ({
      id: field.id,
      label: `${examType.name} / ${field.name}`,
    })),
  );
}

export function examFieldOptionsForItem(
  options: LabDeviceExamFieldOption[],
  examTypeFieldId: string | null,
): LabDeviceExamFieldOption[] {
  if (examTypeFieldId === null) {
    return options;
  }
  if (options.some((option) => option.id === examTypeFieldId)) {
    return options;
  }
  return [...options, { id: examTypeFieldId, label: `欠落フィールド (${examTypeFieldId})` }];
}

export function examFieldSelectValue(examTypeFieldId: string | null): string {
  return examTypeFieldId ?? LAB_DEVICE_UNMAPPED_FIELD;
}

export function parseExamFieldSelectValue(value: string): string | null {
  return value === LAB_DEVICE_UNMAPPED_FIELD ? null : value;
}

export function validateLabDeviceItemMasterDraft(input: {
  displayName: string;
  examTypeFieldId: string | null;
}): string | null {
  if (!input.displayName.trim()) {
    return "表示名は必須です";
  }
  if (input.examTypeFieldId !== null) {
    const id = Number(input.examTypeFieldId);
    if (!Number.isInteger(id) || id <= 0) {
      return "載せる先が不正です";
    }
  }
  return null;
}

export function buildLabDeviceItemMasterUpdateRequest(input: {
  displayName: string;
  unit: string;
  examTypeFieldId: string | null;
  isActive: boolean;
}): UpdateLabDeviceItemMasterRequest {
  return {
    display_name: input.displayName.trim(),
    unit: input.unit,
    exam_type_field_id: input.examTypeFieldId === null ? null : Number(input.examTypeFieldId),
    is_active: input.isActive,
  };
}

export function groupLabDeviceItemMasters(
  items: LabDeviceItemMaster[] | undefined,
): LabDeviceItemMasterGroup[] {
  const groups = new Map<string, LabDeviceItemMaster[]>();
  for (const item of items ?? []) {
    const existing = groups.get(item.sourceType) ?? [];
    groups.set(item.sourceType, [...existing, item]);
  }

  const known = LAB_DEVICE_SOURCE_ORDER.filter((sourceType) => groups.has(sourceType)).map(
    (sourceType) => ({
      sourceType,
      label: labDeviceSourceLabel(sourceType),
      items: groups.get(sourceType) ?? [],
    }),
  );
  const extras = [...groups.keys()]
    .filter((sourceType) => !LAB_DEVICE_SOURCE_ORDER.includes(sourceType as (typeof LAB_DEVICE_SOURCE_ORDER)[number]))
    .sort()
    .map((sourceType) => ({
      sourceType,
      label: labDeviceSourceLabel(sourceType),
      items: groups.get(sourceType) ?? [],
    }));
  return [...known, ...extras];
}
