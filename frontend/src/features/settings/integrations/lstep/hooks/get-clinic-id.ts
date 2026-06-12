/** 現在の選択クリニック ID を localStorage から取得する。 */
export function getClinicId(): string | null {
  try {
    return localStorage.getItem("auth_current_clinic:v1");
  } catch {
    return null;
  }
}
