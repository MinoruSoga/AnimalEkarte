import type {
  CoordinatorTransaction,
  CoordinatorStorage,
  ScheduledRunLedger,
} from "./scheduler-coordinator-records";

export const SCHEDULER_ALERT_PENDING_PREFIX = "scheduler:alert-pending:";
export const SCHEDULER_ALERT_BATCH_SIZE = 20;
const ALERT_CURSOR_KEY = "scheduler:alert-cursor";

// Rotate past failed incidents so one full page cannot starve later incidents.
// At most two bounded reads and one batch of deliveries per invocation.
export async function listPendingSchedulerIncidents(
  storage: CoordinatorStorage,
): Promise<Map<string, ScheduledRunLedger>> {
  const cursor = await storage.get<string>(ALERT_CURSOR_KEY);
  const batch = await storage.list<ScheduledRunLedger>({
    prefix: SCHEDULER_ALERT_PENDING_PREFIX,
    ...(cursor === undefined ? {} : { startAfter: cursor }),
    limit: SCHEDULER_ALERT_BATCH_SIZE,
  });
  const wrapped =
    cursor !== undefined && batch.size < SCHEDULER_ALERT_BATCH_SIZE
      ? await storage.list<ScheduledRunLedger>({
          prefix: SCHEDULER_ALERT_PENDING_PREFIX,
          limit: SCHEDULER_ALERT_BATCH_SIZE - batch.size,
        })
      : new Map<string, ScheduledRunLedger>();
  const incidents = new Map([...batch, ...wrapped]);
  const lastKey = [...incidents.keys()].at(-1);
  if (lastKey !== undefined) await storage.put(ALERT_CURSOR_KEY, lastKey);
  return incidents;
}

export function isDurableSchedulerIncident(ledger: ScheduledRunLedger): boolean {
  return ledger.failureCode === "stuck_lease" || ledger.failureCode === "lease_expired";
}

// Enqueue on the first detection transition, never on replay. This immutable
// snapshot survives ledger pruning and executor/Worker restarts.
export async function enqueueSchedulerIncident(
  transaction: CoordinatorTransaction,
  ledger: ScheduledRunLedger,
): Promise<void> {
  if (!isDurableSchedulerIncident(ledger)) return;
  await transaction.put(
    `${SCHEDULER_ALERT_PENDING_PREFIX}${ledger.runId}:${ledger.failureCode}`,
    ledger,
  );
}
