import { useId } from "react";
import { C, STYLE, PALETTE } from "@/lib/design-tokens";
import { formatJSTWallDate, isPastJSTDate, todayJSTISO, toJSTWallDate } from "@/lib/jst-date";

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
// Validation
// ─────────────────────────────────────────────────

function addYearsIso(iso: string, years: number): string {
  const [yearPart, monthPart, dayPart] = iso.split("-");
  const year = Number(yearPart) + years;
  const month = Number(monthPart);
  const day = Number(dayPart);
  const isLeap = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
  const clampedDay = month === 2 && day === 29 && !isLeap ? 28 : day;
  return `${String(year).padStart(4, "0")}-${monthPart}-${String(clampedDay).padStart(2, "0")}`;
}

function validate(value: string): string | null {
  if (value === "") return null;
  if (isPastJSTDate(value)) return "今日より前の日付は設定できません";
  if (value > addYearsIso(todayJSTISO(), 2)) return "今日から2年以内の日付を設定してください";
  return null;
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
  onValidationChange: (isValid: boolean) => void;
  hasLineIntegration?: boolean;
  disabled?: boolean;
}

// ─────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────

export function NextVisitDateField({
  value,
  onChange,
  onValidationChange,
  hasLineIntegration = false,
  disabled = false,
}: NextVisitDateFieldProps) {
  const inputId = useId();
  const errorMessage = validate(value);
  const isValid = errorMessage === null;

  const handleChange = (newValue: string) => {
    const error = validate(newValue);
    onValidationChange(error === null);
    onChange(newValue);
  };

  const handleQuickSelect = (getValue: () => string) => {
    const newValue = getValue();
    const error = validate(newValue);
    onValidationChange(error === null);
    onChange(newValue);
  };

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
        disabled={disabled}
        className={`${STYLE.formInput} rounded-xs border px-3 w-full max-w-[220px] outline-none focus:ring-2 ${C.focusRingAccent30} ${
          !isValid && value !== "" ? STYLE.formInputError : ""
        }`}
      />

      {/* Quick select buttons */}
      <div className="flex flex-wrap gap-1.5">
        {QUICK_OPTIONS.map((opt) => (
          <button
            key={opt.label}
            type="button"
            disabled={disabled}
            onClick={() => handleQuickSelect(opt.getValue)}
            className={`text-sm px-3 py-1 rounded-xs border ${C.borderMedium} ${C.text60} ${C.hoverBgLight} ${C.hoverText} transition-colors disabled:opacity-40 disabled:cursor-not-allowed`}
          >
            {opt.label}
          </button>
        ))}
        {value !== "" ? (
          <button
            type="button"
            disabled={disabled}
            onClick={() => handleChange("")}
            className={`text-sm px-3 py-1 rounded-xs ${C.text40} ${C.hoverText} ${C.hoverBgLight} transition-colors disabled:opacity-40 disabled:cursor-not-allowed`}
          >
            クリア
          </button>
        ) : null}
      </div>

      {/* Validation error */}
      {errorMessage !== null && value !== "" ? (
        <p className={`text-sm ${C.danger}`} role="alert">
          {errorMessage}
        </p>
      ) : null}
    </div>
  );
}
