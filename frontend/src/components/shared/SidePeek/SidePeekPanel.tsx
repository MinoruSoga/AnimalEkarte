import { memo } from "react";
import type { ReactNode } from "react";
import { STYLE, LAYOUT } from "@/lib/design-tokens";

interface SidePeekPanelProps {
  children: ReactNode;
  className?: string;
  onKeyDown?: (e: React.KeyboardEvent<HTMLDivElement>) => void;
  ariaLabel?: string;
  /**
   * EMR-411: true の場合、viewport <1280px でパネルを absolute overlay にする。
   * workspace（list+panel flex 親が relative）でのみ成立する契約のため opt-in。
   * セクション埋め込みなど非 workspace の利用は既定 false（in-flow）のまま。
   */
  tabletOverlay?: boolean;
}

/**
 * Renders a non-modal side-peek surface, adding a dialog role only when `ariaLabel` is supplied.
 * It stays in flow by default; tablet overlay positioning requires a positioned workspace parent.
 */
export const SidePeekPanel = memo(function SidePeekPanel({
  children,
  className,
  onKeyDown,
  ariaLabel,
  tabletOverlay = false,
}: SidePeekPanelProps) {
  return (
    <div
      data-side-peek
      role={ariaLabel ? "dialog" : undefined}
      aria-label={ariaLabel}
      className={`${STYLE.sidePeekPanel} ${LAYOUT.sidePeek.width} shrink-0 ${tabletOverlay ? LAYOUT.sidePeek.position : "relative"} ${className ?? ""}`}
      onKeyDown={onKeyDown}
    >
      {children}
    </div>
  );
});
