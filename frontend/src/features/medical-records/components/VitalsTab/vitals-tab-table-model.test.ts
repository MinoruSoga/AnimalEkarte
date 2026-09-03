import { describe, expect, it } from "vitest";

import { parseVitalsNumber, toggleWeightValueAndUnit } from "../../lib/vitals-tab-table-model";

describe("toggleWeightValueAndUnit", () => {
  it("converts 5 Kg to 5000 g and back without changing physical mass", () => {
    const toG = toggleWeightValueAndUnit("5", "Kg");
    expect(toG).toEqual({ weight: "5000", weight_unit: "g" });

    const backToKg = toggleWeightValueAndUnit(toG.weight, toG.weight_unit);
    expect(backToKg).toEqual({ weight: "5", weight_unit: "Kg" });
    expect(parseVitalsNumber(backToKg.weight)).toBe(5);
  });

  it("converts 8.523 g to 0.008523 Kg and back without rounding loss", () => {
    const toKg = toggleWeightValueAndUnit("8.523", "g");
    expect(toKg.weight_unit).toBe("Kg");
    expect(parseVitalsNumber(toKg.weight)).toBe(0.008523);

    const backToG = toggleWeightValueAndUnit(toKg.weight, toKg.weight_unit);
    expect(backToG.weight_unit).toBe("g");
    expect(parseVitalsNumber(backToG.weight)).toBe(8.523);
  });

  it("toggles unit only when weight is empty (no physical mass to convert)", () => {
    expect(toggleWeightValueAndUnit("", "Kg")).toEqual({
      weight: "",
      weight_unit: "g",
    });
    expect(toggleWeightValueAndUnit("   ", "g")).toEqual({
      weight: "   ",
      weight_unit: "Kg",
    });
  });

  it("toggles unit only when weight is unparsable", () => {
    expect(toggleWeightValueAndUnit("abc", "Kg")).toEqual({
      weight: "abc",
      weight_unit: "g",
    });
  });

  it("preserves physical mass for the BUG-015 failure mode (8.5 Kg → g)", () => {
    // Old bug: unit label flipped to g while value stayed 8.5 (8.5 g stored).
    const converted = toggleWeightValueAndUnit("8.5", "Kg");
    expect(converted).toEqual({ weight: "8500", weight_unit: "g" });
    expect(parseVitalsNumber(converted.weight)).toBe(8500);
  });
});
