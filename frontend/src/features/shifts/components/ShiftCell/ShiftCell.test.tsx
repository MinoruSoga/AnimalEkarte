import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import type { Shift } from "../../types";
import { ShiftCell } from "./ShiftCell";

// EMR-514 固定API契約の応答DTOをそのまま使う（IDは正十進文字列、時刻は明示date-time）
const OFF_SHIFT: Shift = {
  id: "shift-1",
  clinic_id: "clinic-1",
  staff_id: "staff-1",
  staff_name: "スタッフA",
  date: "2026-07-01",
  shift_type: "off",
  starts_at: null,
  ends_at: null,
  version: 1,
  interval_state: "known",
  notes: "",
  breaks: [],
  created_at: "2026-07-01T00:00:00+09:00",
  updated_at: "2026-07-01T00:00:00+09:00",
};

const TIMED_SHIFT: Shift = {
  ...OFF_SHIFT,
  id: "shift-2",
  shift_type: "morning",
  starts_at: "2026-07-01T09:00:00+09:00",
  ends_at: "2026-07-01T13:00:00+09:00",
};

const UNKNOWN_SHIFT: Shift = {
  ...OFF_SHIFT,
  id: "shift-4",
  shift_type: "morning",
  interval_state: "unknown",
  starts_at: null,
  ends_at: null,
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
    render(<ShiftCell {...defaultProps} shift={undefined} canCreate canEdit />);

    const addButton = screen.getByRole("button", {
      name: "スタッフA 2026-07-01 シフトを追加",
    });
    expect(addButton).toHaveClass("min-h-11");
    expect(addButton).toHaveTextContent("+");
  });

  it("編集可能な休日ボタンは44px以上でスタッフと日付を名前に含める", () => {
    render(<ShiftCell {...defaultProps} shift={OFF_SHIFT} canCreate canEdit />);

    expect(
      screen.getByRole("button", {
        name: "スタッフA 2026-07-01 休日シフト（終日）を編集",
      }),
    ).toHaveClass("min-h-11");
  });

  it("時間付きシフトの編集ボタンは勤務時刻範囲をaccessible nameに含める", () => {
    render(<ShiftCell {...defaultProps} shift={TIMED_SHIFT} canCreate canEdit />);

    expect(
      screen.getByRole("button", {
        name: "スタッフA 2026-07-01 午前シフト（09:00〜13:00）を編集",
      }),
    ).toHaveClass("min-h-11");
  });
});

describe("ShiftCell — BUG-022 filled cell exclusivity / HH:mm / overflow", () => {
  it("時間付きシフトでは追加ボタンを出さず編集ボタンのみ表示する", () => {
    render(<ShiftCell {...defaultProps} shift={TIMED_SHIFT} canCreate canEdit />);

    expect(screen.queryByRole("button", { name: /シフトを追加/ })).not.toBeInTheDocument();
    expect(
      screen.getByRole("button", {
        name: "スタッフA 2026-07-01 午前シフト（09:00〜13:00）を編集",
      }),
    ).toBeInTheDocument();
  });

  it("明示date-timeはJSTのHH:mmで表示しaccessible nameもHH:mmにする", () => {
    render(<ShiftCell {...defaultProps} shift={TIMED_SHIFT} canCreate canEdit />);

    const editButton = screen.getByRole("button", {
      name: "スタッフA 2026-07-01 午前シフト（09:00〜13:00）を編集",
    });
    expect(editButton).toHaveTextContent("09:00");
    expect(editButton).toHaveTextContent("13:00");
    expect(editButton.textContent).not.toContain("+09:00");
  });

  it("編集チップはoverflow-hiddenで隣接セルへのはみ出しを防ぐ", () => {
    render(<ShiftCell {...defaultProps} shift={TIMED_SHIFT} canCreate canEdit />);

    expect(
      screen.getByRole("button", {
        name: "スタッフA 2026-07-01 午前シフト（09:00〜13:00）を編集",
      }),
    ).toHaveClass("overflow-hidden");
  });
});

describe("ShiftCell — EMR-514 interval_state=unknown", () => {
  it("unknown勤務は時刻を補完せず「時刻未確認」と明示する", () => {
    render(<ShiftCell {...defaultProps} shift={UNKNOWN_SHIFT} canCreate canEdit />);

    const editButton = screen.getByRole("button", {
      name: "スタッフA 2026-07-01 午前シフト（時刻未確認）を編集",
    });
    expect(editButton).toHaveTextContent("時刻未確認");
    // 09:00等の推測表示を行わない
    expect(editButton.textContent).not.toContain("09:00");
  });

  it("unknown勤務は未登録セル（+ボタン）と混同しない", () => {
    render(<ShiftCell {...defaultProps} shift={UNKNOWN_SHIFT} canCreate canEdit />);

    expect(screen.queryByRole("button", { name: /シフトを追加/ })).not.toBeInTheDocument();
  });
});
