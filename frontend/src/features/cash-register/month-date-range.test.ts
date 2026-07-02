import { describe, it, expect } from "vitest";
import { monthDateRange } from "./month-date-range";

describe("monthDateRange", () => {
  it("returns the first and last day of a 30-day month", () => {
    expect(monthDateRange(2026, 6)).toEqual({
      startDate: "2026-06-01",
      endDate: "2026-06-30",
    });
  });

  it("returns Dec 31 for December without rolling into the next year", () => {
    expect(monthDateRange(2026, 12)).toEqual({
      startDate: "2026-12-01",
      endDate: "2026-12-31",
    });
  });

  it("returns Feb 29 for a leap-year February", () => {
    expect(monthDateRange(2024, 2)).toEqual({
      startDate: "2024-02-01",
      endDate: "2024-02-29",
    });
  });

  it("returns Feb 28 for a non-leap February", () => {
    expect(monthDateRange(2026, 2)).toEqual({
      startDate: "2026-02-01",
      endDate: "2026-02-28",
    });
  });

  it("zero-pads single-digit months", () => {
    expect(monthDateRange(2026, 1)).toEqual({
      startDate: "2026-01-01",
      endDate: "2026-01-31",
    });
  });
});
