import { describe, expect, it } from "vitest";

import type { Vital } from "../types";
import { latestVisitVitalChips } from "./visit-vital-chips";

function vital(overrides: Partial<Vital>): Vital {
  return {
    id: "1",
    medical_record_id: "9",
    recorded_at: "2026-09-20T10:00:00+09:00",
    weight_unit: "kg",
    version: 1,
    created_at: "2026-09-20T10:00:00+09:00",
    updated_at: "2026-09-20T10:00:00+09:00",
    ...overrides,
  };
}

describe("latestVisitVitalChips", () => {
  it("returns null when there are no vitals", () => {
    expect(latestVisitVitalChips(undefined)).toBeNull();
    expect(latestVisitVitalChips([])).toBeNull();
  });

  it("uses the latest recorded_at row and omits time", () => {
    const chips = latestVisitVitalChips([
      vital({
        id: "old",
        recorded_at: "2026-09-20T09:00:00+09:00",
        temperature: 37.1,
      }),
      vital({
        id: "new",
        recorded_at: "2026-09-20T11:00:00+09:00",
        temperature: 38.5,
        heart_rate: 120,
        respiration_rate: 24,
        weight: 4.2,
      }),
    ]);
    expect(chips).toEqual({
      temperature: 38.5,
      heartRate: 120,
      respirationRate: 24,
      weight: 4.2,
      weightUnit: "kg",
    });
  });

  it("returns null when the latest row has no T/HR/RR/weight", () => {
    expect(latestVisitVitalChips([vital({ note: "only note" })])).toBeNull();
  });
});
