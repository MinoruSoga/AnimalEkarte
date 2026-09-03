import Axios, { type InternalAxiosRequestConfig, type AxiosError } from "axios";
import {
  isAuthPublicPath,
  isPasswordRecoveryPublicPath,
} from "@/lib/auth-route-policy";
import { getStoredClinicId } from "@/lib/current-clinic";
import { parseInternalPath } from "@/lib/internal-navigation";
import { sanitizeNullBytes } from "@/lib/sanitize";
import { paths } from "@/config/paths";

const API_URL = import.meta.env.VITE_API_URL || "/api";

function requestInterceptor(config: InternalAxiosRequestConfig) {
  config.headers ??= new Axios.AxiosHeaders() as typeof config.headers;
  config.headers.Accept = "application/json";
  // H1: CSRF 保護用 X-Requested-With ヘッダを全リクエストに追加（preflight 強制で CSRF 防止）
  // テスト環境では Vitest の MSW 制限のためヘッダを省略（Backend CSRF middleware は preflight OPTIONS をスキップ）
  if (import.meta.env.MODE !== "test") {
    config.headers["X-Requested-With"] = "XMLHttpRequest";
  }
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
  // PR #186 review (P2-11/12/15): 拠点横断で取得したレコード（billing/medical record 等）の
  // 子リソースを操作する場合、呼び出し元が個別に X-Clinic-ID を指定できる必要がある。
  // 呼び出し元が既にヘッダーを設定している場合はそれを優先し、グローバル選択値で上書きしない。
  if (config.headers["X-Clinic-ID"] === undefined) {
    const clinicId = getStoredClinicId();
    if (clinicId !== null) {
      config.headers["X-Clinic-ID"] = clinicId;
    }
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

/**
 * FE-RC-076: 未認証セッションをログインへ退避する共通処理。
 * 401 ハンドリング内の 2 か所（既存パスワード変更 401 とリフレッシュ失敗時）で
 * 同一ロジックが重複していたため抽出（DRY）。`from` は内部相対パスのみを許可し
 * （parseInternalPath）、オープンリダイレクトを防ぐ。
 */
function redirectToLogin(): void {
  const safePath =
    parseInternalPath(`${window.location.pathname}${window.location.search}`) ??
    paths.home.getHref();
  const from = encodeURIComponent(safePath);
  window.location.href = `${paths.auth.login.getHref()}?from=${from}`;
}

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

    // Public auth pages intentionally work without a session. Expected login/recovery
    // 401s belong to the page and must not trigger a refresh request.
    if (
      error.response?.status === 401 &&
      isAuthPublicPath(window.location.pathname)
    ) {
      return Promise.reject(error);
    }

    // --- 2. 401 認証リフレッシュロジック (既存) ---
    const originalRequest = config;
    if (
      error.response?.status !== 401 ||
      originalRequest === undefined ||
      originalRequest._retry === true ||
      originalRequest.url?.includes("/auth/refresh") === true ||
      originalRequest.url?.includes("/users/me/password") === true
    ) {
      // パスワード変更の 401 は「現在のパスワード誤り」でありセッション切れではない（BUG-026）
      if (
        error.response?.status === 401 &&
        window.location.pathname !== "/login" &&
        originalRequest?.url?.includes("/users/me/password") !== true
      ) {
        redirectToLogin();
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
      if (
        window.location.pathname !== "/login" &&
        !isPasswordRecoveryPublicPath(window.location.pathname)
      ) {
        redirectToLogin();
      }
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  },
);
