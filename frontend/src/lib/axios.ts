import Axios, { type InternalAxiosRequestConfig } from "axios";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api";

function authRequestInterceptor(config: InternalAxiosRequestConfig) {
  // httpOnly Cookie は自動送信されるため Authorization ヘッダへの手動注入は不要
  // withCredentials: true によりブラウザがCookieを自動付与する
  config.headers = config.headers || {};
  config.headers.Accept = "application/json";
  config.headers["X-Request-ID"] = crypto.randomUUID();
  return config;
}

export const axios = Axios.create({
  baseURL: API_URL,
  timeout: 60000,
  // httpOnly Cookie を cross-origin リクエストで自動送信するために必須
  withCredentials: true,
  headers: { "Content-Type": "application/json" },
});

axios.interceptors.request.use(authRequestInterceptor);
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (
      error.response?.status === 401 &&
      // すでに /login にいる場合はリダイレクトしない（無限ループ防止）
      // refreshToken() が /me を呼んで 401 を受け取るケースが該当
      window.location.pathname !== "/login"
    ) {
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);
