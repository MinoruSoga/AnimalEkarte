import { describe, expect, it, vi } from "vitest";
import {
  SchedulerCoordinator,
  type CoordinatorStorage,
  type CoordinatorTransaction,
  type ScheduledRunLedger,
} from "./scheduler-coordinator";
import {
  enqueueSchedulerIncident,
  SCHEDULER_ALERT_PENDING_PREFIX,
  SCHEDULER_ALERT_BATCH_SIZE,
} from "./scheduler-alert-outbox";
import { notifySchedulerFailures } from "./scheduler-ops";
import { SCHEDULER_NAME } from "./scheduled-jobs";

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
  allowedHost: "alerts.example.com",
  webhookURL: "https://alerts.example.com/scheduler",
  webhookSecret: "test-secret",
};
function incident(time = 1): ScheduledRunLedger {
  return {
    version: 1,
    scheduler: SCHEDULER_NAME,
    runId: `${SCHEDULER_NAME}:${time}:no_show`,
    runKey: SchedulerCoordinator.runKey("no_show", time),
    job: "no_show",
    cron: "0 1 * * *",
    scheduledTime: time,
    fenceToken: 1,
    status: "skipped",
    failureCode: "stuck_lease",
    startedAt: time,
    finishedAt: time + 1,
  };
}
async function pending(storage: Storage) {
  return storage.list({ prefix: SCHEDULER_ALERT_PENDING_PREFIX });
}

describe("durable scheduler incidents", () => {
  it("uses the same incident key during concurrent drains", async () => {
    const storage = new Storage();
    await storage.transaction((tx) => enqueueSchedulerIncident(tx, incident()));
    let release: () => void = () => undefined;
    let bothStarted: () => void = () => undefined;
    const held = new Promise<void>((resolve) => {
      release = resolve;
    });
    const started = new Promise<void>((resolve) => {
      bothStarted = resolve;
    });
    const received = new Set<string>();
    let attempts = 0;
    const send = vi.fn(async (request: Request) => {
      received.add(request.headers.get("Idempotency-Key") ?? "");
      attempts += 1;
      if (attempts === 2) bothStarted();
      await held;
      return new Response(null, { status: 204 });
    });
    const first = notifySchedulerFailures([], config, send, storage);
    const second = notifySchedulerFailures([], config, send, storage);
    await started;
    release();
    await Promise.all([first, second]);
    expect(send).toHaveBeenCalledTimes(2);
    expect(received.size).toBe(1);
    expect((await pending(storage)).size).toBe(0);
  });
  it("survives detection-before-delivery exit and suppresses acknowledged replay", async () => {
    const storage = new Storage();
    await storage.transaction((tx) => enqueueSchedulerIncident(tx, incident()));
    const send = vi.fn(async () => new Response(null, { status: 204 }));
    await notifySchedulerFailures([], config, send, storage);
    await notifySchedulerFailures(
      [{ disposition: "duplicate", ledger: incident() }],
      config,
      send,
      storage,
    );
    expect(send).toHaveBeenCalledOnce();
    expect((await pending(storage)).size).toBe(0);
  });
  it("retries remote success before ack with receiver dedup and the identical key", async () => {
    const storage = new Storage();
    await storage.transaction((tx) => enqueueSchedulerIncident(tx, incident()));
    const accepted = new Set<string>();
    const send = vi.fn(async (request: Request) => {
      accepted.add(request.headers.get("Idempotency-Key") ?? "");
      return new Response(null, { status: 204 });
    });
    const remove = vi.spyOn(storage, "delete").mockRejectedValueOnce(new Error("ack failed"));
    await expect(notifySchedulerFailures([], config, send, storage)).rejects.toThrow();
    expect((await pending(storage)).size).toBe(1);
    remove.mockRestore();
    await notifySchedulerFailures([], config, send, storage);
    expect(send).toHaveBeenCalledTimes(2);
    expect(accepted.size).toBe(1);
    expect((await pending(storage)).size).toBe(0);
  });
  it("keeps individual failed incidents while acknowledging successes", async () => {
    const storage = new Storage();
    await storage.transaction(async (tx) => {
      await enqueueSchedulerIncident(tx, incident(1));
      await enqueueSchedulerIncident(tx, incident(2));
    });
    const send = vi.fn(
      async (request: Request) =>
        new Response(null, {
          status: request.headers.get("Idempotency-Key")?.includes(":1:") ? 500 : 204,
        }),
    );
    await expect(notifySchedulerFailures([], config, send, storage)).rejects.toThrow();
    expect(send).toHaveBeenCalledTimes(2);
    expect((await pending(storage)).size).toBe(1);
  });
  it("does not starve later incidents behind a full failed batch", async () => {
    const storage = new Storage();
    await storage.transaction(async (tx) => {
      for (let i = 1; i <= SCHEDULER_ALERT_BATCH_SIZE + 1; i++)
        await enqueueSchedulerIncident(tx, incident(i));
    });
    const keys = new Set<string>();
    const send = vi.fn(async (request: Request) => {
      keys.add(request.headers.get("Idempotency-Key") ?? "");
      return new Response(null, { status: 500 });
    });
    await expect(notifySchedulerFailures([], config, send, storage)).rejects.toThrow();
    expect(send).toHaveBeenCalledTimes(SCHEDULER_ALERT_BATCH_SIZE);
    await expect(notifySchedulerFailures([], config, send, storage)).rejects.toThrow();
    expect(keys.size).toBe(SCHEDULER_ALERT_BATCH_SIZE + 1);
  });
  it("keeps pending incidents through history pruning and configured-endpoint failure", async () => {
    const storage = new Storage();
    await storage.transaction((tx) => enqueueSchedulerIncident(tx, incident()));
    await new SchedulerCoordinator(storage, () => Date.UTC(2026, 9, 13)).getStatus(5);
    const send = vi.fn(async () => new Response(null, { status: 204 }));
    await expect(
      notifySchedulerFailures([], { environment: "staging" }, send, storage),
    ).rejects.toThrow();
    expect(send).not.toHaveBeenCalled();
    expect((await pending(storage)).size).toBe(1);
  });
  it("rolls back the incident together with its detection transition", async () => {
    const storage = new Storage();
    await expect(
      storage.transaction(async (tx) => {
        await enqueueSchedulerIncident(tx, incident());
        throw new Error("rollback");
      }),
    ).rejects.toThrow("rollback");
    expect((await pending(storage)).size).toBe(0);
  });
});
