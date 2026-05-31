import { C } from "@/lib/design-tokens";
import type { BodyWeightUnit } from "../../types";

export const EDIT_INPUT_CLASS = `h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 ${C.bgWhite} ${C.text} outline-none ${C.focusBorderAccent} w-full`;
export const ADD_INPUT_CLASS = `h-8 text-sm border ${C.borderMedium} rounded-[3px] px-2 ${C.bgWhite} ${C.text} outline-none ${C.focusBorderAccent}`;

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

export function formatRecordedAt(iso: string): string {
  const d = new Date(iso);
  if (isNaN(d.getTime())) return iso;
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

export function displayNum(value: number | null | undefined): string {
  return value != null ? String(value) : "-";
}
