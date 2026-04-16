import Axios, { type InternalAxiosRequestConfig, type AxiosError } from "axios";

const API_URL = import.meta.env.VITE_API_URL || "/api";

/**
 * BUG-067: NULL バイト（\u0000）を再帰的に除去する。
 * PostgreSQL は NULL バイトを含む文字列を受け付けないため 500 エラーになる。
 */
function sanitizeNullBytes(value: unknown): unknown {
  if (typeof value === "string") {
    // eslint-disable-next-line no-control-regex -- NULL バイト除去のため意図的に制御文字を使用
    return value.replace(/\u0000/g, "");
  }
  if (Array.isArray(value)) {
    return value.map(sanitizeNullBytes);
  }
  if (value !== null && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value as Record<string, unknown>).map(([k, v]) => [k, sanitizeNullBytes(v)])
    );
  }
  return value;
}

function requestInterceptor(config: InternalAxiosRequestConfig) {
  config.headers ??= new Axios.AxiosHeaders() as typeof config.headers;
  config.headers.Accept = "application/json";
  // crypto.randomUUID() は HTTPS または localhost (secure context) でのみ利用可能。
  // Docker内ホスト名 (frontend:3000) など非セキュアコンテキストでは使用不可のため
  // Math.random ベースのフォールバックを用意する。
  config.headers["X-Request-ID"] =
    typeof crypto.randomUUID === "function"
      ? crypto.randomUUID()
      : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
  // Authorization ヘッダは不要 — httpOnly Cookie が自動送信される（withCredentials: true）

  // クリニック切替: localStorage の選択クリニック ID をヘッダーで送信
  // バックエンドの auth ミドルウェアが X-Clinic-ID を優先して clinic_id コンテキストを上書きする
  try {
    const clinicId = localStorage.getItem("auth_current_clinic:v1");
    if (clinicId) {
      config.headers["X-Clinic-ID"] = clinicId;
    }
  } catch {
    /* SSR / localStorage 無効環境では無視 */
  }

  // BUG-067: POST/PATCH/PUT のリクエストボディから NULL バイトを除去
  const method = config.method?.toLowerCase();
  if ((method === "post" || method === "patch" || method === "put") && config.data) {
    config.data = sanitizeNullBytes(config.data);
  }

  return config;
}

export const axios = Axios.create({
  baseURL: API_URL,
  timeout: 60000,
  // httpOnly Cookie を自動送信するために必須
  withCredentials: true,
  headers: { "Content-Type": "application/json" },
});
axios.interceptors.request.use(requestInterceptor);

/** リトライ設定 */
const MAX_RETRIES = 2;
const RETRY_DELAY_MS = 1000;

/** 401 時の自動トークンリフレッシュ用キュー */
let isRefreshing = false;
let pendingRequests: Array<{
  resolve: () => void;
  reject: (reason: unknown) => void;
}> = [];

function processQueue(error: AxiosError | null): void {
  for (const pending of pendingRequests) {
    if (error !== null) {
      pending.reject(error);
    } else {
      pending.resolve();
    }
  }
  pendingRequests = [];
}

type RetryableConfig = InternalAxiosRequestConfig & { 
  _retryCount?: number;
  _retry?: boolean;
};

axios.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const config = error.config as RetryableConfig | undefined;
    if (!config) return Promise.reject(error);

    // --- 1. 自動リトライロジック (GETリクエストのみ) ---
    const isGetRequest = config.method?.toLowerCase() === "get";
    const isNetworkError = !error.response && error.code !== "ERR_CANCELED";
    const isServerError = error.response && error.response.status >= 502 && error.response.status <= 504;

    if (isGetRequest && (isNetworkError || isServerError)) {
      config._retryCount = config._retryCount ?? 0;

      if (config._retryCount < MAX_RETRIES) {
        config._retryCount += 1;
        // 指数バックオフ的な待機
        await new Promise(resolve => setTimeout(resolve, RETRY_DELAY_MS * config._retryCount!));
        return axios(config);
      }
    }

    // --- 2. 401 認証リフレッシュロジック (既存) ---
    const originalRequest = config;
    if (
      error.response?.status !== 401 ||
      originalRequest === undefined ||
      originalRequest._retry === true ||
      originalRequest.url?.includes("/auth/refresh") === true
    ) {
      // 401 で上記条件に当てはまらない場合（リフレッシュ不可）はログインへ
      if (
        error.response?.status === 401 &&
        window.location.pathname !== "/login"
      ) {
        const from = encodeURIComponent(window.location.pathname + window.location.search);
        window.location.href = `/login?from=${from}`;
      }
      return Promise.reject(error);
    }

    // 既にリフレッシュ中の場合はキューに積んで待機
    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        pendingRequests.push({
          resolve: () => resolve(axios(originalRequest)),
          reject,
        });
      });
    }

    originalRequest._retry = true;
    isRefreshing = true;

    try {
      await axios.post("/v1/auth/refresh");
      processQueue(null);
      return axios(originalRequest);
    } catch (refreshError) {
      processQueue(refreshError as AxiosError);
      if (window.location.pathname !== "/login") {
        const from = encodeURIComponent(window.location.pathname + window.location.search);
        window.location.href = `/login?from=${from}`;
      }
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  },
);
