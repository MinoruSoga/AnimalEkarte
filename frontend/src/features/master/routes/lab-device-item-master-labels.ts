/**
 * lab-device-item-master-settings-model の表示ラベル / 一覧行構築を担当。
 * バリデーション・draft 変換は lab-device-item-master-validation.ts /
 * lab-device-item-master-drafts.ts を参照。
 */
import type { LabDevice } from "../api/lab-devices";
import type { LabDeviceItemMaster } from "../api/lab-device-item-masters";
import type { ExaminationTypeMaster } from "../api/exam-types-master";
import {
  LAB_DEVICE_EXAM_MIXED,
  LAB_DEVICE_EXAM_SELECT_UNSET,
  LAB_DEVICE_EXAM_UNSET,
  LAB_DEVICE_SOURCE_LABELS,
  LAB_DEVICE_SOURCE_ORDER,
  LAB_DEVICE_UNMAPPED_FIELD,
  LAB_DEVICE_VALUE_SHAPE_LABELS,
  type LabDeviceExamFieldOption,
  type LabDeviceItemDraft,
  type LabDeviceItemMasterGroup,
  type LabDeviceRow,
} from "./lab-device-item-master-types";

type ExamTypeByFieldId = Map<string, { id: string; name: string }>;

function examTypeByFieldId(examTypes: ExaminationTypeMaster[] | undefined): ExamTypeByFieldId {
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
  const allowed = new Set(
    examFieldOptionsForExamType(examTypes, examTypeId).map((option) => option.id),
  );
  return drafts.map((draft) => ({
    ...draft,
    examTypeFieldId:
      draft.examTypeFieldId !== null && allowed.has(draft.examTypeFieldId)
        ? draft.examTypeFieldId
        : null,
  }));
}

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
  const base =
    examTypeId === null
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
    .filter(
      (sourceType) =>
        !LAB_DEVICE_SOURCE_ORDER.includes(sourceType as (typeof LAB_DEVICE_SOURCE_ORDER)[number]),
    )
    .sort()
    .map((sourceType) => ({
      sourceType,
      label: labDeviceSourceLabel(sourceType),
      items: groups.get(sourceType) ?? [],
    }));
  return [...known, ...extras];
}

function examTypeName(
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
  const index = LAB_DEVICE_SOURCE_ORDER.indexOf(
    sourceType as (typeof LAB_DEVICE_SOURCE_ORDER)[number],
  );
  return index === -1 ? 100 : (index + 1) * 10;
}

export function availableLabDeviceSourceTypes(devices: LabDevice[] | undefined): string[] {
  const used = new Set((devices ?? []).map((device) => device.sourceType));
  return LAB_DEVICE_SOURCE_ORDER.filter((sourceType) => !used.has(sourceType));
}

export function itemsForLabDevice(
  items: LabDeviceItemMaster[] | undefined,
  sourceType: string,
): LabDeviceItemMaster[] {
  return (items ?? [])
    .filter((item) => item.sourceType === sourceType)
    .slice()
    .sort(
      (left, right) =>
        left.sortOrder - right.sortOrder || left.deviceItemCode.localeCompare(right.deviceItemCode),
    );
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
