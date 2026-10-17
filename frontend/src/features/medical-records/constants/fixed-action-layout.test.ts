import { describe, it, expect } from "vitest";

import { MEDICAL_RECORD_FIXED_ACTION_LAYOUT } from "./fixed-action-layout";

/**
 * EMR-432: カルテ編集画面の右下固定ボタン群が、スクロール末尾の入力・表を覆わないこと。
 * 位置・高さのクラスと px 定数の算術がずれると、余白が足りなくなる（または過剰になる）。
 */
describe("MEDICAL_RECORD_FIXED_ACTION_LAYOUT", () => {
  it("footprintPx は position(bottom-N) + buttonHeightClass(h-N) の算術と一致する", () => {
    const bottomOffsetPx =
      Number(MEDICAL_RECORD_FIXED_ACTION_LAYOUT.position.match(/bottom-(\d+)/)?.[1]) * 4;
    const buttonHeightPx =
      Number(MEDICAL_RECORD_FIXED_ACTION_LAYOUT.buttonHeightClass.match(/^h-(\d+)$/)?.[1]) * 4;
    expect(bottomOffsetPx).toBeGreaterThan(0);
    expect(buttonHeightPx).toBeGreaterThan(0);
    expect(bottomOffsetPx).toBe(MEDICAL_RECORD_FIXED_ACTION_LAYOUT.bottomOffsetPx);
    expect(buttonHeightPx).toBe(MEDICAL_RECORD_FIXED_ACTION_LAYOUT.buttonHeightPx);
    expect(MEDICAL_RECORD_FIXED_ACTION_LAYOUT.footprintPx).toBe(bottomOffsetPx + buttonHeightPx);
  });

  it("scrollBottomClearance は footprintPx + slackPx と一致し、pb-N の px 換算とも揃う", () => {
    const clearancePx =
      Number(MEDICAL_RECORD_FIXED_ACTION_LAYOUT.scrollBottomClearance.match(/^pb-(\d+)$/)?.[1]) * 4;
    expect(clearancePx).toBeGreaterThan(0);
    expect(MEDICAL_RECORD_FIXED_ACTION_LAYOUT.scrollBottomClearancePx).toBe(clearancePx);
    expect(MEDICAL_RECORD_FIXED_ACTION_LAYOUT.scrollBottomClearancePx).toBe(
      MEDICAL_RECORD_FIXED_ACTION_LAYOUT.footprintPx + MEDICAL_RECORD_FIXED_ACTION_LAYOUT.slackPx,
    );
  });

  it("下余白はボタン群の占有高さを上回る（末尾の要素がボタン群に隠れない）", () => {
    expect(MEDICAL_RECORD_FIXED_ACTION_LAYOUT.scrollBottomClearancePx).toBeGreaterThan(
      MEDICAL_RECORD_FIXED_ACTION_LAYOUT.footprintPx,
    );
  });
});
