import { AxiosError, AxiosHeaders, type AxiosAdapter } from "axios";
import { afterEach, describe, expect, it, vi } from "vitest";

const recoveryMocks = vi.hoisted(() => ({
  recover: vi.fn().mockResolvedValue(undefined),
}));

vi.mock("./clinic-selection-recovery", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./clinic-selection-recovery")>();
  return {
    ...actual,
    recoverClinicSelectionOnce: (...args: unknown[]) => recoveryMocks.recover(...args),
  };
});

import { axios } from "./axios";
import { attachClinicSelectionInterceptors } from "./clinic-selection-axios";
import {
  CLINIC_SELECTION_UNAVAILABLE,
  areClinicWritesPaused,
  pauseClinicWrites,
  resetClinicSelectionRecoveryForTests,
} from "./clinic-selection-recovery";

function configuredAbsoluteUrl(endpoint: string, query = ""): string {
  const base = axios.defaults.baseURL ?? "/api";
  const path = `${base.replace(/\/+$/, "")}/${endpoint.replace(/^\/+/, "")}${query}`;
  try {
    return new URL(path).href;
  } catch {
    return new URL(path, window.location.origin).href;
  }
}

function foreignOriginSuffixUrl(endpoint: string): string {
  const configured = new URL(configuredAbsoluteUrl(endpoint));
  const foreignHost = configured.hostname === "evil.example" ? "hostile.example" : "evil.example";
  return `${configured.protocol}//${foreignHost}${endpoint}`;
}

function okAdapter(onCall: () => void): AxiosAdapter {
  return async (config) => {
    onCall();
    return {
      config,
      data: {},
      headers: new AxiosHeaders(),
      status: 200,
      statusText: "OK",
    };
  };
}

