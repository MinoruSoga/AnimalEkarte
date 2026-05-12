/** 現在の選択クリニック ID を localStorage から取得する。 */
export function getClinicId(): string {
  return localStorage.getItem("auth_current_clinic:v1") ?? "";
}
