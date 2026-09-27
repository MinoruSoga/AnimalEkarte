/**
 * SupportWidget — 画面右下に常駐するサポート起動ボタン + パネル
 *
 * Layout（認証済みシェル）に1つだけマウントされる。
 * `data-html2canvas-ignore` でボタン・パネル自身はバグ報告スクショの対象外。
 * DOM 順はボタン→パネル（flex-col-reverse で見た目はパネルが上）にして、
 * キーボード操作時にボタンの次の Tab でパネルへ入れるようにする。
 */
import { useCallback, useEffect, useRef, useState } from "react";
import { LifeBuoy, X } from "lucide-react";

import { C, ICON, STYLE, Z_CLASS } from "@/lib/design-tokens";

import { SupportPanel } from "./SupportPanel";

export function SupportWidget() {
  const [open, setOpen] = useState(false);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const toggle = useCallback(() => setOpen((v) => !v), []);
  const close = useCallback(() => {
    setOpen(false);
    // パネル内にフォーカスがあった状態で閉じてもロストしないよう起動ボタンへ戻す
    buttonRef.current?.focus();
  }, []);

  // パネル表示中は Escape で閉じられるようにする（非モーダルのグローバルショートカット）
  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") close();
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [open, close]);

  return (
    <div
      data-html2canvas-ignore
      className={`no-print fixed bottom-4 right-4 ${Z_CLASS.overlay} flex flex-col-reverse items-end gap-3`}
    >
      <button
        ref={buttonRef}
        type="button"
        onClick={toggle}
        aria-expanded={open}
        aria-controls={open ? "support-panel" : undefined}
        aria-label={open ? "サポートを閉じる" : "サポート・ヘルプを開く"}
        className={`size-12 rounded-full ${C.bgActionPrimary} ${C.textOnActionPrimary} ${C.hoverBgActionPrimary} ${STYLE.primaryGlow} flex items-center justify-center transition-colors`}
      >
        {open ? <X className={ICON.lg} /> : <LifeBuoy className={ICON.lg} />}
      </button>
      {open ? <SupportPanel onClose={close} /> : null}
    </div>
  );
}
