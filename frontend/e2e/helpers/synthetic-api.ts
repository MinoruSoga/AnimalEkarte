const READ_METHODS = new Set(["GET", "HEAD", "OPTIONS"]);

interface ExternalAssetStub {
  origin: string;
  pathname: string;
  contentType: string;
  body: string;
}

/**
 * Cross-origin assets the app is allowed to load (frontend/index.html loads
 * Google Fonts CSS intentionally). Allowlisted assets are always answered from
 * these local stubs — in every mode — so E2E runs stay deterministic and never
 * perform real external communication for them.
 */
const EXTERNAL_ASSET_STUBS: readonly ExternalAssetStub[] = [
  {
    origin: "https://fonts.googleapis.com",
    pathname: "/css2",
    contentType: "text/css; charset=utf-8",
    body: "/* synthetic-api local stub: Google Fonts CSS is never fetched externally in E2E */",
  },
];

export type SyntheticHttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

export interface SyntheticInterceptorRequest {
  url(): string;
  method(): string;
  headers(): Record<string, string>;
  postData(): string | null;
}

export type SyntheticFulfillOptions =
  | { json: unknown }
  | { body: string; contentType: string };

export interface SyntheticInterceptorRoute {
  request(): SyntheticInterceptorRequest;
  continue(): Promise<void>;
  abort(errorCode?: "blockedbyclient"): Promise<void>;
  fulfill(options: SyntheticFulfillOptions): Promise<void>;
}

export interface SyntheticInterceptorPage {
  route(
    url: string,
    handler: (route: SyntheticInterceptorRoute) => Promise<void>,
  ): Promise<unknown>;
  unroute(
    url: string,
    handler: (route: SyntheticInterceptorRoute) => Promise<void>,
  ): Promise<void>;
  isClosed(): boolean;
}

type JsonResponseFactory = (
  request: SyntheticInterceptorRequest,
) => unknown | Promise<unknown>;

export interface SyntheticEndpoint {
  method: SyntheticHttpMethod;
  pathname: string | RegExp;
  query?: Readonly<Record<string, string>>;
  response: unknown | JsonResponseFactory;
  validateBody?: (body: unknown, request: SyntheticInterceptorRequest) => void;
}

interface SyntheticApiInterceptorOptions {
  expectedOrigin?: string;
  expectedClinicId?: string;
}

export interface SyntheticRequestLedger {
  attempted: readonly string[];
  locallyFulfilled: readonly string[];
  blocked: readonly string[];
  continuedToBackend: readonly string[];
  validationFailures: readonly string[];
}

export interface SyntheticApiInterceptor {
  ledger: SyntheticRequestLedger;
  reset(): void;
  dispose(): Promise<void>;
}

function requestKey(request: SyntheticInterceptorRequest, withOrigin = false): string {
  const url = new URL(request.url());
  const redactedPathname = url.pathname
    .split("/")
    .map((segment) =>
      /^\d+$/.test(segment) ||
      /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(segment)
        ? ":id"
        : segment,
    )
    .join("/");
  return `${request.method()}:${withOrigin ? url.origin : ""}${redactedPathname}`;
}

function findExternalAssetStub(url: URL): ExternalAssetStub | undefined {
  return EXTERNAL_ASSET_STUBS.find(
    (stub) => stub.origin === url.origin && stub.pathname === url.pathname,
  );
}

function matchesPath(expected: string | RegExp, actual: string): boolean {
  if (typeof expected === "string") return expected === actual;
  expected.lastIndex = 0;
  return expected.test(actual);
}

function matchesQuery(expected: Readonly<Record<string, string>> | undefined, url: URL): boolean {
  const actualEntries = [...url.searchParams.entries()].sort(([left], [right]) => left.localeCompare(right));
  const expectedEntries = Object.entries(expected ?? {}).sort(([left], [right]) => left.localeCompare(right));
  return actualEntries.length === expectedEntries.length &&
    actualEntries.every(([key, value], index) => {
      const expectedEntry = expectedEntries[index];
      return expectedEntry?.[0] === key && expectedEntry[1] === value;
    });
}

function createLedger(): SyntheticRequestLedger {
  return {
    attempted: [],
    locallyFulfilled: [],
    blocked: [],
    continuedToBackend: [],
    validationFailures: [],
  };
}

function findEndpoint(
  endpoints: readonly SyntheticEndpoint[],
  request: SyntheticInterceptorRequest,
): SyntheticEndpoint | undefined {
  const method = request.method();
  const url = new URL(request.url());
  const matches = endpoints.filter(
    (endpoint) =>
      endpoint.method === method &&
      matchesPath(endpoint.pathname, url.pathname) &&
      matchesQuery(endpoint.query, url),
  );
  if (matches.length > 1) {
    throw new Error(`synthetic endpoint matched more than once: ${method}:${url.pathname}`);
  }
  return matches[0];
}

