import { timingSafeEqual } from "./migrate-exec";

export const SCHEDULER_OPS_PRINCIPAL = "scheduler-ops-secret-v1" as const;

const MAX_ACCESS_JWT_BYTES = 16 * 1_024;
const MAX_JWKS_BYTES = 64 * 1_024;
const EXTERNAL_REQUEST_TIMEOUT_MS = 5_000;
const MIN_SCHEDULER_OPS_SECRET_BYTES = 32;
const ACCESS_JWKS_CACHE_TTL_MS = 10 * 60_000;
const ACCESS_JWKS_REFRESH_COOLDOWN_MS = 60_000;

export interface SchedulerOpsAuthConfig {
  automationSecret?: string;
  accessTeamDomain?: string;
  accessAudience?: string;
}

export interface SchedulerOpsAuthenticatedPrincipal {
  actorPrincipal: string;
}

interface CloudflareAccessJWTHeader {
  alg: "RS256";
  kid: string;
}

interface CloudflareAccessJWTPayload {
  iss: string;
  aud: string | readonly string[];
  sub: string;
  exp: number;
  nbf?: number;
}

export type ExternalRequestFetcher = (request: Request) => Promise<Response>;

interface AccessJWKSCacheEntry {
  readonly jwks?: unknown;
  readonly expiresAt?: number;
  readonly refreshAfter: number;
  readonly inFlight?: Promise<unknown>;
}

// Cloudflare Access is the first authentication boundary in production, but
// the Worker still validates its assertion. Cache by the fixed configured team
// domain and transport so an unverified, self-generated JWT cannot force one
// outbound JWKS request per attempt. The cooldown also bounds unknown-kid and
// upstream-failure refreshes, while a later refresh still admits rotated keys.
const accessJWKSCacheByFetcher = new WeakMap<
  ExternalRequestFetcher,
  Map<string, AccessJWKSCacheEntry>
>();

