import { describe, it, expect } from "vitest";

import { PERIOD_OPTIONS, PERIOD_LABELS } from "./constants";

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
