export const SCHEDULER_NAME = "animalekarte-scheduler-v1" as const;

export const SCHEDULED_JOB_GO_TIMEOUT_MS = 100_000;
export const SCHEDULED_JOB_FETCH_TIMEOUT_MS = 110_000;
export const SCHEDULED_JOB_LEASE_MS = 150_000;

export const SCHEDULED_JOBS_INTERNAL_PREFIX = "/_internal/scheduled-jobs";
const MAX_PATH_DECODE_PASSES = 16;

export type ScheduledJobName = "no_show" | "delivery" | "dormant";
export type ScheduledJobOutcomeStatus = "success" | "partial" | "failed";

export interface ScheduledJobRequest {
  scheduler: typeof SCHEDULER_NAME;
  job: ScheduledJobName;
  scheduled_time: number;
  run_id: string;
  fence_token: number;
}

export interface ScheduledJobOutcome {
  outcome: ScheduledJobOutcomeStatus;
  processed: number;
  succeeded: number;
  failed: number;
}

export type ContainerRequest = (request: Request) => Promise<Response>;

const CRON_JOB_PLAN = {
  "0 1 * * *": ["no_show", "delivery"],
  "0 6,11 * * *": ["no_show"],
  "0 17 * * *": ["dormant"],
} as const satisfies Readonly<Record<string, readonly ScheduledJobName[]>>;

function isNonNegativeSafeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isSafeInteger(value) && value >= 0;
}

function isScheduledJobOutcome(value: unknown): value is ScheduledJobOutcome {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  const candidate = value as Record<string, unknown>;
  const outcome = candidate.outcome;
  if (outcome !== "success" && outcome !== "partial" && outcome !== "failed") {
    return false;
  }

  const processed = candidate.processed;
  const succeeded = candidate.succeeded;
  const failed = candidate.failed;
  if (
    !isNonNegativeSafeInteger(processed) ||
    !isNonNegativeSafeInteger(succeeded) ||
    !isNonNegativeSafeInteger(failed) ||
    succeeded + failed !== processed
  ) {
    return false;
  }
  if (outcome === "success") {
    return failed === 0;
  }
  if (outcome === "partial") {
    return succeeded > 0 && failed > 0;
  }
  return succeeded === 0 && failed > 0;
}

export function jobsForCron(cron: string): readonly ScheduledJobName[] {
  if (!Object.hasOwn(CRON_JOB_PLAN, cron)) {
    throw new Error(`unknown cron expression: ${cron}`);
  }

  return [...CRON_JOB_PLAN[cron as keyof typeof CRON_JOB_PLAN]];
}

export function isScheduledJobsInternalPath(pathname: string): boolean {
  let decoded = pathname;
  for (let pass = 0; pass < MAX_PATH_DECODE_PASSES; pass += 1) {
    const segments = decoded.replaceAll("\\", "/").split("/");
    let normalizedSegments: readonly string[] = [];
    for (const segment of segments) {
      if (segment === "" || segment === ".") {
        continue;
      }
      if (segment === "..") {
        normalizedSegments = normalizedSegments.slice(0, -1);
        continue;
      }
      normalizedSegments = [...normalizedSegments, segment];
    }
    const normalized = `/${normalizedSegments.join("/")}`.toLowerCase();
    if (normalized.startsWith(SCHEDULED_JOBS_INTERNAL_PREFIX)) {
      return true;
    }
    if (!decoded.includes("%")) {
      return false;
    }

    try {
      const next = decodeURIComponent(decoded);
      if (next === decoded) {
        return false;
      }
      decoded = next;
    } catch {
      // An ambiguous/malformed encoded path is never forwarded to the
      // Container. This is intentionally broader than the scheduler prefix:
      // the proxy must not guess how a downstream HTTP stack will decode it.
      return true;
    }
  }

  // Deep nesting beyond the normalization budget remains encoded and is
  // therefore ambiguous. It is denied rather than forwarded to a downstream
  // HTTP stack that may apply more decoding passes.
  return true;
}

// Privilege header required by Go requireSchedulerInternalToken (batch_scheduler.go).
// Must match SCHEDULER_INTERNAL_TOKEN injected into the Container envVars.
export const SCHEDULER_INTERNAL_TOKEN_HEADER = "X-Scheduler-Token" as const;

/**
 * Build outbound headers for Worker → Container scheduled-job POSTs.
 *
 * Contract when token is unset/empty: **omit** X-Scheduler-Token entirely.
 * Never attach undefined or empty string (would look like a present-but-wrong secret).
 * Go fails closed if expected token is empty or header mismatches (401).
 *
 * Cutover order (must not reverse):
 * 1. FIRST set the secret: `wrangler secret put SCHEDULER_INTERNAL_TOKEN`.
 *    The currently deployed Worker ignores it, so this step is harmless on its own.
 * 2. THEN deploy this Worker. The deploy is what brings the Go binary whose middleware
 *    enforces the header, and it binds the secret into the Container envVars at the same time.
 * Reversing this (deploying before the secret exists) leaves the Go middleware with an EMPTY
 * expected token, and requireSchedulerInternalToken 401s EVERY request regardless of what the
 * Worker sends -- all-hospital scheduled jobs stop until the secret is put and the Container
 * restarts. The deploy is the cutover event; the secret must already be in place.
 */
export function buildScheduledJobHeaders(
  schedulerInternalToken?: string,
): Record<string, string> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (
    typeof schedulerInternalToken === "string" &&
    schedulerInternalToken.length > 0
  ) {
    headers[SCHEDULER_INTERNAL_TOKEN_HEADER] = schedulerInternalToken;
  }
  return headers;
}

export async function runScheduledJobRequest(
  fetchContainer: ContainerRequest,
  request: ScheduledJobRequest,
  schedulerInternalToken?: string,
): Promise<ScheduledJobOutcome> {
  const abort = new AbortController();
  let timeoutHandle: ReturnType<typeof setTimeout> | undefined;
  const timeoutError = new Error(
    `scheduled job request timed out after ${SCHEDULED_JOB_FETCH_TIMEOUT_MS}ms`,
  );
  const timeout = new Promise<never>((_resolve, reject) => {
    timeoutHandle = setTimeout(() => {
      abort.abort(timeoutError);
      reject(timeoutError);
    }, SCHEDULED_JOB_FETCH_TIMEOUT_MS);
  });

  try {
    const requestAndParse = async (): Promise<ScheduledJobOutcome> => {
      const response = await fetchContainer(
        new Request(
          `http://container.internal${SCHEDULED_JOBS_INTERNAL_PREFIX}/${request.job}:run`,
          {
            method: "POST",
            headers: buildScheduledJobHeaders(schedulerInternalToken),
            body: JSON.stringify(request),
            signal: abort.signal,
          },
        ),
      );

      let body: unknown;
      try {
        body = await response.json();
      } catch {
        throw new Error("invalid scheduled job outcome: response is not JSON");
      }

      if (!isScheduledJobOutcome(body)) {
        throw new Error("invalid scheduled job outcome: response does not match contract");
      }
      if (!response.ok && body.outcome === "success") {
        throw new Error("scheduled job returned non-success HTTP status with success outcome");
      }

      return body;
    };

    return await Promise.race([requestAndParse(), timeout]);
  } finally {
    if (timeoutHandle !== undefined) {
      clearTimeout(timeoutHandle);
    }
  }
}
