import { afterEach, describe, expect, it, vi } from "vitest";
import { AnimalEkarteApiContainer } from "./index";
import {
  SchedulerCoordinator,
  type CoordinatorStorage,
  type CoordinatorTransaction,
  type ScheduledRunLedger,
} from "./scheduler-coordinator";
import { enqueueSchedulerIncident, SCHEDULER_ALERT_PENDING_PREFIX } from "./scheduler-alert-outbox";
import { handleSchedulerOpsRequest, notifySchedulerFailures } from "./scheduler-ops";
import { SCHEDULER_NAME } from "./scheduled-jobs";
vi.mock("@cloudflare/containers", () => ({
  Container: class {},
  getContainer: vi.fn(),
}));
vi.mock("cloudflare:workers", () => ({
  env: {
    SCHEDULER_ENVIRONMENT: "staging",
    SCHEDULER_ALERT_WEBHOOK_URL: "https://alerts.example.com/scheduler",
    SCHEDULER_ALERT_WEBHOOK_SECRET: "test-secret",
    SCHEDULER_ALERT_ALLOWED_HOST: "alerts.example.com",
  },
}));
class Storage implements CoordinatorStorage {
  values = new Map<string, unknown>();
  async get<T>(key: string): Promise<T | undefined> {
    return this.values.get(key) as T | undefined;
  }
  async put<T>(key: string, value: T): Promise<void> {
    this.values.set(key, structuredClone(value));
  }
  async delete(key: string): Promise<boolean> {
    return this.values.delete(key);
  }
  async list<T>(options: {
    prefix: string;
    startAfter?: string;
    reverse?: boolean;
    limit?: number;
  }): Promise<Map<string, T>> {
    const entries = [...this.values.entries()]
      .filter(
        ([key]) =>
          key.startsWith(options.prefix) &&
          (options.startAfter === undefined || key > options.startAfter),
      )
      .sort(([a], [b]) => a.localeCompare(b));
    return new Map((options.reverse ? entries.reverse() : entries).slice(0, options.limit)) as Map<
      string,
      T
    >;
  }
  async transaction<T>(closure: (tx: CoordinatorTransaction) => Promise<T>): Promise<T> {
    const draft = new Storage();
    draft.values = new Map(this.values);
    const result = await closure(draft);
    this.values = draft.values;
    return result;
  }
}

const config = {
  environment: "staging",
  webhookURL: "https://alerts.example.com/scheduler",
  webhookSecret: "test-secret",
  allowedHost: "alerts.example.com",
};
const secret = "scheduler-ops-test-secret-32-bytes-minimum";
// Node-only fallback primitive. Actual workerd primitive is outside this regression.
if (typeof crypto.subtle.timingSafeEqual !== "function") {
  Object.defineProperty(crypto.subtle, "timingSafeEqual", {
    configurable: true,
    value: (a: ArrayBuffer, b: ArrayBuffer) => {
      const left = new Uint8Array(a);
      const right = new Uint8Array(b);
      return left.length === right.length && left.every((v, i) => v === right[i]);
    },
  });
}
class InterruptibleStorage extends Storage {
  failPendingList = false;
  override async list<T>(options: {
    prefix: string;
    startAfter?: string;
    reverse?: boolean;
    limit?: number;
  }): Promise<Map<string, T>> {
    if (this.failPendingList && options.prefix === SCHEDULER_ALERT_PENDING_PREFIX) {
      throw new Error("test durable storage interruption");
    }
    return super.list(options);
  }
}

const SLOT = Date.UTC(2026, 6, 24, 17);
const MANUAL_REQUEST_ID = "de3a2410-0b43-4775-9b8b-b4e3de5a47e8";
const STUCK_INCIDENTS = 7; // 7 unresponsive deliveries x 5s timeout = 35s > the 30s CLI budget.

function makeContainer(
  storage: CoordinatorStorage,
  execute: (request: Request) => Promise<Response>,
): { container: AnimalEkarteApiContainer; waits: Array<Promise<unknown>> } {
  const waits: Array<Promise<unknown>> = [];
  const container = Object.create(AnimalEkarteApiContainer.prototype) as AnimalEkarteApiContainer;
  Object.defineProperties(container, {
    ctx: {
      value: {
        storage,
        waitUntil(promise: Promise<unknown>) {
          waits.push(promise);
        },
      },
    },
    envVars: { value: { SCHEDULER_INTERNAL_TOKEN: "test-token" } },
    containerFetch: { value: execute },
  });
  return { container, waits };
}

