import Axios, { type InternalAxiosRequestConfig } from "axios";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api";

function authRequestInterceptor(config: InternalAxiosRequestConfig) {
  config.headers ??= new Axios.AxiosHeaders() as typeof config.headers;
  config.headers.Accept = "application/json";
  config.headers["X-Request-ID"] = crypto.randomUUID();

  // Authorization Bearer トークンを sessionStorage から取得して設定
  const token = sessionStorage.getItem("auth_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  return config;
}

export const axios = Axios.create({
  baseURL: API_URL,
  timeout: 60000,
  // Authorization Bearer トークンを使うため withCredentials は不要
  headers: { "Content-Type": "application/json" },
});

axios.interceptors.request.use(authRequestInterceptor);
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (
      error.response?.status === 401 &&
      window.location.pathname !== "/login"
    ) {
      // セッション無効 → sessionStorage のトークンをクリア
      sessionStorage.removeItem("auth_token");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  },
);
