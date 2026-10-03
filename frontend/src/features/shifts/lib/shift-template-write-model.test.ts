import { describe, expect, it } from "vitest";

import type { TemplateFormData } from "./shift-template-form-model";
import {
  toShiftTemplateCreateInput,
  toShiftTemplateUpdateInput,
} from "./shift-template-write-model";

function form(overrides: Partial<TemplateFormData> = {}): TemplateFormData {
  return {
    name: "テストテンプレート",
    shift_type: "full",
    start_time: "09:00",
    end_time: "18:00",
    notes: "",
    is_active: true,
    breaks: [],
    ...overrides,
  };
}

// EMR-241: カテゴリ名で時刻・休憩を捨てない。空欄は空のままサーバへ送り、
// 必須判定はバックエンド（sharedkernel.RequiresTimeSlot）に委ねる。
describe("toShiftTemplateCreateInput", () => {
  it("off + 両方空の時刻を省略して送信できる（空欄で保存可能）", () => {
    const input = toShiftTemplateCreateInput(
      form({ shift_type: "off", start_time: "", end_time: "" }),
    );

    expect(input).toMatchObject({ shift_type: "off" });
    expect(input.start_time).toBeUndefined();
    expect(input.end_time).toBeUndefined();
  });

  it("paid_leave + 両方空の時刻を省略して送信できる", () => {
    const input = toShiftTemplateCreateInput(
      form({ shift_type: "paid_leave", start_time: "", end_time: "" }),
    );

    expect(input.start_time).toBeUndefined();
    expect(input.end_time).toBeUndefined();
  });

  it("off/paid_leave に入力された時刻を捨てずに送信する", () => {
    const input = toShiftTemplateCreateInput(
      form({ shift_type: "off", start_time: "10:00", end_time: "15:00" }),
    );

    expect(input.start_time).toBe("10:00");
    expect(input.end_time).toBe("15:00");
  });

  it("off の休憩をカテゴリ名でクリアせずそのまま送信する", () => {
    const input = toShiftTemplateCreateInput(
      form({
        shift_type: "paid_leave",
        start_time: "",
        end_time: "",
        breaks: [{ break_start: "12:00", break_end: "13:00" }],
      }),
    );

    expect(input.breaks).toEqual([{ break_start: "12:00", break_end: "13:00" }]);
  });

  it("勤務種別で空時刻は捏造せず undefined のまま送る（サーバ検証に委ねる）", () => {
    const input = toShiftTemplateCreateInput(
      form({ shift_type: "full", start_time: "", end_time: "" }),
    );

    expect(input.start_time).toBeUndefined();
    expect(input.end_time).toBeUndefined();
  });
});

describe("toShiftTemplateUpdateInput", () => {
  it('off + 両方空の時刻を "" で送りクリアさせる（PATCH の未指定=維持と区別）', () => {
    const input = toShiftTemplateUpdateInput(
      form({ shift_type: "off", start_time: "", end_time: "" }),
    );

    expect(input.start_time).toBe("");
    expect(input.end_time).toBe("");
  });

  it("off/paid_leave に入力された時刻を捨てずに送信する", () => {
    const input = toShiftTemplateUpdateInput(
      form({ shift_type: "paid_leave", start_time: "10:00", end_time: "15:00" }),
    );

    expect(input.start_time).toBe("10:00");
    expect(input.end_time).toBe("15:00");
  });

  it("off の休憩をカテゴリ名でクリアせずそのまま送信する", () => {
    const input = toShiftTemplateUpdateInput(
      form({
        shift_type: "off",
        start_time: "",
        end_time: "",
        breaks: [{ break_start: "12:00", break_end: "13:00" }],
      }),
    );

    expect(input.breaks).toEqual([{ break_start: "12:00", break_end: "13:00" }]);
  });
});
