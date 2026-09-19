import { describe, expect, it } from "vitest";

import { reduceQuantityEnterKey } from "./treatment-quantity-commit";

describe("reduceQuantityEnterKey — quantity Enter arming", () => {
  it("first Enter arms without commit", () => {
    expect(reduceQuantityEnterKey("idle", "Enter")).toEqual({
      phase: "armed",
      action: "none",
    });
  });

  it("second Enter commits and returns to idle", () => {
    expect(reduceQuantityEnterKey("armed", "Enter")).toEqual({
      phase: "idle",
      action: "commit",
    });
  });

  it("Escape cancels from idle or armed", () => {
    expect(reduceQuantityEnterKey("idle", "Escape")).toEqual({
      phase: "idle",
      action: "cancel",
    });
    expect(reduceQuantityEnterKey("armed", "Escape")).toEqual({
      phase: "idle",
      action: "cancel",
    });
  });

  it("other keys leave phase unchanged", () => {
    expect(reduceQuantityEnterKey("armed", "Tab")).toEqual({
      phase: "armed",
      action: "none",
    });
  });
});
