import { AxiosError, AxiosHeaders, type AxiosAdapter } from "axios";
import { afterAll, afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "./axios";
import { CURRENT_CLINIC_STORAGE_KEY } from "./current-clinic";

const originalLocation = window.location;

function setWindowLocation(pathname: string, search = ""): void {
  Object.defineProperty(window, "location", {
    configurable: true,
    writable: true,
    value: {
      ...originalLocation,
      href: `http://localhost${pathname}${search}`,
      pathname,
      search,
    },
  });
}

const unauthorizedAdapter: AxiosAdapter = async (config) => {
  throw new AxiosError("request unauthorized", AxiosError.ERR_BAD_REQUEST, config, undefined, {
    config,
    data: { message: "unauthorized" },
    headers: new AxiosHeaders(),
    status: 401,
    statusText: "Unauthorized",
  });
};

describe("axios 401 route policy", () => {
  beforeEach(() => {
    setWindowLocation("/login");
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  afterAll(() => {
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: originalLocation,
    });
  });

  it.each([
    ["/forgot-password", "", "/v1/auth/forgot-password"],
    ["/forgot-password/", "", "/v1/auth/forgot-password"],
    ["/reset-password", "?token=test-token", "/v1/auth/reset-password"],
    ["/reset-password/", "?token=test-token", "/v1/auth/reset-password"],
  ])(
    "passes through a 401 on public route %s without refresh or login redirect",
    async (pathname, search, requestUrl) => {
      setWindowLocation(pathname, search);
      const refreshSpy = vi
        .spyOn(axios, "post")
        .mockRejectedValue(new AxiosError("refresh unauthorized"));
      const initialHref = window.location.href;

      await expect(
        axios.request({ adapter: unauthorizedAdapter, method: "post", url: requestUrl }),
      ).rejects.toThrow("request unauthorized");

      expect(refreshSpy).not.toHaveBeenCalled();
      expect(window.location.href).toBe(initialHref);
    },
  );

  it("passes through a login 401 without session refresh or forced redirect", async () => {
    const refreshSpy = vi
      .spyOn(axios, "post")
      .mockRejectedValue(new AxiosError("refresh unauthorized"));
    const initialHref = window.location.href;

    await expect(
      axios.request({ adapter: unauthorizedAdapter, method: "post", url: "/v1/login" }),
    ).rejects.toThrow("request unauthorized");

    expect(refreshSpy).not.toHaveBeenCalled();
    expect(window.location.href).toBe(initialHref);
  });

  it("redirects a protected route 401 to login with a safe from value", async () => {
    setWindowLocation("/owners/300588", "?tab=summary");
    vi.spyOn(axios, "post").mockRejectedValue(new AxiosError("refresh unauthorized"));

    await expect(
      axios.request({ adapter: unauthorizedAdapter, method: "get", url: "/v1/owners/300588" }),
    ).rejects.toThrow("refresh unauthorized");

    expect(window.location.href).toBe("/login?from=%2Fowners%2F300588%3Ftab%3Dsummary");
  });

  it.each(["//evil.example", "/\\evil.example", "/%5Cevil.example"])(
    "prevents an unsafe redirect candidate %s in the from value",
    async (pathname) => {
      setWindowLocation(pathname);
      vi.spyOn(axios, "post").mockRejectedValue(new AxiosError("refresh unauthorized"));

      await expect(
        axios.request({ adapter: unauthorizedAdapter, method: "get", url: "/v1/private" }),
      ).rejects.toThrow("refresh unauthorized");

      expect(window.location.href).toBe("/login?from=%2F");
    },
  );

  it("does not redirect when navigation reaches password recovery during an in-flight refresh", async () => {
    setWindowLocation("/owners/300588");
    let rejectRefresh: ((reason: AxiosError) => void) | undefined;
    const refreshPromise = new Promise<never>((_resolve, reject) => {
      rejectRefresh = reject;
    });
    const refreshSpy = vi.spyOn(axios, "post").mockReturnValue(refreshPromise);

    const request = axios.request({
      adapter: unauthorizedAdapter,
      method: "get",
      url: "/v1/owners/300588",
    });
    const rejection = expect(request).rejects.toThrow("refresh unauthorized");
    await vi.waitFor(() => expect(refreshSpy).toHaveBeenCalledOnce());

    setWindowLocation("/reset-password", "?token=test-token");
    const recoveryHref = window.location.href;
    rejectRefresh?.(new AxiosError("refresh unauthorized"));

    await rejection;
    expect(window.location.href).toBe(recoveryHref);
  });
});

describe("axios startupSessionRestore opt-out", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: originalLocation,
    });
  });

  it("skips GET auto-retry on network errors when startupSessionRestore is true", async () => {
    let adapterCalls = 0;
    const adapter: AxiosAdapter = async (config) => {
      adapterCalls += 1;
      throw new AxiosError("Network Error", AxiosError.ERR_NETWORK, config);
    };

    await expect(
      axios.get("/v1/me", { adapter, startupSessionRestore: true }),
    ).rejects.toMatchObject({ code: AxiosError.ERR_NETWORK });
    expect(adapterCalls).toBe(1);
  });

  it("skips GET auto-retry on 502-504 when startupSessionRestore is true", async () => {
    let adapterCalls = 0;
    const adapter: AxiosAdapter = async (config) => {
      adapterCalls += 1;
      throw new AxiosError("bad gateway", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
        config,
        data: { error: "bad gateway" },
        headers: new AxiosHeaders(),
        status: 502,
        statusText: "Bad Gateway",
      });
    };

    await expect(
      axios.get("/v1/me", { adapter, startupSessionRestore: true }),
    ).rejects.toMatchObject({ message: "bad gateway" });
    expect(adapterCalls).toBe(1);
  });

  it("does not refresh or redirect on 401 when startupSessionRestore is true", async () => {
    setWindowLocation("/owners/300588");
    const refreshSpy = vi
      .spyOn(axios, "post")
      .mockRejectedValue(new AxiosError("refresh unauthorized"));
    const initialHref = window.location.href;

    await expect(
      axios.get("/v1/me", { adapter: unauthorizedAdapter, startupSessionRestore: true }),
    ).rejects.toThrow("request unauthorized");

    expect(refreshSpy).not.toHaveBeenCalled();
    expect(window.location.href).toBe(initialHref);
  });
});

