// EMR-474: 再診事前問診フォーム v1 の code → 表示ラベル。
// 正本: docs/spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-line-revisit-intake
// と docs/work/db-contracts/revisit-intake.md §4。v1 のラベルを後から別意味へ差し替えない。
// 値集合は api.yaml の RevisitIntakeEnergyAppetite / Improvement / ExaminationPreference と同一。
import type {
  RevisitIntakeEditBlockReason,
  RevisitIntakeEnergyAppetite,
  RevisitIntakeExaminationPreference,
  RevisitIntakeFullAnswers,
  RevisitIntakeImprovement,
} from "../types/models";

/** この画面が送信・表示に対応するフォーム定義版。 */
export const REVISIT_INTAKE_SUPPORTED_FORM_VERSION = 1;

export interface RevisitIntakeOption<T extends string> {
  value: T;
  label: string;
}

export const ENERGY_APPETITE_OPTIONS: ReadonlyArray<
  RevisitIntakeOption<RevisitIntakeEnergyAppetite>
> = [
  { value: "usual", label: "いつも通り" },
  { value: "slightly_reduced", label: "少し低下" },
  { value: "markedly_reduced", label: "かなり低下" },
];

export const IMPROVEMENT_OPTIONS: ReadonlyArray<RevisitIntakeOption<RevisitIntakeImprovement>> = [
  { value: "improved", label: "良くなった" },
  { value: "slightly_improved", label: "少し良くなった" },
  { value: "unchanged", label: "変わらない" },
  { value: "worse", label: "悪くなった" },
  { value: "new_symptoms", label: "新しい症状が出た" },
];

export const EXAMINATION_PREFERENCE_OPTIONS: ReadonlyArray<
  RevisitIntakeOption<RevisitIntakeExaminationPreference>
> = [
  { value: "consultation_only", label: "診察のみ" },
  { value: "medication_only", label: "お薬のみ希望" },
  { value: "examination", label: "検査希望" },
  { value: "decide_with_vet", label: "獣医師と相談して決めたい" },
];

export type RevisitIntakeFieldName = keyof RevisitIntakeFullAnswers;

/** 表示・フォーカス順（7表示項目 / 8保存値）。 */
export const REVISIT_INTAKE_FIELD_ORDER: readonly RevisitIntakeFieldName[] = [
  "visit_purpose",
  "onset",
  "energy",
  "appetite",
  "current_symptoms",
  "improvement",
  "examination_preference",
  "main_concern",
];

export const REVISIT_INTAKE_FIELD_LABELS: Record<RevisitIntakeFieldName, string> = {
  visit_purpose: "来院目的",
  onset: "いつから",
  energy: "元気",
  appetite: "食欲",
  current_symptoms: "現在の症状",
  improvement: "前回より改善したか",
  examination_preference: "検査希望の有無",
  main_concern: "一番相談したいこと",
};

function labelOf<T extends string>(
  options: ReadonlyArray<RevisitIntakeOption<T>>,
  value: T | null,
): string | null {
  if (value === null) return null;
  return options.find((o) => o.value === value)?.label ?? null;
}

/**
 * 回答1項目の表示文字列。null と空文字はどちらも「未回答」として表示する（症状なし等へ補完しない）。
 */
export function formatRevisitIntakeAnswer(
  answers: RevisitIntakeFullAnswers,
  field: RevisitIntakeFieldName,
): string | null {
  switch (field) {
    case "energy":
      return labelOf(ENERGY_APPETITE_OPTIONS, answers.energy);
    case "appetite":
      return labelOf(ENERGY_APPETITE_OPTIONS, answers.appetite);
    case "improvement":
      return labelOf(IMPROVEMENT_OPTIONS, answers.improvement);
    case "examination_preference":
      return labelOf(EXAMINATION_PREFERENCE_OPTIONS, answers.examination_preference);
    default: {
      const text = answers[field];
      return text === null || text === "" ? null : text;
    }
  }
}

/** editable=false の理由別案内（edit_block_reason / INTAKE_CLOSED details.reason）。 */
export const EDIT_BLOCK_MESSAGES: Record<RevisitIntakeEditBlockReason, string> = {
  reception_started: "受付後の変更はスタッフへお伝えください。",
  cancelled: "この予約はキャンセルされているため、問診の回答・訂正はできません。",
  deleted: "この予約は取り消されているため、問診の回答・訂正はできません。",
  target_unresolved:
    "この予約の対象ペットを確認できないため、問診に回答できません。医院へお問い合わせください。",
  authorization_denied:
    "この予約の問診を回答・訂正する権限を確認できませんでした。医院へお問い合わせください。",
  unsupported_lifecycle:
    "現在の予約状態では問診の回答・訂正ができません。変更があれば医院へお伝えください。",
};

/** reason が null の editable=false（契約上は想定外）に使う安全側の文言。 */
export const EDIT_BLOCK_FALLBACK_MESSAGE =
  "現在は問診の回答・訂正ができません。変更があれば医院へお伝えください。";

/** 入力欄の DOM id（400 時の先頭エラー欄フォーカスに使う）。 */
export function revisitIntakeFieldId(field: RevisitIntakeFieldName): string {
  return `revisit-intake-${field.replace(/_/g, "-")}`;
}

/** server UTC 日時を JST の「YYYY年M月D日 HH:MM」で表示する。 */
export function formatRevisitIntakeDateTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  const parts = new Intl.DateTimeFormat("ja-JP", {
    timeZone: "Asia/Tokyo",
    year: "numeric",
    month: "numeric",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).formatToParts(date);
  const get = (type: Intl.DateTimeFormatPartTypes) =>
    parts.find((p) => p.type === type)?.value ?? "";
  return `${get("year")}年${get("month")}月${get("day")}日 ${get("hour")}:${get("minute")}`;
}
