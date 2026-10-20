import { type Page } from "@playwright/test";

/**
 * design-patterns.md §1.1 / §5.1 のレイアウト信号を実ブラウザから取る。
 *
 * - §1.1: サイドバーの現在地（`[aria-current="page"]`）が初期表示で viewport 内にあること
 * - §5.1: 固定要素（`position: fixed` の button / a）が他の操作要素に重ならないこと
 *
 * ブラウザ側で取るのは矩形だけで、判定は純粋関数（`findOffscreenCurrent` /
 * `findFixedOverlaps`）に寄せる。純粋関数は `ui-design-patterns.spec.ts` で単体検証する。
 */

export interface Rect {
  top: number;
  left: number;
  width: number;
  height: number;
}

export interface MeasuredElement {
  descriptor: string;
  rect: Rect;
}

export interface PatternSignals {
  /** メインナビ（`nav[aria-label="メインナビゲーション"]`）の矩形。アプリシェル外（ログイン等）は null */
  nav: Rect | null;
  /** メインナビ内の `[aria-current="page"]` の矩形。無ければ null */
  current: MeasuredElement | null;
  /** position: fixed の操作要素 */
  fixed: MeasuredElement[];
  /** fixed 以外の可視な操作要素 */
  operable: MeasuredElement[];
  viewport: { width: number; height: number };
}

export interface PatternFindings {
  /** §1.1 違反: 現在地がメインナビの可視範囲または viewport の外、あるいは展開ナビに現在地が無い。null なら合格 */
  offscreenCurrent: string | null;
  /** §5.1 違反: 「fixed要素 ⟂ 操作要素」の重なり */
  fixedOverlaps: string[];
}

/** 矩形が viewport に完全に収まっているか（部分的に見えるだけでは不可: §1.1「スクロールなしで視認」） */
export function isFullyInViewport(rect: Rect, viewport: { width: number; height: number }): boolean {
  return isContained(rect, { top: 0, left: 0, width: viewport.width, height: viewport.height });
}

/** inner が outer に完全に含まれるか（nav のスクロール範囲に対するクリップ判定にも使う）。 */
export function isContained(inner: Rect, outer: Rect): boolean {
  return (
    inner.top >= outer.top &&
    inner.left >= outer.left &&
    inner.top + inner.height <= outer.top + outer.height &&
    inner.left + inner.width <= outer.left + outer.width
  );
}

/**
 * サイドバーが展開される幅（`Sidebar.tsx` の `window.innerWidth < 1280` と同じ境界）。
 * これ未満では折りたたみ表示で現在地の子項目を描画しないため、現在地の有無は判定しない。
 */
export const SIDEBAR_EXPANDED_MIN_WIDTH = 1280;

function intersects(a: Rect, b: Rect): boolean {
  return (
    a.left < b.left + b.width &&
    a.left + a.width > b.left &&
    a.top < b.top + b.height &&
    a.top + a.height > b.top
  );
}

function describeRect(rect: Rect): string {
  return `${Math.round(rect.left)},${Math.round(rect.top)} ${Math.round(rect.width)}x${Math.round(rect.height)}`;
}

/**
 * §1.1: 現在地がメインナビの可視矩形と viewport の両方に収まっていなければ descriptor を返す。
 * メインナビが無いページ（公開ルート・単独ウィンドウ）は対象外。メインナビがあり展開幅
 * （>= 1280px）なのに現在地が無い場合も違反（折りたたみ幅では子項目を描画しないため判定しない）。
 */
export function findOffscreenCurrent(signals: PatternSignals): string | null {
  if (!signals.nav) return null;
  if (!signals.current) {
    return signals.viewport.width >= SIDEBAR_EXPANDED_MIN_WIDTH
      ? "main nav has no [aria-current=page] item"
      : null;
  }
  const { rect } = signals.current;
  if (isFullyInViewport(rect, signals.viewport) && isContained(rect, signals.nav)) return null;
  return `${signals.current.descriptor}@${describeRect(rect)} (nav ${describeRect(signals.nav)})`;
}

