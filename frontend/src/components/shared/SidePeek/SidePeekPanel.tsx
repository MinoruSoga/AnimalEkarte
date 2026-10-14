import { memo } from "react";
import type { ReactNode } from "react";
import { STYLE, LAYOUT } from "@/lib/design-tokens";

interface SidePeekPanelProps {
  children: ReactNode;
  className?: string;
  onKeyDown?: (e: React.KeyboardEvent<HTMLDivElement>) => void;
  ariaLabel?: string;
}

export const SidePeekPanel = memo(function SidePeekPanel({
  children,
  className,
  onKeyDown,
  ariaLabel,
}: SidePeekPanelProps) {
  return (
    <div
      data-side-peek
      role={ariaLabel ? "dialog" : undefined}
      aria-label={ariaLabel}
      className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0 relative ${className ?? ""}`}
      onKeyDown={onKeyDown}
    >
      {children}
    </div>
  );
});
