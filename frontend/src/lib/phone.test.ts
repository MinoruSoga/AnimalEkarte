import { describe, expect, it } from "vitest";
import { isValidOwnerPhone } from "./phone";

describe("isValidOwnerPhone", () => {
  it("accepts digit-rich Japanese-style numbers", () => {
    expect(isValidOwnerPhone("090-1234-5678")).toBe(true);
    expect(isValidOwnerPhone("+81 90 1234 5678")).toBe(true);
  });

  it("rejects too few digits", () => {
    expect(isValidOwnerPhone("123-456")).toBe(false);
  });

  it("rejects invalid characters", () => {
    expect(isValidOwnerPhone("090-1234-5678a")).toBe(false);
  });
});
