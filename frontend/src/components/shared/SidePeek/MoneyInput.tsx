// React/Framework
import type { ComponentProps } from "react";

// Internal
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { C } from "@/lib/design-tokens";

interface MoneyInputBaseProps {
  label?: string;
  min?: number;
  max?: number;
  step?: number | "any";
  placeholder?: string;
  /** Field-level validation message (e.g. negative price) */
  error?: string;
  /** 項目の説明文（ⓘツールチップ） */
  description?: string;
  /**
   * "property"（既定） — サイドピークの PropertyRow + 先頭 ¥ + 固定 w-32/text-sm。
   * "inline" — 呼び元の Label とレイアウトに載る独立入力（NumberInput ベース、suffix/align 対応）。
   */
  variant?: "property" | "inline";
  id?: string;
  name?: string;
  disabled?: boolean;
  required?: boolean;
  readOnly?: boolean;
  /** inline のみ: 入力へ追加するクラス。property variant では適用されない。 */
  className?: string;
  /** inline のみ: 入力右端の単位テキスト（例: "円"） */
  suffix?: string;
  /** inline のみ: 入力のテキスト揃え */
  align?: "left" | "right";
  "aria-label"?: ComponentProps<"input">["aria-label"];
  "aria-required"?: ComponentProps<"input">["aria-required"];
  "aria-invalid"?: ComponentProps<"input">["aria-invalid"];
  "aria-describedby"?: ComponentProps<"input">["aria-describedby"];
  onInvalid?: ComponentProps<"input">["onInvalid"];
}

type MoneyInputProps = MoneyInputBaseProps &
  (
    | {
        value: number;
        onChange: (value: number) => void;
        stringValue?: never;
        onStringChange?: never;
      }
    | {
        value?: never;
        onChange?: never;
        stringValue: string;
        onStringChange: (value: string) => void;
      }
  );

export function MoneyInput({
  label = "単価(税込)",
  value,
  onChange,
  stringValue,
  onStringChange,
  min,
  max,
  step,
  placeholder = "0",
  error,
  description,
  variant = "property",
  id,
  name,
  disabled,
  required,
  readOnly,
  className,
  suffix,
  align,
  "aria-label": ariaLabel,
  "aria-required": ariaRequired,
  "aria-invalid": ariaInvalid,
  "aria-describedby": ariaDescribedBy,
  onInvalid,
}: MoneyInputProps) {
  const rawValue =
    stringValue ?? (value === undefined || Number.isNaN(value) || value === 0 ? "" : value);
  const handleChange = (v: string) => {
    if (onStringChange) {
      onStringChange(v);
    } else if (onChange) {
      const next = v === "" ? Number.NaN : Number(v);
      onChange(Number.isNaN(next) ? 0 : next);
    }
  };

  if (variant === "inline") {
    // 呼び元の Label と紐付く id をそのまま使う。既定 label は aria-label へ降格しない
    // （外部 Label があるとき label prop を渡さないのが前提）。
    return (
      <div className="flex flex-col gap-0.5">
        <NumberInput
          id={id}
          name={name}
          value={rawValue}
          onChange={handleChange}
          placeholder={placeholder}
          step={step}
          min={min}
          max={max}
          suffix={suffix}
          align={align}
          disabled={disabled}
          required={required}
          readOnly={readOnly}
          aria-label={ariaLabel}
          aria-required={ariaRequired}
          aria-invalid={ariaInvalid ?? (error ? true : undefined)}
          aria-describedby={
            error && id
              ? [ariaDescribedBy, `${id}-error`].filter(Boolean).join(" ")
              : ariaDescribedBy
          }
          onInvalid={onInvalid}
          className={`text-base font-normal ${className ?? ""}`.trimEnd()}
        />
        {error ? <FormFieldError id={id ? `${id}-error` : undefined} message={error} /> : null}
      </div>
    );
  }

  return (
    <PropertyRow label={label} description={description}>
      <div className="flex flex-col gap-0.5">
        <div className="flex items-center gap-1">
          <span className={`text-sm ${C.text65} select-none`}>¥</span>
          <input
            type="number"
            id={id}
            name={name}
            min={min}
            max={max}
            aria-label={ariaLabel ?? label}
            aria-invalid={ariaInvalid ?? (error ? true : undefined)}
            aria-required={ariaRequired}
            aria-describedby={ariaDescribedBy}
            onInvalid={onInvalid}
            className={`w-32 bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-xxs ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder} focus-visible:ring-2 ${C.focusRingAccent40}`}
            value={rawValue}
            onChange={(e) => handleChange(e.target.value)}
            placeholder={placeholder}
            disabled={disabled}
            required={required}
            readOnly={readOnly}
            step={step}
          />
        </div>
        {error ? <FormFieldError message={error} /> : null}
      </div>
    </PropertyRow>
  );
}
