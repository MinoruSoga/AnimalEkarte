import { describe, expect, it } from "vitest";

import { resolveLatestVitalWeight } from "./medicine-dose-lookup";
import type { Vital } from "../types";

function makeVital(overrides: Partial<Vital>): Vital {
  return {
    id: "1",
    medical_record_id: "1",
    recorded_at: "2026-07-12T00:00:00Z",
    weight_unit: "Kg",
    created_at: "2026-07-12T00:00:00Z",
    updated_at: "2026-07-12T00:00:00Z",
    ...overrides,
  };
}

describe("resolveLatestVitalWeight", () => {
  it("体重未記録（vitals 空）は null（fail-closed）", () => {
    expect(resolveLatestVitalWeight([])).toBeNull();
  });

  it("weight が全て null/0 以下の場合は null（fail-closed）", () => {
    const vitals = [
      makeVital({ id: "1", weight: null, recorded_at: "2026-07-12T09:00:00Z" }),
      makeVital({ id: "2", weight: 0, recorded_at: "2026-07-12T10:00:00Z" }),
      makeVital({ id: "3", weight: -1, recorded_at: "2026-07-12T11:00:00Z" }),
    ];
    expect(resolveLatestVitalWeight(vitals)).toBeNull();
  });

  it("複数 vital のうち recorded_at が最新のものを採用する", () => {
    const vitals = [
      makeVital({ id: "10", weight: 4.0, weight_unit: "Kg", recorded_at: "2026-07-12T09:00:00Z" }),
      makeVital({ id: "11", weight: 4.5, weight_unit: "Kg", recorded_at: "2026-07-12T15:00:00Z" }),
      makeVital({ id: "12", weight: 4.2, weight_unit: "Kg", recorded_at: "2026-07-12T12:00:00Z" }),
    ];
    const got = resolveLatestVitalWeight(vitals);
    expect(got).not.toBeNull();
    expect(got?.weightKg).toBe(4.5);
    expect(got?.source).toBe("vital_records:11");
  });

  it("最新の記録に体重が無い場合は体重が記録された直近のものへフォールバックする", () => {
    const vitals = [
      makeVital({ id: "20", weight: 4.0, weight_unit: "Kg", recorded_at: "2026-07-12T09:00:00Z" }),
      makeVital({ id: "21", weight: null, recorded_at: "2026-07-12T15:00:00Z" }),
    ];
    const got = resolveLatestVitalWeight(vitals);
    expect(got?.weightKg).toBe(4.0);
    expect(got?.source).toBe("vital_records:20");
  });

  it("g 単位は kg へ正規化する", () => {
    const vitals = [
      makeVital({ id: "30", weight: 850, weight_unit: "g", recorded_at: "2026-07-12T09:00:00Z" }),
    ];
    const got = resolveLatestVitalWeight(vitals);
    expect(got?.weightKg).toBeCloseTo(0.85, 6);
  });

  it("healthcare-review-201 LOW: recorded_at が不正な行を無視し、有効な最新行を採用する", () => {
    const vitals = [
      makeVital({ id: "40", weight: 4.0, weight_unit: "Kg", recorded_at: "2026-07-12T09:00:00Z" }),
      makeVital({ id: "41", weight: 4.9, weight_unit: "Kg", recorded_at: "not-a-date" }),
      makeVital({ id: "42", weight: 4.5, weight_unit: "Kg", recorded_at: "2026-07-12T15:00:00Z" }),
    ];
    const got = resolveLatestVitalWeight(vitals);
    expect(got?.weightKg).toBe(4.5);
    expect(got?.source).toBe("vital_records:42");
  });
});
