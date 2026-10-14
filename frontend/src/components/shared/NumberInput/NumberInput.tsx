// React/Framework
import { memo, type ComponentProps } from "react";

// Internal
import { Input } from "@/components/ui/input";
import { C } from "@/lib/design-tokens";

interface NumberInputProps {
  value: number | string;
  onChange: (value: string) => void;
  placeholder?: string;
  step?: number | "any";
  min?: number;
  max?: number;
  /** 入力フィールド右端に表示する単位テキスト（例: "円", "kg", "℃"） */
  suffix?: string;
  align?: "left" | "right";
  className?: string;
  disabled?: boolean;
  id?: string;
  name?: string;
  required?: boolean;
  readOnly?: boolean;
  "aria-label"?: ComponentProps<"input">["aria-label"];
  "aria-required"?: ComponentProps<"input">["aria-required"];
  "aria-invalid"?: ComponentProps<"input">["aria-invalid"];
  "aria-describedby"?: ComponentProps<"input">["aria-describedby"];
  onInvalid?: ComponentProps<"input">["onInvalid"];
}

export const NumberInput = memo(function NumberInput({
  value,
  onChange,
  placeholder = "0",
  step = 1,
  min,
  max,
  suffix,
  align = "left",
  className,
  disabled,
  id,
  name,
  required,
  readOnly,
  "aria-label": ariaLabel,
  "aria-required": ariaRequired,
  "aria-invalid": ariaInvalid,
  "aria-describedby": ariaDescribedBy,
  onInvalid,
}: NumberInputProps) {
  return (
    <div className="relative">
      <Input
        id={id}
        name={name}
        aria-label={ariaLabel}
        type="number"
        step={step}
        min={min}
        max={max}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled}
        required={required}
        readOnly={readOnly}
        aria-required={ariaRequired}
        aria-invalid={ariaInvalid}
        aria-describedby={ariaDescribedBy}
        onInvalid={onInvalid}
        className={`${suffix ? "pr-10" : ""} ${align === "right" ? "text-right" : ""} ${className ?? ""}`.trimEnd()}
      />
      {suffix ? (
        <span
          className={`pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-sm ${C.text60}`}
        >
          {suffix}
        </span>
      ) : null}
    </div>
  );
});
