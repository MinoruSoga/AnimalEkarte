export const CURRENT_CLINIC_STORAGE_KEY = "auth_current_clinic:v1";

export function normalizeClinicId(value: string | null): string | null {
  const trimmed = value?.trim();
  return trimmed ? trimmed : null;
}

export function getStoredClinicId(): string | null {
  try {
    return normalizeClinicId(localStorage.getItem(CURRENT_CLINIC_STORAGE_KEY));
  } catch {
    return null;
  }
}

/** 選択中クリニックを localStorage に保存する。空値や書込失敗時は false。 */
export function setStoredClinicId(clinicId: string): boolean {
  const normalized = normalizeClinicId(clinicId);
  if (normalized === null) {
    return false;
  }
  try {
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, normalized);
    return true;
  } catch {
    return false;
  }
}

export function requireStoredClinicId(): string {
  const clinicId = getStoredClinicId();
  if (clinicId === null) {
    throw new Error("クリニックが選択されていません。ページをリロードしてください。");
  }
  return clinicId;
}
