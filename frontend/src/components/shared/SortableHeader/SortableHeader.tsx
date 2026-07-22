import { memo } from "react";
import { ArrowUp, ArrowDown, ArrowUpDown } from "lucide-react";
import { C, ICON, STYLE } from "@/lib/design-tokens";

type SortDirection = "ascending" | "descending" | "none";

interface SortableHeaderProps {
  label: string;
  direction: SortDirection;
  onToggle: () => void;
  /**
   * "eyebrow" — DESIGN.md ex-data-table-cell の header 相当（STYLE.sectionLabel: muted + uppercase +
   * typography eyebrow role）。既定の "default"（本文相当の濃い ${C.text}）は他 feature（examinations/checkups/
   * accounting/vaccinations/trimming/inventory/medical-records 等）で広く使われているため変更しない。
   */
  variant?: "default" | "eyebrow";
}

export const SortableHeader = memo(function SortableHeader({ label, direction, onToggle, variant = "default" }: SortableHeaderProps) {
  const Icon =
    direction === "ascending"
      ? ArrowUp
      : direction === "descending"
        ? ArrowDown
        : ArrowUpDown;

  const textClassName = variant === "eyebrow" ? STYLE.sectionLabel : C.text;

  return (
    <button
      type="button"
      onClick={onToggle}
      className={`-my-3 inline-flex min-h-11 min-w-11 items-center gap-1 cursor-pointer select-none ${C.hoverText60} transition-colors ${textClassName}`}
      aria-label={`${label}でソート`}
    >
      {label}
      <Icon className={`${ICON.xs} ${direction === "none" ? C.text30 : C.text}`} />
    </button>
  );
});
