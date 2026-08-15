import {
  cronForManualSlot,
  type SchedulerControlCommand,
  type SchedulerControlOperation,
  type SchedulerManualCommand,
  type SchedulerManualOperation,
  type SchedulerOpsRateLimitDecision,
  type SchedulerStatus,
  type ScheduledRunResult,
  scheduledRunRequiresCronFailure,
  summarizeSchedulerOperation,
} from "./scheduler-coordinator";
import {
  SCHEDULER_NAME,
  type ScheduledJobName,
} from "./scheduled-jobs";
import {
  SCHEDULER_OPS_PRINCIPAL,
  authenticateSchedulerOpsRequest,
  fetchWithTimeout,
  isRedirectResponse,
  isStrictHostname,
  type ExternalRequestFetcher,
  type SchedulerOpsAuthConfig,
} from "./scheduler-access-auth";

export { cronForManualSlot };
export {
  SCHEDULER_OPS_PRINCIPAL,
  authenticateSchedulerOpsRequest,
  isAuthorizedSchedulerOpsRequest,
  type SchedulerOpsAuthConfig,
} from "./scheduler-access-auth";

export const SCHEDULER_OPS_PREFIX = "/_internal/scheduler" as const;

const MAX_BODY_BYTES = 4 * 1_024;
const DEFAULT_STATUS_LIMIT = 20;
const MAX_STATUS_LIMIT = 50;
const MAX_INTERNAL_PATH_DECODE_PASSES = 16;
const UUID_PATTERN =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;

export interface SchedulerOpsBinding {
  consumeScheduledJobsOpsRateLimit(
    actorPrincipal: string,
    now: number,
  ): Promise<SchedulerOpsRateLimitDecision>;
  getScheduledJobsStatus(limit: number): Promise<SchedulerStatus>;
  setScheduledJobsControl(command: SchedulerControlCommand): Promise<SchedulerControlOperation>;
  runScheduledJobManually(command: SchedulerManualCommand): Promise<SchedulerManualOperation>;
}

export interface SchedulerAlertConfig {
  environment: string;
  allowedHost?: string;
  webhookURL?: string;
  webhookSecret?: string;
}

type RequestFetcher = ExternalRequestFetcher;
type ManualOperationObserver = (
  operation: SchedulerManualOperation,
) => Promise<void>;

class SchedulerOpsRequestError extends Error {
  constructor(
    readonly status: number,
    readonly code: string,
  ) {
    super(code);
  }
}

function jsonResponse(status: number, body: unknown, allow?: string): Response {
  const headers = new Headers({ "Content-Type": "application/json" });
  if (allow !== undefined) {
    headers.set("Allow", allow);
  }
  return new Response(JSON.stringify(body), { status, headers });
}

