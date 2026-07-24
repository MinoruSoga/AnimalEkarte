import {
  SCHEDULED_JOB_LEASE_MS,
  SCHEDULER_NAME,
  jobsForCron,
  type ScheduledJobName,
  type ScheduledJobOutcome,
  type ScheduledJobRequest,
} from "./scheduled-jobs";
import {
  cronForManualSlot,
  failureCodeForOutcome,
  internalFailureLedger,
  logLeaseExpired,
  normalizeControl,
  outcomeLedger,
  reconcilePendingManualOperation,
  sameControlIntent,
  sameManualIntent,
  summarizeActive,
  summarizeLedger,
  summarizeSchedulerOperation,
  terminalManualOperation,
  validateOperationIdentity,
  type ActiveScheduledRun,
  type CoordinatorStorage,
  type CoordinatorTransaction,
  type LatestScheduledRun,
  type RunDisposition,
  type RunFailureCode,
  type ScheduledRunLedger,
  type ScheduledRunResult,
  type SchedulerControl,
  type SchedulerControlCommand,
  type SchedulerControlOperation,
  type SchedulerManualCommand,
  type SchedulerManualDriver,
  type SchedulerManualOperation,
  type SchedulerOperation,
  type SchedulerOperationIndex,
  type SchedulerOpsRateLimitDecision,
  type SchedulerStatus,
} from "./scheduler-coordinator-records";
import {
  pruneSchedulerHistory,
  putSchedulerOperation,
  type SchedulerHistoryConfig,
} from "./scheduler-history";
import {
  consumeSchedulerOpsRateLimit,
  SCHEDULER_OPS_RATE_LIMIT,
  SCHEDULER_OPS_RATE_WINDOW_MS,
} from "./scheduler-rate-limit";

export { SCHEDULED_JOB_LEASE_MS };
export { SCHEDULER_OPS_RATE_LIMIT, SCHEDULER_OPS_RATE_WINDOW_MS };
export * from "./scheduler-coordinator-records";

// The fence protects Durable Object ledger finalization; it cannot undo Go
// side effects that already started. Safety therefore also depends on the
// 100s Go context deadline, the 110s Worker fetch timeout, request cancellation,
// and domain-level idempotency/CAS before this 150s lease can be replaced.
const CONTROL_KEY = "scheduler:control";
const ACTIVE_KEY = "scheduler:active";
const FENCE_KEY = "scheduler:fence";
const RUN_KEY_PREFIX = "scheduler:run:";
const LATEST_KEY_PREFIX = "scheduler:latest:";
const OPERATION_KEY_PREFIX = "scheduler:operation:";
const OPERATION_INDEX_PREFIX = "scheduler:operation-index:";
const OPERATION_RESULT_KEY_PREFIX = "scheduler:operation-result:";
const OPERATION_DRIVER_KEY_PREFIX = "scheduler:operation-driver:";
const MANUAL_DRIVER_LEASE_MS = 2 * SCHEDULED_JOB_LEASE_MS;

const HISTORY_CONFIG: SchedulerHistoryConfig = {
  runKeyPrefix: RUN_KEY_PREFIX,
  operationKeyPrefix: OPERATION_KEY_PREFIX,
  operationIndexPrefix: OPERATION_INDEX_PREFIX,
  operationResultKeyPrefix: OPERATION_RESULT_KEY_PREFIX,
  operationDriverKeyPrefix: OPERATION_DRIVER_KEY_PREFIX,
};

type JobExecutor = (request: ScheduledJobRequest) => Promise<ScheduledJobOutcome>;

interface ManualRunContext {
  requestId: string;
  driverKey: string;
  driverToken: string;
  resultKey: string;
}

type ClaimResult =
  | { disposition: "claimed"; active: ActiveScheduledRun; ledger: ScheduledRunLedger }
  | { disposition: Exclude<RunDisposition, "executed" | "fenced">; ledger?: ScheduledRunLedger; active?: ActiveScheduledRun };

export class SchedulerCoordinator {
  constructor(
    private readonly storage: CoordinatorStorage,
    private readonly clock: () => number = Date.now,
  ) {}
  static runKey(job: ScheduledJobName, scheduledTime: number): string {
    return `${RUN_KEY_PREFIX}${scheduledTime.toString().padStart(13, "0")}:${job}`;
  }

  static runId(job: ScheduledJobName, scheduledTime: number): string {
    return `${SCHEDULER_NAME}:${scheduledTime}:${job}`;
  }

