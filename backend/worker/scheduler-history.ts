import {
  OPERATION_AUDIT_RETENTION_MS,
  RUN_LEDGER_RETENTION_MS,
  type CoordinatorStorage,
  type CoordinatorTransaction,
  type ScheduledRunLedger,
  type SchedulerOperation,
  type SchedulerOperationIndex,
} from "./scheduler-coordinator-records";

const HISTORY_PAGE_SIZE = 15;
const HISTORY_MAX_PAGES = 2;

export interface SchedulerHistoryConfig {
  runKeyPrefix: string;
  operationKeyPrefix: string;
  operationIndexPrefix: string;
  operationResultKeyPrefix: string;
  operationDriverKeyPrefix: string;
}

function operationIndexKey(
  indexPrefix: string,
  requestedAt: number,
  requestId: string,
): string {
  return `${indexPrefix}${requestedAt
    .toString()
    .padStart(16, "0")}:${requestId}`;
}

export async function putSchedulerOperation(
  transaction: CoordinatorTransaction,
  operationKey: string,
  indexPrefix: string,
  operation: SchedulerOperation,
): Promise<void> {
  const index: SchedulerOperationIndex = {
    version: 1,
    requestId: operation.requestId,
    requestedAt: operation.requestedAt,
  };
  await transaction.put(operationKey, operation);
  await transaction.put(
    operationIndexKey(
      indexPrefix,
      operation.requestedAt,
      operation.requestId,
    ),
    index,
  );
}

async function listBoundedHistory<T>(
  storage: CoordinatorStorage,
  prefix: string,
): Promise<Map<string, T>> {
  const collected = new Map<string, T>();
  let startAfter: string | undefined;
  for (let page = 0; page < HISTORY_MAX_PAGES; page += 1) {
    const batch = await storage.list<T>({
      prefix,
      ...(startAfter === undefined ? {} : { startAfter }),
      limit: HISTORY_PAGE_SIZE,
    });
    if (batch.size === 0) {
      break;
    }
    for (const [key, value] of batch) {
      collected.set(key, value);
    }
    if (batch.size < HISTORY_PAGE_SIZE) {
      break;
    }
    startAfter = [...batch.keys()].at(-1);
    if (startAfter === undefined) {
      break;
    }
  }
  return collected;
}

export async function pruneSchedulerHistory(
  storage: CoordinatorStorage,
  now: number,
  config: SchedulerHistoryConfig,
): Promise<void> {
  const ledgerCutoff = now - RUN_LEDGER_RETENTION_MS;
  const operationCutoff = now - OPERATION_AUDIT_RETENTION_MS;
  const [ledgers, operationIndexes] = await Promise.all([
    listBoundedHistory<ScheduledRunLedger>(storage, config.runKeyPrefix),
    listBoundedHistory<SchedulerOperationIndex>(
      storage,
      config.operationIndexPrefix,
    ),
  ]);
  const expiredKeys = [...ledgers.entries()]
    .filter(
      ([, ledger]) =>
        ledger.status !== "running" &&
        ledger.scheduledTime < ledgerCutoff,
    )
    .map(([key]) => key)
    .sort((left, right) => left.localeCompare(right));
  const expiredOperations = [...operationIndexes.entries()]
    .filter(([, operation]) => operation.requestedAt < operationCutoff)
    .sort(([left], [right]) => left.localeCompare(right));

  await storage.transaction(async (transaction) => {
    for (const key of expiredKeys) {
      await transaction.delete(key);
    }
    for (const [indexKey, operation] of expiredOperations) {
      await transaction.delete(indexKey);
      await transaction.delete(
        `${config.operationKeyPrefix}${operation.requestId}`,
      );
      await transaction.delete(
        `${config.operationResultKeyPrefix}${operation.requestId}`,
      );
      await transaction.delete(
        `${config.operationDriverKeyPrefix}${operation.requestId}`,
      );
    }
  });
}
