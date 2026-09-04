import { describe, expect, it } from "vitest";
import {
  AUTO_ADVANCE_HELPER_TEXT,
  EXPLICIT_PRIMARY_CTA_LABEL,
  FINAL_CONFIRM_CTA_LABEL,
  getAdvanceMode,
  getAutoAdvanceHelperText,
  type FlowAdvanceStep,
} from "./advance-policy";

/**
 * BUG-030: 予約フロー内で「タップ即遷移」と「次へ必須」が混在して混乱する問題。
 * 選択のみのステップは auto-on-select（ヘルパー文言必須）、
 * 入力・確定が必要なステップは explicit-cta（主CTA「次へ」等）に統一する。
 */
describe("advance-policy（BUG-030: 予約フロー advance ルール）", () => {
  const autoSteps: FlowAdvanceStep[] = [
    "courseSelect",
    "trimmingCourseSelect",
    "staffSelect",
    "timeSelect",
  ];

  const explicitSteps: FlowAdvanceStep[] = [
    "customerInfo",
    "trimmingOptionSelect",
    "dateSelect",
    "request",
    "confirm",
  ];

  it("選択のみのステップは auto-on-select である", () => {
    for (const step of autoSteps) {
      expect(getAdvanceMode(step)).toBe("auto-on-select");
    }
  });

  it("入力・複数選択・確定ステップは explicit-cta である", () => {
    for (const step of explicitSteps) {
      expect(getAdvanceMode(step)).toBe("explicit-cta");
    }
  });

  it("auto-on-select ステップには共通の進む旨ヘルパー文言が付く", () => {
    for (const step of autoSteps) {
      expect(getAutoAdvanceHelperText(step)).toBe(AUTO_ADVANCE_HELPER_TEXT);
    }
  });

  it("explicit-cta ステップには auto ヘルパーを出さない", () => {
    for (const step of explicitSteps) {
      expect(getAutoAdvanceHelperText(step)).toBeNull();
    }
  });

  it("主CTAラベルは「次へ」、最終確定は既存の「予約を確定する」を維持する", () => {
    expect(EXPLICIT_PRIMARY_CTA_LABEL).toBe("次へ");
    expect(FINAL_CONFIRM_CTA_LABEL).toBe("予約を確定する");
  });

  it("ヘルパー文言はタップで進むことを明示する", () => {
    expect(AUTO_ADVANCE_HELPER_TEXT).toMatch(/選択/);
    expect(AUTO_ADVANCE_HELPER_TEXT).toMatch(/進/);
  });
});
