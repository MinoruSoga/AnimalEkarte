import { describe, it, expect } from "vitest";
import { render, screen, fireEvent, within } from "@testing-library/react";

import { StandardClosingTimeSection } from "./StandardClosingTimeSection";
import type { ClinicSettings } from "@/types/generated/models";

// このコンポーネントは closing_* と closed_weekdays のみ参照する。
// ClinicSettings は CPM 閾値など多数のフィールドを持つため、関連フィールドのみの
// 最小 fixture をキャストして使う（無関係フィールドはテスト対象外）。
function makeSettings(overrides: Partial<ClinicSettings> = {}): ClinicSettings {
  return {
    clinic_id: 1,
    closing_am_pm_boundary: "12:00:00",
    closing_weekday_end: "18:30:00",
    closing_sunday_end: "17:30:00",
    closing_am_start: "09:00:00", // #215
    closed_weekdays: [0],
    ...overrides,
  } as ClinicSettings;
}

describe("StandardClosingTimeSection (#151 時間帯プレビュー・#215 越日対応)", () => {
  it("休診曜日checkboxのfocusable targetを44px以上に保つ", () => {
    render(<StandardClosingTimeSection settings={makeSettings()} />);

    const sunday = screen.getByRole("checkbox", { name: "日" });
    expect(sunday).toHaveClass("absolute", "inset-0", "size-full");
    expect(sunday.closest("label")).toHaveClass("relative", "min-h-11", "min-w-11");
  });

  it("初期設定から平日・日曜の AM/PM/EMG レンジを表示する", () => {
    render(<StandardClosingTimeSection settings={makeSettings()} />);

    // AM は am_start(9:00) と境界に依存し平日・日曜で共通 → 同一文字列が 2 セル（#215）
    expect(screen.getAllByText("09:00:00～11:59:59")).toHaveLength(2);

    // PM 行: 平日 18:30 / 日曜 17:30 に追従し、−1秒（:59）で表示
    const pmRow = screen.getByText("PM").closest("tr")!;
    expect(within(pmRow).getByText("12:00:00～18:29:59")).toBeInTheDocument();
    expect(within(pmRow).getByText("12:00:00～17:29:59")).toBeInTheDocument();

    // EMG 行: 翌 am_start−1秒 までの越日表記（#215）
    const emgRow = screen.getByText("EMG").closest("tr")!;
    expect(within(emgRow).getByText("18:30:00～翌08:59:59")).toBeInTheDocument();
    expect(within(emgRow).getByText("17:30:00～翌08:59:59")).toBeInTheDocument();
  });

  it("旧来の曖昧表示（区切り時刻のみ・レンジ無し）ではなく、明示的な開始〜終了レンジを表示する", () => {
    render(<StandardClosingTimeSection settings={makeSettings()} />);

    // 単一の時刻ではなく開始〜終了レンジが存在することを固定（回帰防止）。
    // 同日内: AM(2) + PM(2) = 4 セル、越日: EMG(2) = 2 セル（#215）。
    const sameDayRanges = screen.getAllByText(/^\d{2}:\d{2}:\d{2}～\d{2}:\d{2}:\d{2}$/);
    expect(sameDayRanges).toHaveLength(4);
    const crossDayRanges = screen.getAllByText(/^\d{2}:\d{2}:\d{2}～翌\d{2}:\d{2}:\d{2}$/);
    expect(crossDayRanges).toHaveLength(2);
  });

  it("境界時刻の編集でレンジが即時に再計算される（ライブプレビュー）", () => {
    render(<StandardClosingTimeSection settings={makeSettings()} />);

    expect(screen.getAllByText("09:00:00～11:59:59")).toHaveLength(2);

    // time input はセグメント型で userEvent.type が不安定なため、確実な fireEvent.change を使う。
    const boundaryInput = screen.getByLabelText("午前・午後 区切り時間");
    fireEvent.change(boundaryInput, { target: { value: "10:00" } });

    // AM 終了が即時に 09:59:59 へ更新され、旧レンジは消える
    expect(screen.getAllByText("09:00:00～09:59:59")).toHaveLength(2);
    expect(screen.queryByText("09:00:00～11:59:59")).not.toBeInTheDocument();

    // PM 開始も 10:00:00 に追従
    const pmRow = screen.getByText("PM").closest("tr")!;
    expect(within(pmRow).getByText("10:00:00～18:29:59")).toBeInTheDocument();
  });

  it("終了時刻を空にすると、誤解を招く部分レンジではなく『未設定』を表示する", () => {
    render(<StandardClosingTimeSection settings={makeSettings()} />);

    const weekdayInput = screen.getByLabelText("平日 終了時間");
    fireEvent.change(weekdayInput, { target: { value: "" } });

    // 平日列の PM / EMG は未設定にフォールバックする
    const pmRow = screen.getByText("PM").closest("tr")!;
    const emgRow = screen.getByText("EMG").closest("tr")!;
    expect(within(pmRow).getByText("未設定")).toBeInTheDocument();
    expect(within(emgRow).getByText("未設定")).toBeInTheDocument();

    // 日曜列は引き続きレンジを表示する（部分欠落で全体が壊れない）
    expect(within(pmRow).getByText("12:00:00～17:29:59")).toBeInTheDocument();
  });
});
