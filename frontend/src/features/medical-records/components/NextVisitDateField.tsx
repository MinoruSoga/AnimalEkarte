import { useId } from "react";
import { C, STYLE, PALETTE } from "@/lib/design-tokens";
import { formatJSTWallDate, todayJSTISO, toJSTWallDate } from "@/lib/jst-date";
import { validateNextVisitIntentDate } from "../lib/next-visit";

// ─────────────────────────────────────────────────
// Date helpers (no date-fns dependency)
// ─────────────────────────────────────────────────

function addMonths(date: Date, months: number): string {
  const d = new Date(date);
  d.setMonth(d.getMonth() + months);
  return formatJSTWallDate(d);
}

function addYears(date: Date, years: number): string {
  const d = new Date(date);
  d.setFullYear(d.getFullYear() + years);
  return formatJSTWallDate(d);
}

function today(): string {
  return todayJSTISO();
}

// ─────────────────────────────────────────────────
// Quick-select button definitions
// ─────────────────────────────────────────────────

interface QuickOption {
  label: string;
  getValue: () => string;
}

const QUICK_OPTIONS: QuickOption[] = [
  { label: "1ヶ月後", getValue: () => addMonths(toJSTWallDate(new Date()), 1) },
  { label: "3ヶ月後", getValue: () => addMonths(toJSTWallDate(new Date()), 3) },
  { label: "6ヶ月後", getValue: () => addMonths(toJSTWallDate(new Date()), 6) },
  { label: "1年後", getValue: () => addYears(toJSTWallDate(new Date()), 1) },
];

// ─────────────────────────────────────────────────
// Props
// ─────────────────────────────────────────────────

interface NextVisitDateFieldProps {
  value: string;
  onChange: (date: string) => void;
  hasLineIntegration?: boolean;
  disabled?: boolean;
  setDisabled?: boolean;
  clearDisabled?: boolean;
}

// ─────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────

export function NextVisitDateField({
  value,
  onChange,
  hasLineIntegration = false,
  disabled = false,
  setDisabled = disabled,
  clearDisabled = disabled,
}: NextVisitDateFieldProps) {
  const inputId = useId();
  const errorMessage = validateNextVisitIntentDate(value);
  const isValid = errorMessage === null;

  const handleChange = (newValue: string) => {
    onChange(newValue);
  };

  const handleQuickSelect = (getValue: () => string) => {
    onChange(getValue());
  };

  const isInputInvalid = !isValid && value !== "";
  const showError = errorMessage !== null && value !== "";

  return (
    <div className="flex flex-col gap-2">
      {/* Label row */}
      <div className="flex items-center gap-2">
        <label htmlFor={inputId} className={STYLE.formLabel}>
          次回来院推奨日
        </label>
        {hasLineIntegration ? (
          <span
            className="inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full text-white"
            style={{ backgroundColor: PALETTE.lineGreen }}
          >
            LINEに自動通知
          </span>
        ) : null}
      </div>

      {/* Date input */}
      <input
        id={inputId}
        type="date"
        value={value}
        min={today()}
        onChange={(e) => handleChange(e.target.value)}
        disabled={setDisabled}
        aria-invalid={isInputInvalid}
        aria-describedby={showError ? `${inputId}-error` : undefined}
        className={`${STYLE.formInput} rounded-xs border px-3 w-full max-w-[220px] focus-visible:ring-offset-1 ${
          isInputInvalid ? STYLE.formInputError : ""
        }`}
      />

      {/* Quick select buttons */}
      <div className="flex flex-wrap gap-1.5">
        {QUICK_OPTIONS.map((opt) => (
          <button
            key={opt.label}
            type="button"
            disabled={setDisabled}
            onClick={() => handleQuickSelect(opt.getValue)}
            className={`inline-flex items-center min-h-11 text-sm px-3 py-1 rounded-xs border ${C.borderMedium} ${C.text60} ${C.hoverBgLight} ${C.hoverText} transition-colors disabled:opacity-40 disabled:cursor-not-allowed`}
          >
            {opt.label}
          </button>
        ))}
        {value !== "" ? (
          <button
            type="button"
            disabled={clearDisabled}
            onClick={() => handleChange("")}
            className={`inline-flex items-center min-h-11 text-sm px-3 py-1 rounded-xs ${C.text60} ${C.hoverText} ${C.hoverBgLight} transition-colors disabled:opacity-40 disabled:cursor-not-allowed`}
          >
            クリア
          </button>
        ) : null}
      </div>

      {/* Validation error */}
      {showError ? (
        <p id={`${inputId}-error`} className={`text-sm ${C.danger}`} role="alert">
          {errorMessage}
        </p>
      ) : null}
    </div>
  );
}