describe("axios GET auto-retry", () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it.each([502, 504])("retries a GET once after a %i then succeeds", async (status) => {
    vi.useFakeTimers();
    let adapterCalls = 0;
    const adapter: AxiosAdapter = async (config) => {
      adapterCalls += 1;
      if (adapterCalls === 1) {
        throw new AxiosError("server error", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
          config,
          data: { error: "server error" },
          headers: new AxiosHeaders(),
          status,
          statusText: "",
        });
      }
      return {
        config,
        data: { ok: true },
        headers: new AxiosHeaders(),
        status: 200,
        statusText: "OK",
      };
    };

    const request = axios.get("/v1/example", { adapter });
    await vi.advanceTimersByTimeAsync(1000);
    const response = await request;

    expect(response.status).toBe(200);
    expect(adapterCalls).toBe(2);
  });

  it("does not retry a GET on 503 and propagates the error immediately", async () => {
    vi.useFakeTimers();
    let adapterCalls = 0;
    const adapter: AxiosAdapter = async (config) => {
      adapterCalls += 1;
      throw new AxiosError("service unavailable", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
        config,
        data: { error: "service_unavailable" },
        headers: new AxiosHeaders(),
        status: 503,
        statusText: "Service Unavailable",
      });
    };

    const request = axios.get("/v1/example", { adapter });
    const rejection = expect(request).rejects.toMatchObject({
      message: "service unavailable",
      response: { status: 503 },
    });
    // リトライウィンドウ全体 (1s + 2s) を十分に超える時刻まで進めても再試行されないこと
    await vi.advanceTimersByTimeAsync(10000);
    await rejection;
    expect(adapterCalls).toBe(1);
  });

  it("retries a GET on network error up to MAX_RETRIES with per-attempt backoff", async () => {
    vi.useFakeTimers();
    let adapterCalls = 0;
    const adapter: AxiosAdapter = async (config) => {
      adapterCalls += 1;
      throw new AxiosError("Network Error", AxiosError.ERR_NETWORK, config);
    };

    const request = axios.get("/v1/example", { adapter });
    const rejection = expect(request).rejects.toMatchObject({
      code: AxiosError.ERR_NETWORK,
    });

    // 初回リトライは RETRY_DELAY_MS (1s) 待機
    await vi.advanceTimersByTimeAsync(999);
    expect(adapterCalls).toBe(1);
    await vi.advanceTimersByTimeAsync(2);
    expect(adapterCalls).toBe(2);
    // 2回目は RETRY_DELAY_MS * 2 (2s) 待機 — 1s ではまだ発火しない
    await vi.advanceTimersByTimeAsync(1998);
    expect(adapterCalls).toBe(2);
    await vi.advanceTimersByTimeAsync(10);
    expect(adapterCalls).toBe(3);

    await rejection; // MAX_RETRIES 到達後はエラーをそのまま伝播
  });

  it("never retries a non-GET request on 502", async () => {
    vi.useFakeTimers();
    let adapterCalls = 0;
    const adapter: AxiosAdapter = async (config) => {
      adapterCalls += 1;
      throw new AxiosError("bad gateway", AxiosError.ERR_BAD_RESPONSE, config, undefined, {
        config,
        data: { error: "bad gateway" },
        headers: new AxiosHeaders(),
        status: 502,
        statusText: "Bad Gateway",
      });
    };

    const request = axios.request({ adapter, method: "post", url: "/v1/example", data: {} });
    const rejection = expect(request).rejects.toMatchObject({
      response: { status: 502 },
    });
    await vi.advanceTimersByTimeAsync(10000);
    await rejection;
    expect(adapterCalls).toBe(1);
  });
});

describe("axios clinic boundary", () => {
  afterEach(() => {
    localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
  });

  it("reloadなしの医院切替後も次requestで最新のX-Clinic-IDを読む", async () => {
    const receivedClinicIds: Array<string | undefined> = [];
    const adapter: AxiosAdapter = async (config) => {
      receivedClinicIds.push(config.headers.get("X-Clinic-ID")?.toString());
      return {
        config,
        data: {},
        headers: new AxiosHeaders(),
        status: 200,
        statusText: "OK",
      };
    };

    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "clinic-1");
    await axios.request({ adapter, method: "get", url: "/v1/example" });
    localStorage.setItem(CURRENT_CLINIC_STORAGE_KEY, "clinic-2");
    await axios.request({ adapter, method: "get", url: "/v1/example" });

    expect(receivedClinicIds).toEqual(["clinic-1", "clinic-2"]);
  });
});
