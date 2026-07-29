import { describe, expect, it, vi } from "vitest";
import {
  SCHEDULED_JOB_FETCH_TIMEOUT_MS,
  SCHEDULED_JOB_GO_TIMEOUT_MS,
  SCHEDULED_JOB_LEASE_MS,
  SCHEDULER_NAME,
  jobsForCron,
  isScheduledJobsInternalPath,
  runScheduledJobRequest,
  type ScheduledJobRequest,
} from "./scheduled-jobs";

function requestFor(job: ScheduledJobRequest["job"]): ScheduledJobRequest {
  return {
    scheduler: SCHEDULER_NAME,
    job,
    scheduled_time: 1_721_870_400_000,
    run_id: `${SCHEDULER_NAME}:1721870400000:${job}`,
    fence_token: 7,
  };
}

function encodePath(path: string, passes: number): string {
  return Array.from({ length: passes }).reduce<string>(
    (encoded) => encodeURIComponent(encoded),
    path,
  );
}

describe("scheduled job constants", () => {
  it("keeps the approved stable coordinator name and timeout hierarchy", () => {
    expect(SCHEDULER_NAME).toBe("animalekarte-scheduler-v1");
    expect(SCHEDULED_JOB_GO_TIMEOUT_MS).toBe(100_000);
    expect(SCHEDULED_JOB_FETCH_TIMEOUT_MS).toBe(110_000);
    expect(SCHEDULED_JOB_LEASE_MS).toBe(150_000);
    expect(SCHEDULED_JOB_GO_TIMEOUT_MS).toBeLessThan(SCHEDULED_JOB_FETCH_TIMEOUT_MS);
    expect(SCHEDULED_JOB_FETCH_TIMEOUT_MS).toBeLessThan(SCHEDULED_JOB_LEASE_MS);
  });
});

describe("jobsForCron", () => {
  it("runs no_show then delivery at 01:00 UTC", () => {
    expect(jobsForCron("0 1 * * *")).toEqual(["no_show", "delivery"]);
  });

  it("runs no_show at 06:00 and 11:00 UTC", () => {
    expect(jobsForCron("0 6,11 * * *")).toEqual(["no_show"]);
  });

  it("runs dormant at 17:00 UTC", () => {
    expect(jobsForCron("0 17 * * *")).toEqual(["dormant"]);
  });

  it("fails closed for an unknown cron expression", () => {
    expect(() => jobsForCron("*/5 * * * *")).toThrowError(/unknown cron/i);
  });
});

describe("isScheduledJobsInternalPath", () => {
  it.each([
    "/_internal/scheduled-jobs",
    "/_internal/scheduled-jobs/no_show:run",
    "/_INTERNAL/SCHEDULED-JOBS/delivery:run",
    "/_internal%2Fscheduled-jobs%2Fdormant%3Arun",
    "/%5Finternal/%73cheduled-jobs/no_show:run",
    "/_internal%25252525252Fscheduled-jobs%25252525252Fdelivery%253Arun",
    encodePath("/_internal/scheduled-jobs/no_show:run", 20),
    "/public/../_internal/scheduled-jobs/dormant:run",
    "/_internal/ignored/../scheduled-jobs/no_show:run",
    String.raw`/_internal\scheduled-jobs\delivery:run`,
    "//_internal//scheduled-jobs//dormant:run",
    "/_internal/%ZZ/scheduled-jobs/no_show:run",
    "/api/%ZZ",
  ])("denies the internal scheduler path %s", (path) => {
    expect(isScheduledJobsInternalPath(path)).toBe(true);
  });

  it.each(["/_internal/migrate", "/api/v1/scheduled-jobs", "/_internal/scheduled-job"])(
    "does not claim unrelated path %s",
    (path) => {
      expect(isScheduledJobsInternalPath(path)).toBe(false);
    },
  );

  it("bounds decode work and fails closed for extreme nested encoding", () => {
    const deeplyEncoded = encodePath("/_internal/scheduled-jobs/no_show:run", 128);
    const decode = vi.spyOn(globalThis, "decodeURIComponent");

    expect(isScheduledJobsInternalPath(deeplyEncoded)).toBe(true);
    expect(decode).toHaveBeenCalledTimes(16);
    decode.mockRestore();
  });
});

