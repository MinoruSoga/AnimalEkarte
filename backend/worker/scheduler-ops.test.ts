import { describe, expect, it, vi } from "vitest";
import {
  SCHEDULER_OPS_PRINCIPAL,
  authenticateSchedulerOpsRequest,
  cronForManualSlot,
  handleSchedulerOpsRequest,
  isInternalProxyPath,
  isAuthorizedSchedulerOpsRequest,
  type SchedulerOpsBinding,
} from "./scheduler-ops";
import type {
  SchedulerControlOperation,
  SchedulerManualOperation,
  SchedulerStatus,
} from "./scheduler-coordinator";
import { SCHEDULER_NAME, type ScheduledJobName } from "./scheduled-jobs";

const SECRET = "scheduler-ops-test-secret-32-bytes-minimum";
const NOW = Date.UTC(2026, 6, 24, 18, 0, 0);
const AUTH_CONFIG = {
  automationSecret: SECRET,
  accessTeamDomain: "animalekarte.cloudflareaccess.com",
  accessAudience: "scheduler-access-audience",
} as const;

// The production suite runs under workerd, which provides timingSafeEqual.
// Keep this focused test runnable in the Dockerized Node fallback as well; the
// Cloudflare primitive itself remains covered by migrate-exec.test.ts.
if (typeof crypto.subtle.timingSafeEqual !== "function") {
  Object.defineProperty(crypto.subtle, "timingSafeEqual", {
    configurable: true,
    value(left: BufferSource, right: BufferSource): boolean {
      const leftBytes = new Uint8Array(
        left instanceof ArrayBuffer ? left : left.buffer,
        left instanceof ArrayBuffer ? 0 : left.byteOffset,
        left instanceof ArrayBuffer ? left.byteLength : left.byteLength,
      );
      const rightBytes = new Uint8Array(
        right instanceof ArrayBuffer ? right : right.buffer,
        right instanceof ArrayBuffer ? 0 : right.byteOffset,
        right instanceof ArrayBuffer ? right.byteLength : right.byteLength,
      );
      if (leftBytes.byteLength !== rightBytes.byteLength) {
        return false;
      }
      let difference = 0;
      for (let index = 0; index < leftBytes.byteLength; index += 1) {
        difference |= leftBytes[index] ^ rightBytes[index];
      }
      return difference === 0;
    },
  });
}

function authorizedRequest(
  path: string,
  init: RequestInit = {},
): Request {
  const headers = new Headers(init.headers);
  headers.set("Authorization", `Bearer ${SECRET}`);
  return new Request(`https://api.example.com${path}`, { ...init, headers });
}

function encodeBase64URL(value: Uint8Array | string): string {
  const bytes =
    typeof value === "string" ? new TextEncoder().encode(value) : value;
  let binary = "";
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary)
    .replaceAll("+", "-")
    .replaceAll("/", "_")
    .replace(/=+$/u, "");
}

async function createAccessSigner(keyID = "access-test-key") {
  const keys = (await crypto.subtle.generateKey(
    {
      name: "RSASSA-PKCS1-v1_5",
      modulusLength: 2048,
      publicExponent: new Uint8Array([1, 0, 1]),
      hash: "SHA-256",
    },
    true,
    ["sign", "verify"],
  )) as CryptoKeyPair;
  const publicKey = await crypto.subtle.exportKey("jwk", keys.publicKey);
  const jwk = { ...publicKey, alg: "RS256", kid: keyID, use: "sig" };

  return {
    jwk,
    async sign(
      overrides: Partial<{
        iss: string;
        aud: string | readonly string[];
        sub: string;
        exp: number;
        nbf: number;
      }> = {},
    ): Promise<string> {
      const header = encodeBase64URL(
        JSON.stringify({ alg: "RS256", kid: keyID, typ: "JWT" }),
      );
      const payload = encodeBase64URL(
        JSON.stringify({
          iss: "https://animalekarte.cloudflareaccess.com",
          aud: "scheduler-access-audience",
          sub: "operator-123",
          exp: Math.floor(NOW / 1_000) + 300,
          nbf: Math.floor(NOW / 1_000) - 30,
          ...overrides,
        }),
      );
      const signature = await crypto.subtle.sign(
        "RSASSA-PKCS1-v1_5",
        keys.privateKey,
        new TextEncoder().encode(`${header}.${payload}`),
      );
      return `${header}.${payload}.${encodeBase64URL(
        new Uint8Array(signature),
      )}`;
    },
  };
}

