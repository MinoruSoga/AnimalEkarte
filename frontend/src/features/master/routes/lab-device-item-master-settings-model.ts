import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import { normalizedIncludes } from "@/lib/normalize-kana";

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

export type LabDeviceRow = {
  id: string;
  sourceType: string;
  name: string;
  itemCount: number;
  unmappedCount: number;
  isActive: boolean;
};

export type LabDeviceItemDraft = {
  id: string;
  displayName: string;
  examTypeFieldId: string | null;
  isActive: boolean;
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

function toLabDeviceRow(sourceType: string, list: LabDeviceItemMaster[]): LabDeviceRow {
  return {
    id: sourceType,
    sourceType,
    name: labDeviceSourceLabel(sourceType),
    itemCount: list.length,
    unmappedCount: list.filter((item) => item.examTypeFieldId === null).length,
    isActive: list.length === 0 || list.some((item) => item.isActive),
  };
}

export function toLabDeviceRows(items: LabDeviceItemMaster[] | undefined): LabDeviceRow[] {
  const grouped = new Map(groupLabDeviceItemMasters(items).map((group) => [group.sourceType, group.items]));
  const known = LAB_DEVICE_SOURCE_ORDER.map((sourceType) => toLabDeviceRow(sourceType, grouped.get(sourceType) ?? []));
  const extras = [...grouped.keys()]
    .filter((sourceType) => !LAB_DEVICE_SOURCE_ORDER.includes(sourceType as (typeof LAB_DEVICE_SOURCE_ORDER)[number]))
    .sort()
    .map((sourceType) => toLabDeviceRow(sourceType, grouped.get(sourceType) ?? []));
  return [...known, ...extras];
}

export function itemsForLabDevice(
  items: LabDeviceItemMaster[] | undefined,
  sourceType: string,
): LabDeviceItemMaster[] {
  return (items ?? [])
    .filter((item) => item.sourceType === sourceType)
    .slice()
    .sort((left, right) => left.sortOrder - right.sortOrder || left.deviceItemCode.localeCompare(right.deviceItemCode));
}

export function filterLabDeviceRows(
  rows: LabDeviceRow[],
  searchTerm: string,
  filters: ActiveFilter[],
): LabDeviceRow[] {
  const term = searchTerm.trim();
  return rows.filter((row) => {
    if (term && !normalizedIncludes(row.name, term) && !normalizedIncludes(row.sourceType, term)) {
      return false;
    }
    for (const filter of filters) {
      if (filter.key !== "status" || typeof filter.value !== "string") {
        continue;
      }
      const wantActive = filter.value === "active";
      if (filter.condition === "is" && row.isActive !== wantActive) {
        return false;
      }
      if (filter.condition === "is_not" && row.isActive === wantActive) {
        return false;
      }
    }
    return true;
  });
}

export function itemToLabDeviceDraft(item: LabDeviceItemMaster): LabDeviceItemDraft {
  return {
    id: item.id,
    displayName: item.displayName,
    examTypeFieldId: item.examTypeFieldId,
    isActive: item.isActive,
  };
}

export function isLabDeviceItemDraftDirty(item: LabDeviceItemMaster, draft: LabDeviceItemDraft): boolean {
  return item.displayName !== draft.displayName.trim()
    || item.examTypeFieldId !== draft.examTypeFieldId
    || item.isActive !== draft.isActive;
}

export function collectDirtyLabDeviceUpdates(
  items: LabDeviceItemMaster[],
  drafts: LabDeviceItemDraft[],
): { error: string | null; updates: { id: string; req: UpdateLabDeviceItemMasterRequest }[] } {
  const byId = new Map(items.map((item) => [item.id, item]));
  const updates: { id: string; req: UpdateLabDeviceItemMasterRequest }[] = [];
  for (const draft of drafts) {
    const item = byId.get(draft.id);
    if (item === undefined) {
      continue;
    }
    const error = validateLabDeviceItemMasterDraft(draft);
    if (error !== null) {
      return { error: `${item.deviceItemCode}: ${error}`, updates: [] };
    }
    if (!isLabDeviceItemDraftDirty(item, draft)) {
      continue;
    }
    updates.push({
      id: draft.id,
      req: buildLabDeviceItemMasterUpdateRequest({
        displayName: draft.displayName,
        unit: item.unit,
        examTypeFieldId: draft.examTypeFieldId,
        isActive: draft.isActive,
      }),
    });
  }
  return { error: null, updates };
}
