import { describe, it, expect } from "vitest";

import {
  PERIOD_OPTIONS,
  PERIOD_LABELS,
  CATEGORY_LABELS,
  DISPLAY_CATEGORIES,
  UNCLASSIFIED_OTHER_LABEL,
} from "./constants";

describe("cash-register period constants (#150 EMG)", () => {
  it("区分セレクタは am/pm/emg の3区分を順番に提供する", () => {
    expect(PERIOD_OPTIONS).toEqual(["am", "pm", "emg"]);
  });

  it("各区分に日本語ラベルが定義されている", () => {
    expect(PERIOD_LABELS.am).toBe("午前");
    expect(PERIOD_LABELS.pm).toBe("午後");
    expect(PERIOD_LABELS.emg).toBe("緊急");
  });

  it("全 PERIOD_OPTIONS にラベルが存在する（漏れ防止）", () => {
    for (const p of PERIOD_OPTIONS) {
      expect(PERIOD_LABELS[p]).toBeTruthy();
    }
  });
});

describe("cash-register category labels DEC-40", () => {
  it("other は「未分類・要確認」にリネームされている（別バナーなし）", () => {
    expect(CATEGORY_LABELS.other).toBe("未分類・要確認");
    expect(UNCLASSIFIED_OTHER_LABEL).toBe("未分類・要確認");
    const otherGroup = DISPLAY_CATEGORIES.find((g) => g.keys.includes("other"));
    expect(otherGroup?.label).toBe(UNCLASSIFIED_OTHER_LABEL);
    expect(DISPLAY_CATEGORIES.some((g) => g.label === "その他")).toBe(false);
  });
});
