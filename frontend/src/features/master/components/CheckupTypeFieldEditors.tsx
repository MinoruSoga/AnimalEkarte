import Plus from "lucide-react/dist/esm/icons/plus";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";

import { C, ICON } from "@/lib/design-tokens";
import type { CheckupFieldType } from "@/types/checkup";

import {
  CHECKUP_FIELD_TYPE_OPTIONS,
  type CheckupFieldOptionDraft,
} from "./checkup-type-fields-editor-model";

// EMR-225: 健診フィールド定義フォームの入力部品。
// テキスト入力は ExamTypeFieldEditors.FieldInput を再利用（見た目・高さ規約の単一正本）。

interface CheckupFieldTypeSelectProps {
  value: CheckupFieldType;
  onChange: (value: CheckupFieldType) => void;
}

export function CheckupFieldTypeSelect({ value, onChange }: CheckupFieldTypeSelectProps) {
  return (
    <label className={`block text-sm ${C.text65}`}>
      種別
      <select
        aria-label="健診項目の種別"
        value={value}
        onChange={(event) => onChange(event.target.value as CheckupFieldType)}
        className={`mt-1 min-h-11 w-full rounded-xs border px-2 ${C.borderMedium} ${C.bgWhite} ${C.text}`}
      >
        {CHECKUP_FIELD_TYPE_OPTIONS.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

interface NumberBoundInputProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
}

export function NumberBoundInput({ label, value, onChange }: NumberBoundInputProps) {
  return (
    <label className={`block text-sm ${C.text65}`}>
      {label}
      <input
        type="number"
        step="any"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className={`mt-1 min-h-11 w-full rounded-xs border px-2 ${C.borderMedium} ${C.bgWhite} ${C.text}`}
      />
    </label>
  );
}

interface CheckupFieldOptionsEditorProps {
  options: CheckupFieldOptionDraft[];
  onAdd: () => void;
  onChange: (index: number, patch: Partial<CheckupFieldOptionDraft>) => void;
  onRemove: (index: number) => void;
}

// single_select / multi_select / checklist 用の選択肢エディタ。
// { value, label } は manifest import の永続化形状と同一。
export function CheckupFieldOptionsEditor({
  options,
  onAdd,
  onChange,
  onRemove,
}: CheckupFieldOptionsEditorProps) {
  return (
    <fieldset className={`space-y-2 rounded-xs border p-2 ${C.borderLight}`}>
      <legend className={`px-1 text-xs ${C.text50}`}>選択肢</legend>
      {options.map((option, index) => (
        <div key={index} className="flex items-center gap-2">
          <input
            value={option.value}
            onChange={(event) => onChange(index, { value: event.target.value })}
            placeholder="値"
            aria-label={`選択肢${index + 1}の値`}
            className={`min-h-11 w-full rounded-xs border px-2 text-sm ${C.borderMedium} ${C.bgWhite} ${C.text}`}
          />
          <input
            value={option.label}
            onChange={(event) => onChange(index, { label: event.target.value })}
            placeholder="表示名"
            aria-label={`選択肢${index + 1}の表示名`}
            className={`min-h-11 w-full rounded-xs border px-2 text-sm ${C.borderMedium} ${C.bgWhite} ${C.text}`}
          />
          <button
            type="button"
            onClick={() => onRemove(index)}
            aria-label={`選択肢${index + 1}を削除`}
            className={`inline-flex min-h-11 min-w-11 shrink-0 items-center justify-center rounded-xxs ${C.text50} ${C.hoverTextDanger} ${C.hoverBgLight}`}
          >
            <Trash2 className={ICON.smXs} aria-hidden="true" />
          </button>
        </div>
      ))}
      <button
        type="button"
        onClick={onAdd}
        aria-label="選択肢を追加"
        className={`inline-flex min-h-11 items-center gap-1 rounded-xxs px-2 text-sm ${C.textBrand} ${C.hoverBgLight}`}
      >
        <Plus className={ICON.smXs} aria-hidden="true" />
        選択肢を追加
      </button>
    </fieldset>
  );
}
