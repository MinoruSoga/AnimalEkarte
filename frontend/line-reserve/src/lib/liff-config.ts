// LIFF 設定

export const LIFF_MOCK = import.meta.env.VITE_LIFF_MOCK === "true";
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "";

// URL の pathname から clinicId を取得
// /line-reserve/{clinicId}/... → clinicId を返す
// ローカル開発時（/ 直下）は segments[0] を返す
export function getClinicId(): string {
  const segments = window.location.pathname.split("/").filter(Boolean);
  if (segments[0] === "line-reserve") {
    return segments[1] || "";
  }
  return segments[0] || "";
}
