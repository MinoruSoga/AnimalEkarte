import Axios, { type InternalAxiosRequestConfig, type AxiosError } from "axios";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api";

function requestInterceptor(config: InternalAxiosRequestConfig) {
  config.headers ??= new Axios.AxiosHeaders() as typeof config.headers;
  config.headers.Accept = "application/json";
  config.headers["X-Request-ID"] = crypto.randomUUID();
  // Authorization ヘッダは不要 — httpOnly Cookie が自動送信される（withCredentials: true）
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
