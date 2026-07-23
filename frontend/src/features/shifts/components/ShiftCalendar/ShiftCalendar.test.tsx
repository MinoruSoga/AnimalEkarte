import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { ShiftCalendar, type StaffItem } from "./ShiftCalendar";
import type { Shift } from "../../types";
import type { ClinicHoliday } from "../../api/clinic-holidays";

const STAFFS: StaffItem[] = [{ id: "s1", name: "スタッフA" }];

function renderCalendar(yearMonth: string, overrides: Partial<React.ComponentProps<typeof ShiftCalendar>> = {}) {
  return render(
    <ShiftCalendar
      yearMonth={yearMonth}
      shifts={overrides.shifts ?? ([] as Shift[])}
      staffs={overrides.staffs ?? STAFFS}
      holidays={overrides.holidays ?? ([] as ClinicHoliday[])}
      selectedStaffId={overrides.selectedStaffId ?? "all"}
      canCreate={overrides.canCreate ?? true}
      canEdit={overrides.canEdit ?? true}
      canDelete={overrides.canDelete ?? true}
      onPrevMonth={overrides.onPrevMonth ?? vi.fn()}
      onNextMonth={overrides.onNextMonth ?? vi.fn()}
      onStaffChange={overrides.onStaffChange ?? vi.fn()}
      onDateHeaderClick={overrides.onDateHeaderClick ?? vi.fn()}
    />,
  );
}

// 日付ヘッダーの aria-label（"YYYY-MM-DD の定休日設定"）は canCreate && onDateHeaderClick 時のみ
// 付与される。この個数を数えることで「当月の日数」生成ロジックを間接的に検証する。
function dateHeaderButtons() {
  return screen.getAllByRole("button", { name: /の定休日設定$/ });
}

describe("ShiftCalendar 月内日付生成", () => {
  it("31日ある月（2026年1月）で31日分のヘッダーを生成する", () => {
    renderCalendar("2026-01");
    expect(dateHeaderButtons()).toHaveLength(31);
    expect(screen.getByText("2026年1月")).toBeInTheDocument();
  });

  it("うるう年2月（2024年2月）で29日分のヘッダーを生成する", () => {
    renderCalendar("2024-02");
    expect(dateHeaderButtons()).toHaveLength(29);
  });

  it("平年2月（2023年2月）で28日分のヘッダーを生成する（うるう年と混同しない）", () => {
    renderCalendar("2023-02");
    expect(dateHeaderButtons()).toHaveLength(28);
  });

  it("30日の月（2026年4月）で30日分のヘッダーを生成する", () => {
    renderCalendar("2026-04");
    expect(dateHeaderButtons()).toHaveLength(30);
  });

  it("各日付ヘッダーに正しい曜日ラベルを表示する（2026-01-01は木曜日）", () => {
    renderCalendar("2026-01");
    const firstDayButton = screen.getByRole("button", { name: "2026-01-01 の定休日設定" });
    expect(firstDayButton).toHaveTextContent("木");
  });

  it("うるう日（2024-02-29）も木曜日として表示する", () => {
    renderCalendar("2024-02");
    const leapDayButton = screen.getByRole("button", { name: "2024-02-29 の定休日設定" });
    expect(leapDayButton).toHaveTextContent("木");
  });
});

describe("ShiftCalendar 定休日マーカー", () => {
  it("holidays に含まれる日付のみ定休日マーカーを表示する", () => {
    renderCalendar("2026-01", {
      holidays: [
        { id: 1, clinic_id: 1, date: "2026-01-05", reason: "", created_at: "", updated_at: "" },
      ],
    });

    expect(screen.getAllByLabelText("定休日")).toHaveLength(1);
  });

  it("holidays が空なら定休日マーカーを表示しない", () => {
    renderCalendar("2026-01", { holidays: [] });
    expect(screen.queryByLabelText("定休日")).not.toBeInTheDocument();
  });
});

describe("ShiftCalendar staffs 表示", () => {
  it("スタッフfilterに操作内容を表すaccessible nameがある", () => {
    renderCalendar("2026-01");

    expect(screen.getByRole("combobox", { name: "スタッフ絞り込み" })).toBeInTheDocument();
  });

  it("staffs が空のとき空状態メッセージを表示する", () => {
    renderCalendar("2026-01", { staffs: [] });
    expect(screen.getByText("スタッフが見つかりません")).toBeInTheDocument();
  });

  it("selectedStaffId 指定時は該当スタッフのみ表示する", () => {
    const staffs: StaffItem[] = [
      { id: "s1", name: "スタッフA" },
      { id: "s2", name: "スタッフB" },
    ];
    renderCalendar("2026-01", { staffs, selectedStaffId: "s2" });

    // "スタッフB" は選択済みフィルタ(SearchableSelect のトリガー表示)と行ラベルの
    // 2箇所に出現しうるため、行としての存在は getAllByText の件数で確認する。
    expect(screen.queryByText("スタッフA")).not.toBeInTheDocument();
    expect(screen.getAllByText("スタッフB").length).toBeGreaterThan(0);
  });
});
