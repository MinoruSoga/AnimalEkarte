/**
 * Procedure anesthesia UI options (BUG-028).
 * Values must match BE oneof=none local sedation general — do not invent enums.
 * Labels: Assumption standard 和訳 (none→麻酔なし, local→局所麻酔, sedation→鎮静, general→全身麻酔).
 *
 * Note: intentionally does NOT import @/types/generated/models (TASK-444-S1 allowlist freeze).
 */
export const ANESTHESIA_OPTIONS = [
  { value: "none", label: "麻酔なし" },
  { value: "local", label: "局所麻酔" },
  { value: "sedation", label: "鎮静" },
  { value: "general", label: "全身麻酔" },
] as const;

export type AnesthesiaOptionValue = (typeof ANESTHESIA_OPTIONS)[number]["value"];

export const PRICE_ERROR_MESSAGE = "金額は0以上を入力してください";

export function isAnesthesiaOptionValue(value: string | undefined): value is AnesthesiaOptionValue {
  return value === "none" || value === "local" || value === "sedation" || value === "general";
}

export function initialAnesthesia(value: string | undefined): AnesthesiaOptionValue {
  return isAnesthesiaOptionValue(value) ? value : "none";
}
