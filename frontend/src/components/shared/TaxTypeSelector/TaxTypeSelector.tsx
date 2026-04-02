import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { TaxType } from "@/types/generated/models";
import { STYLE } from "@/lib/design-tokens";

// Static JSX hoisted (rendering-hoist-jsx)
const TAX_TYPE_SELECT_ITEMS = (
  <>
    <SelectItem value="excluded">外税</SelectItem>
    <SelectItem value="included">内税</SelectItem>
    <SelectItem value="exempt">非課税</SelectItem>
  </>
);

interface TaxTypeSelectorProps {
  value: TaxType;
  onChange: (value: TaxType) => void;
  disabled?: boolean;
}

export function TaxTypeSelector({
  value,
  onChange,
  disabled,
}: TaxTypeSelectorProps) {
  return (
    <Select
      value={value}
      onValueChange={(v) => onChange(v as TaxType)}
      disabled={disabled}
    >
      <SelectTrigger className={STYLE.selectCompact}>
        <SelectValue placeholder="選択" />
      </SelectTrigger>
      <SelectContent>{TAX_TYPE_SELECT_ITEMS}</SelectContent>
    </Select>
  );
}
