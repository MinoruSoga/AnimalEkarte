import Axios, { AxiosError } from "axios";
import { axios } from "@/lib/axios";
import { isClinicSelectionUnavailable } from "@/lib/clinic-selection-recovery";
import type { AuthUser } from "../types";
import { type BackendMeResponse, mapMeToAuthUser } from "./transforms";

export type SessionRestoreResult =
  | { kind: "verified200"; user: AuthUser }
  | { kind: "anonymous401" }
  | { kind: "restricted403"; reason: "clinic_selection_unavailable" | "forbidden" }
  | { kind: "timeout" }
  | { kind: "transport"; cause: "network" | "server"; httpStatus?: number }
  | { kind: "cancel" };

function isCanceledError(error: unknown): boolean {
  if (Axios.isCancel(error) || error instanceof Axios.CanceledError) {
    return true;
  }
  if (Axios.isAxiosError(error) && error.code === AxiosError.ERR_CANCELED) {
    return true;
  }
  return error instanceof Error && error.name === "AbortError";
}

export async function restoreSession(options?: {
  signal?: AbortSignal;
}): Promise<SessionRestoreResult> {
  try {
    const { data } = await axios.get<BackendMeResponse>("/v1/me", {
      timeout: 8000,
      signal: options?.signal,
      startupSessionRestore: true,
    });
    try {
      return { kind: "verified200", user: mapMeToAuthUser(data) };
    } catch {
      return { kind: "transport", cause: "server" };
    }
  } catch (error) {
    if (isCanceledError(error)) {
      return { kind: "cancel" };
    }
    if (!Axios.isAxiosError(error)) {
      return { kind: "transport", cause: "network" };
    }
    if (error.code === AxiosError.ECONNABORTED || error.code === "ETIMEDOUT") {
      return { kind: "timeout" };
    }
    const status = error.response?.status;
    if (status === 401) {
      return { kind: "anonymous401" };
    }
    if (status === 403) {
      return {
        kind: "restricted403",
        reason: isClinicSelectionUnavailable(error.response?.data)
          ? "clinic_selection_unavailable"
          : "forbidden",
      };
    }
    if (status !== undefined) {
      return { kind: "transport", cause: "server", httpStatus: status };
    }
    return { kind: "transport", cause: "network" };
  }
}
