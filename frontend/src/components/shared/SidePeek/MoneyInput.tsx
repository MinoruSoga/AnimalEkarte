// Internal
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { C } from "@/lib/design-tokens";

interface MoneyInputProps {
  label?: string;
  value: number;
  onChange: (value: number) => void;
  placeholder?: string;
  /** Field-level validation message (e.g. negative price) */
  error?: string;
}

export function MoneyInput({
  label = "単価(税込)",
  value,
  onChange,
  placeholder = "0",
  error,
}: MoneyInputProps) {
  return (
    <PropertyRow label={label}>
      <div className="flex flex-col gap-0.5">
        <div className="flex items-center gap-1">
          <span className={`text-sm ${C.text65} select-none`}>¥</span>
          <input
            type="number"
            aria-label={label}
            aria-invalid={error ? true : undefined}
            className={`w-32 bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-xxs ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
            value={Number.isNaN(value) || value === 0 ? "" : value}
            onChange={(e) => {
              const next = e.target.valueAsNumber;
              onChange(Number.isNaN(next) ? 0 : next);
            }}
            placeholder={placeholder}
          />
        </div>
        {error ? <FormFieldError message={error} /> : null}
      </div>
    </PropertyRow>
  );
}
