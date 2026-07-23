import { describe, expect, it, vi } from "vitest";
import {
  RUN_LEDGER_RETENTION_MS,
  SCHEDULED_JOB_LEASE_MS,
  SchedulerCoordinator,
  runScheduledPlan,
  scheduledRunRequiresCronFailure,
  type CoordinatorStorage,
  type CoordinatorTransaction,
  type ScheduledRunLedger,
} from "./scheduler-coordinator";
import {
  SCHEDULER_NAME,
  type ScheduledJobOutcome,
  type ScheduledJobRequest,
} from "./scheduled-jobs";

const DAY_MS = 24 * 60 * 60 * 1_000;

class MemoryStorage implements CoordinatorStorage {
  readonly values = new Map<string, unknown>();
  private transactionTail: Promise<void> = Promise.resolve();

  async get<T>(key: string): Promise<T | undefined> {
    return this.values.get(key) as T | undefined;
  }

  async put<T>(key: string, value: T): Promise<void> {
    this.values.set(key, structuredClone(value));
  }

  async delete(key: string): Promise<boolean> {
    return this.values.delete(key);
  }

  async list<T>(options: { prefix: string }): Promise<Map<string, T>> {
    const entries = [...this.values.entries()]
      .filter(([key]) => key.startsWith(options.prefix))
      .sort(([left], [right]) => left.localeCompare(right));
    return new Map(entries) as Map<string, T>;
  }

  async transaction<T>(closure: (transaction: CoordinatorTransaction) => Promise<T>): Promise<T> {
    let release: () => void = () => undefined;
    const previous = this.transactionTail;
    this.transactionTail = new Promise<void>((resolve) => {
      release = resolve;
    });
    await previous;

    const snapshot = new Map(this.values);
    const transaction: CoordinatorTransaction = {
      get: async <V>(key: string) => snapshot.get(key) as V | undefined,
      put: async <V>(key: string, value: V) => {
        snapshot.set(key, structuredClone(value));
      },
      delete: async (key: string) => snapshot.delete(key),
    };

    try {
      const result = await closure(transaction);
      this.values.clear();
      for (const [key, value] of snapshot) {
        this.values.set(key, value);
      }
      return result;
    } finally {
      release();
    }
  }
}

function successOutcome(processed = 1): ScheduledJobOutcome {
  return { outcome: "success", processed, succeeded: processed, failed: 0 };
}

function scheduledTime(day: number): number {
  return 1_721_865_600_000 + day * DAY_MS;
}