  async getControl(): Promise<SchedulerControl> {
    return normalizeControl(await this.storage.get<SchedulerControl>(CONTROL_KEY));
  }

  async getActive(): Promise<ActiveScheduledRun | undefined> {
    return this.storage.get<ActiveScheduledRun>(ACTIVE_KEY);
  }
  async consumeOpsRateLimit(
    actorPrincipal: string,
    now: number = this.clock(),
  ): Promise<SchedulerOpsRateLimitDecision> {
    return consumeSchedulerOpsRateLimit(this.storage, actorPrincipal, now);
  }

  async getStatus(limit: number): Promise<SchedulerStatus> {
    if (!Number.isSafeInteger(limit) || limit < 1 || limit > 50) {
      throw new Error("status limit must be between 1 and 50");
    }
    await pruneSchedulerHistory(this.storage, this.clock(), HISTORY_CONFIG);
    const [control, active, runMap, operationIndexMap] = await Promise.all([
      this.getControl(),
      this.getActive(),
      this.storage.list<ScheduledRunLedger>({
        prefix: RUN_KEY_PREFIX,
        reverse: true,
        limit,
      }),
      this.storage.list<SchedulerOperationIndex>({
        prefix: OPERATION_INDEX_PREFIX,
        reverse: true,
        limit,
      }),
    ]);
    const recentRuns = [...runMap.values()]
      .sort(
        (left, right) =>
          right.scheduledTime - left.scheduledTime ||
          right.startedAt - left.startedAt ||
          right.runId.localeCompare(left.runId),
      )
      .slice(0, limit)
      .map(summarizeLedger);

    const operations: SchedulerOperation[] = [];
    for (const index of operationIndexMap.values()) {
      const operation = await this.storage.get<SchedulerOperation>(
        `${OPERATION_KEY_PREFIX}${index.requestId}`,
      );
      if (operation === undefined) {
        continue;
      }
      if (operation.kind === "manual_run" && operation.status === "pending") {
        const result = await this.storage.get<SchedulerManualOperation>(
          `${OPERATION_RESULT_KEY_PREFIX}${operation.requestId}`,
        );
        operations.push(result ?? operation);
      } else {
        operations.push(operation);
      }
    }
    const recentOperations = operations
      .sort(
        (left, right) =>
          right.requestedAt - left.requestedAt ||
          right.requestId.localeCompare(left.requestId),
      )
      .map(summarizeSchedulerOperation);

    return {
      version: 1,
      scheduler: SCHEDULER_NAME,
      control,
      ...(active === undefined ? {} : { active: summarizeActive(active) }),
      recentRuns,
      recentOperations,
    };
  }

  async setControl(command: SchedulerControlCommand): Promise<SchedulerControlOperation> {
    validateOperationIdentity(command);
    if (typeof command.paused !== "boolean") {
      throw new Error("paused must be a boolean");
    }
    if (!Number.isSafeInteger(command.expectedRevision) || command.expectedRevision < 0) {
      throw new Error("expectedRevision must be a non-negative integer");
    }
    await pruneSchedulerHistory(this.storage, this.clock(), HISTORY_CONFIG);
    const now = this.clock();
    const operationKey = `${OPERATION_KEY_PREFIX}${command.requestId}`;

    return this.storage.transaction(async (transaction) => {
      const existing = await transaction.get<SchedulerOperation>(operationKey);
      if (existing !== undefined) {
        if (existing.kind !== "control" || !sameControlIntent(existing, command)) {
          throw new Error("request_id_conflict");
        }
        return existing;
      }

      const current = normalizeControl(await transaction.get<SchedulerControl>(CONTROL_KEY));
      if (current.revision !== command.expectedRevision) {
        const rejected: SchedulerControlOperation = {
          version: 1,
          kind: "control",
          requestId: command.requestId,
          actorPrincipal: command.actorPrincipal,
          reason: command.reason.trim(),
          requestedAt: now,
          status: "rejected",
          rejectionCode: "revision_conflict",
          expectedRevision: command.expectedRevision,
          requestedPaused: command.paused,
          control: current,
        };
        await putSchedulerOperation(
          transaction,
          operationKey,
          OPERATION_INDEX_PREFIX,
          rejected,
        );
        return rejected;
      }

      const nextControl: SchedulerControl = {
        version: 1,
        revision: current.revision + 1,
        paused: command.paused,
        changedAt: now,
      };
      const operation: SchedulerControlOperation = {
        version: 1,
        kind: "control",
        requestId: command.requestId,
        actorPrincipal: command.actorPrincipal,
        reason: command.reason.trim(),
        requestedAt: now,
        status: "completed",
        expectedRevision: command.expectedRevision,
        requestedPaused: command.paused,
        control: nextControl,
      };
      await transaction.put(CONTROL_KEY, nextControl);
      await putSchedulerOperation(
        transaction,
        operationKey,
        OPERATION_INDEX_PREFIX,
        operation,
      );
      return operation;
    });
  }

