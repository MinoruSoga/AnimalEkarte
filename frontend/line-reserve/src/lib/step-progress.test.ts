import { describe, expect, it } from "vitest";
import { getStepProgress } from "./step-progress";

describe("getStepProgress（SD-16: トリミング分岐でも一貫した進捗表示）", () => {
  it("通常フローは7ステップ+完了で total=8 になる", () => {
    expect(getStepProgress("customerInfo", false)).toEqual({ current: 1, total: 8 });
    expect(getStepProgress("courseSelect", false)).toEqual({ current: 2, total: 8 });
    expect(getStepProgress("staffSelect", false)).toEqual({ current: 3, total: 8 });
    expect(getStepProgress("dateSelect", false)).toEqual({ current: 4, total: 8 });
    expect(getStepProgress("timeSelect", false)).toEqual({ current: 5, total: 8 });
    expect(getStepProgress("request", false)).toEqual({ current: 6, total: 8 });
    expect(getStepProgress("confirm", false)).toEqual({ current: 7, total: 8 });
  });

  it("トリミングフローは9ステップ+完了で total=10 になり、分岐前後で連番が保たれる", () => {
    expect(getStepProgress("customerInfo", true)).toEqual({ current: 1, total: 10 });
    expect(getStepProgress("courseSelect", true)).toEqual({ current: 2, total: 10 });
    expect(getStepProgress("trimmingCourseSelect", true)).toEqual({ current: 3, total: 10 });
    expect(getStepProgress("trimmingOptionSelect", true)).toEqual({ current: 4, total: 10 });
    // SD-16の回帰ポイント: トリミングオプション選択(4/10)の直後がスタッフ選択(5/10)になり、
    // current・total とも後退しない
    expect(getStepProgress("staffSelect", true)).toEqual({ current: 5, total: 10 });
    expect(getStepProgress("dateSelect", true)).toEqual({ current: 6, total: 10 });
    expect(getStepProgress("timeSelect", true)).toEqual({ current: 7, total: 10 });
    expect(getStepProgress("request", true)).toEqual({ current: 8, total: 10 });
    expect(getStepProgress("confirm", true)).toEqual({ current: 9, total: 10 });
  });

  it("トリミング専用ステップを通常フローで指定すると例外を投げる", () => {
    expect(() => getStepProgress("trimmingCourseSelect", false)).toThrow();
    expect(() => getStepProgress("trimmingOptionSelect", false)).toThrow();
  });
});
