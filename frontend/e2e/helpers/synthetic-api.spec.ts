import { expect, test } from "@playwright/test";

import {
  installSyntheticApiInterceptor,
  isBusinessNonGet,
  type SyntheticEndpoint,
  type SyntheticInterceptorPage,
  type SyntheticInterceptorRequest,
  type SyntheticInterceptorRoute,
} from "./synthetic-api";

const ENDPOINTS = [
  {
    method: "GET",
    pathname: "/api/v1/synthetic-read",
    query: { page: "1" },
    response: { source: "synthetic" },
  },
  {
    method: "GET",
    pathname: /^\/api\/v1\/pets\/\d+$/,
    response: { id: 990003, name: "合成監査ペット" },
  },
  {
    method: "POST",
    pathname: "/api/v1/synthetic-write",
    validateBody: (body: unknown) => {
      if (
        typeof body !== "object" ||
        body === null ||
        Array.isArray(body) ||
        Object.keys(body).length !== 1 ||
        !("sentinel" in body) ||
        body.sentinel !== true
      ) {
        throw new Error("synthetic write body mismatch");
      }
    },
    response: { id: 990013 },
  },
] satisfies readonly SyntheticEndpoint[];

type RouteAction =
  | { type: "abort"; errorCode: string | undefined }
  | { type: "continue" }
  | { type: "fulfill"; contentType: string | undefined };

interface RequestInput {
  url: string;
  method?: string;
}

interface InterceptorHarness {
  page: SyntheticInterceptorPage;
  dispatch(input: RequestInput): Promise<readonly RouteAction[]>;
}

function createInterceptorHarness(): InterceptorHarness {
  let routeHandler: ((route: SyntheticInterceptorRoute) => Promise<void>) | undefined;

  const page: SyntheticInterceptorPage = {
    async route(
      _url: string,
      handler: (route: SyntheticInterceptorRoute) => Promise<void>,
    ): Promise<void> {
      routeHandler = handler;
    },
    isClosed: () => false,
    unroute: async () => undefined,
  };

  return {
    page,
    async dispatch({ url, method = "GET" }: RequestInput) {
      if (!routeHandler) throw new Error("synthetic route handler was not installed");

      let actions: readonly RouteAction[] = [];
      const request: SyntheticInterceptorRequest = {
        url: () => url,
        method: () => method,
        headers: () => ({}),
        postData: () => null,
      };
      const route: SyntheticInterceptorRoute = {
        request: () => request,
        continue: async () => {
          actions = [...actions, { type: "continue" }];
        },
        abort: async (errorCode?: string) => {
          actions = [...actions, { type: "abort", errorCode }];
        },
        fulfill: async (options) => {
          actions = [
            ...actions,
            {
              type: "fulfill",
              contentType: "contentType" in options ? options.contentType : undefined,
            },
          ];
        },
      };

      await routeHandler(route);
      return actions;
    },
  };
}

test("pure fully synthetic mode blocks an unregistered same-origin API GET", async () => {
  const harness = createInterceptorHarness();
  const interceptor = await installSyntheticApiInterceptor(harness.page, [], {
    expectedOrigin: "http://frontend.test",
  });

  const actions = await harness.dispatch({
    url: "http://frontend.test/api/v1/unregistered",
  });

  expect(actions).toEqual([{ type: "abort", errorCode: "blockedbyclient" }]);
  expect(interceptor.ledger).toEqual({
    attempted: ["GET:/api/v1/unregistered"],
    locallyFulfilled: [],
    blocked: ["GET:/api/v1/unregistered"],
    continuedToBackend: [],
    validationFailures: [],
  });
});

test("pure fully synthetic mode blocks a cross-origin API GET and records its origin", async () => {
  const harness = createInterceptorHarness();
  const interceptor = await installSyntheticApiInterceptor(harness.page, [], {
    expectedOrigin: "http://frontend.test",
  });

  const actions = await harness.dispatch({
    url: "https://api.evil.test/api/v1/owners",
  });

  expect(actions).toEqual([{ type: "abort", errorCode: "blockedbyclient" }]);
  expect(interceptor.ledger).toEqual({
    attempted: ["GET:https://api.evil.test/api/v1/owners"],
    locallyFulfilled: [],
    blocked: ["GET:https://api.evil.test/api/v1/owners"],
    continuedToBackend: [],
    validationFailures: ["GET:https://api.evil.test/api/v1/owners:unexpected API origin"],
  });
});