  async runManual(
    command: SchedulerManualCommand,
    execute: JobExecutor,
  ): Promise<SchedulerManualOperation> {
    validateOperationIdentity(command);
    if (command.mode !== "catch_up") {
      throw new Error("manual mode must be catch_up");
    }
    await pruneSchedulerHistory(this.storage, this.clock(), HISTORY_CONFIG);
    const operationKey = `${OPERATION_KEY_PREFIX}${command.requestId}`;
    const resultKey = `${OPERATION_RESULT_KEY_PREFIX}${command.requestId}`;
    const driverKey = `${OPERATION_DRIVER_KEY_PREFIX}${command.requestId}`;
    const runKey = SchedulerCoordinator.runKey(command.job, command.scheduledTime);
    const replay = await this.storage.transaction(async (transaction) => {
      const existing = await transaction.get<SchedulerOperation>(operationKey);
      if (existing === undefined) {
        return undefined;
      }
      if (
        existing.kind !== "manual_run" ||
        !sameManualIntent(existing, command)
      ) {
        throw new Error("request_id_conflict");
      }
      const result = await transaction.get<SchedulerManualOperation>(resultKey);
      if (result !== undefined) {
        await transaction.delete(driverKey);
        return result;
      }
      if (existing.status !== "pending") {
        return existing;
      }
      const recovered = reconcilePendingManualOperation(
        existing,
        await transaction.get<ScheduledRunLedger>(runKey),
      );
      if (recovered !== undefined) {
        await transaction.put(resultKey, recovered);
        await transaction.delete(driverKey);
      }
      return recovered;
    });
    if (replay !== undefined) {
      return replay;
    }
    const now = this.clock();
    const cron = cronForManualSlot(command.job, command.scheduledTime, now);
    const driverToken = crypto.randomUUID();
    const begin = await this.storage.transaction<
      | { kind: "existing"; operation: SchedulerManualOperation }
      | { kind: "pending"; operation: SchedulerManualOperation }
    >(async (transaction) => {
      const transactionNow = this.clock();
      const existing = await transaction.get<SchedulerOperation>(operationKey);
      if (existing !== undefined) {
        if (
          existing.kind !== "manual_run" ||
          !sameManualIntent(existing, command)
        ) {
          throw new Error("request_id_conflict");
        }
        const result = await transaction.get<SchedulerManualOperation>(resultKey);
        if (result !== undefined) {
          await transaction.delete(driverKey);
          return { kind: "existing", operation: result };
        }
        if (existing.status !== "pending") {
          return { kind: "existing", operation: existing };
        }
        const recovered = reconcilePendingManualOperation(
          existing,
          await transaction.get<ScheduledRunLedger>(runKey),
        );
        if (recovered !== undefined) {
          await transaction.put(resultKey, recovered);
          await transaction.delete(driverKey);
          return { kind: "existing", operation: recovered };
        }
        const driver = await transaction.get<SchedulerManualDriver>(driverKey);
        if (driver !== undefined && driver.leaseExpiresAt > transactionNow) {
          return { kind: "existing", operation: existing };
        }
        await transaction.put(driverKey, {
          version: 1,
          token: driverToken,
          leaseExpiresAt: transactionNow + MANUAL_DRIVER_LEASE_MS,
        } satisfies SchedulerManualDriver);
        return { kind: "pending", operation: existing };
      }

      const base: Omit<SchedulerManualOperation, "status"> = {
        version: 1,
        kind: "manual_run",
        requestId: command.requestId,
        actorPrincipal: command.actorPrincipal,
        reason: command.reason.trim(),
        requestedAt: now,
        mode: command.mode,
        job: command.job,
        cron,
        scheduledTime: command.scheduledTime,
      };
      if ((await transaction.get<ScheduledRunLedger>(runKey)) !== undefined) {
        const rejected: SchedulerManualOperation = {
          ...base,
          status: "rejected",
          rejectionCode: "slot_already_recorded",
        };
        await putSchedulerOperation(
          transaction,
          operationKey,
          OPERATION_INDEX_PREFIX,
          rejected,
        );
        return { kind: "existing", operation: rejected };
      }
      const control = normalizeControl(await transaction.get<SchedulerControl>(CONTROL_KEY));
      if (control.paused) {
        const rejected: SchedulerManualOperation = {
          ...base,
          status: "rejected",
          rejectionCode: "scheduler_paused",
        };
        await putSchedulerOperation(
          transaction,
          operationKey,
          OPERATION_INDEX_PREFIX,
          rejected,
        );
        return { kind: "existing", operation: rejected };
      }
      const active = await transaction.get<ActiveScheduledRun>(ACTIVE_KEY);
      if (active !== undefined && active.leaseExpiresAt > transactionNow) {
        const rejected: SchedulerManualOperation = {
          ...base,
          status: "rejected",
          rejectionCode: "scheduler_busy",
        };
        await putSchedulerOperation(
          transaction,
          operationKey,
          OPERATION_INDEX_PREFIX,
          rejected,
        );
        return { kind: "existing", operation: rejected };
      }

      const pending: SchedulerManualOperation = {
        ...base,
        status: "pending",
      };
      await putSchedulerOperation(
        transaction,
        operationKey,
        OPERATION_INDEX_PREFIX,
        pending,
      );
      await transaction.put(driverKey, {
        version: 1,
        token: driverToken,
        leaseExpiresAt: transactionNow + MANUAL_DRIVER_LEASE_MS,
      } satisfies SchedulerManualDriver);
      return { kind: "pending", operation: pending };
    });

    if (begin.kind === "existing") {
      return begin.operation;
    }

    const result = await this.run(
      command.job,
      cron,
      command.scheduledTime,
      execute,
      {
        requestId: command.requestId,
        driverKey,
        driverToken,
        resultKey,
      },
    );
    const recovered = reconcilePendingManualOperation(
      begin.operation,
      result.ledger,
    );
    if (
      result.disposition === "duplicate" &&
      result.ledger?.status === "running" &&
      result.ledger.manualRequestId === command.requestId
    ) {
      return begin.operation;
    }
    const terminal =
      recovered ?? terminalManualOperation(begin.operation, result);
    return this.storage.transaction(async (transaction) => {
      const existingResult = await transaction.get<SchedulerManualOperation>(resultKey);
      if (existingResult !== undefined) {
        await transaction.delete(driverKey);
        return existingResult;
      }
      const driver = await transaction.get<SchedulerManualDriver>(driverKey);
      if (
        driver?.token !== driverToken ||
        driver.leaseExpiresAt <= this.clock()
      ) {
        return begin.operation;
      }
      await transaction.put(resultKey, terminal);
      await transaction.delete(driverKey);
      return terminal;
    });
  }

