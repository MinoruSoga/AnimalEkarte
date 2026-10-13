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
import { dispatchScheduledEvent } from "./scheduled-handler";
import { SCHEDULER_ALERT_PENDING_PREFIX } from "./scheduler-alert-outbox";
import { notifySchedulerFailures } from "./scheduler-ops";

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
    const entries = options.limit === undefined ? ordered : ordered.slice(0, options.limit);
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
  it.each(["0 1 * * *", "0 17 * * *"])(
    "alerts when plane_sync blocks business plan %s",
    async (cron) => {
      const storage = new MemoryStorage();
      const now = scheduledTime(75);
      const coordinator = new SchedulerCoordinator(storage, () => now);
      let release = () => {};
      let started = () => {};
      const entered = new Promise<void>((resolve) => {
        started = resolve;
      });
      const held = new Promise<void>((resolve) => {
        release = resolve;
      });
      const plane = coordinator.run("plane_sync", "*/15 * * * *", now, async () => {
        started();
        await held;
        return successOutcome();
      });
      await entered;
      const execute = vi.fn(async () => successOutcome());
      try {
        const blocked = await runScheduledPlan(coordinator, cron, now, execute);
        expect(execute).not.toHaveBeenCalled();
        expect(blocked.every(scheduledRunRequiresCronFailure)).toBe(true);
        expect(blocked.every((r) => r.ledger?.failureCode === "schedule_conflict")).toBe(true);
        const controller = { cron, scheduledTime: now, noRetry: vi.fn() };
        await expect(dispatchScheduledEvent(controller, async () => blocked)).rejects.toThrow(
          "scheduled invocation failed",
        );
        expect(controller.noRetry).toHaveBeenCalledOnce();
        const pending = await storage.list({
          prefix: SCHEDULER_ALERT_PENDING_PREFIX,
        });
        expect(pending.size).toBe(blocked.length);
        const config = {
          environment: "staging",
          allowedHost: "alerts.example.com",
          webhookURL: "https://alerts.example.com/scheduler",
          webhookSecret: "test-secret",
        };
        await expect(
          notifySchedulerFailures(
            blocked,
            config,
            async () => new Response(null, { status: 503 }),
            storage,
          ),
        ).rejects.toThrow();
        expect((await storage.list({ prefix: SCHEDULER_ALERT_PENDING_PREFIX })).size).toBe(
          blocked.length,
        );
        const replay = await runScheduledPlan(coordinator, cron, now, execute);
        expect(replay.every(scheduledRunRequiresCronFailure)).toBe(true);
        const send = vi.fn(async () => new Response(null, { status: 204 }));
        await notifySchedulerFailures(replay, config, send, storage);
        expect(send).toHaveBeenCalledTimes(blocked.length);
        expect((await storage.list({ prefix: SCHEDULER_ALERT_PENDING_PREFIX })).size).toBe(0);
      } finally {
        release();
        await plane;
      }
    },
  );

  it("lets the winning parallel plan execute delivery exactly once", async () => {
    const storage = new MemoryStorage();
    const now = scheduledTime(70);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    let release: () => void = () => undefined;
    let started: () => void = () => undefined;
    const entered = new Promise<void>((resolve) => {
      started = resolve;
    });
    const held = new Promise<void>((resolve) => {
      release = resolve;
    });
    const execute = vi.fn(async (request: ScheduledJobRequest) => {
      if (request.job === "no_show") {
        started();
        await held;
      }
      return successOutcome();
    });
    const winner = runScheduledPlan(coordinator, "0 1 * * *", now, execute);
    await entered;
    const loser = await runScheduledPlan(coordinator, "0 1 * * *", now, execute);
    expect(loser).toHaveLength(2);
    expect(await storage.get(SchedulerCoordinator.runKey("delivery", now))).toBeUndefined();
    release();
    await winner;
    expect(execute.mock.calls.map(([request]) => request.job)).toEqual(["no_show", "delivery"]);
    const restarted = new SchedulerCoordinator(storage, () => now);
    await runScheduledPlan(restarted, "0 1 * * *", now, execute);
    expect(execute).toHaveBeenCalledTimes(2);
    expect(await storage.get(SchedulerCoordinator.runKey("delivery", now))).toMatchObject({
      status: "success",
    });
  });

  it("retries an expired incident after restart and webhook failure with the same key", async () => {
    const storage = new MemoryStorage();
    let now = scheduledTime(71);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    let release: () => void = () => undefined;
    let started: () => void = () => undefined;
    const entered = new Promise<void>((resolve) => {
      started = resolve;
    });
    const held = new Promise<void>((resolve) => {
      release = resolve;
    });
    const old = coordinator.run("plane_sync", "*/15 * * * *", now, async () => {
      started();
      await held;
      return successOutcome();
    });
    await entered;
    now += 15 * 60 * 1_000;
    const replacement = await coordinator.run("plane_sync", "*/15 * * * *", now, async () =>
      successOutcome(),
    );
    expect(replacement.expiredLedger?.failureCode).toBe("lease_expired");
    // Simulate exit after transactional detection, before notification. New instances share only storage.
    const config = {
      environment: "staging",
      allowedHost: "alerts.example.com",
      webhookURL: "https://alerts.example.com/scheduler",
      webhookSecret: "test-secret",
    };
    const send = vi.fn(async (_request: Request) => new Response(null, { status: 500 }));
    await expect(notifySchedulerFailures([], config, send, storage)).rejects.toThrow();
    send.mockImplementation(async () => new Response(null, { status: 204 }));
    const restarted = new SchedulerCoordinator(storage, () => now);
    const replay = await restarted.run("plane_sync", "*/15 * * * *", now, async () =>
      successOutcome(),
    );
    await notifySchedulerFailures([replay], config, send, storage);
    expect(send).toHaveBeenCalledTimes(2);
    expect(send.mock.calls[0]?.[0].headers.get("Idempotency-Key")).toBe(
      send.mock.calls[1]?.[0].headers.get("Idempotency-Key"),
    );
    await notifySchedulerFailures([replay], config, send, storage);
    expect(send).toHaveBeenCalledTimes(2);
    release();
    await old;
  });

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
    expect(await coordinator.getControl()).toMatchObject({
      revision: 1,
      paused: true,
    });
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
    const coordinator = new SchedulerCoordinator(new MemoryStorage(), () => scheduledTime(10));
    const command = {
      paused: true,
      expectedRevision: 7,
      requestId: "ccdfcb43-a170-41f0-b2dc-fee305574d69",
      reason: "INC-134 stale operator retry",
      actorPrincipal: "scheduler-ops-secret-v1",
    } as const;

    const rejected = await coordinator.setControl(command);
    await expect(coordinator.setControl(command)).resolves.toEqual(rejected);
    await expect(coordinator.setControl({ ...command, paused: false })).rejects.toThrow(
      "request_id_conflict",
    );
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
      coordinator.runManual({ ...command, scheduledTime: now + 60 * 60 * 1_000 }, execute),
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
    const coordinator = new SchedulerCoordinator(new MemoryStorage(), () => scheduledTime(10));

    for (let request = 0; request < SCHEDULER_OPS_RATE_LIMIT; request += 1) {
      await expect(
        coordinator.consumeOpsRateLimit("scheduler-ops-secret-v1"),
      ).resolves.toMatchObject({ allowed: true });
    }
    await expect(coordinator.consumeOpsRateLimit("scheduler-ops-secret-v1")).resolves.toMatchObject(
      { allowed: false },
    );
    await expect(
      coordinator.consumeOpsRateLimit("scheduler-ops-secret-v1", scheduledTime(10) + 60_000),
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

    const scheduledRun = coordinator.run("no_show", "0 1 * * *", now, async () => pendingOutcome);
    await vi.waitFor(async () => {
      expect((await coordinator.getActive())?.runId).toBe(`${SCHEDULER_NAME}:${now}:no_show`);
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
    expect(await coordinator.getControl()).toMatchObject({
      paused: true,
      changedAt: now,
    });

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
    expect(await coordinator.getControl()).toMatchObject({
      paused: false,
      changedAt: now,
    });

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
      ledger: { status: "skipped", failureCode: "busy" },
    });
    expect(blocked.ledger).not.toHaveProperty("outcome");
    expect(scheduledRunRequiresCronFailure(blocked)).toBe(false);

    releaseFirst(successOutcome());
    await expect(first).resolves.toMatchObject({
      ledger: { status: "success" },
    });

    const blockedDuplicate = await coordinator.run("delivery", "0 1 * * *", now, async () =>
      successOutcome(),
    );
    expect(blockedDuplicate).toMatchObject({
      disposition: "executed",
      ledger: { status: "success" },
    });
    expect(scheduledRunRequiresCronFailure(blockedDuplicate)).toBe(false);
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
      [...storage.values.keys()].filter((key) => key.startsWith("scheduler:run:")),
    ).toHaveLength(5);
  });

  it("does not log an overlap as skipped when its transaction rolls back", async () => {
    const storage = new MemoryStorage();
    const now = scheduledTime(14);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    const infoLog = vi.spyOn(console, "info").mockImplementation(() => undefined);
    let release: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const pending = new Promise<ScheduledJobOutcome>((resolve) => {
      release = resolve;
    });
    const first = coordinator.run("no_show", "0 1 * * *", now, async () => pending);
    await vi.waitFor(async () => {
      expect((await coordinator.getActive())?.job).toBe("no_show");
    });
    const transaction = storage.transaction.bind(storage);
    const transactionSpy = vi.spyOn(storage, "transaction").mockImplementation((closure) =>
      transaction(async (tx) => {
        const result = await closure(tx);
        const ledger = await tx.get<ScheduledRunLedger>(
          SchedulerCoordinator.runKey("delivery", now + 1),
        );
        if (ledger?.status === "skipped") throw new Error("commit rejected");
        return result;
      }),
    );
    try {
      await expect(
        coordinator.run("delivery", "0 1 * * *", now + 1, async () => successOutcome()),
      ).rejects.toThrow("commit rejected");
      expect(infoLog).not.toHaveBeenCalled();
      expect(await storage.get(SchedulerCoordinator.runKey("delivery", now + 1))).toBeUndefined();
    } finally {
      transactionSpy.mockRestore();
      release(successOutcome());
      await first;
      infoLog.mockRestore();
    }
  });

  it("logs one info line per healthy overlap and never logs a failure", async () => {
    const storage = new MemoryStorage();
    const now = scheduledTime(15);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    const infoLog = vi.spyOn(console, "info").mockImplementation(() => undefined);
    const errorLog = vi.spyOn(console, "error").mockImplementation(() => undefined);
    let release: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const pending = new Promise<ScheduledJobOutcome>((resolve) => {
      release = resolve;
    });

    const first = coordinator.run("no_show", "0 1 * * *", now, async () => pending);
    await vi.waitFor(async () => {
      expect((await coordinator.getActive())?.job).toBe("no_show");
    });

    const skippedExecute = vi.fn(async () => successOutcome());
    const firstOverlap = await coordinator.run("delivery", "0 1 * * *", now, skippedExecute);
    const secondOverlap = await coordinator.run(
      "delivery",
      "0 1 * * *",
      now + 60_000,
      skippedExecute,
    );

    for (const overlap of [firstOverlap, secondOverlap]) {
      expect(overlap).toMatchObject({
        disposition: "busy",
        ledger: { status: "skipped", failureCode: "busy" },
      });
      expect(scheduledRunRequiresCronFailure(overlap)).toBe(false);
    }
    expect(skippedExecute).not.toHaveBeenCalled();
    expect(infoLog).toHaveBeenCalledTimes(2);
    expect(infoLog).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        failure_code: "busy",
        job: "delivery",
        run_id: `${SCHEDULER_NAME}:${now}:delivery`,
      }),
    );
    expect(errorLog).not.toHaveBeenCalled();

    release(successOutcome());
    await expect(first).resolves.toMatchObject({
      ledger: { status: "success" },
    });
    infoLog.mockRestore();
    errorLog.mockRestore();
  });

  it("emits stuck_lease once per active run when the same job's later slot overlaps an unfinished lease", async () => {
    const storage = new MemoryStorage();
    const start = scheduledTime(16);
    let now = start;
    const coordinator = new SchedulerCoordinator(storage, () => now);
    let release: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const pending = new Promise<ScheduledJobOutcome>((resolve) => {
      release = resolve;
    });

    const first = coordinator.run("no_show", "0 1 * * *", start, async () => pending);
    await vi.waitFor(async () => {
      expect((await coordinator.getActive())?.job).toBe("no_show");
    });

    // Different job contending for the same slot is healthy contention,
    // never a stuck lease.
    const contention = await coordinator.run("delivery", "0 1 * * *", start, async () =>
      successOutcome(),
    );
    expect(contention).toMatchObject({
      disposition: "busy",
      ledger: { status: "skipped", failureCode: "busy" },
    });

    const active = await coordinator.getActive();
    if (active === undefined) throw new Error("expected active lease");
    // Simulate the anomalous extended lease needed to cross the real next slot.
    await storage.put("scheduler:active", {
      ...active,
      leaseExpiresAt: start + 11 * 60 * 60 * 1_000,
    });
    // Same job reaching a strictly later scheduled slot while its lease is
    // still valid emits stuck_lease once — even through another cron.
    now = start + 5 * 60 * 60 * 1_000;
    const stuck = await coordinator.run(
      "no_show",
      "0 6,11 * * *",
      start + 5 * 60 * 60 * 1_000,
      async () => successOutcome(),
    );
    expect(stuck).toMatchObject({
      disposition: "busy",
      ledger: { status: "skipped", failureCode: "stuck_lease" },
    });
    expect(scheduledRunRequiresCronFailure(stuck)).toBe(false);

    const config = {
      environment: "staging",
      allowedHost: "alerts.example.com",
      webhookURL: "https://alerts.example.com/scheduler",
      webhookSecret: "test-secret",
    };
    const send = vi.fn(async (_request: Request) => new Response(null, { status: 500 }));
    await expect(notifySchedulerFailures([], config, send, storage)).rejects.toThrow();
    send.mockImplementation(async () => new Response(null, { status: 204 }));
    const restarted = new SchedulerCoordinator(storage, () => now);
    const replay = await restarted.run("no_show", "0 6,11 * * *", now, async () =>
      successOutcome(),
    );
    await notifySchedulerFailures([replay], config, send, storage);
    expect(send).toHaveBeenCalledTimes(2);
    expect(send.mock.calls[0]?.[0].headers.get("Idempotency-Key")).toBe(
      send.mock.calls[1]?.[0].headers.get("Idempotency-Key"),
    );
    await notifySchedulerFailures([replay], config, send, storage);
    expect(send).toHaveBeenCalledTimes(2);

    // A repeated overlap of the same active run stays a plain busy skip.
    now = start + 10 * 60 * 60 * 1_000;
    const repeated = await coordinator.run(
      "no_show",
      "0 6,11 * * *",
      start + 10 * 60 * 60 * 1_000,
      async () => successOutcome(),
    );
    expect(repeated).toMatchObject({
      disposition: "busy",
      ledger: { status: "skipped", failureCode: "busy" },
    });
    expect(scheduledRunRequiresCronFailure(repeated)).toBe(false);

    release(successOutcome());
    await expect(first).resolves.toMatchObject({
      disposition: "executed",
      ledger: { status: "success" },
    });
  });

  it("emits stuck_lease again only when a different active run stalls", async () => {
    const storage = new MemoryStorage();
    const start = scheduledTime(17);
    let now = start;
    const coordinator = new SchedulerCoordinator(storage, () => now);
    let releaseFirst: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const firstExecution = new Promise<ScheduledJobOutcome>((resolve) => {
      releaseFirst = resolve;
    });

    const first = coordinator.run("no_show", "0 1 * * *", start, async () => firstExecution);
    await vi.waitFor(async () => {
      expect((await coordinator.getActive())?.runId).toBe(`${SCHEDULER_NAME}:${start}:no_show`);
    });

    const firstActive = await coordinator.getActive();
    if (firstActive === undefined) throw new Error("expected active lease");
    await storage.put("scheduler:active", {
      ...firstActive,
      leaseExpiresAt: start + 6 * 60 * 60 * 1_000,
    });
    now = start + 5 * 60 * 60 * 1_000;
    const firstStuck = await coordinator.run(
      "no_show",
      "0 6,11 * * *",
      start + 5 * 60 * 60 * 1_000,
      async () => successOutcome(),
    );
    expect(firstStuck.ledger?.failureCode).toBe("stuck_lease");

    releaseFirst(successOutcome());
    await expect(first).resolves.toMatchObject({
      ledger: { status: "success" },
    });

    // A newly claimed run must be able to emit stuck_lease on its own first
    // cross-slot overlap — the one-time marker follows the active run, not
    // the job.
    let releaseSecond: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const secondExecution = new Promise<ScheduledJobOutcome>((resolve) => {
      releaseSecond = resolve;
    });
    now = start + 10 * 60 * 60 * 1_000;
    const second = coordinator.run(
      "no_show",
      "0 6,11 * * *",
      start + 10 * 60 * 60 * 1_000,
      async () => secondExecution,
    );
    await vi.waitFor(async () => {
      expect((await coordinator.getActive())?.runId).toBe(
        `${SCHEDULER_NAME}:${start + 10 * 60 * 60 * 1_000}:no_show`,
      );
    });

    const secondActive = await coordinator.getActive();
    if (secondActive === undefined) throw new Error("expected active lease");
    await storage.put("scheduler:active", {
      ...secondActive,
      leaseExpiresAt: start + 16 * 60 * 60 * 1_000,
    });
    now = start + 15 * 60 * 60 * 1_000;
    const secondStuck = await coordinator.run(
      "no_show",
      "0 6,11 * * *",
      start + 15 * 60 * 60 * 1_000,
      async () => successOutcome(),
    );
    expect(secondStuck).toMatchObject({
      disposition: "busy",
      ledger: { status: "skipped", failureCode: "stuck_lease" },
    });

    releaseSecond(successOutcome());
    await expect(second).resolves.toMatchObject({
      ledger: { status: "success" },
    });
  });

  it("alerts once for a run reclaimed after its real lease expired at the next real cron slot", async () => {
    const storage = new MemoryStorage();
    const firstSlot = scheduledTime(9);
    const slotMs = 15 * 60 * 1_000;
    let now = firstSlot;
    const coordinator = new SchedulerCoordinator(storage, () => now);
    const errorLog = vi.spyOn(console, "error").mockImplementation(() => undefined);
    const send = vi.fn(async (_request: Request) => new Response(null, { status: 204 }));
    let releaseFirst: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const firstExecution = new Promise<ScheduledJobOutcome>((resolve) => {
      releaseFirst = resolve;
    });

    try {
      const first = coordinator.run(
        "plane_sync",
        "*/15 * * * *",
        firstSlot,
        async () => firstExecution,
      );
      await vi.waitFor(async () => {
        expect((await coordinator.getActive())?.job).toBe("plane_sync");
      });

      // The next real */15 slot fires 15 minutes later — past the real
      // 150-second lease — so the stuck run is reclaimed, not skipped.
      now = firstSlot + slotMs;
      const config = {
        environment: "staging",
        allowedHost: "alerts.example.com",
        webhookURL: "https://alerts.example.com/scheduler",
        webhookSecret: "alert-secret",
      };
      const results = await runScheduledPlan(coordinator, "*/15 * * * *", now, async () =>
        successOutcome(),
      );
      const [reclaimed] = results;
      expect(reclaimed).toMatchObject({
        disposition: "executed",
        ledger: {
          status: "success",
          runId: `${SCHEDULER_NAME}:${now}:plane_sync`,
        },
        expiredLedger: {
          status: "failed",
          failureCode: "lease_expired",
          runId: `${SCHEDULER_NAME}:${firstSlot}:plane_sync`,
        },
      });

      await notifySchedulerFailures(results, config, send, storage);
      expect(send).toHaveBeenCalledTimes(1);
      const [request] = send.mock.calls[0] ?? [];
      expect(request).toBeInstanceOf(Request);
      if (!(request instanceof Request)) {
        throw new Error("expected Request");
      }
      expect(request.headers.get("Idempotency-Key")).toBe(
        `staging:${SCHEDULER_NAME}:${firstSlot}:plane_sync:lease_expired`,
      );
      expect(errorLog).toHaveBeenCalledWith(
        "scheduler run lease expired",
        expect.objectContaining({
          event: "scheduler_run_lease_expired",
          run_id: `${SCHEDULER_NAME}:${firstSlot}:plane_sync`,
        }),
      );
      expect(errorLog).toHaveBeenCalledWith(
        "scheduler job failed",
        expect.objectContaining({
          event: "scheduler_job_failed",
          failure_code: "lease_expired",
        }),
      );

      // The reclaim already recorded and reported the failure; the following
      // slot carries nothing and must not re-alert.
      now = firstSlot + 2 * slotMs;
      const later = await runScheduledPlan(coordinator, "*/15 * * * *", now, async () =>
        successOutcome(),
      );
      expect(later.map((result) => result.expiredLedger)).toEqual([undefined]);
      await notifySchedulerFailures(later, config, send, storage);
      expect(send).toHaveBeenCalledTimes(1);

      releaseFirst(successOutcome());
      await expect(first).resolves.toMatchObject({
        disposition: "fenced",
        ledger: { status: "failed", failureCode: "lease_expired" },
      });
    } finally {
      errorLog.mockRestore();
    }
  });

  it("alerts once when a manual catch-up reclaims a run whose real lease expired", async () => {
    const storage = new MemoryStorage();
    const slot = scheduledTime(9) + 60 * 60 * 1_000;
    let now = slot;
    const coordinator = new SchedulerCoordinator(storage, () => now);
    const errorLog = vi.spyOn(console, "error").mockImplementation(() => undefined);
    const send = vi.fn(async (_request: Request) => new Response(null, { status: 204 }));
    const config = {
      environment: "staging",
      allowedHost: "alerts.example.com",
      webhookURL: "https://alerts.example.com/scheduler",
      webhookSecret: "alert-secret",
    };
    let releaseStuck: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const stuckExecution = new Promise<ScheduledJobOutcome>((resolve) => {
      releaseStuck = resolve;
    });

    try {
      const stuck = coordinator.run("no_show", "0 1 * * *", slot, async () => stuckExecution);
      await vi.waitFor(async () => {
        expect((await coordinator.getActive())?.job).toBe("no_show");
      });

      // While the scheduled run's real lease is still valid, a manual catch-up
      // of the missing sibling job is a healthy overlap rejection, not an alert.
      const busy = await coordinator.runManual(
        {
          actorPrincipal: "scheduler-ops-secret-v1",
          job: "delivery",
          mode: "catch_up",
          reason: "INC-201 recover missing delivery slot",
          requestId: "2f4f0a1c-6f0d-4f2b-9a4d-6f0d2f4f0a1c",
          scheduledTime: slot,
        },
        vi.fn(async () => successOutcome()),
      );
      expect(busy).toMatchObject({
        status: "rejected",
        rejectionCode: "scheduler_busy",
      });
      expect(busy.result).toBeUndefined();
      if (busy.result !== undefined) {
        await notifySchedulerFailures([busy.result], config, send);
      }
      expect(send).not.toHaveBeenCalled();

      // Once the real 150-second lease has lapsed, a manual catch-up reclaims
      // the stuck run; its expired ledger must reach the observer notify path.
      now = slot + 15 * 60 * 1_000;
      const command = {
        actorPrincipal: "scheduler-ops-secret-v1",
        job: "delivery",
        mode: "catch_up",
        reason: "INC-202 recover missing delivery slot",
        requestId: "7c3d9b41-2a1f-4c5e-8d6b-1e2f3a4b5c6d",
        scheduledTime: slot,
      } as const;
      const operation = await coordinator.runManual(command, async () => successOutcome());
      expect(operation).toMatchObject({
        status: "completed",
        result: {
          disposition: "executed",
          ledger: {
            status: "success",
            runId: `${SCHEDULER_NAME}:${slot}:delivery`,
          },
          expiredLedger: {
            status: "failed",
            failureCode: "lease_expired",
            runId: `${SCHEDULER_NAME}:${slot}:no_show`,
          },
        },
      });

      if (operation.result !== undefined) {
        await notifySchedulerFailures([operation.result], config, send, storage);
      }
      expect(send).toHaveBeenCalledTimes(1);
      const [request] = send.mock.calls[0] ?? [];
      expect(request).toBeInstanceOf(Request);
      if (!(request instanceof Request)) {
        throw new Error("expected Request");
      }
      expect(request.headers.get("Idempotency-Key")).toBe(
        `staging:${SCHEDULER_NAME}:${slot}:no_show:lease_expired`,
      );
      expect(errorLog).toHaveBeenCalledWith(
        "scheduler run lease expired",
        expect.objectContaining({
          event: "scheduler_run_lease_expired",
          run_id: `${SCHEDULER_NAME}:${slot}:no_show`,
        }),
      );
      expect(errorLog).toHaveBeenCalledWith(
        "scheduler job failed",
        expect.objectContaining({
          event: "scheduler_job_failed",
          failure_code: "lease_expired",
        }),
      );

      // A replayed request returns the stored operation; the alert key stays
      // stable so any downstream re-notify dedupes to the same event.
      const replay = await coordinator.runManual(command, vi.fn());
      expect(replay).toEqual(operation);
      await notifySchedulerFailures(
        replay.result === undefined ? [] : [replay.result],
        config,
        send,
        storage,
      );
      expect(send).toHaveBeenCalledOnce();

      releaseStuck(successOutcome());
      await expect(stuck).resolves.toMatchObject({
        disposition: "fenced",
        ledger: { status: "failed", failureCode: "lease_expired" },
      });
    } finally {
      errorLog.mockRestore();
    }
  });

  it("treats a legacy failed/busy overlap replayed as a duplicate as a nonfailure", async () => {
    const storage = new MemoryStorage();
    const now = scheduledTime(18);
    const coordinator = new SchedulerCoordinator(storage, () => now);
    const legacyBusy: ScheduledRunLedger = {
      version: 1,
      scheduler: SCHEDULER_NAME,
      runId: SchedulerCoordinator.runId("delivery", now),
      runKey: SchedulerCoordinator.runKey("delivery", now),
      cron: "0 1 * * *",
      job: "delivery",
      scheduledTime: now,
      fenceToken: 0,
      status: "failed",
      startedAt: now,
      finishedAt: now,
      failureCode: "busy",
    };
    const genuineFailure: ScheduledRunLedger = {
      ...legacyBusy,
      runId: SchedulerCoordinator.runId("no_show", now),
      runKey: SchedulerCoordinator.runKey("no_show", now),
      job: "no_show",
      failureCode: "job_failed",
    };
    await storage.put(legacyBusy.runKey, legacyBusy);
    await storage.put(genuineFailure.runKey, genuineFailure);

    const replayed = await coordinator.run("delivery", "0 1 * * *", now, async () =>
      successOutcome(),
    );
    expect(replayed).toMatchObject({
      disposition: "duplicate",
      ledger: { status: "failed", failureCode: "busy" },
    });
    expect(scheduledRunRequiresCronFailure(replayed)).toBe(false);

    const failedReplay = await coordinator.run("no_show", "0 1 * * *", now, async () =>
      successOutcome(),
    );
    expect(scheduledRunRequiresCronFailure(failedReplay)).toBe(true);
  });
});
