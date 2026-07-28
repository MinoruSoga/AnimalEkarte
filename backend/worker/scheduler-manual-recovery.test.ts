import { describe, expect, it, vi } from "vitest";
import {
  RUN_LEDGER_RETENTION_MS,
  SCHEDULED_JOB_LEASE_MS,
  SchedulerCoordinator,
  type CoordinatorStorage,
  type CoordinatorTransaction,
  type LatestScheduledRun,
} from "./scheduler-coordinator";
import type { ScheduledJobOutcome } from "./scheduled-jobs";

class RecoveryStorage implements CoordinatorStorage {
  readonly values = new Map<string, unknown>();
  afterPendingOperationCommit?: () => void;
  beforeNextOperationResultRead?: () => Promise<void>;
  beforeNextTransaction?: () => Promise<void>;
  failAfterPendingOperationCommit = false;
  failNextOperationResultPut = false;
  injectOperationResultOnRead?: { key: string; value: unknown };

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
      .filter(([key]) => key.startsWith(options.prefix))
      .filter(([key]) =>
        options.startAfter === undefined ? true : key > options.startAfter
      )
      .sort(([left], [right]) => left.localeCompare(right));
    const ordered = options.reverse ? entries.reverse() : entries;
    return new Map(ordered.slice(0, options.limit)) as Map<string, T>;
  }

  async transaction<T>(
    closure: (transaction: CoordinatorTransaction) => Promise<T>,
  ): Promise<T> {
    if (this.beforeNextTransaction !== undefined) {
      const beforeTransaction = this.beforeNextTransaction;
      this.beforeNextTransaction = undefined;
      await beforeTransaction();
    }
    const snapshot = new Map(this.values);
    const result = await closure({
      get: async <V>(key: string) => {
        if (
          key.startsWith("scheduler:operation-result:") &&
          this.beforeNextOperationResultRead !== undefined
        ) {
          const beforeRead = this.beforeNextOperationResultRead;
          this.beforeNextOperationResultRead = undefined;
          await beforeRead();
        }
        if (this.injectOperationResultOnRead?.key === key) {
          snapshot.set(
            key,
            structuredClone(this.injectOperationResultOnRead.value),
          );
          this.injectOperationResultOnRead = undefined;
        }
        return snapshot.get(key) as V | undefined;
      },
      put: async <V>(key: string, value: V) => {
        if (
          this.failNextOperationResultPut &&
          key.startsWith("scheduler:operation-result:")
        ) {
          this.failNextOperationResultPut = false;
          throw new Error("injected operation result persistence failure");
        }
        snapshot.set(key, structuredClone(value));
      },
      delete: async (key: string) => snapshot.delete(key),
    });
    this.values.clear();
    for (const [key, value] of snapshot) {
      this.values.set(key, value);
    }
    const committedPendingWithoutRun =
      [...snapshot.entries()].some(
        ([key, value]) =>
          key.startsWith("scheduler:operation:") &&
          (value as { status?: string }).status === "pending",
      ) &&
      ![...snapshot.keys()].some((key) => key.startsWith("scheduler:run:"));
    if (this.failAfterPendingOperationCommit && committedPendingWithoutRun) {
      this.failAfterPendingOperationCommit = false;
      throw new Error("injected post-pending crash");
    }
    if (
      this.afterPendingOperationCommit !== undefined &&
      committedPendingWithoutRun
    ) {
      const afterCommit = this.afterPendingOperationCommit;
      this.afterPendingOperationCommit = undefined;
      afterCommit();
    }
    return result;
  }
}

function successOutcome(): ScheduledJobOutcome {
  return { outcome: "success", processed: 1, succeeded: 1, failed: 0 };
}