  async run(
    job: ScheduledJobName,
    cron: string,
    scheduledTime: number,
    execute: JobExecutor,
    manual?: ManualRunContext,
  ): Promise<ScheduledRunResult> {
    this.validateInvocation(job, cron, scheduledTime);
    await pruneSchedulerHistory(this.storage, this.clock(), HISTORY_CONFIG);

    const claim = await this.claim(
      job,
      cron,
      scheduledTime,
      manual,
    );
    if (claim.disposition !== "claimed") {
      return claim;
    }

    let outcome: ScheduledJobOutcome | undefined;
    let failureCode: RunFailureCode | undefined;
    try {
      outcome = await execute({
        scheduler: SCHEDULER_NAME,
        job,
        scheduled_time: scheduledTime,
        run_id: claim.active.runId,
        fence_token: claim.active.fenceToken,
      });
      failureCode = failureCodeForOutcome(outcome);
    } catch {
      failureCode = "transport";
    }

    const finalized = await this.finalize(claim.active, claim.ledger, outcome, failureCode);
    return finalized;
  }

  private validateInvocation(job: ScheduledJobName, cron: string, scheduledTime: number): void {
    const jobs = jobsForCron(cron);
    if (!jobs.includes(job)) {
      throw new Error(`job ${job} is not allowed for cron ${cron}`);
    }
    if (!Number.isSafeInteger(scheduledTime) || scheduledTime <= 0) {
      throw new Error("scheduledTime must be a positive integer");
    }
  }

