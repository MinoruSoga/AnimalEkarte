// LIFF 設定

export const LIFF_MOCK = import.meta.env.VITE_LIFF_MOCK === 'true';
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

// URL の pathname の最初のセグメントから clinicId を取得
export function getClinicId(): string {
  const segments = window.location.pathname.split('/').filter(Boolean);
  return segments[0] || '';
}
