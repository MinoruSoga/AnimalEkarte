import { describe, expect, it } from "vitest";

import { buildProxyObservation } from "./index";

describe("buildProxyObservation", () => {
  it("emits the fixed session success shape without raw request data", () => {
    const observation = buildProxyObservation(
      new Request("https://api.example.test/api/v1/me?owner=private", {
        headers: {
          Authorization: "Bearer secret",
          Cookie: "session=secret",
          "X-Request-ID": "9e4a6f84-8dae-4c30-8b51-a1b3f61b9ca5",
        },
      }),
      10.2,
      22.8,
      { status: 401 },
    );

    expect(observation).toEqual({
      event: "container_fetch_timing",
      method: "get",
      path: "session",
      duration_ms: 13,
      status: 401,
      request_id: "9e4a6f84-8dae-4c30-8b51-a1b3f61b9ca5",
    });
  });

  it("classifies unknown input and omits untrusted correlation IDs on failures", () => {
    const observation = buildProxyObservation(
      new Request("https://api.example.test/private/path?token=secret", {
        method: "POST",
        headers: { "X-Request-ID": "not-a-safe-id" },
      }),
      20,
      10,
      { failureCode: "container_unavailable" },
    );

    expect(observation).toEqual({
      event: "container_fetch_timing",
      method: "other",
      path: "other",
      duration_ms: 0,
      failure_code: "container_unavailable",
    });
  });
});
