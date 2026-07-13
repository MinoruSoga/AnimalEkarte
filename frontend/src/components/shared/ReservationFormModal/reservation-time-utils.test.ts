import { describe, it, expect } from "vitest";
import { slotTimeToSelectValue } from "./reservation-time-utils";

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