function statusFixture(): SchedulerStatus {
  return {
    version: 1,
    scheduler: SCHEDULER_NAME,
    control: { version: 1, revision: 3, paused: false, changedAt: NOW },
    recentRuns: [],
    recentOperations: [],
  };
}

function controlFixture(): SchedulerControlOperation {
  return {
    version: 1,
    kind: "control",
    requestId: "d8dd6534-ef3a-48fc-8a8f-3f9a8d9cd068",
    actorPrincipal: SCHEDULER_OPS_PRINCIPAL,
    reason: "INC-123 scheduled maintenance",
    requestedAt: NOW,
    status: "completed",
    expectedRevision: 3,
    requestedPaused: true,
    control: { version: 1, revision: 4, paused: true, changedAt: NOW },
  };
}

function manualFixture(job: ScheduledJobName = "no_show"): SchedulerManualOperation {
  return {
    version: 1,
    kind: "manual_run",
    requestId: "85679370-8dc1-46f1-93ec-675ead9f0f49",
    actorPrincipal: SCHEDULER_OPS_PRINCIPAL,
    reason: "INC-124 recover missing cron slot",
    requestedAt: NOW,
    status: "completed",
    mode: "catch_up",
    job,
    cron: job === "dormant" ? "0 17 * * *" : "0 1 * * *",
    scheduledTime: Date.UTC(2026, 6, 23, 1, 0, 0),
    result: { disposition: "executed" },
  };
}

function fakeBinding(): SchedulerOpsBinding & {
  statusCalls: number[];
  controlCalls: unknown[];
  manualCalls: unknown[];
  rateLimitCalls: Array<{ actorPrincipal: string; now: number }>;
} {
  const statusCalls: number[] = [];
  const controlCalls: unknown[] = [];
  const manualCalls: unknown[] = [];
  const rateLimitCalls: Array<{ actorPrincipal: string; now: number }> = [];
  return {
    statusCalls,
    controlCalls,
    manualCalls,
    rateLimitCalls,
    async consumeScheduledJobsOpsRateLimit(actorPrincipal, now) {
      rateLimitCalls.push({ actorPrincipal, now });
      return { allowed: true, retryAfterSeconds: 60 };
    },
    async getScheduledJobsStatus(limit) {
      statusCalls.push(limit);
      return statusFixture();
    },
    async setScheduledJobsControl(command) {
      controlCalls.push(command);
      return controlFixture();
    },
    async runScheduledJobManually(command) {
      manualCalls.push(command);
      return manualFixture(command.job);
    },
  };
}

