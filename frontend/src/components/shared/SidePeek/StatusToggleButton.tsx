// Internal
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow";
import { PropertyToggleButton } from "@/components/shared/SidePeek/PropertyToggleButton";
import { FIELD_DESCRIPTIONS } from "@/constants/field-descriptions";

interface StatusToggleButtonProps {
  isActive: boolean;
  onToggle: () => void;
  /** 項目の説明文（ⓘツールチップ）。省略時はステータス共通の説明を表示 */
  description?: string;
}

/**
 * Renders the active state as a toggle control.
 * `isActive` drives the pressed state; the caller updates it in response to `onToggle`.
 */
export function StatusToggleButton({
  isActive,
  onToggle,
  description = FIELD_DESCRIPTIONS.status,
}: StatusToggleButtonProps) {
  return (
    <PropertyRow label="ステータス" description={description}>
      <PropertyToggleButton
        label="ステータス"
        pressed={isActive}
        onPressedChange={() => onToggle()}
        className="px-0.5"
      >
        <StatusPill isActive={isActive} />
      </PropertyToggleButton>
    </PropertyRow>
  );
}
