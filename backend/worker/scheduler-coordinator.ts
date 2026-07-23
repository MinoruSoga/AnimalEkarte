import {
  SCHEDULED_JOB_LEASE_MS,
  SCHEDULER_NAME,
  jobsForCron,
  type ScheduledJobName,
  type ScheduledJobOutcome,
  type ScheduledJobRequest,
} from "./scheduled-jobs";

export { SCHEDULED_JOB_LEASE_MS };

const DAY_MS = 24 * 60 * 60 * 1_000;
export const RUN_LEDGER_RETENTION_MS = 35 * DAY_MS;

// The fence protects Durable Object ledger finalization; it cannot undo Go
// side effects that already started. Safety therefore also depends on the
// 100s Go context deadline, the 110s Worker fetch timeout, request cancellation,
// and domain-level idempotency/CAS before this 150s lease can be replaced.
const CONTROL_KEY = "scheduler:control";
const ACTIVE_KEY = "scheduler:active";
const FENCE_KEY = "scheduler:fence";
const RUN_KEY_PREFIX = "scheduler:run:";
const LATEST_KEY_PREFIX = "scheduler:latest:";

export interface CoordinatorTransaction {
  get<T>(key: string): Promise<T | undefined>;
  put<T>(key: string, value: T): Promise<void>;
  delete(key: string): Promise<boolean>;
}

export interface CoordinatorStorage extends CoordinatorTransaction {
  list<T>(options: { prefix: string }): Promise<Map<string, T>>;
  transaction<T>(closure: (transaction: CoordinatorTransaction) => Promise<T>): Promise<T>;
}

export interface SchedulerControl {
  version: 1;
  paused: boolean;
  changedAt: number;
}

export interface ActiveScheduledRun {
  version: 1;
  scheduler: typeof SCHEDULER_NAME;
  runId: string;
  runKey: string;
  cron: string;
  job: ScheduledJobName;
  scheduledTime: number;
  fenceToken: number;
  claimedAt: number;
  leaseExpiresAt: number;
}

export type RunLedgerStatus = "running" | "paused" | "success" | "partial" | "failed";
export type RunFailureCode =
  | "job_partial"
  | "job_failed"
  | "transport"
  | "stale"
  | "busy"
  | "lease_expired"
  | "fenced";

export interface ScheduledRunLedger {
  version: 1;
  scheduler: typeof SCHEDULER_NAME;
  runId: string;
  runKey: string;
  cron: string;
  job: ScheduledJobName;
  scheduledTime: number;
  fenceToken: number;
  status: RunLedgerStatus;
  startedAt: number;
  finishedAt?: number;
  outcome?: ScheduledJobOutcome;
  failureCode?: RunFailureCode;
}

interface LatestScheduledRun {
  version: 1;
  job: ScheduledJobName;
  scheduledTime: number;
  runId: string;
  updatedAt: number;
}

export type RunDisposition =
  | "executed"
  | "paused"
  | "stale"
  | "duplicate"
  | "busy"
  | "fenced";

export interface ScheduledRunResult {
  disposition: RunDisposition;
  ledger?: ScheduledRunLedger;
  active?: ActiveScheduledRun;
}

type JobExecutor = (request: ScheduledJobRequest) => Promise<ScheduledJobOutcome>;

type ClaimResult =
  | { disposition: "claimed"; active: ActiveScheduledRun; ledger: ScheduledRunLedger }
  | { disposition: Exclude<RunDisposition, "executed" | "fenced">; ledger?: ScheduledRunLedger; active?: ActiveScheduledRun };

function failureCodeForOutcome(outcome: ScheduledJobOutcome): RunFailureCode | undefined {
  if (outcome.outcome === "partial") {
    return "job_partial";
  }
  if (outcome.outcome === "failed") {
    return "job_failed";
  }
  return undefined;
}

