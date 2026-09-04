/**
 * lab-device-item-master 設定の公開 API バレル。
 * FE-RC-045: 元は 1 ファイル 531 行だったため、責務別に 3 分割した:
 * - lab-device-item-master-types.ts: 共有型・定数
 * - lab-device-item-master-labels.ts: 表示ラベル・一覧行構築
 * - lab-device-item-master-drafts.ts: draft ⇄ API リクエスト変換
 * - lab-device-item-master-validation.ts: バリデーション
 * 既存の import パス（このファイル）は互換のため維持する。
 */
export {
  LAB_DEVICE_UNMAPPED_FIELD,
  LAB_DEVICE_EXAM_SELECT_UNSET,
  LAB_DEVICE_EXAM_UNSET,
  LAB_DEVICE_EXAM_MIXED,
  type LabDeviceRow,
  type LabDeviceFormData,
  type LabDeviceItemDraft,
} from "./lab-device-item-master-types";

export {
  labDeviceExamLabel,
  labDeviceExamTypeId,
  labDeviceFieldName,
  examFieldOptionsForExamType,
  restrictDraftsToExamType,
  labDeviceSourceLabel,
  labDeviceValueShapeLabel,
  parseLabDeviceSourceQuery,
  buildExamFieldOptions,
  examFieldOptionsForItem,
  labDeviceItemSelectOptions,
  countDraftsUnmappedByExamChange,
  labDeviceFieldUnit,
  labDeviceUnitMismatch,
  examFieldSelectValue,
  parseExamFieldSelectValue,
  groupLabDeviceItemMasters,
  examTypeSelectOptions,
  examTypeSelectValue,
  parseExamTypeSelectValue,
  availableLabDeviceSourceTypes,
  itemsForLabDevice,
  toLabDeviceRows,
} from "./lab-device-item-master-labels";

export {
  buildLabDeviceItemMasterUpdateRequest,
  labDeviceToFormData,
  buildLabDeviceCreateRequest,
  buildLabDeviceUpdateRequest,
  itemToLabDeviceDraft,
  collectDirtyLabDeviceUpdates,
} from "./lab-device-item-master-drafts";

export {
  validateLabDeviceItemMasterDraft,
  validateLabDeviceDraft,
} from "./lab-device-item-master-validation";
