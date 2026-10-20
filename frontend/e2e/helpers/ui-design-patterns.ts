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
  /** `[aria-current="page"]` の矩形。無いページ（ログイン等）は null */
  current: MeasuredElement | null;
  /** position: fixed の操作要素 */
  fixed: MeasuredElement[];
  /** fixed 以外の可視な操作要素 */
  operable: MeasuredElement[];
  viewport: { width: number; height: number };
}

export interface PatternFindings {
  /** §1.1 違反: 現在地が viewport 外（descriptor と矩形）。null なら合格 */
  offscreenCurrent: string | null;
  /** §5.1 違反: 「fixed要素 ⟂ 操作要素」の重なり */
  fixedOverlaps: string[];
}

/** 矩形が viewport に完全に収まっているか（部分的に見えるだけでは不可: §1.1「スクロールなしで視認」） */
export function isFullyInViewport(rect: Rect, viewport: { width: number; height: number }): boolean {
  return (
    rect.top >= 0 &&
    rect.left >= 0 &&
    rect.top + rect.height <= viewport.height &&
    rect.left + rect.width <= viewport.width
  );
}

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

/** §1.1: 現在地が viewport 外なら descriptor を返す。現在地が無いページは null（判定対象外）。 */
export function findOffscreenCurrent(signals: PatternSignals): string | null {
  if (!signals.current) return null;
  if (isFullyInViewport(signals.current.rect, signals.viewport)) return null;
  return `${signals.current.descriptor}@${describeRect(signals.current.rect)}`;
}

/** §5.1: fixed 操作要素が他の操作要素と矩形で重なる組を返す（重複排除・決定的順序）。 */
export function findFixedOverlaps(signals: PatternSignals): string[] {
  const overlaps: string[] = [];
  for (const fixed of signals.fixed) {
    for (const target of signals.operable) {
      if (!intersects(fixed.rect, target.rect)) continue;
      overlaps.push(
        `${fixed.descriptor}@${describeRect(fixed.rect)} covers ${target.descriptor}@${describeRect(target.rect)}`,
      );
    }
  }
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

    const currentElement = document.querySelector<HTMLElement>("nav [aria-current='page']");
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
      current,
      fixed,
      operable,
      viewport: { width: window.innerWidth, height: window.innerHeight },
    };
  }, OPERABLE_SELECTOR);
}
