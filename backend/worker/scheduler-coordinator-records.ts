import {
  SCHEDULER_NAME,
  type ScheduledJobName,
  type ScheduledJobOutcome,
} from "./scheduled-jobs";

const DAY_MS = 24 * 60 * 60 * 1_000;
export const RUN_LEDGER_RETENTION_MS = 35 * DAY_MS;
export const OPERATION_AUDIT_RETENTION_MS = 400 * DAY_MS;

export interface CoordinatorTransaction {
  get<T>(key: string): Promise<T | undefined>;
  put<T>(key: string, value: T): Promise<void>;
  delete(key: string): Promise<boolean>;
}

export interface CoordinatorStorage extends CoordinatorTransaction {
  list<T>(options: {
    prefix: string;
    startAfter?: string;
    reverse?: boolean;
    limit?: number;
  }): Promise<Map<string, T>>;
  transaction<T>(closure: (transaction: CoordinatorTransaction) => Promise<T>): Promise<T>;
}

export interface SchedulerControl {
  version: 1;
  revision: number;
  paused: boolean;
  changedAt: number;
}

export interface SchedulerControlCommand {
  paused: boolean;
  expectedRevision: number;
  requestId: string;
  reason: string;
  actorPrincipal: string;
}

export interface SchedulerManualCommand {
  mode: "catch_up";
  job: ScheduledJobName;
  scheduledTime: number;
  requestId: string;
  reason: string;
  actorPrincipal: string;
}

export interface SchedulerOperationBase {
  version: 1;
  requestId: string;
  actorPrincipal: string;
  reason: string;
  requestedAt: number;
  status: "pending" | "completed" | "rejected";
  rejectionCode?: "revision_conflict" | "slot_already_recorded" | "scheduler_paused" | "scheduler_busy";
}

export interface SchedulerControlOperation extends SchedulerOperationBase {
  kind: "control";
  expectedRevision: number;
  requestedPaused: boolean;
  control: SchedulerControl;
}

export interface SchedulerManualOperation extends SchedulerOperationBase {
  kind: "manual_run";
  mode: "catch_up";
  job: ScheduledJobName;
  cron: string;
  scheduledTime: number;
  result?: ScheduledRunResult;
}

export interface SchedulerManualDriver {
  version: 1;
  token: string;
  leaseExpiresAt: number;
}

export type SchedulerOperation = SchedulerControlOperation | SchedulerManualOperation;

export interface ScheduledRunSummary {
  version: 1;
  scheduler: typeof SCHEDULER_NAME;
  runId: string;
  cron: string;
  job: ScheduledJobName;
  scheduledTime: number;
  status: RunLedgerStatus;
  startedAt: number;
  finishedAt?: number;
  outcome?: ScheduledJobOutcome;
  failureCode?: RunFailureCode;
}

export interface SchedulerStatus {
  version: 1;
  scheduler: typeof SCHEDULER_NAME;
  control: SchedulerControl;
  active?: ActiveScheduledRunSummary;
  recentRuns: readonly ScheduledRunSummary[];
  recentOperations: readonly SchedulerOperationSummary[];
}

export interface ActiveScheduledRunSummary {
  version: 1;
  scheduler: typeof SCHEDULER_NAME;
  runId: string;
  cron: string;
  job: ScheduledJobName;
  scheduledTime: number;
  claimedAt: number;
  leaseExpiresAt: number;
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
  manualRequestId?: string;
  fenceToken: number;
  status: RunLedgerStatus;
  startedAt: number;
  finishedAt?: number;
  outcome?: ScheduledJobOutcome;
  failureCode?: RunFailureCode;
}

