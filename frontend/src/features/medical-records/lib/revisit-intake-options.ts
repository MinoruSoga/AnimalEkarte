// EMR-477: 再診問診 form_definition_version=1 の code → 回答当時ラベル。
// 正本: docs/work/db-contracts/revisit-intake.md §4 / backend/docs/api.yaml
// （RevisitIntakeEnergyAppetite / Improvement / ExaminationPreference）。
// 未知 code は最初の選択肢へ補完せず、「未対応の選択肢」として明示する（fail-closed）。

export type RevisitIntakeEnergyAppetite = "usual" | "slightly_reduced" | "markedly_reduced";
export type RevisitIntakeImprovement =
  | "improved"
  | "slightly_improved"
  | "unchanged"
  | "worse"
  | "new_symptoms";
export type RevisitIntakeExaminationPreference =
  | "consultation_only"
  | "medication_only"
  | "examination"
  | "decide_with_vet";

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

/** 未知 code（将来の form 版や不正値）の表示。最初の選択肢へ補完しない。 */
export const REVISIT_INTAKE_UNKNOWN_CODE_LABEL = "未対応の選択肢";

/** 未回答（null・未送信）の表示。任意項目の NULL と空文字はどちらも未回答。 */
export const REVISIT_INTAKE_UNANSWERED_LABEL = "未回答";

/**
 * 選択式回答のラベル。null/undefined は null（未回答）、未知 code は未対応表示。
 */
export function resolveRevisitIntakeChoiceLabel(
  labels: Readonly<Record<string, string>>,
  code: string | null | undefined,
): string | null {
  if (code === null || code === undefined) return null;
  return Object.hasOwn(labels, code) ? labels[code] : REVISIT_INTAKE_UNKNOWN_CODE_LABEL;
}
