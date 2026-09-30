// Internal
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { FIELD_DESCRIPTIONS } from "@/constants/field-descriptions";
import { C } from "@/lib/design-tokens";

interface StatusToggleButtonProps {
  isActive: boolean;
  onToggle: () => void;
  /** 項目の説明文（ⓘツールチップ）。省略時はステータス共通の説明を表示 */
  description?: string;
}

export function StatusToggleButton({
  isActive,
  onToggle,
  description = FIELD_DESCRIPTIONS.status,
}: StatusToggleButtonProps) {
  return (
    <PropertyRow label="ステータス" description={description}>
      <button
        type="button"
        onClick={onToggle}
        aria-label="ステータスを切り替え"
        className={`inline-flex min-h-11 items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
      >
        <StatusPill isActive={isActive} />
      </button>
    </PropertyRow>
  );
}