export interface LatestScheduledRun {
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

export interface ScheduledRunResultSummary {
  disposition: RunDisposition;
  ledger?: ScheduledRunSummary;
  active?: ActiveScheduledRunSummary;
}

export type SchedulerOperationSummary =
  | SchedulerControlOperation
  | (Omit<SchedulerManualOperation, "result"> & {
      result?: ScheduledRunResultSummary;
    });

export interface SchedulerOpsRateLimitDecision {
  allowed: boolean;
  retryAfterSeconds: number;
}

export interface SchedulerOperationIndex {
  version: 1;
  requestId: string;
  requestedAt: number;
}

export interface SchedulerOpsRateLimitState {
  version: 1;
  windowStartedAt: number;
  count: number;
}

export function validateOperationIdentity(command: {
  requestId: string;
  reason: string;
  actorPrincipal: string;
}): void {
  if (
    !/^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(
      command.requestId,
    )
  ) {
    throw new Error("requestId must be a UUID");
  }
  const reason = command.reason.trim();
  if (
    reason.length < 4 ||
    reason.length > 200 ||
    /[\u0000-\u001f\u007f]/.test(reason)
  ) {
    throw new Error("reason is invalid");
  }
  if (
    command.actorPrincipal !== "scheduler-ops-secret-v1" &&
    !/^cloudflare-access:[A-Za-z0-9._~:@+-]{1,160}$/.test(
      command.actorPrincipal,
    )
  ) {
    throw new Error("actorPrincipal is invalid");
  }
}

export function sameControlIntent(
  operation: SchedulerControlOperation,
  command: SchedulerControlCommand,
): boolean {
  return (
    operation.requestId === command.requestId &&
    operation.actorPrincipal === command.actorPrincipal &&
    operation.reason === command.reason.trim() &&
    operation.expectedRevision === command.expectedRevision &&
    operation.requestedPaused === command.paused
  );
}

export function sameManualIntent(
  operation: SchedulerManualOperation,
  command: SchedulerManualCommand,
): boolean {
  return (
    operation.requestId === command.requestId &&
    operation.actorPrincipal === command.actorPrincipal &&
    operation.reason === command.reason.trim() &&
    operation.mode === command.mode &&
    operation.job === command.job &&
    operation.scheduledTime === command.scheduledTime
  );
}

export function defaultControl(): SchedulerControl {
  return {
    version: 1,
    revision: 0,
    paused: false,
    changedAt: 0,
  };
}

export function normalizeControl(control: SchedulerControl | undefined): SchedulerControl {
  if (control === undefined) {
    return defaultControl();
  }
  return {
    ...control,
    revision:
      Number.isSafeInteger(control.revision) && control.revision >= 0
        ? control.revision
        : 0,
  };
}

export function logLeaseExpired(ledger: ScheduledRunLedger): void {
  console.error("scheduler run lease expired", {
    event: "scheduler_run_lease_expired",
    scheduler: ledger.scheduler,
    run_id: ledger.runId,
    job: ledger.job,
    scheduled_time: ledger.scheduledTime,
    failure_code: "lease_expired",
  });
}

export function cronForManualSlot(
  job: ScheduledJobName,
  scheduledTime: number,
  now: number,
): string {
  if (
    !Number.isSafeInteger(scheduledTime) ||
    scheduledTime <= 0 ||
    scheduledTime > now ||
    now - scheduledTime > RUN_LEDGER_RETENTION_MS
  ) {
    throw new Error("scheduled_time_outside_recovery_window");
  }
  const scheduled = new Date(scheduledTime);
  if (
    scheduled.getUTCMinutes() !== 0 ||
    scheduled.getUTCSeconds() !== 0 ||
    scheduled.getUTCMilliseconds() !== 0
  ) {
    throw new Error("scheduled_time_is_not_a_cron_slot");
  }
  const hour = scheduled.getUTCHours();
  if ((job === "delivery" || job === "no_show") && hour === 1) {
    return "0 1 * * *";
  }
  if (job === "no_show" && (hour === 6 || hour === 11)) {
    return "0 6,11 * * *";
  }
  if (job === "dormant" && hour === 17) {
    return "0 17 * * *";
  }
  throw new Error("scheduled_time_is_not_a_cron_slot");
}

export function summarizeLedger(ledger: ScheduledRunLedger): ScheduledRunSummary {
  return {
    version: ledger.version,
    scheduler: ledger.scheduler,
    runId: ledger.runId,
    cron: ledger.cron,
    job: ledger.job,
    scheduledTime: ledger.scheduledTime,
    status: ledger.status,
    startedAt: ledger.startedAt,
    ...(ledger.finishedAt === undefined ? {} : { finishedAt: ledger.finishedAt }),
    ...(ledger.outcome === undefined ? {} : { outcome: ledger.outcome }),
    ...(ledger.failureCode === undefined ? {} : { failureCode: ledger.failureCode }),
  };
}

export function summarizeActive(active: ActiveScheduledRun): ActiveScheduledRunSummary {
  return {
    version: active.version,
    scheduler: active.scheduler,
    runId: active.runId,
    cron: active.cron,
    job: active.job,
    scheduledTime: active.scheduledTime,
    claimedAt: active.claimedAt,
    leaseExpiresAt: active.leaseExpiresAt,
  };
}

export function summarizeRunResult(result: ScheduledRunResult): ScheduledRunResultSummary {
  return {
    disposition: result.disposition,
    ...(result.ledger === undefined
      ? {}
      : { ledger: summarizeLedger(result.ledger) }),
    ...(result.active === undefined
      ? {}
      : { active: summarizeActive(result.active) }),
  };
}

export function summarizeSchedulerOperation(
  operation: SchedulerOperation,
): SchedulerOperationSummary {
  if (operation.kind === "control" || operation.result === undefined) {
    return operation;
  }
  return {
    ...operation,
    result: summarizeRunResult(operation.result),
  };
}

export function rejectionCodeForDisposition(
  disposition: Exclude<RunDisposition, "executed">,
): NonNullable<SchedulerManualOperation["rejectionCode"]> {
  if (disposition === "paused") {
    return "scheduler_paused";
  }
  if (disposition === "busy") {
    return "scheduler_busy";
  }
  return "slot_already_recorded";
}

export function terminalManualOperation(
  operation: SchedulerManualOperation,
  result: ScheduledRunResult,
): SchedulerManualOperation {
  if (result.disposition === "executed") {
    return {
      ...operation,
      status: "completed",
      result,
    };
  }
  return {
    ...operation,
    status: "rejected",
    rejectionCode: rejectionCodeForDisposition(result.disposition),
    result,
  };
}

export function reconcilePendingManualOperation(
  operation: SchedulerManualOperation,
  ledger: ScheduledRunLedger | undefined,
): SchedulerManualOperation | undefined {
  if (ledger === undefined || ledger.status === "running") {
    return undefined;
  }
  const disposition: RunDisposition =
    ledger.manualRequestId !== operation.requestId
      ? "duplicate"
      : ledger.status === "paused"
        ? "paused"
        : ledger.failureCode === "busy"
          ? "busy"
          : ledger.failureCode === "stale"
            ? "stale"
            : ledger.failureCode === "fenced" ||
                ledger.failureCode === "lease_expired"
              ? "fenced"
              : "executed";
  return terminalManualOperation(operation, { disposition, ledger });
}

export function failureCodeForOutcome(outcome: ScheduledJobOutcome): RunFailureCode | undefined {
  if (outcome.outcome === "partial") {
    return "job_partial";
  }
  if (outcome.outcome === "failed") {
    return "job_failed";
  }
  return undefined;
}

export function outcomeLedger(
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

export function internalFailureLedger(
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
    ...(ledger.manualRequestId === undefined
      ? {}
      : { manualRequestId: ledger.manualRequestId }),
    fenceToken: ledger.fenceToken,
    status: "failed",
    startedAt: ledger.startedAt,
    finishedAt: now,
    failureCode,
  };
}