type WaitOutcome = "fulfilled" | "rejected";

// Explicitly resolves each registered promise's outcome so a rejected
// registration can never hide behind allSettled.
async function flushWaits(waits: Array<Promise<unknown>>): Promise<WaitOutcome[]> {
  const registered = waits.splice(0, waits.length);
  return Promise.all(
    registered.map((promise) =>
      promise.then(
        () => "fulfilled" as const,
        () => "rejected" as const,
      ),
    ),
  );
}

async function enqueuePendingIncidents(
  storage: CoordinatorStorage,
  slot: number,
  count: number,
): Promise<void> {
  await storage.transaction(async (tx) => {
    for (let index = 0; index < count; index += 1) {
      await enqueueSchedulerIncident(tx, {
        version: 1,
        scheduler: SCHEDULER_NAME,
        runId: `stuck-${index}`,
        runKey: `stuck-run-${index}`,
        cron: "0 1 * * *",
        job: "no_show",
        scheduledTime: slot - 86400000,
        fenceToken: 1,
        status: "failed",
        failureCode: "lease_expired",
        startedAt: slot - 86400000,
        finishedAt: slot,
      });
    }
  });
}

function manualRequest(): Request {
  return new Request("https://api.example.com/_internal/scheduler/runs", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${secret}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      job: "dormant",
      mode: "catch_up",
      scheduled_time_ms: SLOT,
      request_id: MANUAL_REQUEST_ID,
      reason: "retry incident",
    }),
  });
}

async function pendingCount(storage: CoordinatorStorage): Promise<number> {
  return (await storage.list({ prefix: SCHEDULER_ALERT_PENDING_PREFIX })).size;
}

function alertKeys(send: { mock: { calls: Array<[Request]> } }): string[] {
  return send.mock.calls.map(([request]) => request.headers.get("Idempotency-Key") ?? "");
}

const auth = { automationSecret: secret };

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  vi.useRealTimers();
});
describe("manual response during old incident delivery failure", () => {
  it.each(["success", "rejected", "partial", "failed", "transport"] as const)(
    "preserves result and same-ID replay for %s",
    async (mode) => {
      const rejected = mode === "rejected";
      const failure = ["partial", "failed", "transport"].includes(mode);
      const now = Date.UTC(2026, 6, 24, 18);
      vi.spyOn(Date, "now").mockReturnValue(now);
      vi.spyOn(console, "error").mockImplementation(() => {});
      const storage = new Storage();
      const slot = Date.UTC(2026, 6, 24, 17);
      const old: ScheduledRunLedger = {
        version: 1,
        scheduler: SCHEDULER_NAME,
        runId: "old-incident",
        runKey: "old-run",
        cron: "0 1 * * *",
        job: "no_show",
        scheduledTime: slot - 86400000,
        fenceToken: 1,
        status: "failed",
        failureCode: "lease_expired",
        startedAt: slot - 86400000,
        finishedAt: slot,
      };
      await storage.transaction((tx) => enqueueSchedulerIncident(tx, old));
      const coordinator = new SchedulerCoordinator(storage, () => now);
      if (rejected)
        await coordinator.setControl({
          actorPrincipal: "scheduler-ops-secret-v1",
          requestId: "110e8400-e29b-41d4-a716-446655440000",
          reason: "test pause",
          expectedRevision: 0,
          paused: true,
        });
      const execute = vi.fn(async () => {
        if (mode === "transport") throw new Error("test transport failure");
        return Response.json(
          mode === "partial"
            ? { outcome: "partial", processed: 2, succeeded: 1, failed: 1 }
            : mode === "failed"
              ? { outcome: "failed", processed: 1, succeeded: 0, failed: 1 }
              : { outcome: "success", processed: 1, succeeded: 1, failed: 0 },
        );
      });
      const { container, waits } = makeContainer(storage, execute);
      const send = vi.fn(async () => new Response(null, { status: 503 }));
      vi.stubGlobal("fetch", send);
      const request = manualRequest;
      const first = await handleSchedulerOpsRequest(request(), auth, container, now);
      expect(first.status).toBe(rejected ? 409 : failure ? 502 : 200);
      const body = await first.json();
      const replay = await handleSchedulerOpsRequest(request(), auth, container, now);
      expect(replay.status).toBe(first.status);
      expect(await replay.json()).toEqual(body);
      // One registered DurableObjectState lifecycle promise per manual invocation.
      expect(waits.length).toBe(2);
      const outcomes = await flushWaits(waits);
      expect(outcomes.every((outcome) => outcome === "fulfilled")).toBe(true);
      expect(execute).toHaveBeenCalledTimes(rejected ? 0 : 1);
      expect(send).toHaveBeenCalledTimes(failure ? 4 : 2);
      expect((await storage.list({ prefix: SCHEDULER_ALERT_PENDING_PREFIX })).size).toBe(
        failure ? 2 : 1,
      );
      await notifySchedulerFailures(
        [],
        config,
        async () => new Response(null, { status: 204 }),
        storage,
      );
      expect((await storage.list({ prefix: SCHEDULER_ALERT_PENDING_PREFIX })).size).toBe(0);
    },
  );
});

