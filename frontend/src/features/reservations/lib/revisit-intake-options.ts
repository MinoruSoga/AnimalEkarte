/**
 * EMR-475: 飼主事前問診（LINE 再診）フォーム v1 の code → 表示ラベル。
 * 正本: docs/work/db-contracts/revisit-intake.md §4 /
 * docs/spec/domains/reservations/pages/reservation-management.md#line-revisit-intake-20261017。
 * 未知 code は fail-closed で code をそのまま出さず「不明な回答」とする
 * （過去版の回答が将来追加 code を持つ場合に別の意味へ誤読しないため）。
 */
export const REVISIT_INTAKE_UNKNOWN_LABEL = "不明な回答";

export const REVISIT_INTAKE_ENERGY_APPETITE_LABELS: Readonly<Record<string, string>> = {
  usual: "いつも通り",
  slightly_reduced: "少し低下",
  markedly_reduced: "かなり低下",
};

export const REVISIT_INTAKE_IMPROVEMENT_LABELS: Readonly<Record<string, string>> = {
  improved: "良くなった",
  slightly_improved: "少し良くなった",
  unchanged: "変わらない",
  worse: "悪くなった",
  new_symptoms: "新しい症状が出た",
};

export const REVISIT_INTAKE_EXAMINATION_PREFERENCE_LABELS: Readonly<Record<string, string>> = {
  consultation_only: "診察のみ",
  medication_only: "お薬のみ希望",
  examination: "検査希望",
  decide_with_vet: "獣医師と相談して決めたい",
};

function labelOf(
  table: Readonly<Record<string, string>>,
  code: string | null | undefined,
  empty: string,
): string {
  if (code === null || code === undefined) return empty;
  return Object.hasOwn(table, code) ? table[code] : REVISIT_INTAKE_UNKNOWN_LABEL;
}

export function getEnergyAppetiteLabel(code: string | null | undefined, empty: string): string {
  return labelOf(REVISIT_INTAKE_ENERGY_APPETITE_LABELS, code, empty);
}

export function getImprovementLabel(code: string | null | undefined, empty: string): string {
  return labelOf(REVISIT_INTAKE_IMPROVEMENT_LABELS, code, empty);
}

export function getExaminationPreferenceLabel(
  code: string | null | undefined,
  empty: string,
): string {
  return labelOf(REVISIT_INTAKE_EXAMINATION_PREFERENCE_LABELS, code, empty);
}