describe("scheduler ops authentication", () => {
  it("fails closed when the dedicated secret is missing", () => {
    const request = authorizedRequest("/_internal/scheduler/status");
    expect(isAuthorizedSchedulerOpsRequest(request, undefined)).toBe(false);
    expect(isAuthorizedSchedulerOpsRequest(request, "")).toBe(false);
  });

  it("fails closed when the dedicated secret is shorter than 32 UTF-8 bytes", () => {
    const shortASCIISecret = "a".repeat(31);
    const shortUnicodeSecret = "é".repeat(15);

    expect(
      isAuthorizedSchedulerOpsRequest(
        new Request("https://api.example.com/_internal/scheduler/status", {
          headers: { Authorization: `Bearer ${shortASCIISecret}` },
        }),
        shortASCIISecret,
      ),
    ).toBe(false);
    expect(
      isAuthorizedSchedulerOpsRequest(
        new Request("https://api.example.com/_internal/scheduler/status", {
          headers: { Authorization: `Bearer ${shortUnicodeSecret}` },
        }),
        shortUnicodeSecret,
      ),
    ).toBe(false);
  });

  it("accepts only an exact Bearer credential", () => {
    const minimumLengthSecret = "m".repeat(32);
    expect(
      isAuthorizedSchedulerOpsRequest(
        new Request("https://api.example.com/_internal/scheduler/status", {
          headers: { Authorization: `Bearer ${minimumLengthSecret}` },
        }),
        minimumLengthSecret,
      ),
    ).toBe(true);
    expect(
      isAuthorizedSchedulerOpsRequest(
        authorizedRequest("/_internal/scheduler/status"),
        SECRET,
      ),
    ).toBe(true);
    expect(
      isAuthorizedSchedulerOpsRequest(
        new Request("https://api.example.com/_internal/scheduler/status", {
          headers: { Authorization: "Bearer wrong-secret-value" },
        }),
        SECRET,
      ),
    ).toBe(false);
  });

  it("validates a Cloudflare Access JWT and derives a signed per-actor principal", async () => {
    const signer = await createAccessSigner();
    const assertion = await signer.sign();
    const certsFetch = vi.fn(async (request: Request) => {
      expect(request.url).toBe(
        "https://animalekarte.cloudflareaccess.com/cdn-cgi/access/certs",
      );
      expect(request.redirect).toBe("manual");
      return Response.json({ keys: [signer.jwk] });
    });
    const request = new Request(
      "https://api.example.com/_internal/scheduler/status",
      { headers: { "CF-Access-Jwt-Assertion": assertion } },
    );

    await expect(
      authenticateSchedulerOpsRequest(
        request,
        AUTH_CONFIG,
        NOW,
        certsFetch,
      ),
    ).resolves.toEqual({
      actorPrincipal: "cloudflare-access:operator-123",
    });
    expect(certsFetch).toHaveBeenCalledOnce();
  });

  it("bounds forged-kid JWKS refreshes while still accepting a rotated key after cooldown", async () => {
    const currentSigner = await createAccessSigner("current-access-key");
    const rotatedSigner = await createAccessSigner("rotated-access-key");
    const currentAssertion = await currentSigner.sign();
    const rotatedAssertion = await rotatedSigner.sign();
    const certsFetch = vi.fn(async () =>
      Response.json({
        keys:
          certsFetch.mock.calls.length === 1
            ? [currentSigner.jwk]
            : [currentSigner.jwk, rotatedSigner.jwk],
      }),
    );

    await expect(
      authenticateSchedulerOpsRequest(
        new Request("https://api.example.com/_internal/scheduler/status", {
          headers: { "CF-Access-Jwt-Assertion": currentAssertion },
        }),
        AUTH_CONFIG,
        NOW,
        certsFetch,
      ),
    ).resolves.toEqual({
      actorPrincipal: "cloudflare-access:operator-123",
    });

    for (let attempt = 0; attempt < 5; attempt += 1) {
      await expect(
        authenticateSchedulerOpsRequest(
          new Request("https://api.example.com/_internal/scheduler/status", {
            headers: { "CF-Access-Jwt-Assertion": rotatedAssertion },
          }),
          AUTH_CONFIG,
          NOW + 1_000,
          certsFetch,
        ),
      ).resolves.toBeUndefined();
    }
    expect(certsFetch).toHaveBeenCalledOnce();

    await expect(
      authenticateSchedulerOpsRequest(
        new Request("https://api.example.com/_internal/scheduler/status", {
          headers: { "CF-Access-Jwt-Assertion": rotatedAssertion },
        }),
        AUTH_CONFIG,
        NOW + 60_001,
        certsFetch,
      ),
    ).resolves.toEqual({
      actorPrincipal: "cloudflare-access:operator-123",
    });
    expect(certsFetch).toHaveBeenCalledTimes(2);
  });

  it("coalesces concurrent JWKS loads and cools down upstream failures", async () => {
    const signer = await createAccessSigner("concurrent-access-key");
    const assertion = await signer.sign();
    let releaseFetch: (() => void) | undefined;
    const fetchGate = new Promise<void>((resolve) => {
      releaseFetch = resolve;
    });
    const successfulFetch = vi.fn(async () => {
      await fetchGate;
      return Response.json({ keys: [signer.jwk] });
    });
    const request = () =>
      new Request("https://api.example.com/_internal/scheduler/status", {
        headers: { "CF-Access-Jwt-Assertion": assertion },
      });
    const attempts = Array.from({ length: 5 }, () =>
      authenticateSchedulerOpsRequest(
        request(),
        AUTH_CONFIG,
        NOW,
        successfulFetch,
      ),
    );

    try {
      await vi.waitFor(() => {
        expect(successfulFetch).toHaveBeenCalledOnce();
      });
    } finally {
      releaseFetch?.();
    }
    await expect(Promise.all(attempts)).resolves.toEqual(
      Array.from({ length: 5 }, () => ({
        actorPrincipal: "cloudflare-access:operator-123",
      })),
    );

    const failedFetch = vi.fn(async () => new Response(null, { status: 503 }));
    for (let attempt = 0; attempt < 5; attempt += 1) {
      await expect(
        authenticateSchedulerOpsRequest(
          request(),
          AUTH_CONFIG,
          NOW,
          failedFetch,
        ),
      ).resolves.toBeUndefined();
    }
    expect(failedFetch).toHaveBeenCalledOnce();
    await expect(
      authenticateSchedulerOpsRequest(
        request(),
        AUTH_CONFIG,
        NOW + 60_001,
        failedFetch,
      ),
    ).resolves.toBeUndefined();
    expect(failedFetch).toHaveBeenCalledTimes(2);
  });

  it("rejects spoofed actor headers and invalid Access JWT claims or signatures", async () => {
    const signer = await createAccessSigner();
    const certsFetch = vi.fn(async () =>
      Response.json({ keys: [signer.jwk] }),
    );
    const nowSeconds = Math.floor(NOW / 1_000);
    const assertions = [
      await signer.sign({ iss: "https://attacker.example.com" }),
      await signer.sign({ aud: "wrong-audience" }),
      await signer.sign({ exp: nowSeconds }),
      await signer.sign({ nbf: nowSeconds + 1 }),
    ];
    const valid = await signer.sign();
    const signature = valid.split(".").at(-1);
    if (signature === undefined) {
      throw new Error("expected JWT signature");
    }
    assertions.push(
      `${valid.slice(0, valid.length - signature.length)}${
        signature.startsWith("A") ? "B" : "A"
      }${signature.slice(1)}`,
    );

    for (const assertion of assertions) {
      const request = new Request(
        "https://api.example.com/_internal/scheduler/status",
        {
          headers: {
            "CF-Access-Authenticated-User-Email": "spoofed@example.com",
            "CF-Access-Jwt-Assertion": assertion,
          },
        },
      );
      await expect(
        authenticateSchedulerOpsRequest(
          request,
          AUTH_CONFIG,
          NOW,
          certsFetch,
        ),
      ).resolves.toBeUndefined();
    }
    await expect(
      authenticateSchedulerOpsRequest(
        new Request("https://api.example.com/_internal/scheduler/status", {
          headers: {
            "CF-Access-Authenticated-User-Email": "spoofed@example.com",
          },
        }),
        AUTH_CONFIG,
        NOW,
        certsFetch,
      ),
    ).resolves.toBeUndefined();
  });

  it("rejects redirect indicators or oversized Access JWKS responses", async () => {
    const signer = await createAccessSigner();
    const assertion = await signer.sign();
    const request = new Request(
      "https://api.example.com/_internal/scheduler/status",
      { headers: { "CF-Access-Jwt-Assertion": assertion } },
    );
    const redirected = new Response(null, { status: 200 });
    Object.defineProperty(redirected, "redirected", { value: true });

    await expect(
      authenticateSchedulerOpsRequest(
        request,
        AUTH_CONFIG,
        NOW,
        async () => redirected,
      ),
    ).resolves.toBeUndefined();
    await expect(
      authenticateSchedulerOpsRequest(
        request,
        AUTH_CONFIG,
        NOW,
        async (outboundRequest) => {
          expect(outboundRequest.redirect).toBe("manual");
          return new Response(null, {
            status: 302,
            headers: { Location: "https://attacker.example.com/jwks" },
          });
        },
      ),
    ).resolves.toBeUndefined();
    await expect(
      authenticateSchedulerOpsRequest(
        request,
        AUTH_CONFIG,
        NOW,
        async () =>
          Response.json(
            { keys: [signer.jwk] },
            { headers: { Location: "https://attacker.example.com/jwks" } },
          ),
      ),
    ).resolves.toBeUndefined();
    await expect(
      authenticateSchedulerOpsRequest(
        request,
        AUTH_CONFIG,
        NOW,
        async () =>
          new Response("{}", {
            status: 200,
            headers: { "Content-Length": "65537" },
          }),
      ),
    ).resolves.toBeUndefined();
  });

  it("times out while reading a stalled Access JWKS response body", async () => {
    const signer = await createAccessSigner();
    const assertion = await signer.sign();
    const request = new Request(
      "https://api.example.com/_internal/scheduler/status",
      { headers: { "CF-Access-Jwt-Assertion": assertion } },
    );

    vi.useFakeTimers();
    try {
      let cancelled = false;
      const stalledBody = new ReadableStream<Uint8Array>({
        start(controller) {
          setTimeout(() => {
            if (cancelled) {
              return;
            }
            controller.enqueue(
              new TextEncoder().encode(JSON.stringify({ keys: [signer.jwk] })),
            );
            controller.close();
          }, 6_000);
        },
        cancel() {
          cancelled = true;
        },
      });
      const authentication = authenticateSchedulerOpsRequest(
        request,
        AUTH_CONFIG,
        NOW,
        async () => new Response(stalledBody, { status: 200 }),
      );
      await vi.advanceTimersByTimeAsync(6_000);
      await expect(authentication).resolves.toBeUndefined();
    } finally {
      vi.useRealTimers();
    }
  });
});

