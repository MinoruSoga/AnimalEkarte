import { type KeyboardEvent, type ReactNode, useState, useRef } from "react";
import { createPortal } from "react-dom";
import { C, Z } from "@/lib/design-tokens";
import { cn } from "./utils";

interface TooltipProps {
  content: string;
  children: ReactNode;
  className?: string;
}

const VIEWPORT_MARGIN = 8;

export function Tooltip({ content, children, className }: TooltipProps) {
  const [pos, setPos] = useState<{ top: number; left: number; below: boolean } | null>(null);
  const triggerRef = useRef<HTMLSpanElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);

  const show = () => {
    if (!triggerRef.current) return;
    const rect = triggerRef.current.getBoundingClientRect();
    // visibility:hidden でもレイアウトは確定しているため実測できる
    const tipW = tooltipRef.current?.offsetWidth ?? 0;
    const tipH = tooltipRef.current?.offsetHeight ?? 0;
    // 中央寄せしたときに左右の画面外へはみ出さないようクランプ
    const left = Math.min(
      Math.max(rect.left + rect.width / 2, tipW / 2 + VIEWPORT_MARGIN),
      window.innerWidth - tipW / 2 - VIEWPORT_MARGIN,
    );
    // 上に入らない場合はトリガーの下に表示
    const below = rect.top - tipH - 4 < VIEWPORT_MARGIN;
    setPos({ top: below ? rect.bottom : rect.top, left, below });
  };

  const hide = () => setPos(null);

  // WCAG 1.4.13: ホバー/フォーカスで表示する追加コンテンツは Esc で閉じられること
  const handleKeyDown = (e: KeyboardEvent<HTMLSpanElement>) => {
    if (e.key === "Escape") {
      e.stopPropagation();
      hide();
    }
  };

  const isVisible = pos !== null;

  return (
    <span
      ref={triggerRef}
      className={cn("inline-flex", className)}
      onMouseEnter={show}
      onMouseLeave={hide}
      onFocusCapture={show}
      onBlurCapture={hide}
      onKeyDown={handleKeyDown}
    >
      {children}
      {createPortal(
        <div
          ref={tooltipRef}
          role="tooltip"
          aria-hidden={!isVisible}
          style={{
            position: "fixed",
            top: pos?.top ?? 0,
            left: pos?.left ?? 0,
            zIndex: Z.overlay,
            transform: pos?.below
              ? "translateX(-50%) translateY(4px)"
              : "translateX(-50%) translateY(calc(-100% - 4px))",
            visibility: isVisible ? "visible" : "hidden",
            // fixed の shrink-to-fit は left〜ビューポート右端の空き幅に制限される。
            // width:max-content で内容幅（上限 maxWidth）を確保し、はみ出しは位置クランプで防ぐ
            maxWidth: "min(28rem, calc(100vw - 16px))",
          }}
          className={cn(
            `pointer-events-none w-max whitespace-pre-line rounded ${C.bgTooltip} px-2.5 py-1.5 text-xs leading-relaxed ${C.textWhite}`,
            "transition-opacity duration-150",
            isVisible ? "opacity-100" : "opacity-0",
          )}
        >
          {content}
        </div>,
        document.body,
      )}
    </span>
  );
}