test("pure fully synthetic mode fulfills an allowlisted external asset from the local stub", async () => {
  const harness = createInterceptorHarness();
  const interceptor = await installSyntheticApiInterceptor(harness.page, [], {
    expectedOrigin: "http://frontend.test",
  });

  const actions = await harness.dispatch({
    url: "https://fonts.googleapis.com/css2?family=Inter:wght@400;500;700&family=Noto+Sans+JP:wght@400;500;700&display=swap",
  });

  expect(actions).toEqual([{ type: "fulfill", contentType: "text/css; charset=utf-8" }]);
  expect(interceptor.ledger).toEqual({
    attempted: ["GET:https://fonts.googleapis.com/css2"],
    locallyFulfilled: ["GET:https://fonts.googleapis.com/css2"],
    blocked: [],
    continuedToBackend: [],
    validationFailures: [],
  });
});

test("pure fully synthetic mode blocks an unknown cross-origin asset GET", async () => {
  const harness = createInterceptorHarness();
  const interceptor = await installSyntheticApiInterceptor(harness.page, [], {
    expectedOrigin: "http://frontend.test",
  });

  const actions = await harness.dispatch({
    url: "https://cdn.example.test/assets/application.js",
  });

  expect(actions).toEqual([{ type: "abort", errorCode: "blockedbyclient" }]);
  expect(interceptor.ledger).toEqual({
    attempted: ["GET:https://cdn.example.test/assets/application.js"],
    locallyFulfilled: [],
    blocked: ["GET:https://cdn.example.test/assets/application.js"],
    continuedToBackend: [],
    validationFailures: [
      "GET:https://cdn.example.test/assets/application.js:unexpected cross-origin asset",
    ],
  });
});

test("pure fully synthetic mode continues a same-origin non-API asset GET outside the ledger", async () => {
  const harness = createInterceptorHarness();
  const interceptor = await installSyntheticApiInterceptor(harness.page, [], {
    expectedOrigin: "http://frontend.test",
  });

  const actions = await harness.dispatch({
    url: "http://frontend.test/src/main.tsx",
  });

  expect(actions).toEqual([{ type: "continue" }]);
  expect(interceptor.ledger).toEqual({
    attempted: [],
    locallyFulfilled: [],
    blocked: [],
    continuedToBackend: [],
    validationFailures: [],
  });
});

test("pure normal mode fulfills an allowlisted external asset from the local stub", async () => {
  const harness = createInterceptorHarness();
  const interceptor = await installSyntheticApiInterceptor(harness.page, []);

  const actions = await harness.dispatch({
    url: "https://fonts.googleapis.com/css2?family=Inter:wght@400;500;700&display=swap",
  });

  expect(actions).toEqual([{ type: "fulfill", contentType: "text/css; charset=utf-8" }]);
  expect(interceptor.ledger).toEqual({
    attempted: ["GET:https://fonts.googleapis.com/css2"],
    locallyFulfilled: ["GET:https://fonts.googleapis.com/css2"],
    blocked: [],
    continuedToBackend: [],
    validationFailures: [],
  });
});

test("pure normal mode blocks a non-API business write instead of continuing", async () => {
  const harness = createInterceptorHarness();
  const interceptor = await installSyntheticApiInterceptor(harness.page, []);

  const actions = await harness.dispatch({
    url: "http://frontend.test/logout",
    method: "POST",
  });

  expect(actions).toEqual([{ type: "abort", errorCode: "blockedbyclient" }]);
  expect(interceptor.ledger).toEqual({
    attempted: ["POST:/logout"],
    locallyFulfilled: [],
    blocked: ["POST:/logout"],
    continuedToBackend: [],
    validationFailures: [],
  });
});

test("pure normal mode continues an unregistered API GET", async () => {
  const harness = createInterceptorHarness();
  const interceptor = await installSyntheticApiInterceptor(harness.page, []);

  const actions = await harness.dispatch({
    url: "http://frontend.test/api/v1/unregistered",
  });

  expect(actions).toEqual([{ type: "continue" }]);
  expect(interceptor.ledger).toEqual({
    attempted: ["GET:/api/v1/unregistered"],
    locallyFulfilled: [],
    blocked: [],
    continuedToBackend: ["GET:/api/v1/unregistered"],
    validationFailures: [],
  });
});

