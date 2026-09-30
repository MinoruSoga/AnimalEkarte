import { C } from "@/lib/design-tokens";
import { FieldHelp } from "@/components/shared/FieldHelp";

import { QUALITATIVE_VALUES, type ReferenceRangeDraft } from "../lib/exam-type-fields-editor-model";

interface FieldInputProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
  /** 項目の説明文。指定するとラベル横に ⓘ ツールチップを表示する */
  description?: string;
}

export function FieldInput({ label, value, onChange, description }: FieldInputProps) {
  return (
    <div className={`block text-sm ${C.text65}`}>
      <div className="flex items-center gap-1">
        <span>{label}</span>
        {description ? <FieldHelp label={label} content={description} /> : null}
      </div>
      <input
        aria-label={label}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className={`mt-1 min-h-11 w-full rounded-xs border px-2 ${C.borderMedium} ${C.bgWhite} ${C.text}`}
      />
    </div>
  );
}

interface RangeBoundInputProps {
  label: string;
  type: "number" | "text";
  value: string;
  onChange: (value: string) => void;
}

function RangeBoundInput({ label, type, value, onChange }: RangeBoundInputProps) {
  return (
    <div className={`text-xs ${C.text65}`}>
      {label.endsWith("下限") ? "下限" : "上限"}
      <input
        type={type}
        step={type === "number" ? "any" : undefined}
        aria-label={label}
        list={type === "text" ? "exam-qualitative-values" : undefined}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className={`mt-1 min-h-11 w-full rounded-xs border px-2 ${C.borderMedium} ${C.bgWhite} ${C.text}`}
      />
    </div>
  );
}

interface ReferenceRangeInputsProps {
  speciesName: string;
  draft: ReferenceRangeDraft;
  onChange: (update: (draft: ReferenceRangeDraft) => ReferenceRangeDraft) => void;
}

export function ReferenceRangeInputs({ speciesName, draft, onChange }: ReferenceRangeInputsProps) {
  const type = draft.mode === "numeric" ? "number" : "text";
  const rangeKind = draft.mode === "numeric" ? "数値" : "定性";
  return (
    <div className="grid grid-cols-[120px_1fr_1fr] gap-2">
      <div className={`text-xs ${C.text65}`}>
        <span className="inline-flex items-center gap-1">
          種別
          <FieldHelp
            label={`${speciesName}の基準範囲種別`}
            content="基準範囲の入力形式です。「数値」は下限・上限を数値で、「定性」は区分値で指定します。"
          />
        </span>
        <select
          aria-label={`${speciesName}の基準範囲種別`}
          value={draft.mode}
          onChange={(event) =>
            onChange((previous) => ({
              ...previous,
              mode: event.target.value === "qualitative" ? "qualitative" : "numeric",
              min: "",
              max: "",
              qualitativeMin: undefined,
              qualitativeMax: undefined,
            }))
          }
          className={`mt-1 min-h-11 w-full rounded-xs border px-2 ${C.borderMedium} ${C.bgWhite} ${C.text}`}
        >
          <option value="numeric">数値</option>
          <option value="qualitative">定性</option>
        </select>
      </div>
      <RangeBoundInput
        label={`${speciesName}の${rangeKind}下限`}
        type={type}
        value={draft.min}
        onChange={(value) => onChange((previous) => ({ ...previous, min: value }))}
      />
      <RangeBoundInput
        label={`${speciesName}の${rangeKind}上限`}
        type={type}
        value={draft.max}
        onChange={(value) => onChange((previous) => ({ ...previous, max: value }))}
      />
      {draft.mode === "qualitative" ? (
        <p className={`col-span-3 text-xs ${C.text50}`}>
          選択可能: {QUALITATIVE_VALUES.join("、")}
        </p>
      ) : null}
    </div>
  );
}