describe("SchedulerCoordinator manual recovery", () => {
  it("executes an older missing slot without moving the latest scheduled watermark backwards", async () => {
    const storage = new RecoveryStorage();
    const latestTime = Date.UTC(2026, 6, 24, 11);
    const missingTime = Date.UTC(2026, 6, 24, 6);
    const coordinator = new SchedulerCoordinator(storage, () => latestTime + 1_000);
    const execute = vi.fn(async () => successOutcome());

    await coordinator.run("no_show", "0 6,11 * * *", latestTime, execute);
    const operation = await coordinator.runManual(
      {
        actorPrincipal: "scheduler-ops-secret-v1",
        job: "no_show",
        mode: "catch_up",
        reason: "INC-140 recover earlier missing slot",
        requestId: "bb9a61f2-2508-4bbc-bc43-b12de289bb07",
        scheduledTime: missingTime,
      },
      execute,
    );

    expect(operation).toMatchObject({
      status: "completed",
      result: {
        disposition: "executed",
        ledger: { scheduledTime: missingTime, status: "success" },
      },
    });
    expect(execute).toHaveBeenCalledTimes(2);
    await expect(
      storage.get<LatestScheduledRun>("scheduler:latest:no_show"),
    ).resolves.toMatchObject({ scheduledTime: latestTime });
  });

  it("reconciles a pending manual audit from its terminal run ledger on retry", async () => {
    const storage = new RecoveryStorage();
    const scheduledTime = Date.UTC(2026, 6, 23, 17);
    const requestId = "de3a2410-0b43-4775-9b8b-b4e3de5a47e8";
    const runKey = SchedulerCoordinator.runKey("dormant", scheduledTime);
    const command = {
      requestId,
      actorPrincipal: "scheduler-ops-secret-v1",
      reason: "INC-141 retry interrupted catch-up",
      mode: "catch_up",
      job: "dormant",
      scheduledTime,
    } as const;
    const execute = vi.fn(async () => successOutcome());
    let now = scheduledTime + 1_000;
    const coordinator = new SchedulerCoordinator(
      storage,
      () => now,
    );
    storage.failNextOperationResultPut = true;

    await expect(coordinator.runManual(command, execute)).rejects.toThrow(
      "injected operation result persistence failure",
    );
    await expect(storage.get(runKey)).resolves.toMatchObject({
      status: "success",
      manualRequestId: requestId,
    });
    await expect(
      storage.get(`scheduler:operation:${requestId}`),
    ).resolves.toMatchObject({ status: "pending" });
    await expect(
      storage.get(`scheduler:operation-result:${requestId}`),
    ).resolves.toBeUndefined();

    const recovered = await coordinator.runManual(command, execute);

    expect(recovered).toMatchObject({
      status: "completed",
      result: {
        disposition: "executed",
        ledger: { status: "success" },
      },
    });
    expect(execute).toHaveBeenCalledTimes(1);
    await expect(
      storage.get(`scheduler:operation-result:${requestId}`),
    ).resolves.toEqual(recovered);
  });

  it("resumes a pending manual request that crashed before claiming its run", async () => {
    const storage = new RecoveryStorage();
    const scheduledTime = Date.UTC(2026, 6, 22, 17);
    const requestId = "985b5df1-fc99-46f4-8831-74652c18b026";
    const command = {
      actorPrincipal: "scheduler-ops-secret-v1",
      job: "dormant",
      mode: "catch_up",
      reason: "INC-142 resume pre-claim catch-up",
      requestId,
      scheduledTime,
    } as const;
    const execute = vi.fn(async () => successOutcome());
    let now = scheduledTime + 1_000;
    const coordinator = new SchedulerCoordinator(
      storage,
      () => now,
    );
    storage.failAfterPendingOperationCommit = true;

    await expect(coordinator.runManual(command, execute)).rejects.toThrow(
      "injected post-pending crash",
    );
    await expect(
      storage.get(`scheduler:operation:${requestId}`),
    ).resolves.toMatchObject({ status: "pending" });
    await expect(
      storage.get(SchedulerCoordinator.runKey("dormant", scheduledTime)),
    ).resolves.toBeUndefined();

    now += 2 * SCHEDULED_JOB_LEASE_MS + 1;
    await expect(coordinator.runManual(command, execute)).resolves.toMatchObject({
      status: "completed",
      result: { disposition: "executed", ledger: { status: "success" } },
    });
    expect(execute).toHaveBeenCalledTimes(1);
  });

  it("keeps a concurrent retry pending while the same manual request owns the lease", async () => {
    const storage = new RecoveryStorage();
    const scheduledTime = Date.UTC(2026, 6, 20, 17);
    const requestId = "9704894d-64dc-449d-a0e3-287b1b4ec78c";
    const command = {
      actorPrincipal: "scheduler-ops-secret-v1",
      job: "dormant",
      mode: "catch_up",
      reason: "INC-145 concurrent retry",
      requestId,
      scheduledTime,
    } as const;
    let release: (outcome: ScheduledJobOutcome) => void = () => undefined;
    const outcome = new Promise<ScheduledJobOutcome>((resolve) => {
      release = resolve;
    });
    const execute = vi.fn(async () => outcome);
    const coordinator = new SchedulerCoordinator(
      storage,
      () => scheduledTime + 1_000,
    );

    const first = coordinator.runManual(command, execute);
    await vi.waitFor(async () => {
      await expect(storage.get("scheduler:active")).resolves.toMatchObject({
        runId: SchedulerCoordinator.runId("dormant", scheduledTime),
      });
    });
    const concurrent = await coordinator.runManual(command, execute);

    expect(concurrent).toMatchObject({ status: "pending" });
    expect(concurrent).not.toHaveProperty("result");
    await expect(
      storage.get(`scheduler:operation-result:${requestId}`),
    ).resolves.toBeUndefined();
    expect(execute).toHaveBeenCalledTimes(1);

    release(successOutcome());
    await expect(first).resolves.toMatchObject({
      status: "completed",
      result: { disposition: "executed", ledger: { status: "success" } },
    });
  });

  it("prevents a stale driver from executing after a takeover terminalizes", async () => {
    const storage = new RecoveryStorage();
    const scheduledTime = Date.UTC(2026, 6, 19, 17);
    let now = scheduledTime + 1_000;
    const requestId = "bd725295-4df6-4f5f-8fb7-aa19d0a00236";
    const command = {
      actorPrincipal: "scheduler-ops-secret-v1",
      job: "dormant",
      mode: "catch_up",
      reason: "INC-148 stale driver takeover",
      requestId,
      scheduledTime,
    } as const;
    let reachedStall: () => void = () => undefined;
    const stallReached = new Promise<void>((resolve) => {
      reachedStall = resolve;
    });
    let releaseStall: () => void = () => undefined;
    const stallReleased = new Promise<void>((resolve) => {
      releaseStall = resolve;
    });
    storage.afterPendingOperationCommit = () => {
      storage.beforeNextTransaction = async () => {
        reachedStall();
        await stallReleased;
      };
    };
    const execute = vi.fn(async () => successOutcome());
    const coordinator = new SchedulerCoordinator(storage, () => now);

    const staleDriver = coordinator.runManual(command, execute);
    await stallReached;
    now += 2 * SCHEDULED_JOB_LEASE_MS + 1;
    storage.values.set("scheduler:control", {
      version: 1,
      revision: 1,
      paused: true,
      changedAt: now,
    });
    const takeover = await coordinator.runManual(command, execute);

    expect(takeover).toMatchObject({
      status: "rejected",
      rejectionCode: "scheduler_paused",
    });
    releaseStall();
    await expect(staleDriver).resolves.toEqual(takeover);
    expect(execute).not.toHaveBeenCalled();
    await expect(
      storage.get(SchedulerCoordinator.runKey("dormant", scheduledTime)),
    ).resolves.toBeUndefined();
  });

  it("keeps an exactly expired driver pending without recording a false rejection", async () => {
    const storage = new RecoveryStorage();
    const scheduledTime = Date.UTC(2026, 6, 17, 17);
    let now = scheduledTime + 1_000;
    const requestId = "b75aef29-fb55-4ee7-87ed-6c4d0e20c488";
    const command = {
      actorPrincipal: "scheduler-ops-secret-v1",
      job: "dormant",
      mode: "catch_up",
      reason: "INC-149 driver lease boundary",
      requestId,
      scheduledTime,
    } as const;
    let reachedStall: () => void = () => undefined;
    const stallReached = new Promise<void>((resolve) => {
      reachedStall = resolve;
    });
    let releaseStall: () => void = () => undefined;
    const stallReleased = new Promise<void>((resolve) => {
      releaseStall = resolve;
    });
    storage.afterPendingOperationCommit = () => {
      storage.beforeNextTransaction = async () => {
        reachedStall();
        await stallReleased;
      };
    };
    const execute = vi.fn(async () => successOutcome());
    const coordinator = new SchedulerCoordinator(storage, () => now);

    const expiredDriver = coordinator.runManual(command, execute);
    await stallReached;
    now += 2 * SCHEDULED_JOB_LEASE_MS;
    releaseStall();
    await expect(expiredDriver).resolves.toMatchObject({ status: "pending" });
    expect(execute).not.toHaveBeenCalled();
    await expect(
      storage.get(`scheduler:operation-result:${requestId}`),
    ).resolves.toBeUndefined();
    await expect(
      storage.get(SchedulerCoordinator.runKey("dormant", scheduledTime)),
    ).resolves.toBeUndefined();

    await expect(coordinator.runManual(command, execute)).resolves.toMatchObject({
      status: "completed",
      result: { disposition: "executed", ledger: { status: "success" } },
    });
    expect(execute).toHaveBeenCalledTimes(1);
  });

  it("returns the persisted terminal winner from a concurrent idempotent completion", async () => {
    const storage = new RecoveryStorage();
    const scheduledTime = Date.UTC(2026, 6, 19, 17);
    const now = scheduledTime + 1_000;
    const requestId = "e24a1c57-b26f-4262-ae92-b7676c03296f";
    const command = {
      actorPrincipal: "scheduler-ops-secret-v1",
      job: "dormant",
      mode: "catch_up",
      reason: "INC-146 concurrent terminal winner",
      requestId,
      scheduledTime,
    } as const;
    const runKey = SchedulerCoordinator.runKey("dormant", scheduledTime);
    const completed = {
      version: 1,
      kind: "manual_run",
      ...command,
      cron: "0 17 * * *",
      requestedAt: now,
      status: "completed",
      result: {
        disposition: "executed",
        ledger: {
          version: 1,
          scheduler: "animalekarte-scheduler-v1",
          runId: SchedulerCoordinator.runId("dormant", scheduledTime),
          runKey,
          cron: "0 17 * * *",
          job: "dormant",
          scheduledTime,
          manualRequestId: requestId,
          fenceToken: 11,
          status: "success",
          startedAt: now,
          finishedAt: now,
          outcome: successOutcome(),
        },
      },
    } as const;
    storage.afterPendingOperationCommit = () => {
      storage.values.set("scheduler:control", {
        version: 1,
        revision: 1,
        paused: true,
        changedAt: now,
      });
    };
    storage.injectOperationResultOnRead = {
      key: `scheduler:operation-result:${requestId}`,
      value: completed,
    };
    const coordinator = new SchedulerCoordinator(storage, () => now);

    const observed = await coordinator.runManual(
      command,
      vi.fn(async () => successOutcome()),
    );

    expect(observed).toEqual(completed);
    await expect(
      storage.get(`scheduler:operation-result:${requestId}`),
    ).resolves.toEqual(completed);
  });

  it("replays a terminal result after the 35-day execution window", async () => {
    const storage = new RecoveryStorage();
    const scheduledTime = Date.UTC(2026, 5, 1, 17);
    let now = scheduledTime + 1_000;
    const command = {
      actorPrincipal: "scheduler-ops-secret-v1",
      job: "dormant",
      mode: "catch_up",
      reason: "INC-150 retained idempotent replay",
      requestId: "17dd1674-45ec-4e16-b4ec-9718dbd43e77",
      scheduledTime,
    } as const;
    const execute = vi.fn(async () => successOutcome());
    const coordinator = new SchedulerCoordinator(storage, () => now);

    const first = await coordinator.runManual(command, execute);
    now += RUN_LEDGER_RETENTION_MS + 1;
    const replay = await coordinator.runManual(command, execute);

    expect(replay).toEqual(first);
    expect(execute).toHaveBeenCalledTimes(1);
  });

  it("allows only one same-request driver across a paused race", async () => {
    const storage = new RecoveryStorage();
    const scheduledTime = Date.UTC(2026, 6, 18, 17);
    const now = scheduledTime + 1_000;
    const requestId = "1351365d-1f3e-43a4-a3e8-4051d95da47b";
    const command = {
      actorPrincipal: "scheduler-ops-secret-v1",
      job: "dormant",
      mode: "catch_up",
      reason: "INC-147 single manual driver",
      requestId,
      scheduledTime,
    } as const;
    let reachedResultRead: () => void = () => undefined;
    const resultReadReached = new Promise<void>((resolve) => {
      reachedResultRead = resolve;
    });
    let releaseResultRead: () => void = () => undefined;
    const resultReadReleased = new Promise<void>((resolve) => {
      releaseResultRead = resolve;
    });
    storage.afterPendingOperationCommit = () => {
      storage.values.set("scheduler:control", {
        version: 1,
        revision: 1,
        paused: true,
        changedAt: now,
      });
    };
    storage.beforeNextOperationResultRead = async () => {
      reachedResultRead();
      await resultReadReleased;
    };
    const execute = vi.fn(async () => successOutcome());
    const coordinator = new SchedulerCoordinator(storage, () => now);

    const first = coordinator.runManual(command, execute);
    await resultReadReached;
    storage.values.delete("scheduler:control");
    const concurrent = await coordinator.runManual(command, execute);

    expect(concurrent).toMatchObject({ status: "pending" });
    expect(execute).not.toHaveBeenCalled();

    releaseResultRead();
    await expect(first).resolves.toMatchObject({
      status: "rejected",
      rejectionCode: "scheduler_paused",
      result: { disposition: "paused" },
    });
    await expect(
      storage.get(`scheduler:operation-result:${requestId}`),
    ).resolves.toMatchObject({
      status: "rejected",
      rejectionCode: "scheduler_paused",
    });
  });

  it.each([
    {
      disposition: "paused",
      rejectionCode: "scheduler_paused",
      retryRequestId: "c456c15b-1d80-4ae7-81d6-a9c371f129bf",
      arrange(storage: RecoveryStorage, now: number) {
        storage.values.set("scheduler:control", {
          version: 1,
          revision: 1,
          paused: true,
          changedAt: now,
        });
      },
      clear(storage: RecoveryStorage) {
        storage.values.delete("scheduler:control");
      },
      maintain() {},
    },
    {
      disposition: "busy",
      rejectionCode: "scheduler_busy",
      retryRequestId: "b7a07f00-31b9-4e9e-96d6-3c355af4c65c",
      arrange(storage: RecoveryStorage, now: number) {
        storage.values.set("scheduler:active", {
          version: 1,
          scheduler: "animalekarte-scheduler-v1",
          runId: "animalekarte-scheduler-v1:1784700000000:no_show",
          runKey: "scheduler:run:1784700000000:no_show",
          cron: "0 1 * * *",
          job: "no_show",
          scheduledTime: 1_784_700_000_000,
          fenceToken: 91,
          claimedAt: now,
          leaseExpiresAt: now + SCHEDULED_JOB_LEASE_MS,
        });
      },
      clear(storage: RecoveryStorage) {
        storage.values.delete("scheduler:active");
      },
      maintain(storage: RecoveryStorage, now: number) {
        const active = storage.values.get("scheduler:active") as Record<
          string,
          unknown
        >;
        storage.values.set("scheduler:active", {
          ...active,
          leaseExpiresAt: now + SCHEDULED_JOB_LEASE_MS,
        });
      },
    },
  ] as const)(
    "reconciles a $disposition race without burning the missing slot",
    async ({
      disposition,
      rejectionCode,
      retryRequestId,
      arrange,
      clear,
      maintain,
    }) => {
      const storage = new RecoveryStorage();
      const scheduledTime = Date.UTC(2026, 6, 21, 17);
      const requestId =
        disposition === "paused"
          ? "0ad671ec-1816-4c9a-9436-7a174a5ad6ea"
          : "87dc25b6-c2d4-4bdd-b826-c3e69de8e1f7";
      const command = {
        actorPrincipal: "scheduler-ops-secret-v1",
        job: "dormant",
        mode: "catch_up",
        reason: `INC-143 recover ${disposition} race`,
        requestId,
        scheduledTime,
      } as const;
      let now = scheduledTime + 1_000;
      const coordinator = new SchedulerCoordinator(storage, () => now);
      const execute = vi.fn(async () => successOutcome());
      storage.afterPendingOperationCommit = () => arrange(storage, now);
      storage.failNextOperationResultPut = true;

      await expect(coordinator.runManual(command, execute)).rejects.toThrow(
        "injected operation result persistence failure",
      );
      await expect(
        storage.get(SchedulerCoordinator.runKey("dormant", scheduledTime)),
      ).resolves.toBeUndefined();
      now += 2 * SCHEDULED_JOB_LEASE_MS + 1;
      maintain(storage, now);
      const recovered = await coordinator.runManual(command, execute);

      expect(recovered).toMatchObject({
        status: "rejected",
        rejectionCode,
        result: { disposition },
      });
      expect(execute).not.toHaveBeenCalled();

      clear(storage);
      await expect(
        coordinator.runManual(
          {
            ...command,
            requestId: retryRequestId,
            reason: `INC-144 retry after ${disposition} clears`,
          },
          execute,
        ),
      ).resolves.toMatchObject({
        status: "completed",
        result: { disposition: "executed", ledger: { status: "success" } },
      });
      expect(execute).toHaveBeenCalledTimes(1);
    },
  );
});
