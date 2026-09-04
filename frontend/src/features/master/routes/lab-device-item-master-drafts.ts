/**
 * lab-device-item-master-settings-model の draft ⇄ API リクエスト変換を担当。
 */
import type { CreateLabDeviceRequest, LabDevice, UpdateLabDeviceRequest } from "../api/lab-devices";
import type {
  LabDeviceItemMaster,
  UpdateLabDeviceItemMasterRequest,
} from "../api/lab-device-item-masters";
import { labDeviceSourceLabel, defaultLabDeviceSortOrder } from "./lab-device-item-master-labels";
import { validateLabDeviceItemMasterDraft } from "./lab-device-item-master-validation";
import type {
  LabDeviceFormData,
  LabDeviceItemDraft,
  LabDeviceRow,
} from "./lab-device-item-master-types";

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

export function itemToLabDeviceDraft(item: LabDeviceItemMaster): LabDeviceItemDraft {
  return {
    id: item.id,
    examTypeFieldId: item.examTypeFieldId,
    isActive: item.isActive,
  };
}

function isLabDeviceItemDraftDirty(item: LabDeviceItemMaster, draft: LabDeviceItemDraft): boolean {
  return item.examTypeFieldId !== draft.examTypeFieldId || item.isActive !== draft.isActive;
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
