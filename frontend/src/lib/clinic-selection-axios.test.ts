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