  private async claim(
    job: ScheduledJobName,
    cron: string,
    scheduledTime: number,
    manual?: ManualRunContext,
  ): Promise<ClaimResult> {
    const runKey = SchedulerCoordinator.runKey(job, scheduledTime);
    const runId = SchedulerCoordinator.runId(job, scheduledTime);

    return this.storage.transaction(async (transaction) => {
      const now = this.clock();
      if (manual !== undefined) {
        const driver = await transaction.get<SchedulerManualDriver>(
          manual.driverKey,
        );
        const result = await transaction.get<SchedulerManualOperation>(
          manual.resultKey,
        );
        if (
          result !== undefined ||
          driver?.token !== manual.driverToken ||
          driver.leaseExpiresAt <= now
        ) {
          return { disposition: "duplicate" };
        }
        await transaction.put(manual.driverKey, {
          version: 1,
          token: manual.driverToken,
          leaseExpiresAt: now + MANUAL_DRIVER_LEASE_MS,
        } satisfies SchedulerManualDriver);
      }
      const existing = await transaction.get<ScheduledRunLedger>(runKey);
      const currentActive = await transaction.get<ActiveScheduledRun>(ACTIVE_KEY);
      if (existing !== undefined) {
        const reconciled = await this.reconcileDuplicate(
          transaction,
          existing,
          currentActive,
          now,
        );
        return { disposition: "duplicate", ledger: reconciled };
      }

      const latest = await transaction.get<LatestScheduledRun>(`${LATEST_KEY_PREFIX}${job}`);
      if (
        manual === undefined &&
        latest !== undefined &&
        scheduledTime < latest.scheduledTime
      ) {
        const staleLedger: ScheduledRunLedger = {
          version: 1,
          scheduler: SCHEDULER_NAME,
          runId,
          runKey,
          cron,
          job,
          scheduledTime,
          fenceToken: 0,
          status: "failed",
          startedAt: now,
          finishedAt: now,
          failureCode: "stale",
        };
        await transaction.put(runKey, staleLedger);
        return { disposition: "stale", ledger: staleLedger };
      }

      const putLatest = async () => {
        if (latest === undefined || scheduledTime >= latest.scheduledTime) {
          await transaction.put(`${LATEST_KEY_PREFIX}${job}`, {
            version: 1,
            job,
            scheduledTime,
            runId,
            updatedAt: now,
          } satisfies LatestScheduledRun);
        }
      };
      const control = await transaction.get<SchedulerControl>(CONTROL_KEY);
      if (control?.paused === true) {
        if (manual !== undefined) {
          return { disposition: "paused" };
        }
        const pausedLedger: ScheduledRunLedger = {
          version: 1,
          scheduler: SCHEDULER_NAME,
          runId,
          runKey,
          cron,
          job,
          scheduledTime,
          fenceToken: 0,
          status: "paused",
          startedAt: now,
          finishedAt: now,
        };
        await transaction.put(runKey, pausedLedger);
        await putLatest();
        return { disposition: "paused", ledger: pausedLedger };
      }

      if (currentActive !== undefined) {
        if (currentActive.leaseExpiresAt > now) {
          if (manual !== undefined) {
            return { disposition: "busy", active: currentActive };
          }
          const busyLedger: ScheduledRunLedger = {
            version: 1,
            scheduler: SCHEDULER_NAME,
            runId,
            runKey,
            cron,
            job,
            scheduledTime,
            fenceToken: 0,
            status: "failed",
            startedAt: now,
            finishedAt: now,
            failureCode: "busy",
          };
          await transaction.put(runKey, busyLedger);
          await putLatest();
          return { disposition: "busy", ledger: busyLedger, active: currentActive };
        }
        await this.expireActive(transaction, currentActive, now);
      }

      const previousFence = (await transaction.get<number>(FENCE_KEY)) ?? 0;
      if (!Number.isSafeInteger(previousFence) || previousFence < 0) {
        throw new Error("scheduler fence state is invalid");
      }
      const fenceToken = previousFence + 1;
      if (!Number.isSafeInteger(fenceToken)) {
        throw new Error("scheduler fence token exhausted");
      }

      const active: ActiveScheduledRun = {
        version: 1,
        scheduler: SCHEDULER_NAME,
        runId,
        runKey,
        cron,
        job,
        scheduledTime,
        fenceToken,
        claimedAt: now,
        leaseExpiresAt: now + SCHEDULED_JOB_LEASE_MS,
      };
      const ledger: ScheduledRunLedger = {
        version: 1,
        scheduler: SCHEDULER_NAME,
        runId,
        runKey,
        cron,
        job,
        scheduledTime,
        ...(manual === undefined ? {} : { manualRequestId: manual.requestId }),
        fenceToken,
        status: "running",
        startedAt: now,
      };
      await transaction.put(FENCE_KEY, fenceToken);
      await transaction.put(ACTIVE_KEY, active);
      await transaction.put(runKey, ledger);
      await putLatest();
      return { disposition: "claimed", active, ledger };
    });
  }
  private async reconcileDuplicate(
    transaction: CoordinatorTransaction,
    existing: ScheduledRunLedger,
    active: ActiveScheduledRun | undefined,
    now: number,
  ): Promise<ScheduledRunLedger> {
    if (existing.status !== "running") {
      return existing;
    }
    if (
      active !== undefined &&
      active.runId === existing.runId &&
      active.fenceToken === existing.fenceToken &&
      active.leaseExpiresAt > now
    ) {
      return existing;
    }

    const failed = internalFailureLedger(existing, now, "lease_expired");
    await transaction.put(existing.runKey, failed);
    logLeaseExpired(failed);
    if (active?.runId === existing.runId) {
      await transaction.delete(ACTIVE_KEY);
    }
    return failed;
  }