describe("cronForManualSlot", () => {
  it.each([
    ["delivery", Date.UTC(2026, 6, 23, 1), "0 1 * * *"],
    ["no_show", Date.UTC(2026, 6, 23, 1), "0 1 * * *"],
    ["no_show", Date.UTC(2026, 6, 23, 6), "0 6,11 * * *"],
    ["no_show", Date.UTC(2026, 6, 23, 11), "0 6,11 * * *"],
    ["dormant", Date.UTC(2026, 6, 23, 17), "0 17 * * *"],
  ] as const)("derives %s at %s from the server-owned plan", (job, scheduledTime, cron) => {
    expect(cronForManualSlot(job, scheduledTime, NOW)).toBe(cron);
  });

  it.each([
    ["delivery", Date.UTC(2026, 6, 23, 6)],
    ["dormant", Date.UTC(2026, 6, 23, 17, 1)],
    ["no_show", NOW + 60 * 60 * 1_000],
    ["no_show", NOW - 36 * 24 * 60 * 60 * 1_000],
  ] as const)("rejects unsafe %s slot %s", (job, scheduledTime) => {
    expect(() => cronForManualSlot(job, scheduledTime, NOW)).toThrow();
  });
});

describe("isInternalProxyPath", () => {
  it.each([
    "/_internal/migrate",
    "/_INTERNAL/scheduler/status",
    "/%5finternal/scheduler/status",
    "/_internal%2fscheduler%2fstatus",
    "/_internal%252fscheduler%252fstatus",
    String.raw`/_internal\scheduler\status`,
    "/public/../_internal/scheduler/status",
    "//_internal//scheduler//status",
    "/_internal/%ZZ",
  ])("blocks canonical and ambiguous internal path %s", (pathname) => {
    expect(isInternalProxyPath(pathname)).toBe(true);
  });

  it.each(["/api/v1/pets", "/health", "/public/internal"])(
    "allows public path %s",
    (pathname) => {
      expect(isInternalProxyPath(pathname)).toBe(false);
    },
  );
});

