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

type RetryableConfig = InternalAxiosRequestConfig & { _retry?: boolean };

axios.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as RetryableConfig | undefined;

    // 401 以外のエラー、リトライ済み、リフレッシュエンドポイント自体のエラーはスルー
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
        window.location.href = "/login";
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
        window.location.href = "/login";
      }
      return Promise.reject(refreshError);
    } finally {
      isRefreshing = false;
    }
  },
);