describe("SchedulerCoordinator", () => {
  it("persists pause state and intentionally skips work while paused", async () => {
    const storage = new MemoryStorage();
    let now = scheduledTime(10);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    const execute = vi.fn(async () => successOutcome());

    await coordinator.setPaused(true);
    expect(await coordinator.getControl()).toMatchObject({ paused: true, changedAt: now });

    const result = await coordinator.run("no_show", "0 1 * * *", now, execute);
    expect(result.disposition).toBe("paused");
    expect(result.ledger?.status).toBe("paused");
    expect(execute).not.toHaveBeenCalled();

    now += 1;
    await coordinator.setPaused(false);
    expect(await coordinator.getControl()).toMatchObject({ paused: false, changedAt: now });

    const duplicate = await coordinator.run("no_show", "0 1 * * *", now - 1, execute);
    expect(duplicate).toMatchObject({
      disposition: "duplicate",
      ledger: { status: "paused" },
    });
    expect(execute).not.toHaveBeenCalled();
  });

  it("claims a run with a stable run_id and positive fencing token", async () => {
    const storage = new MemoryStorage();
    const time = scheduledTime(1);
    const coordinator = new SchedulerCoordinator(storage, () => time + 1_000);
    const execute = vi.fn(async (request: ScheduledJobRequest) => {
      expect(request).toEqual({
        scheduler: SCHEDULER_NAME,
        job: "no_show",
        scheduled_time: time,
        run_id: `${SCHEDULER_NAME}:${time}:no_show`,
        fence_token: 1,
      });
      return successOutcome(2);
    });

    const result = await coordinator.run("no_show", "0 1 * * *", time, execute);

    expect(result).toMatchObject({
      disposition: "executed",
      ledger: { status: "success", fenceToken: 1, scheduledTime: time },
    });
    expect(execute).toHaveBeenCalledTimes(1);
    expect(await coordinator.getActive()).toBeUndefined();
  });

  it.each(["partial", "failed"] as const)(
    "persists %s as a terminal result instead of retrying",
    async (outcome) => {
      const storage = new MemoryStorage();
      const time = scheduledTime(2);
      const coordinator = new SchedulerCoordinator(storage, () => time + 1_000);
      const execute = vi.fn(async () => ({
        outcome,
        processed: 2,
        succeeded: outcome === "partial" ? 1 : 0,
        failed: outcome === "partial" ? 1 : 2,
      }));

      const first = await coordinator.run("delivery", "0 1 * * *", time, execute);
      const duplicate = await coordinator.run("delivery", "0 1 * * *", time, execute);

      expect(first.ledger?.status).toBe(outcome);
      expect(duplicate).toMatchObject({
        disposition: "duplicate",
        ledger: { status: outcome },
      });
      expect(execute).toHaveBeenCalledTimes(1);
    },
  );

  it("continues the 01:00 UTC plan after no_show is partial, then reports both results", async () => {
    const storage = new MemoryStorage();
    const time = scheduledTime(2);
    const coordinator = new SchedulerCoordinator(storage, () => time + 1_000);
    const jobs: ScheduledJobRequest["job"][] = [];

    const results = await runScheduledPlan(coordinator, "0 1 * * *", time, async (request) => {
      jobs.push(request.job);
      return request.job === "no_show"
        ? { outcome: "partial", processed: 2, succeeded: 1, failed: 1 }
        : successOutcome(3);
    });

    expect(jobs).toEqual(["no_show", "delivery"]);
    expect(results.map((result) => result.ledger?.status)).toEqual(["partial", "success"]);
    expect(results.some(scheduledRunRequiresCronFailure)).toBe(true);
  });

  it("does not fail Cron for an intentional pause or a successful duplicate", async () => {
    const storage = new MemoryStorage();
    const time = scheduledTime(2);
    const coordinator = new SchedulerCoordinator(storage, () => time + 1_000);
    const execute = vi.fn(async () => successOutcome());

    await coordinator.setPaused(true);
    const paused = await coordinator.run("no_show", "0 1 * * *", time, execute);
    expect(scheduledRunRequiresCronFailure(paused)).toBe(false);

    await coordinator.setPaused(false);
    await coordinator.run("no_show", "0 1 * * *", time, execute);
    const duplicate = await coordinator.run("no_show", "0 1 * * *", time, execute);
    expect(scheduledRunRequiresCronFailure(duplicate)).toBe(false);
  });

  it("turns transport failure into a terminal failed ledger without storing error text", async () => {
    const storage = new MemoryStorage();
    const time = scheduledTime(3);
    const coordinator = new SchedulerCoordinator(storage, () => time + 1_000);
    const execute = vi.fn(async () => {
      throw new Error("sensitive upstream detail");
    });

    const result = await coordinator.run("dormant", "0 17 * * *", time, execute);
    expect(result).toMatchObject({
      disposition: "executed",
      ledger: { status: "failed", failureCode: "transport" },
    });
    expect(result.ledger).not.toHaveProperty("outcome");
    expect(JSON.stringify(result)).not.toContain("sensitive upstream detail");
  });

  it("deduplicates an exact successful run_id", async () => {
    const storage = new MemoryStorage();
    const time = scheduledTime(4);
    const coordinator = new SchedulerCoordinator(storage, () => time + 1_000);
    const execute = vi.fn(async () => successOutcome());

    await coordinator.run("no_show", "0 6,11 * * *", time, execute);
    const duplicate = await coordinator.run("no_show", "0 6,11 * * *", time, execute);

    expect(duplicate).toMatchObject({
      disposition: "duplicate",
      ledger: { status: "success" },
    });
    expect(execute).toHaveBeenCalledTimes(1);
  });

  it("rejects an older event after a newer scheduledTime was claimed", async () => {
    const storage = new MemoryStorage();
    let now = scheduledTime(7);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    const execute = vi.fn(async () => successOutcome());

    await coordinator.run("no_show", "0 6,11 * * *", scheduledTime(6), execute);
    now += 1_000;
    const stale = await coordinator.run("no_show", "0 1 * * *", scheduledTime(5), execute);

    expect(stale).toMatchObject({
      disposition: "stale",
      ledger: { status: "failed", failureCode: "stale" },
    });
    expect(execute).toHaveBeenCalledTimes(1);
  });

  it("does not use elapsed wall-clock time as a stale criterion", async () => {
    const storage = new MemoryStorage();
    const oldSchedule = scheduledTime(0);
    const coordinator = new SchedulerCoordinator(storage, () => oldSchedule + 365 * DAY_MS);
    const execute = vi.fn(async () => successOutcome());

    const result = await coordinator.run("dormant", "0 17 * * *", oldSchedule, execute);

    expect(result.ledger?.status).toBe("success");
    expect(execute).toHaveBeenCalledTimes(1);
  });

  it.each([0, -1, Number.NaN, 1.5])("rejects invalid scheduledTime %s", async (time) => {
    const coordinator = new SchedulerCoordinator(new MemoryStorage(), () => scheduledTime(1));

    await expect(
      coordinator.run("no_show", "0 1 * * *", time, async () => successOutcome()),
    ).rejects.toThrowError(/positive integer/i);
  });

  it("enforces global single-flight across different jobs", async () => {
    const storage = new MemoryStorage();
    let now = scheduledTime(8);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    let releaseFirst: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const firstExecution = new Promise<ScheduledJobOutcome>((resolve) => {
      releaseFirst = resolve;
    });

    const first = coordinator.run("no_show", "0 1 * * *", now, async () => firstExecution);
    await vi.waitFor(async () => {
      expect((await coordinator.getActive())?.job).toBe("no_show");
    });

    const blocked = await coordinator.run("delivery", "0 1 * * *", now, async () =>
      successOutcome(),
    );
    expect(blocked).toMatchObject({
      disposition: "busy",
      ledger: { status: "failed", failureCode: "busy" },
    });
    expect(blocked.ledger).not.toHaveProperty("outcome");

    releaseFirst(successOutcome());
    await expect(first).resolves.toMatchObject({ ledger: { status: "success" } });

    const blockedDuplicate = await coordinator.run("delivery", "0 1 * * *", now, async () =>
      successOutcome(),
    );
    expect(blockedDuplicate).toMatchObject({
      disposition: "duplicate",
      ledger: { status: "failed", failureCode: "busy" },
    });
  });

  it("marks an expired active lease failed before granting a newer fence", async () => {
    const storage = new MemoryStorage();
    let now = scheduledTime(9);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    let releaseFirst: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const firstExecution = new Promise<ScheduledJobOutcome>((resolve) => {
      releaseFirst = resolve;
    });

    const first = coordinator.run("no_show", "0 1 * * *", now, async () => firstExecution);
    await vi.waitFor(async () => {
      expect((await coordinator.getActive())?.fenceToken).toBe(1);
    });

    now += SCHEDULED_JOB_LEASE_MS + 1;
    const second = await coordinator.run("delivery", "0 1 * * *", now, async () =>
      successOutcome(),
    );
    expect(second).toMatchObject({
      ledger: { status: "success", fenceToken: 2 },
    });

    releaseFirst(successOutcome());
    const fenced = await first;
    // The old executor did return success, but its expired fence may no longer
    // overwrite the replacement run's ledger. This does not claim to roll back
    // side effects already performed by that executor.
    expect(fenced).toMatchObject({
      disposition: "fenced",
      ledger: { status: "failed", failureCode: "lease_expired" },
    });
  });

  it("treats an exact in-progress run as a duplicate and never executes twice", async () => {
    const storage = new MemoryStorage();
    const now = scheduledTime(10);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    let release: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const pending = new Promise<ScheduledJobOutcome>((resolve) => {
      release = resolve;
    });
    const execute = vi.fn(async () => pending);

    const first = coordinator.run("no_show", "0 6,11 * * *", now, execute);
    await vi.waitFor(async () => {
      expect((await coordinator.getActive())?.job).toBe("no_show");
    });
    const duplicate = await coordinator.run("no_show", "0 6,11 * * *", now, execute);

    expect(duplicate).toMatchObject({
      disposition: "duplicate",
      ledger: { status: "running" },
    });
    expect(execute).toHaveBeenCalledTimes(1);
    release(successOutcome());
    await first;
  });

  it("deterministically prunes run ledgers older than 35 days", async () => {
    const storage = new MemoryStorage();
    const now = scheduledTime(50);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    const oldTime = now - RUN_LEDGER_RETENTION_MS - 1;
    const boundaryTime = now - RUN_LEDGER_RETENTION_MS;

    const oldLedger: ScheduledRunLedger = {
      version: 1,
      scheduler: SCHEDULER_NAME,
      runId: `${SCHEDULER_NAME}:${oldTime}:no_show`,
      runKey: SchedulerCoordinator.runKey("no_show", oldTime),
      cron: "0 1 * * *",
      job: "no_show",
      scheduledTime: oldTime,
      fenceToken: 1,
      status: "success",
      startedAt: oldTime,
      finishedAt: oldTime,
      outcome: successOutcome(),
    };
    const boundaryLedger: ScheduledRunLedger = {
      ...oldLedger,
      runId: `${SCHEDULER_NAME}:${boundaryTime}:delivery`,
      runKey: SchedulerCoordinator.runKey("delivery", boundaryTime),
      job: "delivery",
      scheduledTime: boundaryTime,
    };
    await storage.put(oldLedger.runKey, oldLedger);
    await storage.put(boundaryLedger.runKey, boundaryLedger);

    await coordinator.run("dormant", "0 17 * * *", now, async () => successOutcome());

    expect(storage.values.has(oldLedger.runKey)).toBe(false);
    expect(storage.values.has(boundaryLedger.runKey)).toBe(true);
  });
});