  private async expireActive(
    transaction: CoordinatorTransaction,
    active: ActiveScheduledRun,
    now: number,
  ): Promise<void> {
    const ledger = await transaction.get<ScheduledRunLedger>(active.runKey);
    if (ledger?.status === "running") {
      const failed = internalFailureLedger(ledger, now, "lease_expired");
      await transaction.put(active.runKey, failed);
      logLeaseExpired(failed);
    }
    await transaction.delete(ACTIVE_KEY);
  }

  private async finalize(
    active: ActiveScheduledRun,
    claimedLedger: ScheduledRunLedger,
    outcome: ScheduledJobOutcome | undefined,
    failureCode: RunFailureCode | undefined,
  ): Promise<ScheduledRunResult> {
    const now = this.clock();
    return this.storage.transaction(async (transaction) => {
      const currentActive = await transaction.get<ActiveScheduledRun>(ACTIVE_KEY);
      const currentLedger =
        (await transaction.get<ScheduledRunLedger>(claimedLedger.runKey)) ?? claimedLedger;
      const ownsLease =
        currentActive?.runId === active.runId &&
        currentActive.fenceToken === active.fenceToken &&
        currentActive.leaseExpiresAt > now;

      if (!ownsLease) {
        const fencedLedger =
          currentLedger.status === "running"
            ? internalFailureLedger(currentLedger, now, "fenced")
            : currentLedger;
        if (currentLedger.status === "running") {
          await transaction.put(claimedLedger.runKey, fencedLedger);
        }
        if (
          currentActive?.runId === active.runId &&
          currentActive.fenceToken === active.fenceToken
        ) {
          await transaction.delete(ACTIVE_KEY);
        }
        return { disposition: "fenced", ledger: fencedLedger };
      }

      const finalized =
        outcome === undefined
          ? internalFailureLedger(currentLedger, now, failureCode ?? "transport")
          : outcomeLedger(currentLedger, now, outcome, failureCode);
      await transaction.put(claimedLedger.runKey, finalized);
      await transaction.delete(ACTIVE_KEY);
      return { disposition: "executed", ledger: finalized };
    });
  }

}

export async function runScheduledPlan(
  coordinator: SchedulerCoordinator,
  cron: string,
  scheduledTime: number,
  execute: JobExecutor,
): Promise<readonly ScheduledRunResult[]> {
  let results: readonly ScheduledRunResult[] = [];
  for (const job of jobsForCron(cron)) {
    results = [...results, await coordinator.run(job, cron, scheduledTime, execute)];
  }
  return results;
}

export function scheduledRunRequiresCronFailure(result: ScheduledRunResult): boolean {
  if (result.disposition === "paused") {
    return false;
  }
  if (result.disposition === "duplicate") {
    return result.ledger?.status === "partial" || result.ledger?.status === "failed";
  }
  if (result.disposition !== "executed") {
    return true;
  }
  return result.ledger?.status === "partial" || result.ledger?.status === "failed";
}