export function isRedirectResponse(response: Response): boolean {
  return (
    response.redirected ||
    (response.status >= 300 && response.status < 400) ||
    response.headers.has("Location")
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function decodeBase64URL(value: string): Uint8Array {
  if (
    value.length === 0 ||
    value.length % 4 === 1 ||
    !/^[A-Za-z0-9_-]+$/.test(value)
  ) {
    throw new Error("invalid_base64url");
  }
  const padded = `${value.replaceAll("-", "+").replaceAll("_", "/")}${"=".repeat(
    (4 - (value.length % 4)) % 4,
  )}`;
  const decoded = atob(padded);
  return Uint8Array.from(decoded, (character) => character.charCodeAt(0));
}

function parseJWTPart(value: string): unknown {
  const bytes = decodeBase64URL(value);
  const text = new TextDecoder("utf-8", { fatal: true }).decode(bytes);
  return JSON.parse(text) as unknown;
}

function isCloudflareAccessTeamDomain(value: string): boolean {
  return (
    value.length <= 253 &&
    /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.cloudflareaccess\.com$/.test(
      value,
    )
  );
}

export function isStrictHostname(value: string): boolean {
  const labels = value.split(".");
  return (
    value.length <= 253 &&
    labels.length >= 2 &&
    labels.every(
      (label) =>
        label.length <= 63 &&
        /^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/.test(label),
    )
  );
}

export async function fetchWithTimeout(
  request: Request,
  send: ExternalRequestFetcher,
): Promise<Response> {
  return await fetchWithTimeoutAndConsume(request, send, async (response) => response);
}

async function fetchWithTimeoutAndConsume<T>(
  request: Request,
  send: ExternalRequestFetcher,
  consume: (response: Response, signal: AbortSignal) => Promise<T>,
): Promise<T> {
  const abort = new AbortController();
  const boundedRequest = new Request(request, { signal: abort.signal });
  let timeoutHandle: ReturnType<typeof setTimeout> | undefined;
  const timeout = new Promise<never>((_, reject) => {
    timeoutHandle = setTimeout(() => {
      abort.abort();
      reject(new Error("external_request_timeout"));
    }, EXTERNAL_REQUEST_TIMEOUT_MS);
  });
  try {
    const requestAndConsumption = (async () => {
      const response = await send(boundedRequest);
      return await consume(response, abort.signal);
    })();
    return await Promise.race([requestAndConsumption, timeout]);
  } finally {
    if (timeoutHandle !== undefined) {
      clearTimeout(timeoutHandle);
    }
  }
}

async function readResponseChunk(
  reader: ReadableStreamDefaultReader<Uint8Array>,
  signal: AbortSignal,
): Promise<ReadableStreamReadResult<Uint8Array>> {
  if (signal.aborted) {
    throw new Error("external_request_timeout");
  }
  let rejectAbort: ((reason?: unknown) => void) | undefined;
  const aborted = new Promise<never>((_, reject) => {
    rejectAbort = reject;
  });
  const onAbort = () => {
    rejectAbort?.(new Error("external_request_timeout"));
  };
  signal.addEventListener("abort", onAbort, { once: true });
  try {
    return await Promise.race([reader.read(), aborted]);
  } finally {
    signal.removeEventListener("abort", onAbort);
  }
}

async function readBoundedResponseJSON(
  response: Response,
  maxBytes: number,
  signal: AbortSignal,
): Promise<unknown> {
  const declaredLength = response.headers.get("Content-Length");
  if (
    declaredLength !== null &&
    (!/^\d+$/.test(declaredLength) || Number(declaredLength) > maxBytes)
  ) {
    throw new Error("external_response_too_large");
  }
  if (response.body === null) {
    throw new Error("external_response_body_required");
  }
  const reader = response.body.getReader();
  const chunks: Uint8Array[] = [];
  let total = 0;
  try {
    while (true) {
      const next = await readResponseChunk(reader, signal);
      if (next.done) {
        break;
      }
      total += next.value.byteLength;
      if (total > maxBytes) {
        throw new Error("external_response_too_large");
      }
      chunks.push(next.value);
    }
  } catch (error) {
    try {
      await reader.cancel(error);
    } catch {
      // The aborted network stream can already be closed during cleanup.
    }
    throw error;
  } finally {
    reader.releaseLock();
  }
  const bytes = new Uint8Array(total);
  let offset = 0;
  for (const chunk of chunks) {
    bytes.set(chunk, offset);
    offset += chunk.byteLength;
  }
  const text = new TextDecoder("utf-8", { fatal: true }).decode(bytes);
  return JSON.parse(text) as unknown;
}

function isAccessJWTHeader(value: unknown): value is CloudflareAccessJWTHeader {
  return (
    isRecord(value) &&
    value.alg === "RS256" &&
    typeof value.kid === "string" &&
    /^[A-Za-z0-9._~-]{1,200}$/.test(value.kid)
  );
}

function isAccessJWTPayload(
  value: unknown,
  issuer: string,
  audience: string,
  nowSeconds: number,
): value is CloudflareAccessJWTPayload {
  if (
    !isRecord(value) ||
    value.iss !== issuer ||
    typeof value.sub !== "string" ||
    !/^[A-Za-z0-9._~:@+-]{1,160}$/.test(value.sub) ||
    !Number.isSafeInteger(value.exp) ||
    Number(value.exp) <= nowSeconds ||
    (value.nbf !== undefined &&
      (!Number.isSafeInteger(value.nbf) || Number(value.nbf) > nowSeconds))
  ) {
    return false;
  }
  const audiences =
    typeof value.aud === "string"
      ? [value.aud]
      : Array.isArray(value.aud) &&
          value.aud.every((entry) => typeof entry === "string")
        ? value.aud
        : [];
  return audiences.includes(audience);
}

function selectAccessVerificationJWK(
  value: unknown,
  keyID: string,
): JsonWebKey | undefined {
  if (!isRecord(value) || !Array.isArray(value.keys)) {
    return undefined;
  }
  return value.keys.find(
    (candidate): candidate is JsonWebKey =>
      isRecord(candidate) &&
      candidate.kid === keyID &&
      candidate.kty === "RSA" &&
      (candidate.alg === undefined || candidate.alg === "RS256") &&
      (candidate.use === undefined || candidate.use === "sig"),
  );
}

function accessJWKSCacheFor(
  send: ExternalRequestFetcher,
): Map<string, AccessJWKSCacheEntry> {
  const existing = accessJWKSCacheByFetcher.get(send);
  if (existing !== undefined) {
    return existing;
  }
  const created = new Map<string, AccessJWKSCacheEntry>();
  accessJWKSCacheByFetcher.set(send, created);
  return created;
}

async function fetchAccessJWKS(
  teamDomain: string,
  send: ExternalRequestFetcher,
): Promise<unknown> {
  return await fetchWithTimeoutAndConsume(
    new Request(`https://${teamDomain}/cdn-cgi/access/certs`, {
      method: "GET",
      redirect: "manual",
      headers: { Accept: "application/json" },
    }),
    send,
    async (response, signal) => {
      if (!response.ok || isRedirectResponse(response)) {
        throw new Error("invalid_access_jwks_response");
      }
      return await readBoundedResponseJSON(response, MAX_JWKS_BYTES, signal);
    },
  );
}

async function resolveAccessVerificationJWK(
  teamDomain: string,
  keyID: string,
  now: number,
  send: ExternalRequestFetcher,
): Promise<JsonWebKey | undefined> {
  const cache = accessJWKSCacheFor(send);
  const cached = cache.get(teamDomain);
  if (
    cached?.jwks !== undefined &&
    cached.expiresAt !== undefined &&
    now < cached.expiresAt
  ) {
    const cachedKey = selectAccessVerificationJWK(cached.jwks, keyID);
    if (cachedKey !== undefined) {
      return cachedKey;
    }
  }
  if (cached?.inFlight !== undefined) {
    return selectAccessVerificationJWK(await cached.inFlight, keyID);
  }
  if (cached !== undefined && now < cached.refreshAfter) {
    return undefined;
  }

  const inFlight = fetchAccessJWKS(teamDomain, send);
  cache.set(teamDomain, {
    ...cached,
    refreshAfter: now + ACCESS_JWKS_REFRESH_COOLDOWN_MS,
    inFlight,
  });
  try {
    const jwks = await inFlight;
    cache.set(teamDomain, {
      jwks,
      expiresAt: now + ACCESS_JWKS_CACHE_TTL_MS,
      refreshAfter: now + ACCESS_JWKS_REFRESH_COOLDOWN_MS,
    });
    return selectAccessVerificationJWK(jwks, keyID);
  } catch (error) {
    cache.set(teamDomain, {
      jwks: cached?.jwks,
      expiresAt: cached?.expiresAt,
      refreshAfter: now + ACCESS_JWKS_REFRESH_COOLDOWN_MS,
    });
    throw error;
  }
}

export function isAuthorizedSchedulerOpsRequest(
  request: Request,
  secret: string | undefined,
): boolean {
  if (
    secret === undefined ||
    new TextEncoder().encode(secret).byteLength <
      MIN_SCHEDULER_OPS_SECRET_BYTES
  ) {
    return false;
  }
  const authorization = request.headers.get("Authorization") ?? "";
  return timingSafeEqual(authorization, `Bearer ${secret}`);
}

export async function authenticateSchedulerOpsRequest(
  request: Request,
  config: SchedulerOpsAuthConfig,
  now: number = Date.now(),
  send: ExternalRequestFetcher = fetch,
): Promise<SchedulerOpsAuthenticatedPrincipal | undefined> {
  if (isAuthorizedSchedulerOpsRequest(request, config.automationSecret)) {
    return { actorPrincipal: SCHEDULER_OPS_PRINCIPAL };
  }
  const assertion = request.headers.get("CF-Access-Jwt-Assertion");
  const teamDomain = config.accessTeamDomain?.trim().toLowerCase() ?? "";
  const audience = config.accessAudience?.trim() ?? "";
  if (
    !assertion ||
    new TextEncoder().encode(assertion).byteLength > MAX_ACCESS_JWT_BYTES ||
    !isCloudflareAccessTeamDomain(teamDomain) ||
    audience.length < 1 ||
    audience.length > 256 ||
    !Number.isSafeInteger(now) ||
    now <= 0
  ) {
    return undefined;
  }

  try {
    const parts = assertion.split(".");
    if (parts.length !== 3) {
      return undefined;
    }
    const [encodedHeader, encodedPayload, encodedSignature] = parts;
    if (
      encodedHeader === undefined ||
      encodedPayload === undefined ||
      encodedSignature === undefined
    ) {
      return undefined;
    }
    const header = parseJWTPart(encodedHeader);
    const payload = parseJWTPart(encodedPayload);
    const issuer = `https://${teamDomain}`;
    const nowSeconds = Math.floor(now / 1_000);
    if (
      !isAccessJWTHeader(header) ||
      !isAccessJWTPayload(payload, issuer, audience, nowSeconds)
    ) {
      return undefined;
    }

    const jwk = await resolveAccessVerificationJWK(
      teamDomain,
      header.kid,
      now,
      send,
    );
    if (jwk === undefined) {
      return undefined;
    }
    const verificationKey = await crypto.subtle.importKey(
      "jwk",
      jwk,
      { name: "RSASSA-PKCS1-v1_5", hash: "SHA-256" },
      false,
      ["verify"],
    );
    const verified = await crypto.subtle.verify(
      "RSASSA-PKCS1-v1_5",
      verificationKey,
      decodeBase64URL(encodedSignature).buffer as ArrayBuffer,
      new TextEncoder().encode(`${encodedHeader}.${encodedPayload}`)
        .buffer as ArrayBuffer,
    );
    return verified
      ? { actorPrincipal: `cloudflare-access:${payload.sub}` }
      : undefined;
  } catch {
    return undefined;
  }
}