export function isInternalProxyPath(pathname: string): boolean {
  let decoded = pathname;
  for (let pass = 0; pass < MAX_INTERNAL_PATH_DECODE_PASSES; pass += 1) {
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
    if (
      normalized === "/_internal" ||
      normalized.startsWith("/_internal/")
    ) {
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
      return true;
    }
  }
  return true;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function hasExactKeys(
  record: Record<string, unknown>,
  expected: readonly string[],
): boolean {
  const actual = Object.keys(record).sort();
  const wanted = [...expected].sort();
  return (
    actual.length === wanted.length &&
    actual.every((key, index) => key === wanted[index])
  );
}

function validateRequestIdentity(
  requestId: unknown,
  reason: unknown,
): requestId is string {
  return (
    typeof requestId === "string" &&
    UUID_PATTERN.test(requestId) &&
    typeof reason === "string" &&
    reason === reason.trim() &&
    reason.length >= 4 &&
    reason.length <= 200 &&
    !/[\u0000-\u001f\u007f]/.test(reason)
  );
}

async function readBoundedJSON(request: Request): Promise<unknown> {
  const declaredLength = request.headers.get("Content-Length");
  if (
    declaredLength !== null &&
    !/^\d+$/.test(declaredLength)
  ) {
    throw new SchedulerOpsRequestError(400, "invalid_request");
  }
  if (
    declaredLength !== null &&
    Number(declaredLength) > MAX_BODY_BYTES
  ) {
    throw new SchedulerOpsRequestError(413, "request_body_too_large");
  }
  if (request.body === null) {
    throw new SchedulerOpsRequestError(400, "request_body_required");
  }

  const reader = request.body.getReader();
  const chunks: Uint8Array[] = [];
  let total = 0;
  try {
    while (true) {
      const next = await reader.read();
      if (next.done) {
        break;
      }
      total += next.value.byteLength;
      if (total > MAX_BODY_BYTES) {
        await reader.cancel();
        throw new SchedulerOpsRequestError(413, "request_body_too_large");
      }
      chunks.push(next.value);
    }
  } finally {
    reader.releaseLock();
  }

  const bytes = new Uint8Array(total);
  let offset = 0;
  for (const chunk of chunks) {
    bytes.set(chunk, offset);
    offset += chunk.byteLength;
  }
  try {
    const text = new TextDecoder("utf-8", { fatal: true }).decode(bytes);
    return JSON.parse(text) as unknown;
  } catch {
    throw new SchedulerOpsRequestError(400, "invalid_json");
  }
}

function hasJSONContentType(request: Request): boolean {
  return (
    (request.headers.get("Content-Type") ?? "")
      .split(";", 1)[0]
      ?.trim()
      .toLowerCase() === "application/json"
  );
}

function parseStatusLimit(url: URL): number | undefined {
  let invalidKey = false;
  let limitCount = 0;
  url.searchParams.forEach((_value, key) => {
    if (key === "limit") {
      limitCount += 1;
      return;
    }
    invalidKey = true;
  });
  if (invalidKey || limitCount > 1) {
    return undefined;
  }
  const raw = url.searchParams.get("limit");
  if (raw === null) {
    return DEFAULT_STATUS_LIMIT;
  }
  if (!/^\d+$/.test(raw)) {
    return undefined;
  }
  const limit = Number(raw);
  return Number.isSafeInteger(limit) && limit >= 1 && limit <= MAX_STATUS_LIMIT
    ? limit
    : undefined;
}

function parseControlCommand(
  value: unknown,
  actorPrincipal: string,
): SchedulerControlCommand | undefined {
  if (
    !isRecord(value) ||
    !hasExactKeys(value, ["paused", "expected_revision", "request_id", "reason"]) ||
    typeof value.paused !== "boolean" ||
    !Number.isSafeInteger(value.expected_revision) ||
    Number(value.expected_revision) < 0 ||
    !validateRequestIdentity(value.request_id, value.reason)
  ) {
    return undefined;
  }
  return {
    paused: value.paused,
    expectedRevision: Number(value.expected_revision),
    requestId: value.request_id,
    reason: String(value.reason),
    actorPrincipal,
  };
}

function isScheduledJobName(value: unknown): value is ScheduledJobName {
  return value === "no_show" || value === "delivery" || value === "dormant";
}

function parseManualCommand(
  value: unknown,
  now: number,
  actorPrincipal: string,
): SchedulerManualCommand | undefined {
  if (
    !isRecord(value) ||
    !hasExactKeys(value, [
      "job",
      "scheduled_time_ms",
      "mode",
      "request_id",
      "reason",
    ]) ||
    !isScheduledJobName(value.job) ||
    value.mode !== "catch_up" ||
    !Number.isSafeInteger(value.scheduled_time_ms) ||
    !validateRequestIdentity(value.request_id, value.reason)
  ) {
    return undefined;
  }
  try {
    cronForManualSlot(value.job, Number(value.scheduled_time_ms), now);
  } catch {
    return undefined;
  }
  return {
    actorPrincipal,
    job: value.job,
    mode: "catch_up",
    reason: String(value.reason),
    requestId: value.request_id,
    scheduledTime: Number(value.scheduled_time_ms),
  };
}

export async function handleSchedulerOpsRequest(
  request: Request,
  authConfig: SchedulerOpsAuthConfig,
  binding: SchedulerOpsBinding,
  now: number = Date.now(),
  observeManualOperation?: ManualOperationObserver,
  accessJWKSFetch: RequestFetcher = fetch,
): Promise<Response> {
  const authenticated = await authenticateSchedulerOpsRequest(
    request,
    authConfig,
    now,
    accessJWKSFetch,
  );
  if (authenticated === undefined) {
    return jsonResponse(401, { error: "unauthorized" });
  }

  const url = new URL(request.url);
  try {
    const rateLimit = await binding.consumeScheduledJobsOpsRateLimit(
      authenticated.actorPrincipal,
      now,
    );
    if (!rateLimit.allowed) {
      const response = jsonResponse(429, { error: "rate_limited" });
      response.headers.set(
        "Retry-After",
        rateLimit.retryAfterSeconds.toString(),
      );
      return response;
    }
    if (url.pathname === `${SCHEDULER_OPS_PREFIX}/status`) {
      if (request.method !== "GET") {
        return jsonResponse(405, { error: "method_not_allowed" }, "GET");
      }
      const limit = parseStatusLimit(url);
      if (limit === undefined) {
        return jsonResponse(400, { error: "invalid_request" });
      }
      return jsonResponse(200, await binding.getScheduledJobsStatus(limit));
    }

    if (
      url.pathname === `${SCHEDULER_OPS_PREFIX}/control` ||
      url.pathname === `${SCHEDULER_OPS_PREFIX}/runs`
    ) {
      const expectedMethod =
        url.pathname === `${SCHEDULER_OPS_PREFIX}/control` ? "PUT" : "POST";
      if (request.method !== expectedMethod) {
        return jsonResponse(
          405,
          { error: "method_not_allowed" },
          expectedMethod,
        );
      }
      if (!hasJSONContentType(request)) {
        return jsonResponse(415, { error: "unsupported_media_type" });
      }
      const body = await readBoundedJSON(request);
      if (expectedMethod === "PUT") {
        const command = parseControlCommand(
          body,
          authenticated.actorPrincipal,
        );
        if (command === undefined) {
          return jsonResponse(400, { error: "invalid_request" });
        }
        const operation = await binding.setScheduledJobsControl(command);
        return jsonResponse(operation.status === "rejected" ? 409 : 200, operation);
      }
      const command = parseManualCommand(
        body,
        now,
        authenticated.actorPrincipal,
      );
      if (command === undefined) {
        return jsonResponse(400, { error: "invalid_request" });
      }
      const operation = await binding.runScheduledJobManually(command);
      await observeManualOperation?.(operation);
      const status =
        operation.status === "rejected"
          ? 409
          : operation.status === "pending"
            ? 202
            : operation.result !== undefined &&
                scheduledRunRequiresCronFailure(operation.result)
              ? 502
              : 200;
      return jsonResponse(status, summarizeSchedulerOperation(operation));
    }
    return jsonResponse(404, { error: "not_found" });
  } catch (error) {
    if (error instanceof SchedulerOpsRequestError) {
      return jsonResponse(error.status, { error: error.code });
    }
    const conflict =
      error instanceof Error && error.message === "request_id_conflict";
    console.error("scheduler ops request failed", {
      event: "scheduler_ops_request_failed",
      path: url.pathname,
      failure_code: conflict ? "request_id_conflict" : "internal",
    });
    return jsonResponse(conflict ? 409 : 500, {
      error: conflict ? "request_id_conflict" : "scheduler_ops_failed",
    });
  }
}

function failureAlert(
  result: ScheduledRunResult,
  environment: string,
): {
  key: string;
  payload: Record<string, unknown>;
} | undefined {
  const ledger = result.ledger;
  if (ledger === undefined) {
    return undefined;
  }
  const isFailure =
    ledger.status === "partial" ||
    ledger.status === "failed" ||
    result.disposition === "busy" ||
    result.disposition === "stale" ||
    result.disposition === "fenced";
  if (!isFailure) {
    return undefined;
  }
  const failureCode = ledger.failureCode ?? (
    ledger.status === "partial" ? "job_partial" : "job_failed"
  );
  const key = `${environment}:${ledger.runId}:${failureCode}`;
  return {
    key,
    payload: {
      version: 1,
      alert_key: key,
      scheduler: ledger.scheduler,
      run_id: ledger.runId,
      job: ledger.job,
      scheduled_time: ledger.scheduledTime,
      status: ledger.status,
      failure_code: failureCode,
      ...(ledger.outcome === undefined
        ? {}
        : {
            counts: {
              processed: ledger.outcome.processed,
              succeeded: ledger.outcome.succeeded,
              failed: ledger.outcome.failed,
            },
          }),
    },
  };
}

function alertEndpoint(config: SchedulerAlertConfig): URL | undefined {
  if (
    !config.webhookURL ||
    !config.webhookSecret ||
    !config.allowedHost
  ) {
    return undefined;
  }
  try {
    const endpoint = new URL(config.webhookURL);
    const allowedHost = config.allowedHost.trim().toLowerCase();
    if (
      !isStrictHostname(allowedHost) ||
      endpoint.protocol !== "https:" ||
      endpoint.hostname.toLowerCase() !== allowedHost ||
      endpoint.username !== "" ||
      endpoint.password !== "" ||
      (endpoint.port !== "" && endpoint.port !== "443") ||
      allowedHost === ""
    ) {
      return undefined;
    }
    return endpoint;
  } catch {
    return undefined;
  }
}

export async function notifySchedulerFailures(
  results: readonly ScheduledRunResult[],
  config: SchedulerAlertConfig,
  send: RequestFetcher = fetch,
): Promise<void> {
  const environment = config.environment.trim().toLowerCase();
  const alerts = results
    .map((result) => failureAlert(result, environment))
    .filter((alert) => alert !== undefined);
  if (alerts.length === 0) {
    return;
  }

  for (const alert of alerts) {
    console.error("scheduler job failed", {
      event: "scheduler_job_failed",
      alert_key: alert.key,
      environment,
      failure_code: alert.payload.failure_code,
    });
  }

  const endpoint =
    /^[a-z0-9](?:[a-z0-9._-]{0,62}[a-z0-9])?$/.test(environment)
      ? alertEndpoint(config)
      : undefined;
  let deliveryFailed = false;
  for (const alert of alerts) {
    if (endpoint === undefined) {
      console.error("scheduler alert is not configured", {
        event: "scheduler_alert_not_configured",
        alert_key: alert.key,
        environment,
        failure_code: "alert_not_configured",
      });
      deliveryFailed = true;
      continue;
    }
    const payload = {
      ...alert.payload,
      environment,
    };
    let response: Response;
    try {
      response = await fetchWithTimeout(
        new Request(endpoint, {
          method: "POST",
          redirect: "manual",
          headers: {
            Authorization: `Bearer ${config.webhookSecret}`,
            "Content-Type": "application/json",
            "Idempotency-Key": alert.key,
          },
          body: JSON.stringify(payload),
        }),
        send,
      );
    } catch {
      console.error("scheduler alert delivery failed", {
        event: "scheduler_alert_delivery_failed",
        alert_key: alert.key,
        environment,
        failure_code: "alert_delivery_transport",
      });
      deliveryFailed = true;
      continue;
    }
    const redirectResponse = isRedirectResponse(response);
    if (redirectResponse || !response.ok) {
      console.error("scheduler alert delivery failed", {
        event: "scheduler_alert_delivery_failed",
        alert_key: alert.key,
        environment,
        failure_code: redirectResponse
          ? "alert_delivery_redirect"
          : "alert_delivery_http_status",
        status: response.status,
      });
      deliveryFailed = true;
    }
  }
  if (deliveryFailed) {
    throw new Error("scheduler alert delivery failed");
  }
}