describe("runScheduledJobRequest", () => {
  it("posts the typed request to the job-specific internal endpoint", async () => {
    const fetcher = vi.fn(async (request: Request) => {
      expect(request.method).toBe("POST");
      expect(new URL(request.url).pathname).toBe("/_internal/scheduled-jobs/no_show:run");
      expect(request.headers.get("Content-Type")).toBe("application/json");
      await expect(request.json()).resolves.toEqual(requestFor("no_show"));
      return Response.json({
        outcome: "success",
        processed: 4,
        succeeded: 4,
        failed: 0,
      });
    });

    await expect(runScheduledJobRequest(fetcher, requestFor("no_show"))).resolves.toEqual({
      outcome: "success",
      processed: 4,
      succeeded: 4,
      failed: 0,
    });
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  // Go requireSchedulerInternalToken expects header X-Scheduler-Token matching
  // SCHEDULER_INTERNAL_TOKEN (batch_scheduler.go). Worker must send it when configured.
  it("sends X-Scheduler-Token when schedulerInternalToken is configured", async () => {
    const token = "test-scheduler-internal-token-32b!!";
    const fetcher = vi.fn(async (request: Request) => {
      expect(request.headers.get("Content-Type")).toBe("application/json");
      expect(request.headers.get("X-Scheduler-Token")).toBe(token);
      return Response.json({
        outcome: "success",
        processed: 1,
        succeeded: 1,
        failed: 0,
      });
    });

    await expect(
      runScheduledJobRequest(fetcher, requestFor("no_show"), token),
    ).resolves.toMatchObject({ outcome: "success" });
    expect(fetcher).toHaveBeenCalledTimes(1);
  });

  // Explicit contract when unset: omit the privilege header entirely.
  // Never put undefined/empty into the header map (silent wrong auth).
  // Go fails closed on empty expected token (401); Worker must not invent a value.
  it.each([undefined, ""] as const)(
    "omits X-Scheduler-Token when schedulerInternalToken is %s",
    async (token) => {
      const fetcher = vi.fn(async (request: Request) => {
        expect(request.headers.get("Content-Type")).toBe("application/json");
        expect(request.headers.get("X-Scheduler-Token")).toBeNull();
        return Response.json({
          outcome: "success",
          processed: 0,
          succeeded: 0,
          failed: 0,
        });
      });

      await expect(
        runScheduledJobRequest(fetcher, requestFor("delivery"), token),
      ).resolves.toMatchObject({ outcome: "success" });
      expect(fetcher).toHaveBeenCalledTimes(1);
    },
  );

  it.each(["partial", "failed"] as const)("preserves a structured %s outcome", async (outcome) => {
    const fetcher = vi.fn(async () =>
      Response.json(
        {
          outcome,
          processed: 3,
          succeeded: outcome === "partial" ? 2 : 0,
          failed: outcome === "partial" ? 1 : 3,
        },
        { status: outcome === "failed" ? 500 : 200 },
      ),
    );

    await expect(runScheduledJobRequest(fetcher, requestFor("delivery"))).resolves.toMatchObject({
      outcome,
    });
  });

  it("rejects an HTTP error that falsely claims success", async () => {
    const fetcher = vi.fn(async () =>
      Response.json(
        { outcome: "success", processed: 0, succeeded: 0, failed: 0 },
        { status: 500 },
      ),
    );

    await expect(runScheduledJobRequest(fetcher, requestFor("dormant"))).rejects.toThrowError(
      /non-success HTTP status/i,
    );
  });

  it.each([
    {},
    { outcome: "unknown", processed: 0, succeeded: 0, failed: 0 },
    { outcome: "success", processed: -1, succeeded: 0, failed: 0 },
    { outcome: "success", processed: 1, succeeded: 2, failed: 0 },
    { outcome: "success", processed: 1, succeeded: 0, failed: 1 },
    { outcome: "partial", processed: 1, succeeded: 1, failed: 0 },
    { outcome: "partial", processed: 1, succeeded: 0, failed: 1 },
    { outcome: "failed", processed: 2, succeeded: 1, failed: 1 },
    { outcome: "failed", processed: 0, succeeded: 0, failed: 0 },
    { outcome: "failed", processed: 3, succeeded: 0, failed: 2 },
  ])("rejects malformed structured output %#", async (body) => {
    const fetcher = vi.fn(async () => Response.json(body));

    await expect(runScheduledJobRequest(fetcher, requestFor("no_show"))).rejects.toThrowError(
      /invalid scheduled job outcome/i,
    );
  });

  it("aborts containerFetch at the Worker timeout", async () => {
    vi.useFakeTimers();
    const fetcher = vi.fn(
      (request: Request) =>
        new Promise<Response>((_resolve, reject) => {
          request.signal.addEventListener("abort", () => reject(request.signal.reason));
        }),
    );

    const result = runScheduledJobRequest(fetcher, requestFor("no_show"));
    const rejection = expect(result).rejects.toThrowError(/timed out/i);
    await vi.advanceTimersByTimeAsync(SCHEDULED_JOB_FETCH_TIMEOUT_MS);

    await rejection;
    vi.useRealTimers();
  });

  it("rejects at the Worker timeout even when containerFetch ignores AbortSignal", async () => {
    vi.useFakeTimers();
    const fetcher = vi.fn(() => new Promise<Response>(() => undefined));

    const result = runScheduledJobRequest(fetcher, requestFor("delivery"));
    const rejection = expect(result).rejects.toThrowError(/timed out/i);
    await vi.advanceTimersByTimeAsync(SCHEDULED_JOB_FETCH_TIMEOUT_MS);

    await rejection;
    vi.useRealTimers();
  });
});
