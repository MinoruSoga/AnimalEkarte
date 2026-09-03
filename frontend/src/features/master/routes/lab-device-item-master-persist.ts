import { toast } from "sonner";

import type { LabDeviceItemMaster } from "../api/lab-device-item-masters";
import type { CreateLabDeviceRequest, SaveLabDeviceConfigurationRequest } from "../api/lab-devices";
import {
  buildLabDeviceCreateRequest,
  buildLabDeviceUpdateRequest,
  collectDirtyLabDeviceUpdates,
  validateLabDeviceDraft,
  type LabDeviceFormData,
  type LabDeviceItemDraft,
  type LabDeviceRow,
} from "./lab-device-item-master-settings-model";

interface PersistLabDeviceItemMasterOptions {
  form: LabDeviceFormData;
  drafts: LabDeviceItemDraft[];
  isCreating: boolean;
  canCreate: boolean;
  canEdit: boolean;
  selectedRow: LabDeviceRow | null;
  selectedItems: LabDeviceItemMaster[];
  createDevice: (req: CreateLabDeviceRequest) => Promise<unknown>;
  saveConfiguration: (args: {
    id: string;
    req: SaveLabDeviceConfigurationRequest;
  }) => Promise<unknown>;
}

export async function persistLabDeviceItemMaster({
  form,
  drafts,
  isCreating,
  canCreate,
  canEdit,
  selectedRow,
  selectedItems,
  createDevice,
  saveConfiguration,
}: PersistLabDeviceItemMasterOptions): Promise<"saved" | "aborted"> {
  const error = validateLabDeviceDraft({
    name: form.name,
    sourceType: form.sourceType,
    examTypeId: form.examTypeId,
    requireSourceType: isCreating,
  });
  if (error !== null) {
    toast.error(error);
    return "aborted";
  }
  try {
    if (isCreating) {
      if (canCreate !== true) {
        return "aborted";
      }
      await createDevice(buildLabDeviceCreateRequest(form));
      toast.success("登録しました");
    } else if (selectedRow !== null) {
      if (canEdit !== true) {
        return "aborted";
      }
      const itemChanges = collectDirtyLabDeviceUpdates(selectedItems, drafts);
      if (itemChanges.error !== null) {
        toast.error(itemChanges.error);
        return "aborted";
      }
      await saveConfiguration({
        id: selectedRow.id,
        req: {
          device: buildLabDeviceUpdateRequest(form),
          items: itemChanges.updates.map((update) => ({
            id: Number(update.id),
            ...update.req,
          })),
        },
      });
      toast.success("更新しました");
    }
  } catch {
    return "aborted";
  }
  return "saved";
}
