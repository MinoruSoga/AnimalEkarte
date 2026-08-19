import type {
  CreateLabDeviceRequest,
  LabDevice,
  UpdateLabDeviceRequest,
} from "../api/lab-devices";
import type {
  LabDeviceItemMaster,
  UpdateLabDeviceItemMasterRequest,
} from "../api/lab-device-item-masters";
import type { ExaminationTypeMaster } from "../api/exam-types-master";

export const LAB_DEVICE_UNMAPPED_FIELD = "__unmapped__";
export const LAB_DEVICE_EXAM_SELECT_UNSET = "__unset__";
export const LAB_DEVICE_EXAM_UNSET = "未設定";
export const LAB_DEVICE_EXAM_MIXED = "複数の検査";

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
  examTypeId: string | null;
  examLabel: string;
  isActive: boolean;
  sortOrder: number;
  itemCount: number;
  unmappedCount: number;
};

export type LabDeviceFormData = {
  name: string;
  sourceType: string;
  examTypeId: string | null;
  isActive: boolean;
  sortOrder: number;
};

type ExamTypeByFieldId = Map<string, { id: string; name: string }>;

function examTypeByFieldId(
  examTypes: ExaminationTypeMaster[] | undefined,
): ExamTypeByFieldId {
  const byField = new Map<string, { id: string; name: string }>();
  for (const examType of examTypes ?? []) {
    for (const field of examType.items ?? []) {
      byField.set(field.id, { id: examType.id, name: examType.name });
    }
  }
  return byField;
}

function uniqueExamTypesForFields(
  fieldIds: Array<string | null | undefined>,
  examTypes: ExaminationTypeMaster[] | undefined,
): Array<{ id: string; name: string }> {
  const byField = examTypeByFieldId(examTypes);
  const seen = new Set<string>();
  const unique: Array<{ id: string; name: string }> = [];
  for (const fieldId of fieldIds) {
    if (fieldId == null) {
      continue;
    }
    const examType = byField.get(fieldId);
    if (examType === undefined || seen.has(examType.id)) {
      continue;
    }
    seen.add(examType.id);
    unique.push(examType);
  }
  return unique;
}

export function labDeviceExamLabel(
  items: Array<{ examTypeFieldId: string | null }>,
  examTypes: ExaminationTypeMaster[] | undefined,
): string {
  const unique = uniqueExamTypesForFields(
    items.map((item) => item.examTypeFieldId),
    examTypes,
  );
  if (unique.length === 0) {
    return LAB_DEVICE_EXAM_UNSET;
  }
  if (unique.length === 1) {
    return unique[0]!.name;
  }
  return LAB_DEVICE_EXAM_MIXED;
}

export function labDeviceExamTypeId(
  items: Array<{ examTypeFieldId: string | null }>,
  examTypes: ExaminationTypeMaster[] | undefined,
): string | null {
  const unique = uniqueExamTypesForFields(
    items.map((item) => item.examTypeFieldId),
    examTypes,
  );
  return unique.length === 1 ? unique[0]!.id : null;
}

export function labDeviceFieldName(
  examTypeFieldId: string | null,
  examTypes: ExaminationTypeMaster[] | undefined,
): string {
  if (examTypeFieldId === null) {
    return "";
  }
  for (const examType of examTypes ?? []) {
    const field = (examType.items ?? []).find((candidate) => candidate.id === examTypeFieldId);
    if (field !== undefined) {
      return field.name;
    }
  }
  return "";
}

export function examFieldOptionsForExamType(
  examTypes: ExaminationTypeMaster[] | undefined,
  examTypeId: string | null,
): LabDeviceExamFieldOption[] {
  if (examTypeId === null) {
    return [];
  }
  const examType = (examTypes ?? []).find((candidate) => candidate.id === examTypeId);
  return (examType?.items ?? []).map((field) => ({
    id: field.id,
    label: field.name,
  }));
}

export function restrictDraftsToExamType(
  drafts: LabDeviceItemDraft[],
  examTypeId: string | null,
  examTypes: ExaminationTypeMaster[] | undefined,
): LabDeviceItemDraft[] {
  const allowed = new Set(examFieldOptionsForExamType(examTypes, examTypeId).map((option) => option.id));
  return drafts.map((draft) => ({
    ...draft,
    examTypeFieldId:
      draft.examTypeFieldId !== null && allowed.has(draft.examTypeFieldId)
        ? draft.examTypeFieldId
        : null,
  }));
}

export type LabDeviceItemDraft = {
  id: string;
  examTypeFieldId: string | null;
  isActive: boolean;
};

export function labDeviceSourceLabel(sourceType: string): string {
  return LAB_DEVICE_SOURCE_LABELS[sourceType] ?? sourceType;
}

