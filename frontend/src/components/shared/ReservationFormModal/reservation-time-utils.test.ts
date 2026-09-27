import { describe, it, expect } from "vitest";
import {
  resolveEndTimeOnStartChange,
  slotTimeToSelectValue,
  slotVacancyLabel,
} from "./reservation-time-utils";

describe("slotTimeToSelectValue", () => {
  it("returns HH:mm unchanged when already normalized", () => {
    expect(slotTimeToSelectValue("09:00")).toBe("09:00");
    expect(slotTimeToSelectValue("09:45")).toBe("09:45");
  });

  it("pads a single-digit hour (H:mm)", () => {
    expect(slotTimeToSelectValue("9:00")).toBe("09:00");
  });

  it("inserts a colon for a colon-less HHmm value", () => {
    expect(slotTimeToSelectValue("0900")).toBe("09:00");
  });

  it("defaults an empty minute segment to 00 instead of producing an invalid value", () => {
    expect(slotTimeToSelectValue("09:")).toBe("09:00");
  });

  it("pads a single-digit minute segment", () => {
    expect(slotTimeToSelectValue("09:5")).toBe("09:05");
  });
});

describe("slotVacancyLabel (EMR-170)", () => {
  it("maps each status to a symbol + text label (not color-only)", () => {
    expect(slotVacancyLabel("available")).toBe("〇 空きあり");
    expect(slotVacancyLabel("low")).toBe("△ 残り1枠");
    expect(slotVacancyLabel("full")).toBe("✕ 満員");
  });
});

describe("resolveEndTimeOnStartChange (EMR-191)", () => {
  const start = new Date(2026, 8, 27, 10, 0, 0, 0);

  it("adds the type duration to the start time when no slot end exists", () => {
    const end = resolveEndTimeOnStartChange(start, 15, undefined);
    expect(end.getHours()).toBe(10);
    expect(end.getMinutes()).toBe(15);
  });

  it("uses the LINE slot end when one exists", () => {
    const end = resolveEndTimeOnStartChange(start, 15, "10:30");
    expect(end.getHours()).toBe(10);
    expect(end.getMinutes()).toBe(30);
  });

  it("falls back to the 15-minute default when duration is undefined", () => {
    const end = resolveEndTimeOnStartChange(start, undefined, undefined);
    expect(end.getHours()).toBe(10);
    expect(end.getMinutes()).toBe(15);
  });

  it("falls back to the 15-minute default when duration is 0", () => {
    const end = resolveEndTimeOnStartChange(start, 0, undefined);
    expect(end.getHours()).toBe(10);
    expect(end.getMinutes()).toBe(15);
  });

  it("rolls over into the next day past midnight", () => {
    const lateStart = new Date(2026, 8, 27, 23, 50, 0, 0);
    const end = resolveEndTimeOnStartChange(lateStart, 30, undefined);
    expect(end.getDate()).toBe(28);
    expect(end.getHours()).toBe(0);
    expect(end.getMinutes()).toBe(20);
  });

  it("does not mutate the input start date", () => {
    const input = new Date(2026, 8, 27, 10, 0, 0, 0);
    resolveEndTimeOnStartChange(input, 45, undefined);
    expect(input.getMinutes()).toBe(0);
  });
});
