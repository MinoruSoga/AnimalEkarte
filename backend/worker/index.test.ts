import { afterEach, describe, expect, it, vi } from "vitest";

import { buildProxyObservation, forwardContainerFetch } from "./index";

describe("buildProxyObservation", () => {
  it("emits the fixed session success shape without raw request data", () => {
    const observation = buildProxyObservation(
      new Request("https://api.example.test/api/v1/me?owner=private", {
        headers: {
          Authorization: "Bearer secret",
          Cookie: "session=secret",
          "X-Request-ID": "9e4a6f84-8dae-4c30-8b51-a1b3f61b9ca5",
        },
      }),
      10.2,
      22.8,
      { status: 401 },
    );

    expect(observation).toEqual({
      event: "container_fetch_timing",
      method: "get",
      path: "session",
      duration_ms: 13,
      status: 401,
      request_id: "9e4a6f84-8dae-4c30-8b51-a1b3f61b9ca5",
    });
  });

  it("classifies unknown input and omits untrusted correlation IDs on failures", () => {
    const observation = buildProxyObservation(
      new Request("https://api.example.test/private/path?token=secret", {
        method: "POST",
        headers: { "X-Request-ID": "not-a-safe-id" },
      }),
      20,
      10,
      { failureCode: "container_unavailable" },
    );

    expect(observation).toEqual({
      event: "container_fetch_timing",
      method: "other",
      path: "other",
      duration_ms: 0,
      failure_code: "container_unavailable",
    });
  });
});

describe("forwardContainerFetch observation", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  function sensitiveRequest(): Request {
    return new Request(
      "https://api.example.test/api/v1/me?owner=private&token=secret-query",
      {
        method: "GET",
        headers: {
          Authorization: "Bearer secret-token",
          Cookie: "session=secret-cookie",
          "CF-Connecting-IP": "203.0.113.50",
          "X-Request-ID": "9e4a6f84-8dae-4c30-8b51-a1b3f61b9ca5",
        },
      },
    );
  }

  function assertNoSensitiveLogPayload(args: unknown[]): void {
    const serialized = JSON.stringify(args);
    expect(serialized).not.toContain("secret-token");
    expect(serialized).not.toContain("secret-cookie");
    expect(serialized).not.toContain("secret-query");
    expect(serialized).not.toContain("owner=private");
    expect(serialized).not.toContain("203.0.113.50");
    expect(serialized).not.toContain("Authorization");
    expect(serialized).not.toContain("Cookie");
    expect(serialized).not.toContain("boom-exception-body");
  }

  it("forwards success responses and logs allowed observation fields only", async () => {
    const request = sensitiveRequest();
    const forwardedRequest = new Request(request);
    const upstream = new Response(JSON.stringify({ ok: true }), {
      status: 200,
    });
    const container = {
      fetch: vi.fn(async () => upstream),
    };
    const infoSpy = vi.spyOn(console, "info").mockImplementation(() => {});

    const response = await forwardContainerFetch(
      request,
      forwardedRequest,
      container,
    );

    expect(response.status).toBe(200);
    expect(await response.json()).toEqual({ ok: true });
    expect(container.fetch).toHaveBeenCalledTimes(1);
    expect(container.fetch).toHaveBeenCalledWith(forwardedRequest);

    expect(infoSpy).toHaveBeenCalledTimes(1);
    expect(infoSpy.mock.calls[0]?.[0]).toBe("container fetch timing");
    const observation = infoSpy.mock.calls[0]?.[1] as Record<string, unknown>;
    expect(observation).toEqual({
      event: "container_fetch_timing",
      method: "get",
      path: "session",
      duration_ms: expect.any(Number),
      status: 200,
      request_id: "9e4a6f84-8dae-4c30-8b51-a1b3f61b9ca5",
    });
    expect(observation.duration_ms).toBeGreaterThanOrEqual(0);
    expect(Number.isInteger(observation.duration_ms)).toBe(true);
    expect(Object.keys(observation).sort()).toEqual([
      "duration_ms",
      "event",
      "method",
      "path",
      "request_id",
      "status",
    ]);
    assertNoSensitiveLogPayload(infoSpy.mock.calls[0] ?? []);
  });

  it("returns 503 container_unavailable and omits exception bodies from logs", async () => {
    const request = sensitiveRequest();
    const forwardedRequest = new Request(request);
    const container = {
      fetch: vi.fn(async () => {
        throw new Error("boom-exception-body with token=secret-query");
      }),
    };
    const infoSpy = vi.spyOn(console, "info").mockImplementation(() => {});
    const errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});

    const response = await forwardContainerFetch(
      request,
      forwardedRequest,
      container,
    );

    expect(response.status).toBe(503);
    expect(await response.json()).toEqual({ error: "service_unavailable" });
    expect(container.fetch).toHaveBeenCalledTimes(1);

    expect(errorSpy).toHaveBeenCalledTimes(1);
    expect(errorSpy.mock.calls[0]?.[0]).toBe("container fetch failed");
    expect(errorSpy.mock.calls[0]?.[1]).toEqual({
      event: "container_fetch_failed",
      failure_code: "container_unavailable",
    });
    assertNoSensitiveLogPayload(errorSpy.mock.calls[0] ?? []);

    expect(infoSpy).toHaveBeenCalledTimes(1);
    expect(infoSpy.mock.calls[0]?.[0]).toBe("container fetch timing");
    const observation = infoSpy.mock.calls[0]?.[1] as Record<string, unknown>;
    expect(observation).toEqual({
      event: "container_fetch_timing",
      method: "get",
      path: "session",
      duration_ms: expect.any(Number),
      failure_code: "container_unavailable",
      request_id: "9e4a6f84-8dae-4c30-8b51-a1b3f61b9ca5",
    });
    expect(observation.duration_ms).toBeGreaterThanOrEqual(0);
    expect(Object.keys(observation).sort()).toEqual([
      "duration_ms",
      "event",
      "failure_code",
      "method",
      "path",
      "request_id",
    ]);
    assertNoSensitiveLogPayload(infoSpy.mock.calls[0] ?? []);
  });
});
