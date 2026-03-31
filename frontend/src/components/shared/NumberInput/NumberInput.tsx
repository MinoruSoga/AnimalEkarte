// Internal
import { Input } from "@/components/ui/input";

interface NumberInputProps {
  value: number | string;
  onChange: (value: string) => void;
  placeholder?: string;
  step?: number;
  min?: number;
  max?: number;
  /** 入力フィールド右端に表示する単位テキスト（例: "円", "kg", "℃"） */
  suffix?: string;
  align?: "left" | "right";
  className?: string;
  disabled?: boolean;
  id?: string;
}

export function NumberInput({
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
}: NumberInputProps) {
  return (
    <div className="relative">
      <Input
        id={id}
        type="number"
        step={step}
        min={min}
        max={max}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled}
        className={`${suffix ? "pr-10" : ""} ${align === "right" ? "text-right" : ""} ${className ?? ""}`.trimEnd()}
      />
      {suffix ? (
        <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-sm text-[#37352F]/60">
          {suffix}
        </span>
      ) : null}
    </div>
  );
}
