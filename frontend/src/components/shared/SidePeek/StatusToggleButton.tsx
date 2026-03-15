// Internal
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
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
        className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
      >
        <NotionStatusPill isActive={isActive} />
      </button>
    </PropertyRow>
  );
}
