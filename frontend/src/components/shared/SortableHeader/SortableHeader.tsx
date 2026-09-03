import { memo } from "react";
import { ArrowUp, ArrowDown, ArrowUpDown } from "lucide-react";
import { C, ICON, STYLE } from "@/lib/design-tokens";

type SortDirection = "ascending" | "descending" | "none";

/**
 * a11y (FE-RC-044): この button は `<th>` を伴わない単体使用（DataTable の column.header へ差し込み）
 * のため、th 側に `aria-sort` を付与できない。代わりに現在の並び順を aria-label へ含めて
 * スクリーンリーダーに状態を伝える。direction="none" のときは既定文言（`${label}でソート`）を維持し、
 * 既存の呼び出し側テストの accessible name を変えない。
 */
const DIRECTION_ANNOUNCEMENT: Record<SortDirection, string> = {
  ascending: "（昇順）",
  descending: "（降順）",
  none: "",
};

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
      aria-label={`${label}でソート${DIRECTION_ANNOUNCEMENT[direction]}`}
    >
      {label}
      <Icon aria-hidden="true" className={`${ICON.xs} ${direction === "none" ? C.text30 : C.text}`} />
    </button>
  );
});
