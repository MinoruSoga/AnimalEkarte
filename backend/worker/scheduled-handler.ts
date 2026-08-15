import { jobsForCron } from "./scheduled-jobs";
import {
  scheduledRunRequiresCronFailure,
  type ScheduledRunResult,
} from "./scheduler-coordinator";

export interface ScheduledControllerInput {
  readonly scheduledTime: number;
  readonly cron: string;
  noRetry(): void;
}

export type ScheduledCoordinatorInvocation = (
  cron: string,
  scheduledTime: number,
) => Promise<readonly ScheduledRunResult[]>;

export async function dispatchScheduledEvent(
  controller: ScheduledControllerInput,
  invoke: ScheduledCoordinatorInvocation,
): Promise<readonly ScheduledRunResult[]> {
  let expectedJobs: ReturnType<typeof jobsForCron>;
  try {
    expectedJobs = jobsForCron(controller.cron);
  } catch {
    controller.noRetry();
    throw new Error("scheduled invocation rejected");
  }

  let results: readonly ScheduledRunResult[];
  try {
    results = await invoke(controller.cron, controller.scheduledTime);
  } catch {
    controller.noRetry();
    throw new Error("scheduled invocation rejected");
  }

  const matchesPlan =
    results.length === expectedJobs.length &&
    results.every(
      (result, index) =>
        result.ledger?.job === expectedJobs[index] &&
        result.ledger.scheduledTime === controller.scheduledTime &&
        result.ledger.cron === controller.cron,
    );
  if (!matchesPlan || results.some(scheduledRunRequiresCronFailure)) {
    controller.noRetry();
    throw new Error("scheduled invocation failed");
  }
  return results;
}
