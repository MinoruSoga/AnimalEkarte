import {
  validateOperationIdentity,
  type CoordinatorStorage,
  type SchedulerOpsRateLimitDecision,
  type SchedulerOpsRateLimitState,
} from "./scheduler-coordinator-records";

const OPS_RATE_LIMIT_KEY_PREFIX = "scheduler:ops-rate-limit:";
export const SCHEDULER_OPS_RATE_LIMIT = 60;
export const SCHEDULER_OPS_RATE_WINDOW_MS = 60_000;

export async function consumeSchedulerOpsRateLimit(
  storage: CoordinatorStorage,
  actorPrincipal: string,
  now: number,
): Promise<SchedulerOpsRateLimitDecision> {
  validateOperationIdentity({
    actorPrincipal,
    reason: "rate limit check",
    requestId: "00000000-0000-4000-8000-000000000000",
  });
  if (!Number.isSafeInteger(now) || now <= 0) {
    throw new Error("rate limit time must be a positive integer");
  }
  const windowStartedAt =
    Math.floor(now / SCHEDULER_OPS_RATE_WINDOW_MS) *
    SCHEDULER_OPS_RATE_WINDOW_MS;
  const key = `${OPS_RATE_LIMIT_KEY_PREFIX}${actorPrincipal.replaceAll(
    ":",
    "_",
  )}`;
  return storage.transaction(async (transaction) => {
    const existing = await transaction.get<SchedulerOpsRateLimitState>(key);
    const state =
      existing?.windowStartedAt === windowStartedAt
        ? existing
        : { version: 1 as const, windowStartedAt, count: 0 };
    const retryAfterSeconds = Math.max(
      1,
      Math.ceil(
        (windowStartedAt + SCHEDULER_OPS_RATE_WINDOW_MS - now) / 1_000,
      ),
    );
    if (state.count >= SCHEDULER_OPS_RATE_LIMIT) {
      return { allowed: false, retryAfterSeconds };
    }
    await transaction.put(key, { ...state, count: state.count + 1 });
    return { allowed: true, retryAfterSeconds };
  });
}
