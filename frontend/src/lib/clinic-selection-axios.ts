import Axios, { type AxiosError, type AxiosInstance, type InternalAxiosRequestConfig } from "axios";
import {
  areClinicWritesPaused,
  isClinicSelectionUnavailable,
  recoverClinicSelectionOnce,
} from "@/lib/clinic-selection-recovery";

let attached = false;

const PRE_SESSION_AUTH_ENDPOINTS = [
  "/v1/login",
  "/v1/auth/forgot-password",
  "/v1/auth/reset-password",
] as const;

function requestPath(config: InternalAxiosRequestConfig): string {
  return (config.url ?? "").split("?")[0];
}

function isSessionLogoutRequest(config: InternalAxiosRequestConfig): boolean {
  const method = config.method?.toLowerCase();
  const path = requestPath(config);
  return method === "post" && path.includes("/auth/refresh/logout");
}

function relativeBaseOrigin(): string {
  if (
    typeof window !== "undefined" &&
    window.location?.origin &&
    window.location.origin !== "null"
  ) {
    return window.location.origin;
  }
  return "https://ae-relative-base.invalid";
}

function isAbsoluteUrl(url: string): boolean {
  return /^(?:[a-z][a-z\d+\-.]*:)?\/\//i.test(url);
}

function combineUrls(baseURL: string, requestedURL: string): string {
  if (requestedURL === "") {
    return baseURL;
  }
  if (baseURL === "") {
    return requestedURL;
  }
  return `${baseURL.replace(/\/+$/, "")}/${requestedURL.replace(/^\/+/, "")}`;
}

function joinBasePath(basePath: string, endpoint: string): string {
  const normalized = basePath.replace(/\/+$/, "");
  if (normalized === "" || normalized === "/") {
    return endpoint;
  }
  return `${normalized}${endpoint}`;
}

function resolveAgainstConfiguredBase(
  config: InternalAxiosRequestConfig,
  client: AxiosInstance,
): { request: URL; base: URL } | null {
  const requestedURL = config.url ?? "";
  const baseURL = config.baseURL ?? client.defaults.baseURL ?? "";
  const origin = relativeBaseOrigin();
  try {
    const base = new URL(baseURL === "" ? "/" : baseURL, origin);
    const fullPath =
      baseURL !== "" && !isAbsoluteUrl(requestedURL)
        ? combineUrls(baseURL, requestedURL)
        : requestedURL;
    const request = new URL(fullPath === "" ? "/" : fullPath, origin);
    return { request, base };
  } catch {
    return null;
  }
}

function isPreSessionAuthRequest(
  config: InternalAxiosRequestConfig,
  client: AxiosInstance,
): boolean {
  const method = config.method?.toLowerCase();
  if (method !== "post") {
    return false;
  }
  const resolved = resolveAgainstConfiguredBase(config, client);
  if (resolved === null) {
    return false;
  }
  if (resolved.request.origin !== resolved.base.origin) {
    return false;
  }
  return PRE_SESSION_AUTH_ENDPOINTS.some(
    (endpoint) => resolved.request.pathname === joinBasePath(resolved.base.pathname, endpoint),
  );
}

function isWriteMethod(method: string | undefined): boolean {
  const normalized = method?.toLowerCase();
  return (
    normalized === "post" ||
    normalized === "put" ||
    normalized === "patch" ||
    normalized === "delete"
  );
}

export function attachClinicSelectionInterceptors(client: AxiosInstance): void {
  if (attached) {
    return;
  }
  attached = true;

  client.interceptors.request.use((config) => {
    if (
      areClinicWritesPaused() &&
      isWriteMethod(config.method) &&
      !isSessionLogoutRequest(config) &&
      !isPreSessionAuthRequest(config, client)
    ) {
      return Promise.reject(new Axios.CanceledError("clinic writes paused"));
    }
    return config;
  });

  client.interceptors.response.use(
    (response) => response,
    (error: AxiosError) => {
      const config = error.config;
      if (config === undefined) {
        return Promise.reject(error);
      }
      if (
        isClinicSelectionUnavailable(error.response?.data) &&
        !isPreSessionAuthRequest(config, client)
      ) {
        // Startup restore classifies clinic403 itself and owns recovery under the 8s budget.
        // Awaiting recovery here would leave AuthProvider pending until recovery finishes/aborts.
        if (config.startupSessionRestore === true) {
          return Promise.reject(error);
        }
        if (isWriteMethod(config.method)) {
          void recoverClinicSelectionOnce();
          return Promise.reject(error);
        }
        return recoverClinicSelectionOnce().then(() => Promise.reject(error));
      }
      if (
        areClinicWritesPaused() &&
        isWriteMethod(config.method) &&
        !isSessionLogoutRequest(config) &&
        !isPreSessionAuthRequest(config, client)
      ) {
        return Promise.reject(error);
      }
      return Promise.reject(error);
    },
  );
}
