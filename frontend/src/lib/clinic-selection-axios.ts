import Axios, { type AxiosError, type AxiosInstance, type InternalAxiosRequestConfig } from "axios";
import {
  areClinicWritesPaused,
  isClinicSelectionUnavailable,
  recoverClinicSelectionOnce,
} from "@/lib/clinic-selection-recovery";

let attached = false;

function isSessionLogoutRequest(config: InternalAxiosRequestConfig): boolean {
  const method = config.method?.toLowerCase();
  const url = config.url ?? "";
  return method === "post" && url.includes("/auth/refresh/logout");
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
      !isSessionLogoutRequest(config)
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
      if (isClinicSelectionUnavailable(error.response?.data)) {
        if (isWriteMethod(config.method)) {
          void recoverClinicSelectionOnce();
          return Promise.reject(error);
        }
        return recoverClinicSelectionOnce().then(() => Promise.reject(error));
      }
      if (
        areClinicWritesPaused() &&
        isWriteMethod(config.method) &&
        !isSessionLogoutRequest(config)
      ) {
        return Promise.reject(error);
      }
      return Promise.reject(error);
    },
  );
}
