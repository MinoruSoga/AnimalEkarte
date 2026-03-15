// Internal
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { C } from "@/lib/design-tokens";

interface MoneyInputProps {
  label?: string;
  value: number;
  onChange: (value: number) => void;
  placeholder?: string;
}

export function MoneyInput({ label = "単価(税込)", value, onChange, placeholder = "0" }: MoneyInputProps) {
  return (
    <PropertyRow label={label}>
      <div className="flex items-center gap-1">
        <span className={`text-sm ${C.text65} select-none`}>¥</span>
        <input
          type="number"
          min={0}
          className={`w-32 bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-[3px] ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
          value={value === 0 ? "" : value}
          onChange={(e) => onChange(e.target.valueAsNumber || 0)}
          placeholder={placeholder}
        />
      </div>
    </PropertyRow>
  );
}
