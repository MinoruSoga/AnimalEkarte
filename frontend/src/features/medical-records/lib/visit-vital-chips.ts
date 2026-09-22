import type { Vital } from "../types";

export interface VisitVitalChipValues {
  temperature?: number;
  heartRate?: number;
  respirationRate?: number;
  weight?: number;
  weightUnit?: string;
}

function isPresent(value: number | null | undefined): value is number {
  return typeof value === "number" && Number.isFinite(value);
}

export function latestVisitVitalChips(
  vitals: readonly Vital[] | undefined,
): VisitVitalChipValues | null {
  if (!vitals || vitals.length === 0) {
    return null;
  }
  const sorted = [...vitals].sort((a, b) => b.recorded_at.localeCompare(a.recorded_at));
  const latest = sorted[0];
  const chips: VisitVitalChipValues = {};
  if (isPresent(latest.temperature)) {
    chips.temperature = latest.temperature;
  }
  if (isPresent(latest.heart_rate)) {
    chips.heartRate = latest.heart_rate;
  }
  if (isPresent(latest.respiration_rate)) {
    chips.respirationRate = latest.respiration_rate;
  }
  if (isPresent(latest.weight)) {
    chips.weight = latest.weight;
    chips.weightUnit = latest.weight_unit;
  }
  if (
    chips.temperature == null &&
    chips.heartRate == null &&
    chips.respirationRate == null &&
    chips.weight == null
  ) {
    return null;
  }
  return chips;
}
