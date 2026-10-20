import { describe, expect, it } from "vitest";

import {
  isRevisitIntakeAppointmentIdUsable,
  toRevisitIntakeAppointmentId,
} from "./revisit-intake-appointment-id";

describe("revisit-intake appointment id guard", () => {
  it("正の安全整数だけを decimal string にする", () => {
    expect(toRevisitIntakeAppointmentId(12001)).toBe("12001");
    expect(toRevisitIntakeAppointmentId(Number.MAX_SAFE_INTEGER)).toBe("9007199254740991");
  });

  it.each([
    ["MAX_SAFE_INTEGER 超", Number.MAX_SAFE_INTEGER + 1],
    ["丸められた巨大値", 9223372036854775807],
    ["0", 0],
    ["負数", -1],
    ["小数", 1.5],
    ["NaN", Number.NaN],
    ["Infinity", Number.POSITIVE_INFINITY],
  ])("%s は使えず null", (_label, id) => {
    expect(isRevisitIntakeAppointmentIdUsable(id)).toBe(false);
    expect(toRevisitIntakeAppointmentId(id)).toBeNull();
  });
});
