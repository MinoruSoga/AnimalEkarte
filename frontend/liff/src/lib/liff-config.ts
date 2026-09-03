export const LIFF_MOCK = import.meta.env.VITE_LIFF_MOCK === 'true';
export const LIFF_ID = import.meta.env.VITE_LIFF_ID || '';
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '';

/** モック連携成功を表示するまでの待ち時間 (FE5-6) */
export const LINK_SUCCESS_DISPLAY_MS = 800;
