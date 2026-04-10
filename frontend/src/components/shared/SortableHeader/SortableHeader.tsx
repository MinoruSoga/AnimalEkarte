import { memo } from "react";
import { ArrowUp, ArrowDown, ArrowUpDown } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";

type SortDirection = "ascending" | "descending" | "none";

interface SortableHeaderProps {
  label: string;
  direction: SortDirection;
  onToggle: () => void;
}

export const SortableHeader = memo(function SortableHeader({ label, direction, onToggle }: SortableHeaderProps) {
  const Icon =
    direction === "ascending"
      ? ArrowUp
      : direction === "descending"
        ? ArrowDown
        : ArrowUpDown;

  return (
    <button
      type="button"
      onClick={onToggle}
      className={`inline-flex items-center gap-1 cursor-pointer select-none ${C.hoverText60} transition-colors ${C.text}`}
      aria-label={`${label}でソート`}
    >
      {label}
      <Icon className={`${ICON.xs} ${direction === "none" ? C.text30 : C.text}`} />
    </button>
  );
});
