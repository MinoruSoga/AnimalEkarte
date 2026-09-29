import type { CheckupFieldType } from "@/types/checkup";
import type { CheckupTypeFieldRow } from "@/hooks/use-checkup-fields";

import type {
  CheckupFieldOption,
  CreateCheckupTypeFieldRequest,
  UpdateCheckupTypeFieldRequest,
} from "../api/checkup-type-fields";

// EMR-225: 定期健診フィールド定義エディタの純粋ロジック（draft ↔ request 変換・検証）。
// BE の検証と対になる: field_type 妥当性・選択式の非空 options・min<=max・options 重複。

export const CHECKUP_FIELD_TYPE_OPTIONS: ReadonlyArray<{
  value: CheckupFieldType;
  label: string;
}> = [
  { value: "number", label: "数値" },
  { value: "single_select", label: "単一選択" },
  { value: "multi_select", label: "複数選択" },
  { value: "boolean", label: "はい/いいえ" },
  { value: "checklist", label: "チェックリスト" },
  { value: "text", label: "テキスト" },
];

export const CHECKUP_FIELD_TYPE_LABELS: Readonly<Record<CheckupFieldType, string>> = {
  number: "数値",
  single_select: "単一選択",
  multi_select: "複数選択",
  boolean: "はい/いいえ",
  checklist: "チェックリスト",
  text: "テキスト",
};

// options が必須の選択式フィールド型（BE isCheckupSelectFieldType と同一集合）。
export function isCheckupSelectFieldType(fieldType: CheckupFieldType): boolean {
  return fieldType === "single_select" || fieldType === "multi_select" || fieldType === "checklist";
}

export type CheckupFieldOptionDraft = CheckupFieldOption;

export interface CheckupFieldDraft {
  name: string;
  fieldType: CheckupFieldType;
  unit: string;
  /** number 型のみ意味を持つ。空文字 = 基準値なし（min_value/max_value 省略 or クリア）。 */
  minValue: string;
  maxValue: string;
  options: CheckupFieldOptionDraft[];
}

export const emptyCheckupFieldDraft = (): CheckupFieldDraft => ({
  name: "",
  fieldType: "number",
  unit: "",
  minValue: "",
  maxValue: "",
  options: [],
});

export const checkupFieldToDraft = (field: CheckupTypeFieldRow): CheckupFieldDraft => ({
  name: field.name,
  fieldType: field.fieldType,
  unit: field.unit,
  minValue: field.minValue !== undefined ? String(field.minValue) : "",
  maxValue: field.maxValue !== undefined ? String(field.maxValue) : "",
  options: field.options.map((option) => ({ value: option.value, label: option.label })),
});

function parseBound(raw: string): number | undefined | null {
  const trimmed = raw.trim();
  if (trimmed === "") return undefined;
  const value = Number(trimmed);
  return Number.isFinite(value) ? value : null;
}

function normalizedOptions(options: CheckupFieldOptionDraft[]): CheckupFieldOption[] {
  return options.map((option) => ({ value: option.value.trim(), label: option.label.trim() }));
}

export function validateCheckupFieldDraft(draft: CheckupFieldDraft): string | null {
  if (!draft.name.trim()) {
    return "項目名を入力してください";
  }
  if (draft.fieldType === "number") {
    const min = parseBound(draft.minValue);
    const max = parseBound(draft.maxValue);
    if (min === null || max === null) {
      return "基準値には有限の数値を入力してください";
    }
    if (min !== undefined && max !== undefined && min > max) {
      return "基準値の下限は上限以下にしてください";
    }
  }
  if (isCheckupSelectFieldType(draft.fieldType)) {
    const options = normalizedOptions(draft.options);
    if (options.length === 0) {
      return "選択式の項目には選択肢を1件以上登録してください";
    }
    const seen = new Set<string>();
    for (const option of options) {
      if (option.value === "" || option.label === "") {
        return "選択肢の値と表示名を入力してください";
      }
      if (seen.has(option.value)) {
        return "選択肢の値は一意にしてください";
      }
      seen.add(option.value);
    }
  }
  return null;
}

export function buildCheckupFieldCreateRequest(
  draft: CheckupFieldDraft,
  sortOrder?: number,
): CreateCheckupTypeFieldRequest {
  const req: CreateCheckupTypeFieldRequest = {
    name: draft.name.trim(),
    field_type: draft.fieldType,
    unit: draft.unit,
    ...(sortOrder !== undefined ? { sort_order: sortOrder } : {}),
  };
  if (draft.fieldType === "number") {
    const min = parseBound(draft.minValue);
    const max = parseBound(draft.maxValue);
    // != null: undefined（空欄）と null（非数値 — validate が先に弾く）の双方を除外し、
    // 0 という有効な基準値を落とさない。
    if (min != null) req.min_value = min;
    if (max != null) req.max_value = max;
  }
  if (isCheckupSelectFieldType(draft.fieldType)) {
    req.options = normalizedOptions(draft.options);
  }
  return req;
}

// PATCH は「フォームが表す定義全体」を送る（name/field_type/unit/options は常に上書き、
// min/max は空欄なら clear_* で NULL に戻す）。field_type 変更もそのまま送る
// （checkup_field_results はスナップショット保持のため履歴は保全される）。
export function buildCheckupFieldUpdateRequest(
  draft: CheckupFieldDraft,
  original: CheckupTypeFieldRow,
): UpdateCheckupTypeFieldRequest {
  const req: UpdateCheckupTypeFieldRequest = {
    name: draft.name.trim(),
    field_type: draft.fieldType,
    unit: draft.unit,
  };
  // parseBound の null（非数値文字列入力）は validateCheckupFieldDraft が先に弾くため
  // ここでは到達不能。型上は undefined と同じ「クリア」経路に畳み込んでおく。
  const min = parseBound(draft.minValue);
  const max = parseBound(draft.maxValue);
  if (min == null) {
    req.clear_min_value = true;
  } else {
    req.min_value = min;
  }
  if (max == null) {
    req.clear_max_value = true;
  } else {
    req.max_value = max;
  }
  if (isCheckupSelectFieldType(draft.fieldType)) {
    req.options = normalizedOptions(draft.options);
  } else if (original.options.length > 0) {
    // 選択式→非選択式への型変更時は残存 options を明示クリアする。
    req.options = [];
  }
  return req;
}