function outcomeLedger(
  ledger: ScheduledRunLedger,
  now: number,
  outcome: ScheduledJobOutcome,
  failureCode?: RunFailureCode,
): ScheduledRunLedger {
  return {
    ...ledger,
    status: outcome.outcome,
    finishedAt: now,
    outcome,
    ...(failureCode === undefined ? {} : { failureCode }),
  };
}

function internalFailureLedger(
  ledger: ScheduledRunLedger,
  now: number,
  failureCode: RunFailureCode,
): ScheduledRunLedger {
  return {
    version: ledger.version,
    scheduler: ledger.scheduler,
    runId: ledger.runId,
    runKey: ledger.runKey,
    cron: ledger.cron,
    job: ledger.job,
    scheduledTime: ledger.scheduledTime,
    fenceToken: ledger.fenceToken,
    status: "failed",
    startedAt: ledger.startedAt,
    finishedAt: now,
    failureCode,
  };
}

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
    return (
      (await this.storage.get<SchedulerControl>(CONTROL_KEY)) ?? {
        version: 1,
        paused: false,
        changedAt: 0,
      }
    );
  }

  async setPaused(paused: boolean): Promise<SchedulerControl> {
    const control: SchedulerControl = {
      version: 1,
      paused,
      changedAt: this.clock(),
    };
    await this.storage.put(CONTROL_KEY, control);
    return control;
  }

  async getActive(): Promise<ActiveScheduledRun | undefined> {
    return this.storage.get<ActiveScheduledRun>(ACTIVE_KEY);
  }

  async run(
    job: ScheduledJobName,
    cron: string,
    scheduledTime: number,
    execute: JobExecutor,
  ): Promise<ScheduledRunResult> {
    this.validateInvocation(job, cron, scheduledTime);
    const now = this.clock();
    await this.pruneLedgers(now);

    const claim = await this.claim(job, cron, scheduledTime, now);
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
    now: number,
  ): Promise<ClaimResult> {
    const runKey = SchedulerCoordinator.runKey(job, scheduledTime);
    const runId = SchedulerCoordinator.runId(job, scheduledTime);

    return this.storage.transaction(async (transaction) => {
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
      if (latest !== undefined && scheduledTime < latest.scheduledTime) {
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

      const control = await transaction.get<SchedulerControl>(CONTROL_KEY);
      if (control?.paused === true) {
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
        await transaction.put(`${LATEST_KEY_PREFIX}${job}`, {
          version: 1,
          job,
          scheduledTime,
          runId,
          updatedAt: now,
        } satisfies LatestScheduledRun);
        return { disposition: "paused", ledger: pausedLedger };
      }

      if (currentActive !== undefined) {
        if (currentActive.leaseExpiresAt > now) {
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
          await transaction.put(`${LATEST_KEY_PREFIX}${job}`, {
            version: 1,
            job,
            scheduledTime,
            runId,
            updatedAt: now,
          } satisfies LatestScheduledRun);
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
        fenceToken,
        status: "running",
        startedAt: now,
      };
      const latestRun: LatestScheduledRun = {
        version: 1,
        job,
        scheduledTime,
        runId,
        updatedAt: now,
      };

      await transaction.put(FENCE_KEY, fenceToken);
      await transaction.put(ACTIVE_KEY, active);
      await transaction.put(runKey, ledger);
      await transaction.put(`${LATEST_KEY_PREFIX}${job}`, latestRun);
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
      await transaction.put(
        active.runKey,
        internalFailureLedger(ledger, now, "lease_expired"),
      );
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

  private async pruneLedgers(now: number): Promise<void> {
    const cutoff = now - RUN_LEDGER_RETENTION_MS;
    const ledgers = await this.storage.list<ScheduledRunLedger>({ prefix: RUN_KEY_PREFIX });
    const expiredKeys = [...ledgers.entries()]
      .filter(([, ledger]) => ledger.status !== "running" && ledger.scheduledTime < cutoff)
      .map(([key]) => key)
      .sort((left, right) => left.localeCompare(right));

    for (const key of expiredKeys) {
      await this.storage.delete(key);
    }
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
