import { C } from "@/lib/design-tokens";
import { formatJSTWallDate, formatJSTWallTime, toJSTWallDate } from "@/lib/jst-date";
import type { BodyWeightUnit } from "../../types";

export const EDIT_INPUT_CLASS = `h-8 text-sm border ${C.borderMedium} rounded-xxs px-2 ${C.bgWhite} ${C.text} outline-none ${C.focusBorderAccent} w-full`;
export const ADD_INPUT_CLASS = `h-8 text-sm border ${C.borderMedium} rounded-xxs px-2 ${C.bgWhite} ${C.text} outline-none ${C.focusBorderAccent}`;

export interface VitalsAddFormState {
  recorded_at: string;
  temperature: string;
  heart_rate: string;
  respiration_rate: string;
  weight: string;
  weight_unit: BodyWeightUnit;
  note: string;
}

export const EMPTY_VITALS_ADD_FORM: VitalsAddFormState = {
  recorded_at: "",
  temperature: "",
  heart_rate: "",
  respiration_rate: "",
  weight: "",
  weight_unit: "Kg",
  note: "",
};

export function parseVitalsNumber(value: string): number | null {
  if (value.trim() === "") return null;
  const n = Number(value);
  return isNaN(n) ? null : n;
}

/** Kg↔g トグル結果（値と単位を原子的に更新するためのペア）。 */
export interface WeightUnitToggleResult {
  weight: string;
  weight_unit: BodyWeightUnit;
}

/**
 * 体重の単位トグル時に数値を物理量として換算する pure helper。
 * - 空欄/非数値: 単位ラベルのみ切替（保存すべき物理量がない）
 * - Kg→g: ×1000、g→Kg: ÷1000（丸めなし・厳密値）
 */
export function toggleWeightValueAndUnit(
  weight: string,
  currentUnit: BodyWeightUnit
): WeightUnitToggleResult {
  const nextUnit: BodyWeightUnit = currentUnit === "Kg" ? "g" : "Kg";
  const parsed = parseVitalsNumber(weight);
  if (parsed === null) {
    return { weight, weight_unit: nextUnit };
  }
  const converted = currentUnit === "Kg" ? parsed * 1000 : parsed / 1000;
  return { weight: String(converted), weight_unit: nextUnit };
}

export function formatRecordedAt(iso: string): string {
  const d = toJSTWallDate(iso);
  if (isNaN(d.getTime())) return iso;
  return `${formatJSTWallDate(d)} ${formatJSTWallTime(d)}`;
}

export function displayNum(value: number | null | undefined): string {
  return value != null ? String(value) : "-";
}
