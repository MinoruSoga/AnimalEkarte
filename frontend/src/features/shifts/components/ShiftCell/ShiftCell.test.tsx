import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { Shift } from "../../types";
import { ShiftCell } from "./ShiftCell";

const OFF_SHIFT: Shift = {
  id: "shift-1",
  clinic_id: "clinic-1",
  staff_id: "staff-1",
  staff_name: "スタッフA",
  date: "2026-07-01",
  shift_type: "off",
  start_time: "",
  end_time: "",
  notes: "",
  breaks: [],
  created_at: "",
  updated_at: "",
};

const TIMED_SHIFT: Shift = {
  ...OFF_SHIFT,
  id: "shift-2",
  shift_type: "morning",
  start_time: "09:00",
  end_time: "13:00",
};

const defaultProps = {
  staffId: "staff-1",
  staffName: "スタッフA",
  dateStr: "2026-07-01",
  onAddShift: vi.fn(),
  onEditShift: vi.fn(),
};

describe("ShiftCell — P9 hit target / accessible name", () => {
  it("空セルの追加ボタンは44px以上でスタッフと日付を名前に含める", () => {
    render(
      <ShiftCell
        {...defaultProps}
        shift={undefined}
        canCreate
        canEdit
      />,
    );

    expect(
      screen.getByRole("button", {
        name: "スタッフA 2026-07-01 シフトを追加",
      }),
    ).toHaveClass("min-h-11");
  });

  it("編集可能な休日ボタンは44px以上でスタッフと日付を名前に含める", () => {
    render(
      <ShiftCell
        {...defaultProps}
        shift={OFF_SHIFT}
        canCreate
        canEdit
      />,
    );

    expect(
      screen.getByRole("button", {
        name: "スタッフA 2026-07-01 休日シフト（時刻なし）を編集",
      }),
    ).toHaveClass("min-h-11");
  });

  it("時間付きシフトの編集ボタンは勤務時刻範囲をaccessible nameに含める", () => {
    render(
      <ShiftCell
        {...defaultProps}
        shift={TIMED_SHIFT}
        canCreate
        canEdit
      />,
    );

    expect(
      screen.getByRole("button", {
        name: "スタッフA 2026-07-01 午前シフト（09:00〜13:00）を編集",
      }),
    ).toHaveClass("min-h-11");
  });
});
