import {
  AxiosError,
  AxiosHeaders,
  type AxiosAdapter,
} from "axios";
import { afterAll, afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { axios } from "./axios";

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
  throw new AxiosError(
    "request unauthorized",
    AxiosError.ERR_BAD_REQUEST,
    config,
    undefined,
    {
      config,
      data: { message: "unauthorized" },
      headers: new AxiosHeaders(),
      status: 401,
      statusText: "Unauthorized",
    },
  );
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

  it("keeps the existing login behavior without a forced redirect", async () => {
    const refreshSpy = vi
      .spyOn(axios, "post")
      .mockRejectedValue(new AxiosError("refresh unauthorized"));
    const initialHref = window.location.href;

    await expect(
      axios.request({ adapter: unauthorizedAdapter, method: "post", url: "/v1/login" }),
    ).rejects.toThrow("refresh unauthorized");

    expect(refreshSpy).toHaveBeenCalledWith("/v1/auth/refresh");
    expect(window.location.href).toBe(initialHref);
  });

  it("redirects a protected route 401 to login with a safe from value", async () => {
    setWindowLocation("/owners/300588", "?tab=summary");
    vi.spyOn(axios, "post").mockRejectedValue(
      new AxiosError("refresh unauthorized"),
    );

    await expect(
      axios.request({ adapter: unauthorizedAdapter, method: "get", url: "/v1/owners/300588" }),
    ).rejects.toThrow("refresh unauthorized");

    expect(window.location.href).toBe(
      "/login?from=%2Fowners%2F300588%3Ftab%3Dsummary",
    );
  });

  it.each(["//evil.example", "/\\evil.example", "/%5Cevil.example"])(
    "prevents an unsafe redirect candidate %s in the from value",
    async (pathname) => {
      setWindowLocation(pathname);
      vi.spyOn(axios, "post").mockRejectedValue(
        new AxiosError("refresh unauthorized"),
      );

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
