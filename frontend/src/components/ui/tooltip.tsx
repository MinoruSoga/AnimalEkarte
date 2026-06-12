import { type ReactNode, useState, useRef } from "react";
import { createPortal } from "react-dom";
import { C } from "@/lib/design-tokens";
import { cn } from "./utils";

interface TooltipProps {
  content: string;
  children: ReactNode;
  className?: string;
}

export function Tooltip({ content, children, className }: TooltipProps) {
  const [pos, setPos] = useState<{ top: number; left: number } | null>(null);
  const triggerRef = useRef<HTMLDivElement>(null);

  const handleMouseEnter = () => {
    if (!triggerRef.current) return;
    const rect = triggerRef.current.getBoundingClientRect();
    setPos({ top: rect.top, left: rect.left + rect.width / 2 });
  };

  const handleMouseLeave = () => setPos(null);

  const isVisible = pos !== null;

  return (
    <div
      ref={triggerRef}
      className={cn("inline-flex", className)}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      {children}
      {createPortal(
        <div
          role="tooltip"
          aria-hidden={!isVisible}
          style={{
            position: "fixed",
            top: pos?.top ?? 0,
            left: pos?.left ?? 0,
            zIndex: 9999,
            transform: "translateX(-50%) translateY(calc(-100% - 4px))",
            visibility: isVisible ? "visible" : "hidden",
          }}
          className={cn(
            `pointer-events-none whitespace-nowrap rounded ${C.bgTooltip} px-2 py-1 text-xs ${C.textWhite}`,
            "transition-opacity duration-150",
            isVisible ? "opacity-100" : "opacity-0",
          )}
        >
          {content}
        </div>,
        document.body,
      )}
    </div>
  );
}
