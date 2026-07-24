import { describe, expect, it, vi } from "vitest";
import {
  SCHEDULER_OPS_PRINCIPAL,
  notifySchedulerFailures,
} from "./scheduler-ops";
import type { SchedulerManualOperation } from "./scheduler-coordinator";
import {
  SCHEDULER_NAME,
  type ScheduledJobName,
} from "./scheduled-jobs";

const NOW = Date.UTC(2026, 6, 24, 18, 0, 0);

function manualFixture(job: ScheduledJobName = "no_show"): SchedulerManualOperation {
  return {
    version: 1,
    kind: "manual_run",
    requestId: "85679370-8dc1-46f1-93ec-675ead9f0f49",
    actorPrincipal: SCHEDULER_OPS_PRINCIPAL,
    reason: "INC-124 recover missing cron slot",
    requestedAt: NOW,
    status: "completed",
    mode: "catch_up",
    job,
    cron: job === "dormant" ? "0 17 * * *" : "0 1 * * *",
    scheduledTime: Date.UTC(2026, 6, 23, 1, 0, 0),
    result: { disposition: "executed" },
  };
}

describe("notifySchedulerFailures", () => {
  it("sends only sanitized terminal failures with a stable idempotency key", async () => {
    const send = vi.fn(async (_request: Request) => new Response(null, { status: 204 }));
    const results = [
      {
        disposition: "executed" as const,
        ledger: {
          version: 1 as const,
          scheduler: SCHEDULER_NAME,
          runId: `${SCHEDULER_NAME}:1:no_show`,
          runKey: "internal-key",
          cron: "0 1 * * *",
          job: "no_show" as const,
          scheduledTime: 1,
          fenceToken: 2,
          status: "failed" as const,
          startedAt: 1,
          finishedAt: 2,
          failureCode: "transport" as const,
        },
      },
    ];

    await notifySchedulerFailures(
      results,
      {
        environment: "staging",
        allowedHost: "alerts.example.com",
        webhookURL: "https://alerts.example.com/scheduler",
        webhookSecret: "alert-secret",
      },
      send,
    );

    expect(send).toHaveBeenCalledOnce();
    const [request] = send.mock.calls[0] ?? [];
    expect(request).toBeInstanceOf(Request);
    if (!(request instanceof Request)) {
      throw new Error("expected Request");
    }
    expect(request.headers.get("Idempotency-Key")).toBe(
      `staging:${SCHEDULER_NAME}:1:no_show:transport`,
    );
    expect(request.redirect).toBe("manual");
    expect(request.headers.get("Authorization")).toBe("Bearer alert-secret");
    const payload = await request.json();
    expect(payload).toEqual({
      version: 1,
      alert_key: `staging:${SCHEDULER_NAME}:1:no_show:transport`,
      environment: "staging",
      scheduler: SCHEDULER_NAME,
      run_id: `${SCHEDULER_NAME}:1:no_show`,
      job: "no_show",
      scheduled_time: 1,
      status: "failed",
      failure_code: "transport",
    });
    expect(JSON.stringify(payload)).not.toContain("internal-key");
  });

  it("rejects plaintext or non-allowlisted alert endpoints before sending", async () => {
    const send = vi.fn(async (_request: Request) => new Response(null, { status: 204 }));
    const failure = {
      disposition: "executed" as const,
      ledger: {
        version: 1 as const,
        scheduler: SCHEDULER_NAME,
        runId: `${SCHEDULER_NAME}:1:no_show`,
        runKey: "internal-key",
        cron: "0 1 * * *",
        job: "no_show" as const,
        scheduledTime: 1,
        fenceToken: 2,
        status: "failed" as const,
        startedAt: 1,
        finishedAt: 2,
        failureCode: "transport" as const,
      },
    };

    await expect(
      notifySchedulerFailures(
        [failure],
        {
          environment: "staging",
          allowedHost: "alerts.example.com",
          webhookURL: "http://alerts.example.com/scheduler",
          webhookSecret: "alert-secret",
        },
        send,
      ),
    ).rejects.toThrow("scheduler alert delivery failed");
    await expect(
      notifySchedulerFailures(
        [failure],
        {
          environment: "staging",
          allowedHost: "alerts.example.com",
          webhookURL: "https://attacker.example.net/scheduler",
          webhookSecret: "alert-secret",
        },
        send,
      ),
    ).rejects.toThrow("scheduler alert delivery failed");
    expect(send).not.toHaveBeenCalled();
  });

  it("fails closed with a fixed code when alert configuration is missing", async () => {
    const errorLog = vi.spyOn(console, "error").mockImplementation(() => undefined);
    const failure = {
      disposition: "executed" as const,
      ledger: {
        version: 1 as const,
        scheduler: SCHEDULER_NAME,
        runId: `${SCHEDULER_NAME}:1:no_show`,
        runKey: "internal-key",
        cron: "0 1 * * *",
        job: "no_show" as const,
        scheduledTime: 1,
        fenceToken: 2,
        status: "failed" as const,
        startedAt: 1,
        finishedAt: 2,
        failureCode: "transport" as const,
      },
    };

    await expect(
      notifySchedulerFailures([failure], { environment: "staging" }),
    ).rejects.toThrow("scheduler alert delivery failed");
    expect(errorLog).toHaveBeenCalledWith(
      "scheduler alert is not configured",
      expect.objectContaining({
        event: "scheduler_alert_not_configured",
        failure_code: "alert_not_configured",
      }),
    );
    errorLog.mockRestore();
  });

  it("attempts every terminal alert before returning a delivery failure", async () => {
    const errorLog = vi.spyOn(console, "error").mockImplementation(() => undefined);
    const send = vi
      .fn<(request: Request) => Promise<Response>>()
      .mockResolvedValueOnce(new Response(null, { status: 503 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    const first = {
      disposition: "executed" as const,
      ledger: {
        version: 1 as const,
        scheduler: SCHEDULER_NAME,
        runId: `${SCHEDULER_NAME}:1:no_show`,
        runKey: "first-internal",
        cron: "0 1 * * *",
        job: "no_show" as const,
        scheduledTime: 1,
        fenceToken: 1,
        status: "failed" as const,
        startedAt: 1,
        finishedAt: 2,
        failureCode: "transport" as const,
      },
    };
    const second = {
      ...first,
      ledger: {
        ...first.ledger,
        runId: `${SCHEDULER_NAME}:1:delivery`,
        runKey: "second-internal",
        job: "delivery" as const,
      },
    };

    await expect(
      notifySchedulerFailures(
        [first, second],
        {
          environment: "staging",
          allowedHost: "alerts.example.com",
          webhookURL: "https://alerts.example.com/scheduler",
          webhookSecret: "alert-secret",
        },
        send,
      ),
    ).rejects.toThrow("scheduler alert delivery failed");
    expect(send).toHaveBeenCalledTimes(2);
    const terminalLogs = errorLog.mock.calls.filter(
      ([, fields]) =>
        typeof fields === "object" &&
        fields !== null &&
        "event" in fields &&
        fields.event === "scheduler_job_failed",
    );
    expect(terminalLogs).toHaveLength(2);
    expect(errorLog.mock.invocationCallOrder[1]).toBeLessThan(
      send.mock.invocationCallOrder[0] ?? Number.POSITIVE_INFINITY,
    );
    errorLog.mockRestore();
  });

  it("includes environment in alert idempotency and rejects all redirect indicators", async () => {
    const failure = {
      disposition: "executed" as const,
      ledger: {
        version: 1 as const,
        scheduler: SCHEDULER_NAME,
        runId: `${SCHEDULER_NAME}:2:no_show`,
        runKey: "internal-key",
        cron: "0 1 * * *",
        job: "no_show" as const,
        scheduledTime: 2,
        fenceToken: 2,
        status: "failed" as const,
        startedAt: 2,
        finishedAt: 3,
        failureCode: "transport" as const,
      },
    };
    const requests: Request[] = [];
    const followedRedirect = new Response(null, { status: 204 });
    Object.defineProperty(followedRedirect, "redirected", { value: true });
    const responses = [
      followedRedirect,
      new Response(null, {
        status: 302,
        headers: { Location: "https://attacker.example.com/alerts" },
      }),
      new Response(null, {
        status: 200,
        headers: { Location: "https://attacker.example.com/alerts" },
      }),
    ];
    const send = vi.fn(async (request: Request) => {
      requests.push(request);
      const response = responses.shift();
      if (response === undefined) {
        throw new Error("unexpected alert request");
      }
      return response;
    });
    const errorLog = vi.spyOn(console, "error").mockImplementation(() => undefined);

    try {
      for (let attempt = 0; attempt < 3; attempt += 1) {
        await expect(
          notifySchedulerFailures(
            [failure],
            {
              environment: "production",
              allowedHost: "alerts.example.com",
              webhookURL: "https://alerts.example.com/scheduler",
              webhookSecret: "alert-secret",
            },
            send,
          ),
        ).rejects.toThrow("scheduler alert delivery failed");
      }
      expect(send).toHaveBeenCalledTimes(3);
      expect(requests).toHaveLength(3);
      for (const request of requests) {
        expect(request.redirect).toBe("manual");
        expect(request.headers.get("Idempotency-Key")).toBe(
          `production:${SCHEDULER_NAME}:2:no_show:transport`,
        );
      }
      expect(errorLog).toHaveBeenCalledWith(
        "scheduler alert delivery failed",
        expect.objectContaining({
          event: "scheduler_alert_delivery_failed",
          failure_code: "alert_delivery_redirect",
        }),
      );
    } finally {
      errorLog.mockRestore();
    }
  });

  it("times out alert delivery and emits a fixed transport failure code", async () => {
    vi.useFakeTimers();
    const errorLog = vi.spyOn(console, "error").mockImplementation(() => undefined);
    const failure = {
      disposition: "executed" as const,
      ledger: {
        version: 1 as const,
        scheduler: SCHEDULER_NAME,
        runId: `${SCHEDULER_NAME}:3:no_show`,
        runKey: "internal-key",
        cron: "0 1 * * *",
        job: "no_show" as const,
        scheduledTime: 3,
        fenceToken: 3,
        status: "failed" as const,
        startedAt: 3,
        finishedAt: 4,
        failureCode: "transport" as const,
      },
    };
    const pending = notifySchedulerFailures(
      [failure],
      {
        environment: "staging",
        allowedHost: "alerts.example.com",
        webhookURL: "https://alerts.example.com/scheduler",
        webhookSecret: "alert-secret",
      },
      async () => new Promise<Response>(() => undefined),
    );
    const rejected = expect(pending).rejects.toThrow(
      "scheduler alert delivery failed",
    );

    await vi.advanceTimersByTimeAsync(5_001);
    await rejected;
    expect(errorLog).toHaveBeenCalledWith(
      "scheduler alert delivery failed",
      expect.objectContaining({
        event: "scheduler_alert_delivery_failed",
        failure_code: "alert_delivery_transport",
      }),
    );
    errorLog.mockRestore();
    vi.useRealTimers();
  });

  it("does not send success, paused, or successful duplicate events", async () => {
    const send = vi.fn(async () => new Response(null, { status: 204 }));
    const success = manualFixture().result;
    if (success === undefined) {
      throw new Error("expected manual result");
    }
    success.ledger = {
      version: 1,
      scheduler: SCHEDULER_NAME,
      runId: `${SCHEDULER_NAME}:1:no_show`,
      runKey: "internal-success-key",
      cron: "0 1 * * *",
      job: "no_show",
      scheduledTime: 1,
      fenceToken: 1,
      status: "success",
      startedAt: 1,
      finishedAt: 2,
      outcome: {
        outcome: "success",
        processed: 1,
        succeeded: 1,
        failed: 0,
      },
    };

    await notifySchedulerFailures(
      [
        success,
        { disposition: "paused" },
        { disposition: "duplicate" },
      ],
      {
        environment: "staging",
        allowedHost: "alerts.example.com",
        webhookURL: "https://alerts.example.com/scheduler",
        webhookSecret: "alert-secret",
      },
      send,
    );

    expect(send).not.toHaveBeenCalled();
  });
});
