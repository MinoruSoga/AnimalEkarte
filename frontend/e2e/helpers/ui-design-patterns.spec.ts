import { expect, test } from "@playwright/test";

import {
  evaluatePatternSignals,
  findFixedOverlaps,
  findOffscreenCurrent,
  isFullyInViewport,
  type PatternSignals,
} from "./ui-design-patterns";

const VIEWPORT = { width: 1280, height: 900 };

function signals(overrides: Partial<PatternSignals> = {}): PatternSignals {
  return { current: null, fixed: [], operable: [], viewport: VIEWPORT, ...overrides };
}

test.describe("isFullyInViewport", () => {
  test("inside the viewport passes", () => {
    expect(isFullyInViewport({ top: 10, left: 10, width: 100, height: 40 }, VIEWPORT)).toBe(true);
  });

  test("partially below the fold fails (§1.1 requires no scroll)", () => {
    expect(isFullyInViewport({ top: 880, left: 10, width: 100, height: 40 }, VIEWPORT)).toBe(
      false,
    );
  });

  test("above the top edge fails", () => {
    expect(isFullyInViewport({ top: -5, left: 10, width: 100, height: 40 }, VIEWPORT)).toBe(false);
  });
});

test.describe("findOffscreenCurrent (design-patterns §1.1)", () => {
  test("no current item (login) is not a finding", () => {
    expect(findOffscreenCurrent(signals())).toBeNull();
  });

  test("visible current item passes", () => {
    const result = findOffscreenCurrent(
      signals({
        current: { descriptor: "a(会計)", rect: { top: 200, left: 0, width: 240, height: 44 } },
      }),
    );
    expect(result).toBeNull();
  });

  test("current item below the viewport is reported with its rect", () => {
    const result = findOffscreenCurrent(
      signals({
        current: {
          descriptor: "a(スタッフマスタ)",
          rect: { top: 1683, left: 0, width: 240, height: 44 },
        },
      }),
    );
    expect(result).toBe("a(スタッフマスタ)@0,1683 240x44");
  });
});

test.describe("findFixedOverlaps (design-patterns §5.1)", () => {
  const fab = {
    descriptor: "button[aria-label=サポート・ヘルプを開く]",
    rect: { top: 836, left: 1216, width: 48, height: 48 },
  };

  test("no fixed elements means no findings", () => {
    expect(
      findFixedOverlaps(
        signals({
          operable: [{ descriptor: "button(最後のページ)", rect: { top: 856, left: 1200, width: 44, height: 44 } }],
        }),
      ),
    ).toEqual([]);
  });

  test("fixed button covering a pagination button is reported", () => {
    const result = findFixedOverlaps(
      signals({
        fixed: [fab],
        operable: [
          { descriptor: "button(最後のページ)", rect: { top: 856, left: 1200, width: 44, height: 44 } },
          { descriptor: "a(飼主一覧)", rect: { top: 100, left: 0, width: 240, height: 44 } },
        ],
      }),
    );
    expect(result).toEqual([
      "button[aria-label=サポート・ヘルプを開く]@1216,836 48x48 covers button(最後のページ)@1200,856 44x44",
    ]);
  });

  test("touching edges without intersection is not an overlap", () => {
    const result = findFixedOverlaps(
      signals({
        fixed: [fab],
        operable: [
          { descriptor: "button(次へ)", rect: { top: 836, left: 1168, width: 48, height: 48 } },
        ],
      }),
    );
    expect(result).toEqual([]);
  });

  test("fixed elements do not count as overlapping each other", () => {
    const result = findFixedOverlaps(
      signals({
        fixed: [fab, { descriptor: "button(保存)", rect: { top: 840, left: 1220, width: 40, height: 40 } }],
      }),
    );
    expect(result).toEqual([]);
  });
});

test("evaluatePatternSignals combines both findings", () => {
  const findings = evaluatePatternSignals(
    signals({
      current: { descriptor: "a(設定)", rect: { top: 1200, left: 0, width: 240, height: 44 } },
      fixed: [{ descriptor: "button(help)", rect: { top: 836, left: 1216, width: 48, height: 48 } }],
      operable: [{ descriptor: "button(last)", rect: { top: 856, left: 1200, width: 44, height: 44 } }],
    }),
  );
  expect(findings.offscreenCurrent).toBe("a(設定)@0,1200 240x44");
  expect(findings.fixedOverlaps).toHaveLength(1);
});
