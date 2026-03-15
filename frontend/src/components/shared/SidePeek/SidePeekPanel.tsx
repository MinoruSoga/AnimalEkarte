import type { ReactNode } from "react";
import { STYLE, LAYOUT } from "@/lib/design-tokens";

interface SidePeekPanelProps {
  children: ReactNode;
  className?: string;
}

export function SidePeekPanel({ children, className }: SidePeekPanelProps) {
  return (
    <div className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0 ${className ?? ""}`}>
      {children}
    </div>
  );
}
