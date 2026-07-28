import { describe, it, expect } from "vitest";
import { isEstimateLockedStatus } from "./is-estimate-locked-status";

describe("isEstimateLockedStatus", () => {
  it("approved → true", () => {
    expect(isEstimateLockedStatus("approved")).toBe(true);
  });

  it("rejected → true", () => {
    expect(isEstimateLockedStatus("rejected")).toBe(true);
  });

  it("draft → false", () => {
    expect(isEstimateLockedStatus("draft")).toBe(false);
  });

  it("sent → false", () => {
    expect(isEstimateLockedStatus("sent")).toBe(false);
  });
});
