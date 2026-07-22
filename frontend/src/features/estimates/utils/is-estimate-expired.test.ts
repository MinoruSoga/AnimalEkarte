import { describe, expect, it } from "vitest";

import { isEstimateExpired } from "./is-estimate-expired";

describe("isEstimateExpired", () => {
  it.each([
    [null, "2026-07-22", false],
    ["", "2026-07-22", false],
    ["2026-07-21", "2026-07-22", true],
    ["2026-07-22", "2026-07-22", false],
    ["2026-07-23", "2026-07-22", false],
  ])("validUntil=%s, today=%s → %s", (validUntil, today, expected) => {
    expect(isEstimateExpired(validUntil, today)).toBe(expected);
  });
});
