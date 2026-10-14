// Internal
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { C } from "@/lib/design-tokens";

interface MoneyInputBaseProps {
  label?: string;
  min?: number;
  placeholder?: string;
  /** Field-level validation message (e.g. negative price) */
  error?: string;
  /** 項目の説明文（ⓘツールチップ） */
  description?: string;
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
  placeholder = "0",
  error,
  description,
}: MoneyInputProps) {
  return (
    <PropertyRow label={label} description={description}>
      <div className="flex flex-col gap-0.5">
        <div className="flex items-center gap-1">
          <span className={`text-sm ${C.text65} select-none`}>¥</span>
          <input
            type="number"
            min={min}
            aria-label={label}
            aria-invalid={error ? true : undefined}
            className={`w-32 bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-xxs ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder} focus-visible:ring-2 ${C.focusRingAccent40}`}
            value={stringValue ?? (Number.isNaN(value) || value === 0 ? "" : value)}
            onChange={(e) => {
              if (onStringChange) {
                onStringChange(e.target.value);
              } else if (onChange) {
                const next = e.target.valueAsNumber;
                onChange(Number.isNaN(next) ? 0 : next);
              }
            }}
            placeholder={placeholder}
          />
        </div>
        {error ? <FormFieldError message={error} /> : null}
      </div>
    </PropertyRow>
  );
}
