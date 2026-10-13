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
      const container = Object.create(
        AnimalEkarteApiContainer.prototype,
      ) as AnimalEkarteApiContainer;
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
      Object.defineProperties(container, {
        ctx: { value: { storage } },
        envVars: { value: { SCHEDULER_INTERNAL_TOKEN: "test-token" } },
        containerFetch: { value: execute },
      });
      const send = vi.fn(async () => new Response(null, { status: 503 }));
      vi.stubGlobal("fetch", send);
      const request = () =>
        new Request("https://api.example.com/_internal/scheduler/runs", {
          method: "POST",
          headers: {
            Authorization: `Bearer ${secret}`,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            job: "dormant",
            mode: "catch_up",
            scheduled_time_ms: slot,
            request_id: "de3a2410-0b43-4775-9b8b-b4e3de5a47e8",
            reason: "retry incident",
          }),
        });
      const auth = { automationSecret: secret };
      const first = await handleSchedulerOpsRequest(request(), auth, container, now);
      expect(first.status).toBe(rejected ? 409 : failure ? 502 : 200);
      const body = await first.json();
      const replay = await handleSchedulerOpsRequest(request(), auth, container, now);
      expect(replay.status).toBe(first.status);
      expect(await replay.json()).toEqual(body);
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
