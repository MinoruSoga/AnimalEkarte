import Axios, { type InternalAxiosRequestConfig } from "axios";

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
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (
      error.response?.status === 401 &&
      window.location.pathname !== "/login"
    ) {
      // Cookie は httpOnly のため JavaScript からは削除不可 — /login へリダイレクトするのみ
      window.location.href = "/login";
    }
    return Promise.reject(error);
  },
);
