import { describe, expect, it, vi } from "vitest";
import { dispatchScheduledEvent } from "./scheduled-handler";
import type { ScheduledRunResult } from "./scheduler-coordinator";

function controller(cron: string, scheduledTime = 1_721_870_400_000) {
  return {
    cron,
    scheduledTime,
    noRetry: vi.fn(),
  };
}

function result(status: "running" | "success" | "partial" | "failed"): ScheduledRunResult {
  return {
    disposition: status === "running" ? "duplicate" : "executed",
    ledger: {
      version: 1,
      scheduler: "animalekarte-scheduler-v1",
      runId: `run-${status}`,
      runKey: `scheduler:run:${status}`,
      cron: "0 1 * * *",
      job: "no_show",
      scheduledTime: 1_721_870_400_000,
      fenceToken: 1,
      status,
      startedAt: 1_721_870_400_000,
    },
  };
}

function resultFor(
  job: "no_show" | "delivery" | "dormant",
  cron: string,
  scheduledTime: number,
): ScheduledRunResult {
  return {
    ...result("success"),
    ledger: {
      ...result("success").ledger!,
      job,
      cron,
      scheduledTime,
    },
  };
}

describe("dispatchScheduledEvent", () => {
  it("passes the exact cron and scheduledTime to the named coordinator", async () => {
    const event = controller("0 6,11 * * *");
    const invoke = vi.fn(async () => [resultFor("no_show", event.cron, event.scheduledTime)]);

    await expect(dispatchScheduledEvent(event, invoke)).resolves.toHaveLength(1);
    expect(invoke).toHaveBeenCalledWith(event.cron, event.scheduledTime);
    expect(event.noRetry).not.toHaveBeenCalled();
  });

  it("fails closed before RPC for an unknown cron", async () => {
    const event = controller("*/5 * * * *");
    const invoke = vi.fn(async () => [result("success")]);

    await expect(dispatchScheduledEvent(event, invoke)).rejects.toThrowError(
      /scheduled invocation rejected/i,
    );
    expect(invoke).not.toHaveBeenCalled();
    expect(event.noRetry).toHaveBeenCalledTimes(1);
  });

  it.each(["partial", "failed"] as const)(
    "marks a terminal %s result as a failed Cron invocation with no retry",
    async (status) => {
      const event = controller("0 1 * * *");
      const invoke = vi.fn(async () => [
        { ...resultFor("no_show", event.cron, event.scheduledTime), ledger: result(status).ledger },
        resultFor("delivery", event.cron, event.scheduledTime),
      ]);

      await expect(dispatchScheduledEvent(event, invoke)).rejects.toThrowError(
        /scheduled invocation failed/i,
      );
      expect(event.noRetry).toHaveBeenCalledTimes(1);
    },
  );

  it("accepts an exact in-progress duplicate without running it again", async () => {
    const event = controller("0 1 * * *");
    const invoke = vi.fn(async () => [
      {
        ...result("running"),
        ledger: {
          ...result("running").ledger!,
          cron: event.cron,
          scheduledTime: event.scheduledTime,
        },
      },
      resultFor("delivery", event.cron, event.scheduledTime),
    ]);

    await expect(dispatchScheduledEvent(event, invoke)).resolves.toHaveLength(2);
    expect(event.noRetry).not.toHaveBeenCalled();
  });

  it.each([
    { results: [] },
    { results: [resultFor("delivery", "0 6,11 * * *", 1_721_870_400_000)] },
    { results: [resultFor("no_show", "0 17 * * *", 1_721_870_400_000)] },
  ])("fails closed for a malformed coordinator plan %#", async ({ results }) => {
    const event = controller("0 6,11 * * *");
    const invoke = vi.fn(async () => results);

    await expect(dispatchScheduledEvent(event, invoke)).rejects.toThrowError(
      /scheduled invocation failed/i,
    );
    expect(event.noRetry).toHaveBeenCalledTimes(1);
  });

  it("prevents retry when coordinator RPC itself fails", async () => {
    const event = controller("0 17 * * *");
    const invoke = vi.fn(async (): Promise<readonly ScheduledRunResult[]> => {
      throw new Error("RPC details must not escape");
    });

    await expect(dispatchScheduledEvent(event, invoke)).rejects.toThrowError(
      /^scheduled invocation rejected$/,
    );
    expect(event.noRetry).toHaveBeenCalledTimes(1);
  });
});
