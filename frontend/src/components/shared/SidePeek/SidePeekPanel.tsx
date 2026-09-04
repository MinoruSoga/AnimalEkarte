import { memo } from "react";
import type { ReactNode } from "react";
import { STYLE, LAYOUT } from "@/lib/design-tokens";

interface SidePeekPanelProps {
  children: ReactNode;
  className?: string;
  onKeyDown?: (e: React.KeyboardEvent<HTMLDivElement>) => void;
}

export const SidePeekPanel = memo(function SidePeekPanel({
  children,
  className,
  onKeyDown,
}: SidePeekPanelProps) {
  return (
    <div
      className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0 ${className ?? ""}`}
      onKeyDown={onKeyDown}
    >
      {children}
    </div>
  );
});
