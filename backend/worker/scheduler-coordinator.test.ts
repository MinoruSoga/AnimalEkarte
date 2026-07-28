import { describe, expect, it, vi } from "vitest";
import {
  OPERATION_AUDIT_RETENTION_MS,
  RUN_LEDGER_RETENTION_MS,
  SCHEDULED_JOB_LEASE_MS,
  SCHEDULER_OPS_RATE_LIMIT,
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
  readonly listCalls: Array<{
    prefix: string;
    startAfter?: string;
    reverse?: boolean;
    limit?: number;
  }> = [];
  readonly directDeleteCalls: string[] = [];
  private transactionTail: Promise<void> = Promise.resolve();

  async get<T>(key: string): Promise<T | undefined> {
    return this.values.get(key) as T | undefined;
  }

  async put<T>(key: string, value: T): Promise<void> {
    this.values.set(key, structuredClone(value));
  }

  async delete(key: string): Promise<boolean> {
    this.directDeleteCalls.push(key);
    return this.values.delete(key);
  }

  async list<T>(options: {
    prefix: string;
    startAfter?: string;
    reverse?: boolean;
    limit?: number;
  }): Promise<Map<string, T>> {
    this.listCalls.push({ ...options });
    const sorted = [...this.values.entries()]
      .filter(([key]) => key.startsWith(options.prefix))
      .sort(([left], [right]) => left.localeCompare(right));
    const afterCursor =
      options.startAfter === undefined
        ? sorted
        : sorted.filter(([key]) => key > options.startAfter!);
    const ordered = options.reverse ? [...afterCursor].reverse() : afterCursor;
    const entries =
      options.limit === undefined ? ordered : ordered.slice(0, options.limit);
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
  it("audits revision-checked pause controls and replays a request_id idempotently", async () => {
    const storage = new MemoryStorage();
    const now = scheduledTime(10);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    const command = {
      paused: true,
      expectedRevision: 0,
      requestId: "d8dd6534-ef3a-48fc-8a8f-3f9a8d9cd068",
      reason: "INC-123 scheduled maintenance",
      actorPrincipal: "scheduler-ops-secret-v1",
    } as const;

    const first = await coordinator.setControl(command);
    const duplicate = await coordinator.setControl(command);

    expect(first).toMatchObject({
      status: "completed",
      requestedPaused: true,
      control: { paused: true, revision: 1, changedAt: now },
    });
    expect(duplicate).toEqual(first);
    expect(await coordinator.getControl()).toMatchObject({ revision: 1, paused: true });
  });

  it("records a rejected control operation when the expected revision is stale", async () => {
    const coordinator = new SchedulerCoordinator(new MemoryStorage(), () => scheduledTime(10));
    const operation = await coordinator.setControl({
      paused: true,
      expectedRevision: 7,
      requestId: "28d8c19d-3f74-49c3-9a70-2077b16846f2",
      reason: "INC-125 stale operator window",
      actorPrincipal: "scheduler-ops-secret-v1",
    });

    expect(operation).toMatchObject({
      status: "rejected",
      rejectionCode: "revision_conflict",
      requestedPaused: true,
      control: { revision: 0, paused: false },
    });
  });

  it("replays a rejected CAS using the persisted requested pause intent", async () => {
    const coordinator = new SchedulerCoordinator(
      new MemoryStorage(),
      () => scheduledTime(10),
    );
    const command = {
      paused: true,
      expectedRevision: 7,
      requestId: "ccdfcb43-a170-41f0-b2dc-fee305574d69",
      reason: "INC-134 stale operator retry",
      actorPrincipal: "scheduler-ops-secret-v1",
    } as const;

    const rejected = await coordinator.setControl(command);
    await expect(coordinator.setControl(command)).resolves.toEqual(rejected);
    await expect(
      coordinator.setControl({ ...command, paused: false }),
    ).rejects.toThrow("request_id_conflict");
  });

  it("exposes no unaudited pause mutation API", () => {
    const coordinator = new SchedulerCoordinator(new MemoryStorage());

    expect("setPaused" in coordinator).toBe(false);
  });

  it("enforces the manual recovery slot and window inside the coordinator", async () => {
    const now = scheduledTime(50) + 17 * 60 * 60 * 1_000;
    const coordinator = new SchedulerCoordinator(new MemoryStorage(), () => now);
    const execute = vi.fn(async () => successOutcome());
    const command = {
      actorPrincipal: "scheduler-ops-secret-v1",
      job: "dormant",
      mode: "catch_up",
      reason: "INC-135 recover missing slot",
      requestId: "817e6e53-37de-4486-bb12-472e343e5178",
    } as const;

    await expect(
      coordinator.runManual(
        { ...command, scheduledTime: now + 60 * 60 * 1_000 },
        execute,
      ),
    ).rejects.toThrow("scheduled_time_outside_recovery_window");
    await expect(
      coordinator.runManual(
        {
          ...command,
          requestId: "3ba93288-1a41-4444-81c2-e49484398115",
          scheduledTime: now - RUN_LEDGER_RETENTION_MS - 1,
        },
        execute,
      ),
    ).rejects.toThrow("scheduled_time_outside_recovery_window");
    await expect(
      coordinator.runManual(
        {
          ...command,
          requestId: "207cbcb9-078a-4fb7-a7da-a18588221982",
          scheduledTime: now - 30 * 60 * 1_000,
        },
        execute,
      ),
    ).rejects.toThrow("scheduled_time_is_not_a_cron_slot");
    expect(execute).not.toHaveBeenCalled();
  });

  it("persists a per-principal scheduler operations rate limit", async () => {
    const coordinator = new SchedulerCoordinator(
      new MemoryStorage(),
      () => scheduledTime(10),
    );

    for (let request = 0; request < SCHEDULER_OPS_RATE_LIMIT; request += 1) {
      await expect(
        coordinator.consumeOpsRateLimit("scheduler-ops-secret-v1"),
      ).resolves.toMatchObject({ allowed: true });
    }
    await expect(
      coordinator.consumeOpsRateLimit("scheduler-ops-secret-v1"),
    ).resolves.toMatchObject({ allowed: false });
    await expect(
      coordinator.consumeOpsRateLimit(
        "scheduler-ops-secret-v1",
        scheduledTime(10) + 60_000,
      ),
    ).resolves.toMatchObject({ allowed: true });
  });

  it("never replays an existing failed slot through manual catch-up", async () => {
    const storage = new MemoryStorage();
    const time = scheduledTime(11) + 60 * 60 * 1_000;
    const coordinator = new SchedulerCoordinator(storage, () => time + 1_000);
    await coordinator.run("delivery", "0 1 * * *", time, async () => {
      throw new Error("ambiguous external side effect");
    });
    const execute = vi.fn(async () => successOutcome());

    const operation = await coordinator.runManual(
      {
        actorPrincipal: "scheduler-ops-secret-v1",
        job: "delivery",
        mode: "catch_up",
        reason: "INC-126 investigate failed delivery",
        requestId: "b74170bf-8a53-44f2-b66b-71b90609a54a",
        scheduledTime: time,
      },
      execute,
    );

    expect(operation).toMatchObject({
      status: "rejected",
      rejectionCode: "slot_already_recorded",
    });
    expect(execute).not.toHaveBeenCalled();
  });

  it("executes one missing manual slot and returns the immutable result idempotently", async () => {
    const storage = new MemoryStorage();
    const time = scheduledTime(12) + 17 * 60 * 60 * 1_000;
    const coordinator = new SchedulerCoordinator(storage, () => time + 1_000);
    const execute = vi.fn(async () => successOutcome());
    const command = {
      actorPrincipal: "scheduler-ops-secret-v1",
      job: "dormant",
      mode: "catch_up",
      reason: "INC-127 recover missing dormant slot",
      requestId: "7e7d0f27-e351-4e2f-9871-d3fed48e8af7",
      scheduledTime: time,
    } as const;

    const first = await coordinator.runManual(command, execute);
    const duplicate = await coordinator.runManual(command, execute);

    expect(first).toMatchObject({
      status: "completed",
      result: { disposition: "executed", ledger: { status: "success" } },
    });
    expect(duplicate).toEqual(first);
    expect(execute).toHaveBeenCalledTimes(1);
  });

  it("sanitizes active state and manual operation results in operator status", async () => {
    const storage = new MemoryStorage();
    let now = scheduledTime(13);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    let release: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const pendingOutcome = new Promise<ScheduledJobOutcome>((resolve) => {
      release = resolve;
    });

    const scheduledRun = coordinator.run(
      "no_show",
      "0 1 * * *",
      now,
      async () => pendingOutcome,
    );
    await vi.waitFor(async () => {
      expect((await coordinator.getActive())?.runId).toBe(
        `${SCHEDULER_NAME}:${now}:no_show`,
      );
    });

    const activeStatus = await coordinator.getStatus(20);
    expect(activeStatus.active).toEqual({
      version: 1,
      scheduler: SCHEDULER_NAME,
      runId: `${SCHEDULER_NAME}:${now}:no_show`,
      cron: "0 1 * * *",
      job: "no_show",
      scheduledTime: now,
      claimedAt: now,
      leaseExpiresAt: now + SCHEDULED_JOB_LEASE_MS,
    });
    expect(JSON.stringify(activeStatus)).not.toMatch(/runKey|fenceToken/);

    release(successOutcome());
    await scheduledRun;
    now += DAY_MS + 17 * 60 * 60 * 1_000;
    await coordinator.runManual(
      {
        actorPrincipal: "scheduler-ops-secret-v1",
        job: "dormant",
        mode: "catch_up",
        reason: "INC-128 recover missing dormant slot",
        requestId: "b4aad99b-0fa0-493e-80c6-d804f73ba077",
        scheduledTime: now,
      },
      async () => successOutcome(),
    );

    const terminalStatus = await coordinator.getStatus(20);
    expect(terminalStatus.recentOperations[0]).toMatchObject({
      kind: "manual_run",
      status: "completed",
      result: {
        disposition: "executed",
        ledger: { status: "success" },
      },
    });
    expect(JSON.stringify(terminalStatus)).not.toMatch(/runKey|fenceToken/);
    expect(storage.listCalls).toContainEqual({
      prefix: "scheduler:run:",
      reverse: true,
      limit: 20,
    });
    expect(storage.listCalls).toContainEqual({
      prefix: "scheduler:operation-index:",
      reverse: true,
      limit: 20,
    });
  });

  it("retains operation audits beyond the recovery window and prunes after 400 days", async () => {
    const storage = new MemoryStorage();
    let now = scheduledTime(14);
    const coordinator = new SchedulerCoordinator(storage, () => now);

    await coordinator.setControl({
      paused: true,
      expectedRevision: 0,
      requestId: "030033ed-ff75-4f7f-91b0-ac2332f7eea4",
      reason: "INC-129 old maintenance window",
      actorPrincipal: "scheduler-ops-secret-v1",
    });
    now += RUN_LEDGER_RETENTION_MS + 1;
    expect((await coordinator.getStatus(20)).recentOperations).toHaveLength(1);

    now += OPERATION_AUDIT_RETENTION_MS - RUN_LEDGER_RETENTION_MS;
    await coordinator.setControl({
      paused: false,
      expectedRevision: 1,
      requestId: "f014a62f-b818-4c3c-924b-8654ea948ca8",
      reason: "INC-130 current maintenance window",
      actorPrincipal: "scheduler-ops-secret-v1",
    });

    const status = await coordinator.getStatus(20);
    expect(status.recentOperations.map((operation) => operation.requestId)).toEqual([
      "f014a62f-b818-4c3c-924b-8654ea948ca8",
    ]);
    expect(storage.directDeleteCalls).toEqual([]);
    expect(
      [...storage.values.keys()].some((key) =>
        key.includes("030033ed-ff75-4f7f-91b0-ac2332f7eea4"),
      ),
    ).toBe(false);
  });

  it("persists pause state and intentionally skips work while paused", async () => {
    const storage = new MemoryStorage();
    let now = scheduledTime(10);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    const execute = vi.fn(async () => successOutcome());

    await coordinator.setControl({
      paused: true,
      expectedRevision: 0,
      requestId: "677205c8-7824-4dfd-b1f9-31427f616dc6",
      reason: "INC-132 pause scheduler",
      actorPrincipal: "scheduler-ops-secret-v1",
    });
    expect(await coordinator.getControl()).toMatchObject({ paused: true, changedAt: now });

    const result = await coordinator.run("no_show", "0 1 * * *", now, execute);
    expect(result.disposition).toBe("paused");
    expect(result.ledger?.status).toBe("paused");
    expect(execute).not.toHaveBeenCalled();

    now += 1;
    await coordinator.setControl({
      paused: false,
      expectedRevision: 1,
      requestId: "436e2a42-e3b8-4593-b80a-698250e5ad92",
      reason: "INC-132 resume scheduler",
      actorPrincipal: "scheduler-ops-secret-v1",
    });
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

    await coordinator.setControl({
      paused: true,
      expectedRevision: 0,
      requestId: "9a5f2bc5-efb4-41a3-9398-6217e5a26934",
      reason: "INC-133 pause scheduler",
      actorPrincipal: "scheduler-ops-secret-v1",
    });
    const paused = await coordinator.run("no_show", "0 1 * * *", time, execute);
    expect(scheduledRunRequiresCronFailure(paused)).toBe(false);

    await coordinator.setControl({
      paused: false,
      expectedRevision: 1,
      requestId: "84037c00-7abd-48f4-8486-7a1a9d6a5e47",
      reason: "INC-133 resume scheduler",
      actorPrincipal: "scheduler-ops-secret-v1",
    });
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
    const errorLog = vi.spyOn(console, "error").mockImplementation(() => undefined);
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
    expect(errorLog).toHaveBeenCalledWith(
      "scheduler run lease expired",
      expect.objectContaining({
        event: "scheduler_run_lease_expired",
        failure_code: "lease_expired",
        run_id: `${SCHEDULER_NAME}:${scheduledTime(9)}:no_show`,
      }),
    );
    errorLog.mockRestore();
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

  it("bounds history pruning to two storage pages per request", async () => {
    const storage = new MemoryStorage();
    const now = scheduledTime(500);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    for (let index = 0; index < 35; index += 1) {
      const time = now - RUN_LEDGER_RETENTION_MS - 1_000 - index;
      const ledger: ScheduledRunLedger = {
        version: 1,
        scheduler: SCHEDULER_NAME,
        runId: `${SCHEDULER_NAME}:${time}:no_show`,
        runKey: SchedulerCoordinator.runKey("no_show", time),
        cron: "0 1 * * *",
        job: "no_show",
        scheduledTime: time,
        fenceToken: 1,
        status: "success",
        startedAt: time,
        finishedAt: time,
        outcome: successOutcome(),
      };
      await storage.put(ledger.runKey, ledger);
    }

    await coordinator.getStatus(1);

    const prunePages = storage.listCalls.filter(
      (call) => call.prefix === "scheduler:run:" && call.limit === 15,
    );
    expect(prunePages).toHaveLength(2);
    expect(prunePages[1]?.startAfter).toBeTypeOf("string");
    expect(
      [...storage.values.keys()].filter((key) =>
        key.startsWith("scheduler:run:"),
      ),
    ).toHaveLength(5);
  });
});
