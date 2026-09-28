import { describe, it, expect } from "vitest";
import {
  resolveEndTimeOnStartChange,
  slotTimeToSelectValue,
  slotVacancyLabel,
  withCurrentTimeOption,
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

describe("withCurrentTimeOption (BUG-015/EMR-191)", () => {
  it("appends an off-grid current value at the end", () => {
    expect(withCurrentTimeOption(["10:00", "10:15"], "10:20")).toEqual(["10:00", "10:15", "10:20"]);
  });

  it("does not grow options when the value is already on the 15-minute grid", () => {
    const options = ["10:00", "10:15"];
    expect(withCurrentTimeOption(options, "10:15")).toBe(options);
  });

  it("returns the options unchanged when there is no current value", () => {
    const options = ["10:00"];
    expect(withCurrentTimeOption(options, undefined)).toBe(options);
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
