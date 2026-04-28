import { C } from "@/lib/design-tokens";

interface PetDeceasedBadgeProps {
  deceasedAt: string;
}

export function PetDeceasedBadge({ deceasedAt: _deceasedAt }: PetDeceasedBadgeProps) {
  return (
    <span className={`text-xs ${C.text50}`}>
      （故）
    </span>
  );
}