test("synthetic API interceptor classifies every request and never continues a business write", async ({
  page,
}) => {
  await page.goto("/@vite/client", { waitUntil: "domcontentloaded" });
  const interceptor = await installSyntheticApiInterceptor(page, ENDPOINTS);

  try {
    const results = await page.evaluate(async () => {
      const syntheticRead = await fetch("/api/v1/synthetic-read?page=1").then((response) =>
        response.json(),
      );
      const redactedRead = await fetch("/api/v1/pets/123456").then((response) => response.json());
      const syntheticWrite = await fetch("/api/v1/synthetic-write", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ sentinel: true }),
      }).then((response) => response.json());
      const continuedReadStatus = await fetch("/api/v1/me").then((response) => response.status);
      const blockedUnexpectedWrite = await fetch("/api/v1/unexpected-write", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ sentinel: true }),
      }).then(
        () => false,
        () => true,
      );
      const blockedInvalidWrite = await fetch("/api/v1/synthetic-write", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ sentinel: false }),
      }).then(
        () => false,
        () => true,
      );

      return {
        syntheticRead,
        redactedRead,
        syntheticWrite,
        continuedReadStatus,
        blockedUnexpectedWrite,
        blockedInvalidWrite,
      };
    });

    expect(results).toEqual({
      syntheticRead: { source: "synthetic" },
      redactedRead: { id: 990003, name: "合成監査ペット" },
      syntheticWrite: { id: 990013 },
      continuedReadStatus: 401,
      blockedUnexpectedWrite: true,
      blockedInvalidWrite: true,
    });
    expect(interceptor.ledger).toEqual({
      attempted: [
        "GET:/api/v1/synthetic-read",
        "GET:/api/v1/pets/:id",
        "POST:/api/v1/synthetic-write",
        "GET:/api/v1/me",
        "PATCH:/api/v1/unexpected-write",
        "POST:/api/v1/synthetic-write",
      ],
      locallyFulfilled: [
        "GET:/api/v1/synthetic-read",
        "GET:/api/v1/pets/:id",
        "POST:/api/v1/synthetic-write",
      ],
      blocked: ["PATCH:/api/v1/unexpected-write", "POST:/api/v1/synthetic-write"],
      continuedToBackend: ["GET:/api/v1/me"],
      validationFailures: ["POST:/api/v1/synthetic-write:synthetic write body mismatch"],
    });
    expect(interceptor.ledger.continuedToBackend.filter(isBusinessNonGet)).toEqual([]);
    expect(interceptor.ledger.attempted.length).toBe(
      interceptor.ledger.locallyFulfilled.length +
        interceptor.ledger.blocked.length +
        interceptor.ledger.continuedToBackend.length,
    );
  } finally {
    await page.close();
    await interceptor.dispose();
  }
});

test("fully synthetic mode rejects cross-origin and wrong-clinic requests", async ({ page }) => {
  await page.goto("/@vite/client", { waitUntil: "domcontentloaded" });
  const origin = new URL(page.url()).origin;
  const endpoints = [
    { method: "GET", pathname: "/api/v1/synthetic-read", response: { ok: true } },
    { method: "POST", pathname: "/api/v1/synthetic-write", response: { ok: true } },
  ] satisfies readonly SyntheticEndpoint[];
  const interceptor = await installSyntheticApiInterceptor(page, endpoints, {
    expectedOrigin: origin,
    expectedClinicId: "990001",
  });

  try {
    const results = await page.evaluate(async () => {
      const validHeaders = {
        "X-Clinic-ID": "990001",
        "X-Requested-With": "XMLHttpRequest",
      };
      const validRead = await fetch("/api/v1/synthetic-read", {
        headers: validHeaders,
      }).then((response) => response.ok);
      const validWrite = await fetch("/api/v1/synthetic-write", {
        method: "POST",
        headers: validHeaders,
      }).then((response) => response.ok);
      const wrongClinicBlocked = await fetch("/api/v1/synthetic-read", {
        headers: { "X-Clinic-ID": "990999" },
      }).then(
        () => false,
        () => true,
      );
      const crossOriginBlocked = await fetch("https://example.invalid/api/v1/synthetic-read").then(
        () => false,
        () => true,
      );
      return { validRead, validWrite, wrongClinicBlocked, crossOriginBlocked };
    });

    expect(results).toEqual({
      validRead: true,
      validWrite: true,
      wrongClinicBlocked: true,
      crossOriginBlocked: true,
    });
    expect(interceptor.ledger.locallyFulfilled).toEqual([
      "GET:/api/v1/synthetic-read",
      "POST:/api/v1/synthetic-write",
    ]);
    expect(interceptor.ledger.blocked).toEqual([
      "GET:/api/v1/synthetic-read",
      "GET:https://example.invalid/api/v1/synthetic-read",
    ]);
    expect(interceptor.ledger.continuedToBackend).toEqual([]);
    expect(interceptor.ledger.validationFailures).toEqual([
      "GET:/api/v1/synthetic-read:synthetic clinic header mismatch",
      "GET:https://example.invalid/api/v1/synthetic-read:unexpected API origin",
    ]);
  } finally {
    await interceptor.dispose();
  }
});
