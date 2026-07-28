// Internal
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { C } from "@/lib/design-tokens";

interface StatusToggleButtonProps {
  isActive: boolean;
  onToggle: () => void;
}

export function StatusToggleButton({ isActive, onToggle }: StatusToggleButtonProps) {
  return (
    <PropertyRow label="ステータス">
      <button
        type="button"
        onClick={onToggle}
        aria-label="ステータスを切り替え"
        className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
      >
        <StatusPill isActive={isActive} />
      </button>
    </PropertyRow>
  );
}
