import Axios, { AxiosError, AxiosHeaders } from "axios";
import { beforeEach, describe, expect, it, vi } from "vitest";

const { axiosGet } = vi.hoisted(() => ({
  axiosGet: vi.fn(),
}));

vi.mock("@/lib/axios", () => ({
  axios: {
    get: (...args: unknown[]) => axiosGet(...args),
  },
}));

import { restoreSession } from "./restore-session";
import { mapMeToAuthUser } from "./transforms";

const RAW_ME = {
  id: "staff-1",
  email: "staff@example.com",
  display_name: "Test Staff",
  is_system_admin: false,
  main_clinic_id: "clinic-1",
  clinic: null,
  clinics: [{ clinic_id: "clinic-1", clinic_name: "Test Clinic", is_main: true }],
  permissions: {},
};

function requestConfig() {
  return { headers: new AxiosHeaders() };
}

function httpError(status: number, data: unknown = { message: "error" }): AxiosError {
  const config = requestConfig();
  return new AxiosError("request failed", AxiosError.ERR_BAD_REQUEST, config, undefined, {
    config,
    data,
    headers: new AxiosHeaders(),
    status,
    statusText: "Error",
  });
}

describe("restoreSession", () => {
  beforeEach(() => {
    axiosGet.mockReset();
  });

  it("maps a valid 200 /me payload to verified200 AuthUser", async () => {
    axiosGet.mockResolvedValueOnce({ data: RAW_ME, status: 200 });

    await expect(restoreSession()).resolves.toEqual({
      kind: "verified200",
      user: mapMeToAuthUser(RAW_ME),
    });
  });

  it("classifies 401 as anonymous401", async () => {
    axiosGet.mockRejectedValueOnce(httpError(401, { message: "unauthorized" }));

    await expect(restoreSession()).resolves.toEqual({ kind: "anonymous401" });
  });

  it("classifies 403 clinic_selection_unavailable as restricted403", async () => {
    axiosGet.mockRejectedValueOnce(httpError(403, { error_code: "clinic_selection_unavailable" }));

    await expect(restoreSession()).resolves.toEqual({
      kind: "restricted403",
      reason: "clinic_selection_unavailable",
    });
  });

  it("classifies a generic 403 as restricted403 forbidden", async () => {
    axiosGet.mockRejectedValueOnce(httpError(403, { error: "forbidden" }));

    await expect(restoreSession()).resolves.toEqual({
      kind: "restricted403",
      reason: "forbidden",
    });
  });

  it("classifies ECONNABORTED as timeout, not anonymous401", async () => {
    const error = new AxiosError(
      "timeout of 8000ms exceeded",
      AxiosError.ECONNABORTED,
      requestConfig(),
    );
    axiosGet.mockRejectedValueOnce(error);

    const result = await restoreSession();
    expect(result).toEqual({ kind: "timeout" });
    expect(result.kind).not.toBe("anonymous401");
  });

  it("classifies a network error as transport network, not anonymous401", async () => {
    const error = new AxiosError("Network Error", AxiosError.ERR_NETWORK, requestConfig());
    axiosGet.mockRejectedValueOnce(error);

    const result = await restoreSession();
    expect(result).toEqual({ kind: "transport", cause: "network" });
    expect(result.kind).not.toBe("anonymous401");
  });

  it("classifies 5xx as transport server with httpStatus, not anonymous401", async () => {
    axiosGet.mockRejectedValueOnce(httpError(503, { error: "unavailable" }));

    const result = await restoreSession();
    expect(result).toEqual({ kind: "transport", cause: "server", httpStatus: 503 });
    expect(result.kind).not.toBe("anonymous401");
  });

  it("classifies AbortError/CanceledError as cancel, not anonymous401", async () => {
    axiosGet.mockRejectedValueOnce(new Axios.CanceledError("canceled"));

    const result = await restoreSession();
    expect(result).toEqual({ kind: "cancel" });
    expect(result.kind).not.toBe("anonymous401");
  });

  it("sends timeout 8000 and startupSessionRestore true on GET /v1/me", async () => {
    axiosGet.mockResolvedValueOnce({ data: RAW_ME, status: 200 });
    const controller = new AbortController();

    await restoreSession({ signal: controller.signal });

    expect(axiosGet).toHaveBeenCalledOnce();
    expect(axiosGet).toHaveBeenCalledWith(
      "/v1/me",
      expect.objectContaining({
        timeout: 8000,
        signal: controller.signal,
        startupSessionRestore: true,
      }),
    );
  });

  it("does not treat a malformed 200 as verified200 or anonymous401", async () => {
    axiosGet.mockResolvedValueOnce({ data: { id: 1, email: null }, status: 200 });

    const result = await restoreSession();
    expect(result.kind).not.toBe("verified200");
    expect(result.kind).not.toBe("anonymous401");
    expect(result).toEqual({ kind: "transport", cause: "server" });
  });
});