describe("handleSchedulerOpsRequest", () => {
  it("returns sanitized status with a bounded limit", async () => {
    const binding = fakeBinding();
    const response = await handleSchedulerOpsRequest(
      authorizedRequest("/_internal/scheduler/status?limit=7"),
      AUTH_CONFIG,
      binding,
      NOW,
    );

    expect(response.status).toBe(200);
    expect(binding.rateLimitCalls).toEqual([
      { actorPrincipal: SCHEDULER_OPS_PRINCIPAL, now: NOW },
    ]);
    expect(binding.statusCalls).toEqual([7]);
    expect(await response.json()).toEqual(statusFixture());
  });

  it.each([
    "/_internal/scheduler/status?limit=1&limit=2",
    "/_internal/scheduler/status?cursor=untrusted",
  ])("rejects ambiguous status query %s", async (path) => {
    const binding = fakeBinding();
    const response = await handleSchedulerOpsRequest(
      authorizedRequest(path),
      AUTH_CONFIG,
      binding,
      NOW,
    );

    expect(response.status).toBe(400);
    expect(await response.json()).toEqual({ error: "invalid_request" });
    expect(binding.statusCalls).toEqual([]);
  });

  it("audits the verified Cloudflare Access subject and rate limits that principal", async () => {
    const signer = await createAccessSigner();
    const assertion = await signer.sign({ sub: "operator-456" });
    const binding = fakeBinding();
    const request = new Request(
      "https://api.example.com/_internal/scheduler/control",
      {
        method: "PUT",
        headers: {
          "CF-Access-Jwt-Assertion": assertion,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          paused: true,
          expected_revision: 3,
          request_id: "36bb8148-c36d-476a-850b-ae296abe297a",
          reason: "INC-136 verified operator pause",
        }),
      },
    );

    const response = await handleSchedulerOpsRequest(
      request,
      AUTH_CONFIG,
      binding,
      NOW,
      undefined,
      async () => Response.json({ keys: [signer.jwk] }),
    );

    expect(response.status).toBe(200);
    expect(binding.rateLimitCalls).toEqual([
      { actorPrincipal: "cloudflare-access:operator-456", now: NOW },
    ]);
    expect(binding.controlCalls).toEqual([
      expect.objectContaining({
        actorPrincipal: "cloudflare-access:operator-456",
      }),
    ]);
  });

  it("returns 429 before a coordinator read or mutation when the durable rate limit is exhausted", async () => {
    const binding = fakeBinding();
    binding.consumeScheduledJobsOpsRateLimit = async () => ({
      allowed: false,
      retryAfterSeconds: 17,
    });

    const response = await handleSchedulerOpsRequest(
      authorizedRequest("/_internal/scheduler/status"),
      AUTH_CONFIG,
      binding,
      NOW,
    );

    expect(response.status).toBe(429);
    expect(response.headers.get("Retry-After")).toBe("17");
    expect(binding.statusCalls).toEqual([]);
  });

  it("rejects unauthorized requests before invoking the coordinator", async () => {
    const binding = fakeBinding();
    const response = await handleSchedulerOpsRequest(
      new Request("https://api.example.com/_internal/scheduler/status"),
      AUTH_CONFIG,
      binding,
      NOW,
    );

    expect(response.status).toBe(401);
    expect(binding.rateLimitCalls).toEqual([]);
    expect(binding.statusCalls).toEqual([]);
  });

  it("rejects unknown control fields and non-JSON bodies", async () => {
    const binding = fakeBinding();
    const unknownField = await handleSchedulerOpsRequest(
      authorizedRequest("/_internal/scheduler/control", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          paused: true,
          expected_revision: 3,
          request_id: "d8dd6534-ef3a-48fc-8a8f-3f9a8d9cd068",
          reason: "INC-123 scheduled maintenance",
          actor: "untrusted-user",
        }),
      }),
      AUTH_CONFIG,
      binding,
      NOW,
    );
    const wrongType = await handleSchedulerOpsRequest(
      authorizedRequest("/_internal/scheduler/control", {
        method: "PUT",
        headers: { "Content-Type": "text/plain" },
        body: "{}",
      }),
      AUTH_CONFIG,
      binding,
      NOW,
    );

    expect(unknownField.status).toBe(400);
    expect(wrongType.status).toBe(415);
    expect(binding.controlCalls).toEqual([]);
  });

  it("returns bounded client errors for malformed and oversized JSON", async () => {
    const binding = fakeBinding();
    const malformed = await handleSchedulerOpsRequest(
      authorizedRequest("/_internal/scheduler/control", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: "{",
      }),
      AUTH_CONFIG,
      binding,
      NOW,
    );
    const oversized = await handleSchedulerOpsRequest(
      authorizedRequest("/_internal/scheduler/control", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ reason: "x".repeat(4_097) }),
      }),
      AUTH_CONFIG,
      binding,
      NOW,
    );

    expect(malformed.status).toBe(400);
    expect(await malformed.json()).toEqual({ error: "invalid_json" });
    expect(oversized.status).toBe(413);
    expect(await oversized.json()).toEqual({ error: "request_body_too_large" });
    expect(binding.controlCalls).toEqual([]);
  });

  it("injects the trusted credential principal into a control command", async () => {
    const binding = fakeBinding();
    const response = await handleSchedulerOpsRequest(
      authorizedRequest("/_internal/scheduler/control", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          paused: true,
          expected_revision: 3,
          request_id: "d8dd6534-ef3a-48fc-8a8f-3f9a8d9cd068",
          reason: "INC-123 scheduled maintenance",
        }),
      }),
      AUTH_CONFIG,
      binding,
      NOW,
    );

    expect(response.status).toBe(200);
    expect(binding.controlCalls).toEqual([
      {
        paused: true,
        expectedRevision: 3,
        requestId: "d8dd6534-ef3a-48fc-8a8f-3f9a8d9cd068",
        reason: "INC-123 scheduled maintenance",
        actorPrincipal: SCHEDULER_OPS_PRINCIPAL,
      },
    ]);
  });

  it("derives cron and accepts only catch-up manual runs", async () => {
    const binding = fakeBinding();
    const observedOperations: SchedulerManualOperation[] = [];
    const scheduledTime = Date.UTC(2026, 6, 23, 17);
    const response = await handleSchedulerOpsRequest(
      authorizedRequest("/_internal/scheduler/runs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          job: "dormant",
          scheduled_time_ms: scheduledTime,
          mode: "catch_up",
          request_id: "85679370-8dc1-46f1-93ec-675ead9f0f49",
          reason: "INC-124 recover missing cron slot",
        }),
      }),
      AUTH_CONFIG,
      binding,
      NOW,
      async (operation) => {
        observedOperations.push(operation);
      },
    );

    expect(response.status).toBe(200);
    expect(binding.manualCalls).toEqual([
      {
        actorPrincipal: SCHEDULER_OPS_PRINCIPAL,
        job: "dormant",
        mode: "catch_up",
        reason: "INC-124 recover missing cron slot",
        requestId: "85679370-8dc1-46f1-93ec-675ead9f0f49",
        scheduledTime,
      },
    ]);
    expect(observedOperations).toEqual([manualFixture("dormant")]);
  });

  it("returns a sanitized non-success response for a failed manual run", async () => {
    const binding = fakeBinding();
    binding.runScheduledJobManually = async () => ({
      ...manualFixture("no_show"),
      result: {
        disposition: "executed",
        ledger: {
          version: 1,
          scheduler: SCHEDULER_NAME,
          runId: `${SCHEDULER_NAME}:1784768400000:no_show`,
          runKey: "scheduler:run:internal",
          cron: "0 1 * * *",
          job: "no_show",
          scheduledTime: Date.UTC(2026, 6, 23, 1),
          fenceToken: 7,
          status: "failed",
          startedAt: NOW,
          finishedAt: NOW,
          failureCode: "transport",
        },
      },
    });

    const response = await handleSchedulerOpsRequest(
      authorizedRequest("/_internal/scheduler/runs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          job: "no_show",
          scheduled_time_ms: Date.UTC(2026, 6, 23, 1),
          mode: "catch_up",
          request_id: "045d245a-c7bf-438e-bd12-e2b61d364081",
          reason: "INC-131 inspect failed catch-up",
        }),
      }),
      AUTH_CONFIG,
      binding,
      NOW,
    );

    expect(response.status).toBe(502);
    const body = await response.json();
    expect(body).toMatchObject({
      status: "completed",
      result: {
        disposition: "executed",
        ledger: { status: "failed", failureCode: "transport" },
      },
    });
    expect(JSON.stringify(body)).not.toMatch(/runKey|fenceToken/);
  });
});