describe("manual response detached from slow durable incident drain", () => {
  it("resolves a successful manual response strictly before the 30s budget while seven unresponsive durable incidents drain", async () => {
    const fixedNow = Date.UTC(2026, 6, 24, 18);
    vi.useFakeTimers({ now: fixedNow });
    vi.spyOn(console, "error").mockImplementation(() => {});
    const storage = new Storage();
    await enqueuePendingIncidents(storage, SLOT, STUCK_INCIDENTS);
    const execute = vi.fn(async () =>
      Response.json({ outcome: "success", processed: 1, succeeded: 1, failed: 0 }),
    );
    const { container, waits } = makeContainer(storage, execute);
    const send = vi.fn(async (_request: Request) => new Promise<Response>(() => {}));
    vi.stubGlobal("fetch", send);
    const responseP = handleSchedulerOpsRequest(manualRequest(), auth, container, fixedNow);
    let settled = false;
    void responseP.then(() => {
      settled = true;
    });
    await vi.advanceTimersByTimeAsync(29_900);
    await Promise.resolve();
    expect(settled).toBe(true);
    const first = await responseP;
    expect(first.status).toBe(200);
    const firstBody = await first.json();
    expect(execute).toHaveBeenCalledTimes(1);
    expect(await pendingCount(storage)).toBe(STUCK_INCIDENTS);
    expect(waits.length).toBe(1);
    const replayP = handleSchedulerOpsRequest(manualRequest(), auth, container, fixedNow);
    let replaySettled = false;
    void replayP.then(() => {
      replaySettled = true;
    });
    await vi.advanceTimersByTimeAsync(29_900);
    await Promise.resolve();
    expect(replaySettled).toBe(true);
    const replay = await replayP;
    expect(replay.status).toBe(200);
    expect(await replay.json()).toEqual(firstBody);
    expect(execute).toHaveBeenCalledTimes(1);
    expect(waits.length).toBe(2);
    // Flush the registered DurableObjectState lifecycle work: transport timeouts keep every incident pending.
    await vi.runAllTimersAsync();
    const outcomes = await flushWaits(waits);
    expect(outcomes.every((outcome) => outcome === "fulfilled")).toBe(true);
    expect(await pendingCount(storage)).toBe(STUCK_INCIDENTS);
    // Simulated restart/interruption: retained durable storage drains later with stable idempotency keys.
    const ackSend = vi.fn(async (_request: Request) => new Response(null, { status: 204 }));
    await notifySchedulerFailures([], config, ackSend, storage);
    expect(await pendingCount(storage)).toBe(0);
    const drainedKeys = new Set(alertKeys(send));
    expect(drainedKeys.size).toBe(STUCK_INCIDENTS);
    expect(new Set(alertKeys(ackSend))).toEqual(drainedKeys);
  });

  it("returns a refused manual response strictly before the 30s budget while seven unresponsive durable incidents drain", async () => {
    const fixedNow = Date.UTC(2026, 6, 24, 18);
    vi.useFakeTimers({ now: fixedNow });
    vi.spyOn(console, "error").mockImplementation(() => {});
    const storage = new Storage();
    await enqueuePendingIncidents(storage, SLOT, STUCK_INCIDENTS);
    const coordinator = new SchedulerCoordinator(storage, () => fixedNow);
    await coordinator.setControl({
      actorPrincipal: "scheduler-ops-secret-v1",
      requestId: "110e8400-e29b-41d4-a716-446655440000",
      reason: "test pause",
      expectedRevision: 0,
      paused: true,
    });
    const execute = vi.fn(async () =>
      Response.json({ outcome: "success", processed: 1, succeeded: 1, failed: 0 }),
    );
    const { container, waits } = makeContainer(storage, execute);
    const send = vi.fn(async (_request: Request) => new Promise<Response>(() => {}));
    vi.stubGlobal("fetch", send);
    const responseP = handleSchedulerOpsRequest(manualRequest(), auth, container, fixedNow);
    let settled = false;
    void responseP.then(() => {
      settled = true;
    });
    await vi.advanceTimersByTimeAsync(29_900);
    await Promise.resolve();
    expect(settled).toBe(true);
    const response = await responseP;
    expect(response.status).toBe(409);
    expect(execute).toHaveBeenCalledTimes(0);
    expect(await pendingCount(storage)).toBe(STUCK_INCIDENTS);
    expect(waits.length).toBe(1);
    // Slow refused same-ID replay also resolves strictly before the 30s budget.
    const replayP = handleSchedulerOpsRequest(manualRequest(), auth, container, fixedNow);
    let replaySettled = false;
    void replayP.then(() => {
      replaySettled = true;
    });
    await vi.advanceTimersByTimeAsync(29_900);
    await Promise.resolve();
    expect(replaySettled).toBe(true);
    const replay = await replayP;
    expect(replay.status).toBe(409);
    expect(execute).toHaveBeenCalledTimes(0);
    expect(await pendingCount(storage)).toBe(STUCK_INCIDENTS);
    expect(waits.length).toBe(2);
    await vi.runAllTimersAsync();
    const outcomes = await flushWaits(waits);
    expect(outcomes.every((outcome) => outcome === "fulfilled")).toBe(true);
    expect(await pendingCount(storage)).toBe(STUCK_INCIDENTS);
    const ackSend = vi.fn(async (_request: Request) => new Response(null, { status: 204 }));
    await notifySchedulerFailures([], config, ackSend, storage);
    expect(await pendingCount(storage)).toBe(0);
    const drainedKeys = new Set(alertKeys(send));
    expect(drainedKeys.size).toBe(STUCK_INCIDENTS);
    expect(new Set(alertKeys(ackSend))).toEqual(drainedKeys);
  });

  it("keeps the manual response independent when the background drain throws and records the deferral", async () => {
    const fixedNow = Date.UTC(2026, 6, 24, 18);
    vi.useFakeTimers({ now: fixedNow });
    const errorLog = vi.spyOn(console, "error").mockImplementation(() => {});
    const storage = new InterruptibleStorage();
    await enqueuePendingIncidents(storage, SLOT, 1);
    const execute = vi.fn(async () =>
      Response.json({ outcome: "success", processed: 1, succeeded: 1, failed: 0 }),
    );
    const { container, waits } = makeContainer(storage, execute);
    const send = vi.fn(async (_request: Request) => new Response(null, { status: 503 }));
    vi.stubGlobal("fetch", send);
    const first = await handleSchedulerOpsRequest(manualRequest(), auth, container, fixedNow);
    expect(first.status).toBe(200);
    expect(execute).toHaveBeenCalledTimes(1);
    expect(waits.length).toBe(1);
    expect(await pendingCount(storage)).toBe(1);
    storage.failPendingList = true;
    const replay = await handleSchedulerOpsRequest(manualRequest(), auth, container, fixedNow);
    expect(replay.status).toBe(200);
    expect(execute).toHaveBeenCalledTimes(1);
    expect(waits.length).toBe(2);
    const outcomes = await flushWaits(waits);
    // The catch guarantees every registration settles fulfilled even when the drain throws.
    expect(outcomes.every((outcome) => outcome === "fulfilled")).toBe(true);
    const drainDeferred = errorLog.mock.calls.some((args) => {
      const payload = args[1] as Record<string, unknown> | undefined;
      return payload?.event === "scheduler_manual_alert_drain_deferred";
    });
    expect(drainDeferred).toBe(true);
    storage.failPendingList = false;
    const ackSend = vi.fn(async (_request: Request) => new Response(null, { status: 204 }));
    await notifySchedulerFailures([], config, ackSend, storage);
    expect(await pendingCount(storage)).toBe(0);
  });
});
