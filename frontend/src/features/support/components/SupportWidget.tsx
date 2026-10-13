/**
 * SupportWidget — 画面右下に常駐するサポート起動ボタン + パネル
 *
 * Layout（認証済みシェル）に1つだけマウントされる。
 * `data-html2canvas-ignore` でボタン・パネル自身はバグ報告スクショの対象外。
 * DOM 順はボタン→パネル（flex-col-reverse で見た目はパネルが上）にして、
 * キーボード操作時にボタンの次の Tab でパネルへ入れるようにする。
 * EMR-387: パネル表示中のみ data-support-widget-open を付与する。無関係な
 * モーダル Dialog や SidePeek が開いている間は globals.css の body:has(...)
 * ルールが閉じた起動ボタンを display:none で隠し、確定/編集などの操作要素と
 * 重ならないようにする。パネルが開いているウィジェットは退避・非表示の対象外
 * （既存のモーダル中操作を維持する）。
 */
import { useCallback, useEffect, useRef, useState } from "react";
import { LifeBuoy, X } from "lucide-react";

import { C, ICON, STYLE, Z_CLASS } from "@/lib/design-tokens";
import { SUPPORT_WIDGET_LAYOUT } from "@/constants/support-widget-layout";

import { HelpChatHistoryProvider } from "./HelpChatHistoryProvider";

import { SupportPanel } from "./SupportPanel";

// FORMS-A11Y-01: 外部 SidePeek/dialog/alertdialog が開いている間、閉じた起動
// ボタンは globals.css で display:none となりフォーカスを保持できない。
// パネルを閉じる際は外部オーバーレイ内の適切な要素へフォーカスを返す。
const FOREIGN_OVERLAY_SELECTOR =
  '[data-side-peek], [role="dialog"][data-state="open"], [role="alertdialog"][data-state="open"]';
const FOCUSABLE_SELECTOR =
  'a[href], button, input, select, textarea, summary, [contenteditable="true"], [tabindex]';

const isVisibleEnabled = (el: Element | null): el is HTMLElement => {
  if (!(el instanceof HTMLElement) || !el.isConnected) return false;
  if (el.hasAttribute("disabled") || el.getAttribute("aria-disabled") === "true") return false;
  // type="hidden" の input はフォーカス不能
  if (el instanceof HTMLInputElement && el.type === "hidden") return false;
  // display:none は自身に限らず祖先が非表示でもフォーカスできないため、
  // inert/hidden 属性や visibility と併せて祖先チェーンを確認する。
  // （getClientRects は jsdom で常に空のため算出スタイルの実値で判定する）
  for (let node: HTMLElement | null = el; node; node = node.parentElement) {
    if (node.hasAttribute("hidden") || node.hasAttribute("inert")) return false;
    const style = window.getComputedStyle(node);
    if (style.display === "none") return false;
    if (style.visibility === "hidden" || style.visibility === "collapse") return false;
  }
  return true;
};

// tabindex="-1" はタブ順から外れた要素を含むため、タブ順で辿れる復帰先としては不適。
const isTabbableTarget = (el: Element | null): el is HTMLElement =>
  isVisibleEnabled(el) && el.tabIndex !== -1;

// ウィジェットが portal した dialog（HelpChat の確認ダイアログ等）は DOM 上は
// body 直下で [data-support-widget] の子孫ではない。Radix はトリガーに
// aria-controls="<content id>" を付けるため、それでウィジェット所有を判定し、
// 外部オーバーレイ判定から除外して自己参照を防ぐ。
const isWidgetOwnedElement = (el: HTMLElement) => {
  if (el.closest("[data-support-widget]") !== null) return true;
  const { id } = el;
  return (
    id !== "" &&
    document.querySelector(`[data-support-widget] [aria-controls="${CSS.escape(id)}"]`) !== null
  );
};

export function SupportWidget() {
  const [open, setOpen] = useState(false);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const lastExternalFocusRef = useRef<HTMLElement | null>(null);

  const restoreFocusAfterClose = useCallback(() => {
    // 最後に開いた外部オーバーレイ（DOM 順で最後）を復帰先パネルとする
    const overlay = Array.from(document.querySelectorAll<HTMLElement>(FOREIGN_OVERLAY_SELECTOR))
      .filter((el) => !isWidgetOwnedElement(el))
      .at(-1);
    if (!overlay) {
      // パネル内にフォーカスがあった状態で閉じてもロストしないよう起動ボタンへ戻す
      buttonRef.current?.focus();
      return;
    }
    // そのオーバーレイ内で直前にフォーカスされていた要素が有効ならそこへ戻し、
    // 無効/不明ならパネル内の先頭コントロールへフォールバックする。
    // focus() が実際に受理されなかった候補（モーダルの奪回・レースによる
    // 非活性化等）は activeElement で検証して次の候補へ進む。
    const previous = lastExternalFocusRef.current;
    const targets: HTMLElement[] = [];
    if (previous && overlay.contains(previous) && isTabbableTarget(previous)) {
      targets.push(previous);
    }
    for (const el of overlay.querySelectorAll(FOCUSABLE_SELECTOR)) {
      if (el !== previous && isTabbableTarget(el)) targets.push(el as HTMLElement);
    }
    // Radix DialogContent 等、tabindex="-1" でプログラム的にのみ受け取れる
    // コンテナを最後の受け皿とする（display:none/非接続にはフォーカスしない）
    if (isVisibleEnabled(overlay) && overlay.hasAttribute("tabindex")) {
      targets.push(overlay);
    }
    for (const target of targets) {
      target.focus();
      if (document.activeElement === target) return;
    }
  }, []);

  const close = useCallback(() => {
    setOpen(false);
    restoreFocusAfterClose();
  }, [restoreFocusAfterClose]);

  const toggle = useCallback(() => {
    if (open) {
      close();
    } else {
      setOpen(true);
    }
  }, [open, close]);

  // パネル表示中は Escape で閉じられるようにする（非モーダルのグローバルショートカット）
  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") close();
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [open, close]);

  // FORMS-A11Y-01: 閉じる際に「外部オーバーレイ内で直前にフォーカスされていた
  // 要素」へ戻せるよう、ウィジェット外の focusin を追跡しておく。
  // （ウィジェット内の focusin は除外。documentElement 上の他リスナーと同様に
  // 登録・解除だけを行う）
  useEffect(() => {
    const track = (event: Event) => {
      if (
        event.target instanceof HTMLElement &&
        event.target.closest("[data-support-widget]") === null
      ) {
        lastExternalFocusRef.current = event.target;
      }
    };
    document.documentElement.addEventListener("focusin", track);
    return () => document.documentElement.removeEventListener("focusin", track);
  }, []);

  // モーダル Dialog を開いたまま「既に開いている」パネルを操作できるようにする。
  // EMR-387: 閉じた起動ボタン自体はモーダル中に CSS で隠れるため、この遮蔽が
  // 意味を持つのはパネル表示中（data-support-widget-open）のみ。
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
      // EMR-387: パネル表示中のみ付与する開状態属性。globals.css の
      // body:has(...) ルールがこれを参照し、閉じた起動ボタンだけを隠す。
      data-support-widget-open={open ? "" : undefined}
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
