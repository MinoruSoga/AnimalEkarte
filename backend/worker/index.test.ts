import { afterEach, describe, expect, it, vi } from "vitest";

import worker, {
  buildCorsPreflightResponse,
  buildProxyObservation,
  forwardContainerFetch,
} from "./index";

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

// CORS preflight contract mirrors backend/internal/middleware/cors.go —
// keep these expected values byte-identical with that middleware.
const EDGE_CORS_ALLOW_METHODS = "GET, POST, PUT, PATCH, DELETE, OPTIONS";
const EDGE_CORS_ALLOW_HEADERS =
  "Content-Type, Authorization, X-Request-ID, X-Clinic-ID, X-Requested-With, Idempotency-Key";
const EDGE_CORS_TEST_ORIGINS =
  "https://stg.example.test, https://admin.example.test";

function optionsRequest(origin?: string): Request {
  return new Request("https://api.example.test/api/v1/pets", {
    method: "OPTIONS",
    headers: origin === undefined ? {} : { Origin: origin },
  });
}

describe("buildCorsPreflightResponse", () => {
  it("returns 204 with the full cors.go header contract for an allowed origin", async () => {
    const response = buildCorsPreflightResponse(
      optionsRequest("https://stg.example.test"),
      EDGE_CORS_TEST_ORIGINS,
    );

    expect(response.status).toBe(204);
    expect(await response.text()).toBe("");
    expect(response.headers.get("Access-Control-Allow-Origin")).toBe(
      "https://stg.example.test",
    );
    expect(response.headers.get("Timing-Allow-Origin")).toBe(
      "https://stg.example.test",
    );
    expect(response.headers.get("Vary")).toBe("Origin");
    expect(response.headers.get("Access-Control-Allow-Credentials")).toBe(
      "true",
    );
    expect(response.headers.get("Access-Control-Allow-Methods")).toBe(
      EDGE_CORS_ALLOW_METHODS,
    );
    expect(response.headers.get("Access-Control-Allow-Headers")).toBe(
      EDGE_CORS_ALLOW_HEADERS,
    );
    expect(response.headers.get("Access-Control-Max-Age")).toBe("86400");
  });

  it("omits origin-scoped headers for a disallowed origin but keeps fixed fields", () => {
    const response = buildCorsPreflightResponse(
      optionsRequest("https://evil.example.test"),
      EDGE_CORS_TEST_ORIGINS,
    );

    expect(response.status).toBe(204);
    expect(response.headers.get("Access-Control-Allow-Origin")).toBeNull();
    expect(response.headers.get("Timing-Allow-Origin")).toBeNull();
    expect(response.headers.get("Vary")).toBeNull();
    expect(response.headers.get("Access-Control-Allow-Credentials")).toBe(
      "true",
    );
    expect(response.headers.get("Access-Control-Allow-Methods")).toBe(
      EDGE_CORS_ALLOW_METHODS,
    );
    expect(response.headers.get("Access-Control-Allow-Headers")).toBe(
      EDGE_CORS_ALLOW_HEADERS,
    );
    expect(response.headers.get("Access-Control-Max-Age")).toBe("86400");
  });

  it("omits origin-scoped headers when the Origin header is absent", () => {
    const response = buildCorsPreflightResponse(
      optionsRequest(),
      EDGE_CORS_TEST_ORIGINS,
    );

    expect(response.status).toBe(204);
    expect(response.headers.get("Access-Control-Allow-Origin")).toBeNull();
    expect(response.headers.get("Timing-Allow-Origin")).toBeNull();
    expect(response.headers.get("Vary")).toBeNull();
  });

  it("matches a trimmed entry in the comma-separated allowlist", () => {
    const response = buildCorsPreflightResponse(
      optionsRequest("https://admin.example.test"),
      EDGE_CORS_TEST_ORIGINS,
    );

    expect(response.headers.get("Access-Control-Allow-Origin")).toBe(
      "https://admin.example.test",
    );
  });

  it("falls back to the cors.go development defaults when the allowlist is empty", () => {
    const response = buildCorsPreflightResponse(
      optionsRequest("http://localhost:3000"),
      "",
    );

    expect(response.headers.get("Access-Control-Allow-Origin")).toBe(
      "http://localhost:3000",
    );
  });
});