/**
 * §5.1: fixed 操作要素が他の操作要素（通常要素と、自分以外の fixed 要素）と矩形で重なる組を返す
 * （重複排除・決定的順序）。固定アクションバーのボタンと FAB の衝突も検出する。
 */
export function findFixedOverlaps(signals: PatternSignals): string[] {
  const overlaps: string[] = [];
  signals.fixed.forEach((fixed, index) => {
    const targets = [
      ...signals.operable,
      ...signals.fixed.filter((_, otherIndex) => otherIndex !== index),
    ];
    for (const target of targets) {
      if (!intersects(fixed.rect, target.rect)) continue;
      overlaps.push(
        `${fixed.descriptor}@${describeRect(fixed.rect)} covers ${target.descriptor}@${describeRect(target.rect)}`,
      );
    }
  });
  return [...new Set(overlaps)].sort();
}

export function evaluatePatternSignals(signals: PatternSignals): PatternFindings {
  return {
    offscreenCurrent: findOffscreenCurrent(signals),
    fixedOverlaps: findFixedOverlaps(signals),
  };
}

const OPERABLE_SELECTOR =
  "a[href], button, input:not([type='hidden']), select, textarea, [role='button'], [role='link'], [role='tab'], [role='switch']";

/** ページから矩形だけを採取する。判定はしない（`evaluatePatternSignals` へ渡す）。 */
export async function collectPatternSignals(page: Page): Promise<PatternSignals> {
  return page.evaluate((selector) => {
    const toRect = (element: Element) => {
      const r = element.getBoundingClientRect();
      return { top: r.top, left: r.left, width: r.width, height: r.height };
    };
    const describe = (element: HTMLElement) =>
      [
        element.tagName.toLowerCase(),
        element.id ? `#${element.id}` : "",
        element.getAttribute("aria-label") ? `[aria-label=${element.getAttribute("aria-label")}]` : "",
        element.dataset.testid ? `[testid=${element.dataset.testid}]` : "",
        element.textContent?.trim() ? `(${element.textContent.trim().slice(0, 20)})` : "",
      ].join("");
    const isVisible = (element: HTMLElement) => {
      const style = getComputedStyle(element);
      const r = element.getBoundingClientRect();
      return (
        element.getAttribute("aria-hidden") !== "true" &&
        style.display !== "none" &&
        style.visibility !== "hidden" &&
        r.width > 0 &&
        r.height > 0
      );
    };
    const isFixed = (element: HTMLElement) => {
      let node: HTMLElement | null = element;
      while (node && node !== document.body) {
        if (getComputedStyle(node).position === "fixed") return true;
        node = node.parentElement;
      }
      return false;
    };

    // メインナビだけを見る（ページネーション等の nav[aria-current] を拾わない）。
    const navElement = document.querySelector<HTMLElement>("nav[aria-label='メインナビゲーション']");
    const nav = navElement && isVisible(navElement) ? toRect(navElement) : null;
    const currentElement = navElement?.querySelector<HTMLElement>("[aria-current='page']") ?? null;
    const current =
      currentElement && isVisible(currentElement)
        ? { descriptor: describe(currentElement), rect: toRect(currentElement) }
        : null;

    const fixed: { descriptor: string; rect: ReturnType<typeof toRect> }[] = [];
    const operable: { descriptor: string; rect: ReturnType<typeof toRect> }[] = [];
    for (const element of document.querySelectorAll<HTMLElement>(selector)) {
      if (!isVisible(element)) continue;
      const measured = { descriptor: describe(element), rect: toRect(element) };
      if (isFixed(element)) fixed.push(measured);
      else operable.push(measured);
    }
    return {
      nav,
      current,
      fixed,
      operable,
      viewport: { width: window.innerWidth, height: window.innerHeight },
    };
  }, OPERABLE_SELECTOR);
}
