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

export function requireStoredClinicId(): string {
  const clinicId = getStoredClinicId();
  if (clinicId === null) {
    throw new Error("クリニックが選択されていません。ページをリロードしてください。");
  }
  return clinicId;
}
