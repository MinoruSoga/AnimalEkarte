import { C } from "@/lib/design-tokens";

import { QUALITATIVE_VALUES, type ReferenceRangeDraft } from "./exam-type-fields-editor-model";

interface FieldInputProps {
  label: string;
  value: string;
  onChange: (value: string) => void;
}

export function FieldInput({ label, value, onChange }: FieldInputProps) {
  return (
    <label className={`block text-sm ${C.text65}`}>
      {label}
      <input
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className={`mt-1 min-h-11 w-full rounded-xs border px-2 ${C.borderMedium} ${C.bgWhite} ${C.text}`}
      />
    </label>
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
    <label className={`text-xs ${C.text65}`}>
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
    </label>
  );
}

interface ReferenceRangeInputsProps {
  speciesName: string;
  draft: ReferenceRangeDraft;
  onChange: (update: (draft: ReferenceRangeDraft) => ReferenceRangeDraft) => void;
}

export function ReferenceRangeInputs({
  speciesName,
  draft,
  onChange,
}: ReferenceRangeInputsProps) {
  const type = draft.mode === "numeric" ? "number" : "text";
  const rangeKind = draft.mode === "numeric" ? "数値" : "定性";
  return (
    <div className="grid grid-cols-[120px_1fr_1fr] gap-2">
      <label className={`text-xs ${C.text65}`}>
        種別
        <select
          aria-label={`${speciesName}の基準範囲種別`}
          value={draft.mode}
          onChange={(event) => onChange((previous) => ({
            ...previous,
            mode: event.target.value === "qualitative" ? "qualitative" : "numeric",
            min: "",
            max: "",
            qualitativeMin: undefined,
            qualitativeMax: undefined,
          }))}
          className={`mt-1 min-h-11 w-full rounded-xs border px-2 ${C.borderMedium} ${C.bgWhite} ${C.text}`}
        >
          <option value="numeric">数値</option>
          <option value="qualitative">定性</option>
        </select>
      </label>
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
