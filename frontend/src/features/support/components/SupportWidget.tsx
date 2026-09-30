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
import { SUPPORT_WIDGET_LAYOUT } from "@/constants/support-widget-layout";

import { HelpChatHistoryProvider } from "./HelpChatHistoryProvider";

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

  // モーダル Dialog を開いたままウィジェットを操作できるようにする。
  // Radix は document で pointerdown（外側クリック判定）と focusin/focusout
  // （FocusScope のフォーカス回収）を監視するため、ウィジェット内で発生した
  // イベントが document に届かないよう documentElement のバブルで遮断する。
  // focusout は target ではなく relatedTarget（フォーカスの移動先）が
  // ウィジェット内の場合に止める必要がある（target はダイアログ内要素のため）。
  // コンテナ自身で止めると #root の React 委譲リスナーに届かずウィジェット
  // 自身の onClick 等が動かなくなるため、html 要素で止める（#root 内蔵の
  // React リスナーはバブル順序上先に実行済み）。
  useEffect(() => {
    const SHIELDED_EVENTS = ["pointerdown", "focusin", "focusout"] as const;
    const insideWidget = (node: unknown) =>
      node instanceof Element && node.closest("[data-support-widget]") !== null;
    const shield = (event: Event) => {
      if (
        insideWidget(event.target) ||
        (event instanceof FocusEvent && insideWidget(event.relatedTarget))
      ) {
        event.stopPropagation();
      }
    };
    for (const name of SHIELDED_EVENTS) {
      document.documentElement.addEventListener(name, shield);
    }
    return () => {
      for (const name of SHIELDED_EVENTS) {
        document.documentElement.removeEventListener(name, shield);
      }
    };
  }, []);

  return (
    <div
      data-html2canvas-ignore
      data-support-widget
      // pointer-events-auto: モーダル Dialog は body を pointer-events:none にする。
      // ダイアログ外のウィジェットは見えているのにクリックが背面へ透過してしまうため、
      // portaled Popover/Select と同じく明示的にイベントを復帰させる。
      className={`no-print pointer-events-auto ${SUPPORT_WIDGET_LAYOUT.position} ${Z_CLASS.overlay} flex flex-col-reverse items-end gap-3`}
    >
      <button
        ref={buttonRef}
        type="button"
        onClick={toggle}
        aria-expanded={open}
        aria-controls={open ? "support-panel" : undefined}
        aria-label={open ? "サポートを閉じる" : "サポート・ヘルプを開く"}
        className={`${SUPPORT_WIDGET_LAYOUT.buttonSizeClass} rounded-full ${C.bgActionPrimarySolid} ${C.textOnActionPrimary} ${C.hoverBgActionPrimarySolid} ${C.hoverTextOnActionPrimary} ${C.activeBgActionPrimarySolid} ${STYLE.primaryGlow} flex items-center justify-center transition-colors`}
      >
        {open ? <X className={ICON.lg} /> : <LifeBuoy className={ICON.lg} />}
      </button>
      <HelpChatHistoryProvider>
        {open ? <SupportPanel onClose={close} /> : null}
      </HelpChatHistoryProvider>
    </div>
  );
}
