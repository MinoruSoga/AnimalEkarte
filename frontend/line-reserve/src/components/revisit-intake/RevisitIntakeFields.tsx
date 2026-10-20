// EMR-474: 再診問診7項目（8保存値）の入力欄。値は呼び出し側の draft state に保持する（制御入力）。
// 自由文: 未入力は null、入力して消した場合は空文字のまま（区別して送る）。
// 単一選択: 任意。「選択を解除」で null に戻せる（最初の選択肢へ補完しない）。
import type {
  RevisitIntakeEnergyAppetite,
  RevisitIntakeExaminationPreference,
  RevisitIntakeFullAnswers,
  RevisitIntakeImprovement,
} from "../../types/models";
import {
  ENERGY_APPETITE_OPTIONS,
  EXAMINATION_PREFERENCE_OPTIONS,
  IMPROVEMENT_OPTIONS,
  REVISIT_INTAKE_FIELD_LABELS,
  revisitIntakeFieldId,
  type RevisitIntakeFieldName,
  type RevisitIntakeOption,
} from "../../lib/revisit-intake-options";
import type { RevisitIntakeFieldErrors } from "../../lib/revisit-intake-submit";

type TextFieldName = "visit_purpose" | "onset" | "current_symptoms" | "main_concern";

interface FieldsProps {
  values: RevisitIntakeFullAnswers;
  errors: RevisitIntakeFieldErrors;
  disabled: boolean;
  onChange: <K extends RevisitIntakeFieldName>(
    field: K,
    value: RevisitIntakeFullAnswers[K],
  ) => void;
}

const inputClass =
  "w-full rounded-lg border border-noah-border-input bg-white px-3 py-2.5 text-base text-noah-text focus:border-noah-teal focus:outline-none disabled:bg-noah-surface-muted";

function FieldError({ field, message }: { field: RevisitIntakeFieldName; message?: string }) {
  if (message === undefined) return null;
  return (
    <p id={`${revisitIntakeFieldId(field)}-error`} className="mt-1 text-sm text-noah-danger-text">
      {message}
    </p>
  );
}

function TextField({
  field,
  multiline,
  required,
  values,
  errors,
  disabled,
  onChange,
}: FieldsProps & { field: TextFieldName; multiline: boolean; required?: boolean }) {
  const id = revisitIntakeFieldId(field);
  const value = values[field] ?? "";
  const error = errors[field];
  const common = {
    id,
    value,
    disabled,
    "aria-invalid": error !== undefined ? true : undefined,
    "aria-describedby": error !== undefined ? `${id}-error` : undefined,
    "aria-required": required ? true : undefined,
    className: inputClass,
  };
  return (
    <div>
      <label htmlFor={id} className="mb-1 block text-sm font-medium text-noah-text">
        {REVISIT_INTAKE_FIELD_LABELS[field]}
        {required ? (
          <span className="ml-1 text-xs text-noah-danger-text">（送信時必須）</span>
        ) : (
          <span className="ml-1 text-xs text-noah-text-sub">（任意）</span>
        )}
      </label>
      {multiline ? (
        <textarea {...common} rows={3} onChange={(e) => onChange(field, e.target.value)} />
      ) : (
        <input
          {...common}
          type="text"
          onChange={(e) => onChange(field, e.target.value)}
          onKeyDown={(e) => {
            // 単一行欄の Enter で問診を暗黙送信しない（IME 変換確定の Enter は対象外）。
            if (e.key === "Enter" && !e.nativeEvent.isComposing) e.preventDefault();
          }}
        />
      )}
      <FieldError field={field} message={error} />
    </div>
  );
}

function ChoiceField<T extends string>({
  field,
  options,
  value,
  errors,
  disabled,
  onSelect,
}: {
  field: RevisitIntakeFieldName;
  options: ReadonlyArray<RevisitIntakeOption<T>>;
  value: T | null;
  errors: RevisitIntakeFieldErrors;
  disabled: boolean;
  onSelect: (value: T | null) => void;
}) {
  const id = revisitIntakeFieldId(field);
  const error = errors[field];
  return (
    <fieldset aria-describedby={error !== undefined ? `${id}-error` : undefined}>
      <legend className="mb-1 text-sm font-medium text-noah-text">
        {REVISIT_INTAKE_FIELD_LABELS[field]}
        <span className="ml-1 text-xs text-noah-text-sub">（任意）</span>
      </legend>
      <div className="space-y-1">
        {options.map((option, index) => {
          const optionId = index === 0 ? id : `${id}-${option.value}`;
          return (
            <label
              key={option.value}
              htmlFor={optionId}
              className="flex min-h-11 items-center gap-3 rounded-lg border border-noah-border bg-white px-3 text-sm text-noah-text"
            >
              <input
                id={optionId}
                type="radio"
                name={id}
                value={option.value}
                checked={value === option.value}
                disabled={disabled}
                onChange={() => onSelect(option.value)}
                className="h-5 w-5 accent-noah-teal"
              />
              {option.label}
            </label>
          );
        })}
      </div>
      {value !== null && !disabled ? (
        <button
          type="button"
          onClick={() => onSelect(null)}
          className="mt-1 min-h-11 px-2 text-sm text-noah-teal-dark underline"
        >
          {`${REVISIT_INTAKE_FIELD_LABELS[field]}の選択を解除`}
        </button>
      ) : null}
      <FieldError field={field} message={error} />
    </fieldset>
  );
}

export function RevisitIntakeFields(props: FieldsProps) {
  const { values, errors, disabled, onChange } = props;
  return (
    <div className="space-y-5">
      <TextField {...props} field="visit_purpose" multiline required />
      <TextField {...props} field="onset" multiline={false} />
      <ChoiceField<RevisitIntakeEnergyAppetite>
        field="energy"
        options={ENERGY_APPETITE_OPTIONS}
        value={values.energy}
        errors={errors}
        disabled={disabled}
        onSelect={(v) => onChange("energy", v)}
      />
      <ChoiceField<RevisitIntakeEnergyAppetite>
        field="appetite"
        options={ENERGY_APPETITE_OPTIONS}
        value={values.appetite}
        errors={errors}
        disabled={disabled}
        onSelect={(v) => onChange("appetite", v)}
      />
      <TextField {...props} field="current_symptoms" multiline />
      <ChoiceField<RevisitIntakeImprovement>
        field="improvement"
        options={IMPROVEMENT_OPTIONS}
        value={values.improvement}
        errors={errors}
        disabled={disabled}
        onSelect={(v) => onChange("improvement", v)}
      />
      <ChoiceField<RevisitIntakeExaminationPreference>
        field="examination_preference"
        options={EXAMINATION_PREFERENCE_OPTIONS}
        value={values.examination_preference}
        errors={errors}
        disabled={disabled}
        onSelect={(v) => onChange("examination_preference", v)}
      />
      <TextField {...props} field="main_concern" multiline />
    </div>
  );
}