export function labDeviceValueShapeLabel(valueShape: string): string {
  return LAB_DEVICE_VALUE_SHAPE_LABELS[valueShape] ?? valueShape;
}

export function parseLabDeviceSourceQuery(value: string | null): string | null {
  const source = value?.trim() ?? "";
  return source === "" ? null : source;
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

// 機器の「検査」が選ばれていればその検査の項目だけ、未選択なら全検査の項目を出す。
// 現在値が候補に無い（検査変更や field 削除の後）行は欠落フィールドとして残し、黙って消さない。
export function labDeviceItemSelectOptions(
  examTypes: ExaminationTypeMaster[] | undefined,
  examTypeId: string | null,
  currentFieldId: string | null,
): Array<{ value: string; label: string }> {
  const base = examTypeId === null
    ? buildExamFieldOptions(examTypes)
    : examFieldOptionsForExamType(examTypes, examTypeId);
  const withCurrent = examFieldOptionsForItem(base, currentFieldId);
  return [
    { value: LAB_DEVICE_UNMAPPED_FIELD, label: LAB_DEVICE_EXAM_UNSET },
    ...withCurrent.map((option) => ({ value: option.id, label: option.label })),
  ];
}

export function countDraftsUnmappedByExamChange(
  drafts: LabDeviceItemDraft[],
  nextExamTypeId: string | null,
  examTypes: ExaminationTypeMaster[] | undefined,
): number {
  if (nextExamTypeId === null) {
    return 0;
  }
  const restricted = restrictDraftsToExamType(drafts, nextExamTypeId, examTypes);
  let count = 0;
  for (let i = 0; i < drafts.length; i += 1) {
    if (drafts[i]!.examTypeFieldId !== null && restricted[i]!.examTypeFieldId === null) {
      count += 1;
    }
  }
  return count;
}

export function labDeviceFieldUnit(
  examTypeFieldId: string | null,
  examTypes: ExaminationTypeMaster[] | undefined,
): string {
  if (examTypeFieldId === null) {
    return "";
  }
  for (const examType of examTypes ?? []) {
    const field = (examType.items ?? []).find((candidate) => candidate.id === examTypeFieldId);
    if (field !== undefined) {
      return field.unit;
    }
  }
  return "";
}

// 単位が両方入っていて食い違う場合だけ true（大文字小文字と空白の揺れは不一致にしない）
export function labDeviceUnitMismatch(deviceUnit: string, fieldUnit: string): boolean {
  const left = deviceUnit.trim().toLowerCase();
  const right = fieldUnit.trim().toLowerCase();
  return left !== "" && right !== "" && left !== right;
}

export function examFieldSelectValue(examTypeFieldId: string | null): string {
  return examTypeFieldId ?? "";
}

export function parseExamFieldSelectValue(value: string): string | null {
  return value === "" || value === LAB_DEVICE_UNMAPPED_FIELD ? null : value;
}

export function validateLabDeviceItemMasterDraft(input: {
  examTypeFieldId: string | null;
}): string | null {
  if (input.examTypeFieldId !== null) {
    const id = Number(input.examTypeFieldId);
    if (!Number.isInteger(id) || id <= 0) {
      return "検査項目が不正です";
    }
  }
  return null;
}

export function buildLabDeviceItemMasterUpdateRequest(input: {
  unit: string;
  examTypeFieldId: string | null;
  isActive: boolean;
}): UpdateLabDeviceItemMasterRequest {
  return {
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

export function examTypeName(
  examTypeId: string | null,
  examTypes: ExaminationTypeMaster[] | undefined,
): string | null {
  if (examTypeId === null) {
    return null;
  }
  return (examTypes ?? []).find((examType) => examType.id === examTypeId)?.name ?? null;
}

export function examTypeSelectOptions(
  examTypes: ExaminationTypeMaster[] | undefined,
): Array<{ value: string; label: string }> {
  const named = (examTypes ?? [])
    .slice()
    .sort((left, right) => left.sortOrder - right.sortOrder || left.name.localeCompare(right.name))
    .map((examType) => ({ value: examType.id, label: examType.name }));
  return [{ value: LAB_DEVICE_EXAM_SELECT_UNSET, label: LAB_DEVICE_EXAM_UNSET }, ...named];
}

export function examTypeSelectValue(examTypeId: string | null): string {
  return examTypeId ?? LAB_DEVICE_EXAM_SELECT_UNSET;
}

export function parseExamTypeSelectValue(value: string): string | null {
  return value === "" || value === LAB_DEVICE_EXAM_SELECT_UNSET ? null : value;
}

export function defaultLabDeviceSortOrder(sourceType: string): number {
  const index = LAB_DEVICE_SOURCE_ORDER.indexOf(sourceType as (typeof LAB_DEVICE_SOURCE_ORDER)[number]);
  return index === -1 ? 100 : (index + 1) * 10;
}

export function availableLabDeviceSourceTypes(devices: LabDevice[] | undefined): string[] {
  const used = new Set((devices ?? []).map((device) => device.sourceType));
  return LAB_DEVICE_SOURCE_ORDER.filter((sourceType) => !used.has(sourceType));
}

export function labDeviceToFormData(
  device: LabDevice | LabDeviceRow | null,
  unusedSourceTypes: string[],
): LabDeviceFormData {
  if (device !== null) {
    return {
      name: device.name,
      sourceType: device.sourceType,
      examTypeId: device.examTypeId,
      isActive: device.isActive,
      sortOrder: device.sortOrder,
    };
  }
  const sourceType = unusedSourceTypes[0] ?? "";
  return {
    name: sourceType === "" ? "" : labDeviceSourceLabel(sourceType),
    sourceType,
    examTypeId: null,
    isActive: true,
    sortOrder: defaultLabDeviceSortOrder(sourceType),
  };
}

export function validateLabDeviceDraft(input: {
  name: string;
  sourceType: string;
  examTypeId: string | null;
  requireSourceType: boolean;
}): string | null {
  if (input.name.trim() === "") {
    return "機器名は必須です";
  }
  if (input.name.trim().length > 100) {
    return "機器名が長すぎます";
  }
  if (input.requireSourceType && !LAB_DEVICE_SOURCE_ORDER.includes(input.sourceType as (typeof LAB_DEVICE_SOURCE_ORDER)[number])) {
    return "プロトコルを選んでください";
  }
  if (input.examTypeId !== null) {
    const id = Number(input.examTypeId);
    if (!Number.isInteger(id) || id <= 0) {
      return "検査が不正です";
    }
  }
  return null;
}

export function buildLabDeviceCreateRequest(input: LabDeviceFormData): CreateLabDeviceRequest {
  return {
    name: input.name.trim(),
    source_type: input.sourceType,
    exam_type_id: input.examTypeId === null ? null : Number(input.examTypeId),
    is_active: input.isActive,
    sort_order: input.sortOrder,
  };
}

export function buildLabDeviceUpdateRequest(input: LabDeviceFormData): UpdateLabDeviceRequest {
  return {
    name: input.name.trim(),
    exam_type_id: input.examTypeId === null ? null : Number(input.examTypeId),
    is_active: input.isActive,
    sort_order: input.sortOrder,
  };
}

// 一覧の「検査」列は persist の実挙動（項目対応から導出）を正とする。
// 項目が未対応のときだけ機器に設定した検査名を意図として見せる。
function deviceExamLabel(
  device: LabDevice,
  list: LabDeviceItemMaster[],
  examTypes: ExaminationTypeMaster[] | undefined,
): string {
  const derived = labDeviceExamLabel(list, examTypes);
  if (derived !== LAB_DEVICE_EXAM_UNSET) {
    return derived;
  }
  return examTypeName(device.examTypeId, examTypes) ?? LAB_DEVICE_EXAM_UNSET;
}

export function toLabDeviceRows(
  devices: LabDevice[] | undefined,
  items: LabDeviceItemMaster[] | undefined,
  examTypes?: ExaminationTypeMaster[],
): LabDeviceRow[] {
  return (devices ?? [])
    .slice()
    .sort((left, right) => left.sortOrder - right.sortOrder || Number(left.id) - Number(right.id))
    .map((device) => {
      const list = itemsForLabDevice(items, device.sourceType);
      return {
        id: device.id,
        sourceType: device.sourceType,
        name: device.name,
        examTypeId: device.examTypeId,
        examLabel: deviceExamLabel(device, list, examTypes),
        isActive: device.isActive,
        sortOrder: device.sortOrder,
        itemCount: list.length,
        unmappedCount: list.filter((item) => item.examTypeFieldId === null).length,
      };
    });
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

export function itemToLabDeviceDraft(item: LabDeviceItemMaster): LabDeviceItemDraft {
  return {
    id: item.id,
    examTypeFieldId: item.examTypeFieldId,
    isActive: item.isActive,
  };
}

export function isLabDeviceItemDraftDirty(item: LabDeviceItemMaster, draft: LabDeviceItemDraft): boolean {
  return item.examTypeFieldId !== draft.examTypeFieldId
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
        unit: item.unit,
        examTypeFieldId: draft.examTypeFieldId,
        isActive: draft.isActive,
      }),
    });
  }
  return { error: null, updates };
}