function readJsonBody(request: SyntheticInterceptorRequest): unknown {
  const raw = request.postData();
  if (raw === null || raw === "") return undefined;
  return JSON.parse(raw) as unknown;
}

export function isBusinessNonGet(request: string): boolean {
  return !request.startsWith("GET:") && !request.startsWith("HEAD:") && !request.startsWith("OPTIONS:");
}

export async function installSyntheticApiInterceptor(
  page: SyntheticInterceptorPage,
  endpoints: readonly SyntheticEndpoint[],
  options: SyntheticApiInterceptorOptions = {},
): Promise<SyntheticApiInterceptor> {
  let ledger = createLedger();

  const record = (field: keyof SyntheticRequestLedger, value: string): void => {
    ledger = { ...ledger, [field]: [...ledger[field], value] };
  };

  const isUnexpectedOrigin = (requestUrl: URL): boolean =>
    options.expectedOrigin !== undefined &&
    requestUrl.origin !== options.expectedOrigin;

  // Non-API surface: assets never consult synthetic endpoints, clinic/CSRF
  // headers, or the business non-GET allowlist — those stay API-only.
  const handleAssetRequest = async (
    route: SyntheticInterceptorRoute,
    request: SyntheticInterceptorRequest,
    requestUrl: URL,
  ): Promise<void> => {
    const stub = findExternalAssetStub(requestUrl);
    if (stub && READ_METHODS.has(request.method())) {
      const key = requestKey(request, true);
      record("attempted", key);
      await route.fulfill({ body: stub.body, contentType: stub.contentType });
      record("locallyFulfilled", key);
      return;
    }

    if (!READ_METHODS.has(request.method())) {
      const key = requestKey(request, isUnexpectedOrigin(requestUrl));
      record("attempted", key);
      record("blocked", key);
      await route.abort("blockedbyclient");
      return;
    }

    if (isUnexpectedOrigin(requestUrl)) {
      const key = requestKey(request, true);
      record("attempted", key);
      record("blocked", key);
      record("validationFailures", `${key}:unexpected cross-origin asset`);
      await route.abort("blockedbyclient");
      return;
    }

    // Same-origin app asset (or normal mode): outside the API audit ledger.
    await route.continue();
  };

  const handleApiRequest = async (
    route: SyntheticInterceptorRoute,
    request: SyntheticInterceptorRequest,
    requestUrl: URL,
  ): Promise<void> => {
    if (isUnexpectedOrigin(requestUrl)) {
      const key = requestKey(request, true);
      record("attempted", key);
      record("blocked", key);
      record("validationFailures", `${key}:unexpected API origin`);
      await route.abort("blockedbyclient");
      return;
    }

    const key = requestKey(request);
    record("attempted", key);

    let endpoint: SyntheticEndpoint | undefined;
    try {
      endpoint = findEndpoint(endpoints, request);
    } catch (error) {
      record("blocked", key);
      record("validationFailures", `${key}:${error instanceof Error ? error.message : "matcher error"}`);
      await route.abort("blockedbyclient");
      return;
    }

    if (endpoint) {
      let response: unknown;
      try {
        if (options.expectedClinicId) {
          const headers = request.headers();
          if (headers["x-clinic-id"] !== options.expectedClinicId) {
            throw new Error("synthetic clinic header mismatch");
          }
          if (isBusinessNonGet(key) && headers["x-requested-with"] !== "XMLHttpRequest") {
            throw new Error("synthetic CSRF header mismatch");
          }
        }
        if (endpoint.validateBody) endpoint.validateBody(readJsonBody(request), request);
        response = typeof endpoint.response === "function"
          ? await endpoint.response(request)
          : endpoint.response;
      } catch (error) {
        record("blocked", key);
        record("validationFailures", `${key}:${error instanceof Error ? error.message : "fixture error"}`);
        await route.abort("blockedbyclient");
        return;
      }
      await route.fulfill({ json: response });
      record("locallyFulfilled", key);
      return;
    }

    if (READ_METHODS.has(request.method())) {
      if (options.expectedOrigin !== undefined) {
        record("blocked", key);
        await route.abort("blockedbyclient");
        return;
      }
      record("continuedToBackend", key);
      await route.continue();
      return;
    }

    record("blocked", key);
    await route.abort("blockedbyclient");
  };

  const handler = async (route: SyntheticInterceptorRoute): Promise<void> => {
    const request = route.request();
    const requestUrl = new URL(request.url());
    if (requestUrl.pathname.startsWith("/api/")) {
      await handleApiRequest(route, request, requestUrl);
      return;
    }
    await handleAssetRequest(route, request, requestUrl);
  };

  await page.route("**/*", handler);

  return {
    get ledger() {
      return ledger;
    },
    reset() {
      ledger = createLedger();
    },
    async dispose() {
      if (!page.isClosed()) await page.unroute("**/*", handler);
    },
  };
}