describe("worker fetch OPTIONS edge responder", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  function fakeEnv(allowedOrigin = "https://stg.example.test") {
    const containerFetch = vi.fn(
      async (_request: Request) =>
        new Response(JSON.stringify({ ok: true }), { status: 200 }),
    );
    const binding = {
      idFromName: vi.fn((name: string) => `id:${name}`),
      get: vi.fn(() => ({ fetch: containerFetch })),
    };
    const env = {
      API_CONTAINER: binding,
      CORS_ALLOWED_ORIGIN: allowedOrigin,
    } as unknown as Env;
    return { env, binding, containerFetch };
  }

  it("answers OPTIONS at the edge without touching the container binding", async () => {
    const { env, binding, containerFetch } = fakeEnv();

    const response = await worker.fetch(
      optionsRequest("https://stg.example.test"),
      env,
    );

    expect(response.status).toBe(204);
    expect(response.headers.get("Access-Control-Allow-Origin")).toBe(
      "https://stg.example.test",
    );
    expect(binding.idFromName).not.toHaveBeenCalled();
    expect(binding.get).not.toHaveBeenCalled();
    expect(containerFetch).not.toHaveBeenCalled();
  });

  it("returns a preflight without Allow-Origin for a disallowed origin", async () => {
    const { env, containerFetch } = fakeEnv();

    const response = await worker.fetch(
      optionsRequest("https://evil.example.test"),
      env,
    );

    expect(response.status).toBe(204);
    expect(response.headers.get("Access-Control-Allow-Origin")).toBeNull();
    expect(containerFetch).not.toHaveBeenCalled();
  });

  it("keeps /_internal guards ahead of the OPTIONS responder", async () => {
    const { env, containerFetch } = fakeEnv();

    const migrate = await worker.fetch(
      new Request("https://api.example.test/_internal/migrate", {
        method: "OPTIONS",
        headers: { Origin: "https://stg.example.test" },
      }),
      env,
    );
    expect(migrate.status).toBe(405);
    expect(migrate.headers.get("Allow")).toBe("POST");

    const internal = await worker.fetch(
      new Request("https://api.example.test/_internal/jobs/daily", {
        method: "OPTIONS",
        headers: { Origin: "https://stg.example.test" },
      }),
      env,
    );
    expect(internal.status).toBe(404);

    expect(containerFetch).not.toHaveBeenCalled();
  });

  it("still proxies non-OPTIONS requests to the container with X-Forwarded-For", async () => {
    const { env, binding, containerFetch } = fakeEnv();
    const infoSpy = vi.spyOn(console, "info").mockImplementation(() => {});

    const response = await worker.fetch(
      new Request("https://api.example.test/api/v1/pets", {
        method: "GET",
        headers: { "CF-Connecting-IP": "203.0.113.10" },
      }),
      env,
    );

    expect(response.status).toBe(200);
    expect(binding.get).toHaveBeenCalledTimes(1);
    expect(containerFetch).toHaveBeenCalledTimes(1);
    const forwarded = containerFetch.mock.calls[0]?.[0];
    expect(forwarded.headers.get("X-Forwarded-For")).toBe("203.0.113.10");
    expect(infoSpy).toHaveBeenCalledWith(
      "container fetch timing",
      expect.objectContaining({ method: "get", status: 200 }),
    );
  });
});

describe("worker fetch /_internal/migrate", () => {
  // Dummy fixture only — not a real credential. Must stay >= 32 UTF-8 bytes.
  const MIGRATE_SECRET = "t".repeat(48);

  function migrateEnv() {
    const runMigrate = vi.fn(
      async (_expected?: string | null) => ({
        exitCode: 0,
        stdout: "",
        stderr: "",
      }),
    );
    const binding = {
      idFromName: vi.fn((name: string) => `id:${name}`),
      get: vi.fn(() => ({ runMigrate })),
    };
    const env = {
      API_CONTAINER: binding,
      MIGRATE_RUN_SECRET: MIGRATE_SECRET,
    } as unknown as Env;
    return { env, binding, runMigrate };
  }

  function migrateRequest(headers: Record<string, string>): Request {
    return new Request("https://api.example.test/_internal/migrate", {
      method: "POST",
      headers: { Authorization: `Bearer ${MIGRATE_SECRET}`, ...headers },
    });
  }

  it("routes to the dedicated runner instance and forwards the expected migration", async () => {
    const { env, binding, runMigrate } = migrateEnv();

    const response = await worker.fetch(
      migrateRequest({
        "X-Expected-Migration": "006_accounts_rls_ops_bypass.sql",
      }),
      env,
    );

    expect(response.status).toBe(200);
    expect(binding.idFromName).toHaveBeenCalledWith(
      "animalekarte-migrate-runner-v1",
    );
    expect(runMigrate).toHaveBeenCalledWith("006_accounts_rls_ops_bypass.sql");
  });

  it("passes null when the expected-migration header is absent", async () => {
    const { env, runMigrate } = migrateEnv();

    const response = await worker.fetch(migrateRequest({}), env);

    expect(response.status).toBe(200);
    expect(runMigrate).toHaveBeenCalledWith(null);
  });

  it("rejects a malformed expected migration without touching the runner", async () => {
    const { env, binding, runMigrate } = migrateEnv();

    const response = await worker.fetch(
      migrateRequest({ "X-Expected-Migration": "../evil.sql" }),
      env,
    );

    expect(response.status).toBe(400);
    expect(binding.get).not.toHaveBeenCalled();
    expect(runMigrate).not.toHaveBeenCalled();
  });
});