describe("clinic selection axios guards", () => {
  afterEach(() => {
    recoveryMocks.recover.mockReset().mockResolvedValue(undefined);
    resetClinicSelectionRecoveryForTests();
  });

  it("does not replay a failed write after clinic selection recovery", async () => {
    attachClinicSelectionInterceptors(axios);
    let adapterCalls = 0;
    const adapter: AxiosAdapter = async (config) => {
      adapterCalls += 1;
      throw new AxiosError("clinic unavailable", AxiosError.ERR_BAD_REQUEST, config, undefined, {
        config,
        data: { error_code: CLINIC_SELECTION_UNAVAILABLE },
        headers: new AxiosHeaders(),
        status: 403,
        statusText: "Forbidden",
      });
    };

    await expect(
      axios.request({ adapter, method: "post", url: "/v1/owners", data: { name: "x" } }),
    ).rejects.toThrow("clinic unavailable");
    expect(adapterCalls).toBe(1);
    expect(recoveryMocks.recover).toHaveBeenCalledOnce();
  });

  it("does not recover from a generic forbidden response", async () => {
    attachClinicSelectionInterceptors(axios);
    const adapter: AxiosAdapter = async (config) => {
      throw new AxiosError("forbidden", AxiosError.ERR_BAD_REQUEST, config, undefined, {
        config,
        data: { error: "forbidden" },
        headers: new AxiosHeaders(),
        status: 403,
        statusText: "Forbidden",
      });
    };

    await expect(axios.request({ adapter, method: "get", url: "/v1/staffs" })).rejects.toThrow(
      "forbidden",
    );
    expect(recoveryMocks.recover).not.toHaveBeenCalled();
  });

  it("rejects subsequent writes while clinic writes are paused", async () => {
    attachClinicSelectionInterceptors(axios);
    pauseClinicWrites();
    let adapterCalls = 0;
    const adapter: AxiosAdapter = async (config) => {
      adapterCalls += 1;
      return {
        config,
        data: {},
        headers: new AxiosHeaders(),
        status: 200,
        statusText: "OK",
      };
    };

    await expect(
      axios.request({ adapter, method: "post", url: "/v1/owners", data: { name: "x" } }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    expect(adapterCalls).toBe(0);
  });

  it("does not recover from a failed login without an issued session", async () => {
    attachClinicSelectionInterceptors(axios);
    const adapter: AxiosAdapter = async (config) => {
      throw new AxiosError("clinic unavailable", AxiosError.ERR_BAD_REQUEST, config, undefined, {
        config,
        data: { error_code: CLINIC_SELECTION_UNAVAILABLE },
        headers: new AxiosHeaders(),
        status: 403,
        statusText: "Forbidden",
      });
    };

    await expect(
      axios.request({
        adapter,
        method: "post",
        url: "/v1/login",
        data: { email: "a", password: "b" },
      }),
    ).rejects.toThrow("clinic unavailable");
    expect(recoveryMocks.recover).not.toHaveBeenCalled();
    expect(areClinicWritesPaused()).toBe(false);
  });

  it("still allows login while clinic writes are paused", async () => {
    attachClinicSelectionInterceptors(axios);
    pauseClinicWrites();
    let adapterCalls = 0;
    const adapter: AxiosAdapter = async (config) => {
      adapterCalls += 1;
      return {
        config,
        data: {},
        headers: new AxiosHeaders(),
        status: 200,
        statusText: "OK",
      };
    };

    await axios.request({
      adapter,
      method: "post",
      url: "/v1/login",
      data: { email: "a", password: "b" },
    });
    expect(adapterCalls).toBe(1);
  });

  it("still allows exact POST /v1/auth/forgot-password and /v1/auth/reset-password while paused and still rejects business writes and near-match /auth paths", async () => {
    attachClinicSelectionInterceptors(axios);
    pauseClinicWrites();
    let adapterCalls = 0;
    const adapter: AxiosAdapter = async (config) => {
      adapterCalls += 1;
      return {
        config,
        data: {},
        headers: new AxiosHeaders(),
        status: 200,
        statusText: "OK",
      };
    };

    await axios.request({
      adapter,
      method: "post",
      url: "/v1/auth/forgot-password",
      data: { email: "a" },
    });
    await axios.request({
      adapter,
      method: "post",
      url: "/v1/auth/reset-password",
      data: { token: "t", password: "p" },
    });
    expect(adapterCalls).toBe(2);

    await expect(
      axios.request({
        adapter,
        method: "post",
        url: foreignOriginSuffixUrl("/v1/auth/forgot-password"),
        data: { email: "a" },
      }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({
        adapter,
        method: "post",
        url: foreignOriginSuffixUrl("/v1/auth/reset-password"),
        data: { token: "t", password: "p" },
      }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    expect(adapterCalls).toBe(2);

    await expect(
      axios.request({ adapter, method: "post", url: "/v1/owners", data: { name: "x" } }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({ adapter, method: "put", url: "/v1/owners/1", data: { name: "y" } }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({ adapter, method: "patch", url: "/v1/pets/1", data: { name: "z" } }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({ adapter, method: "delete", url: "/v1/owners/1" }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({ adapter, method: "post", url: "/v1/auth/refresh" }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({ adapter, method: "post", url: "/v1/auth/forgot-password/extra" }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({ adapter, method: "put", url: "/v1/auth/forgot-password" }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    expect(adapterCalls).toBe(2);
  });

  it("paused pre-session bypass is POST-only exact configured-base paths and rejects hostile identity", async () => {
    attachClinicSelectionInterceptors(axios);
    pauseClinicWrites();
    let adapterCalls = 0;
    const adapter = okAdapter(() => {
      adapterCalls += 1;
    });

    const endpoints = ["/v1/login", "/v1/auth/forgot-password", "/v1/auth/reset-password"] as const;
    for (const endpoint of endpoints) {
      await axios.request({ adapter, method: "post", url: endpoint, data: {} });
      await axios.request({ adapter, method: "post", url: `${endpoint}?q=1`, data: {} });
      await axios.request({
        adapter,
        method: "post",
        url: configuredAbsoluteUrl(endpoint),
        data: {},
      });
    }
    expect(adapterCalls).toBe(9);

    const allowed = adapterCalls;
    await expect(
      axios.request({
        adapter,
        method: "post",
        url: foreignOriginSuffixUrl("/v1/login"),
        data: {},
      }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({ adapter, method: "post", url: "/evil/v1/login", data: {} }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({ adapter, method: "post", url: "/v1/login/extra", data: {} }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({ adapter, method: "post", url: "/v1/auth/forgot-password/extra", data: {} }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({ adapter, method: "put", url: "/v1/login", data: {} }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({ adapter, method: "post", url: "/v1/owners", data: { name: "x" } }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({
        adapter,
        baseURL: foreignOriginSuffixUrl(""),
        method: "post",
        url: "/v1/login",
        data: {},
      }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    expect(adapterCalls).toBe(allowed);
  });

  it("still allows logout while clinic writes are paused", async () => {
    attachClinicSelectionInterceptors(axios);
    pauseClinicWrites();
    let adapterCalls = 0;
    const adapter: AxiosAdapter = async (config) => {
      adapterCalls += 1;
      return {
        config,
        data: {},
        headers: new AxiosHeaders(),
        status: 200,
        statusText: "OK",
      };
    };

    await axios.request({ adapter, method: "post", url: "/v1/auth/refresh/logout" });
    await expect(
      axios.request({ adapter, method: "post", url: "/evil/v1/auth/refresh/logout/trailing" }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    await expect(
      axios.request({
        adapter,
        baseURL: foreignOriginSuffixUrl(""),
        method: "post",
        url: "/v1/auth/refresh/logout",
      }),
    ).rejects.toMatchObject({ message: "clinic writes paused" });
    expect(adapterCalls).toBe(1);
  });

  it("does not start clinic recovery for startupSessionRestore GET /v1/me", async () => {
    attachClinicSelectionInterceptors(axios);
    let recoverStarted = false;
    recoveryMocks.recover.mockImplementation(
      () =>
        new Promise((resolve) => {
          recoverStarted = true;
          resolve(undefined);
        }),
    );
    const adapter: AxiosAdapter = async (config) => {
      throw new AxiosError("clinic unavailable", AxiosError.ERR_BAD_REQUEST, config, undefined, {
        config,
        data: { error_code: CLINIC_SELECTION_UNAVAILABLE },
        headers: new AxiosHeaders(),
        status: 403,
        statusText: "Forbidden",
      });
    };

    await expect(
      axios.request({
        adapter,
        method: "get",
        url: "/v1/me",
        startupSessionRestore: true,
      }),
    ).rejects.toThrow("clinic unavailable");
    expect(recoverStarted).toBe(false);
    expect(recoveryMocks.recover).not.toHaveBeenCalled();
  });

  it("still recovers ordinary GET clinic_selection_unavailable responses", async () => {
    attachClinicSelectionInterceptors(axios);
    const adapter: AxiosAdapter = async (config) => {
      throw new AxiosError("clinic unavailable", AxiosError.ERR_BAD_REQUEST, config, undefined, {
        config,
        data: { error_code: CLINIC_SELECTION_UNAVAILABLE },
        headers: new AxiosHeaders(),
        status: 403,
        statusText: "Forbidden",
      });
    };

    await expect(axios.request({ adapter, method: "get", url: "/v1/me" })).rejects.toThrow(
      "clinic unavailable",
    );
    expect(recoveryMocks.recover).toHaveBeenCalledOnce();
  });
});
