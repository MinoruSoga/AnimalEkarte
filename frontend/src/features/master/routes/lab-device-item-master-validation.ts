/**
 * lab-device-item-master-settings-model のバリデーションのみを担当。
 */
import { LAB_DEVICE_SOURCE_ORDER } from "./lab-device-item-master-types";

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
  if (
    input.requireSourceType &&
    !LAB_DEVICE_SOURCE_ORDER.includes(input.sourceType as (typeof LAB_DEVICE_SOURCE_ORDER)[number])
  ) {
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
