import { memo } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { STYLE } from "@/lib/design-tokens";

// Static JSX hoisted (rendering-hoist-jsx)
const TAX_RATE_SELECT_ITEMS_STANDARD = (
  <>
    <SelectItem value="0.1">10%（通常課税）</SelectItem>
    <SelectItem value="0.08">8%（軽減税率）</SelectItem>
  </>
);

interface TaxRateSelectorProps {
  value: number;
  onChange: (value: number) => void;
  disabled?: boolean;
  ariaLabel?: string;
  className?: string;
}

export const TaxRateSelector = memo(function TaxRateSelector({
  value,
  onChange,
  disabled,
  ariaLabel,
  className,
}: TaxRateSelectorProps) {
  return (
    <Select value={String(value)} onValueChange={(v) => onChange(Number(v))} disabled={disabled}>
      <SelectTrigger
        aria-label={ariaLabel}
        className={`${STYLE.selectCompact} ${className ?? ""}`.trimEnd()}
      >
        <SelectValue placeholder="選択" />
      </SelectTrigger>
      <SelectContent>{TAX_RATE_SELECT_ITEMS_STANDARD}</SelectContent>
    </Select>
  );
});
