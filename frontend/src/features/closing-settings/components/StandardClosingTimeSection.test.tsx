import { describe, it, expect, vi, beforeEach } from "vitest";
import { act, render, screen, fireEvent, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { StandardClosingTimeSection } from "./StandardClosingTimeSection";
import type { ClinicSettings } from "@/types/generated/models";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const { mockUpdateClosingSettings, mockToastError } = vi.hoisted(() => ({
  mockUpdateClosingSettings: vi.fn(),
  mockToastError: vi.fn(),
}));

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: mockToastError } }));

vi.mock("../api/update-closing-settings", () => ({
  updateClosingSettings: (...args: unknown[]) => mockUpdateClosingSettings(...args),
}));

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
    closed_weekdays: [],
    ...overrides,
  } as ClinicSettings;
}

describe("StandardClosingTimeSection (#151 時間帯プレビュー・#215 越日対応)", () => {
  beforeEach(() => {
    mockUpdateClosingSettings.mockReset().mockResolvedValue(undefined);
    mockToastError.mockClear();
  });

  it("Sunday closure hides ranges immediately but retains the configured submitted end time", async () => {
    render(
      <StandardClosingTimeSection settings={makeSettings({ closed_weekdays: [0] })} canEdit />,
    );
    expect(screen.getAllByRole("cell", { name: "休診" })).toHaveLength(3);
    const end = screen.getByLabelText("日曜 終了時間");
    expect(end).toHaveValue("17:30");
    expect(end).toHaveAttribute("aria-describedby", "sunday-closed-guidance");
    expect(
      screen.getByText("日曜は休診日のため、この終了時間は適用されません"),
    ).toBeInTheDocument();
    fireEvent.invalid(end);
    expect(end.getAttribute("aria-describedby")).toContain("closing-sunday-end-error");
    fireEvent.change(end, { target: { value: "17:30" } });
    await act(async () => end.closest("form")?.requestSubmit());
    await waitFor(() =>
      expect(mockUpdateClosingSettings).toHaveBeenCalledWith(
        expect.objectContaining({ closing_sunday_end: "17:30", closed_weekdays: [0] }),
      ),
    );
    fireEvent.click(screen.getByRole("checkbox", { name: "日" }));
    expect(screen.queryByRole("cell", { name: "休診" })).not.toBeInTheDocument();
    expect(screen.getByText("12:00:00～17:29:59")).toBeInTheDocument();
    expect(end).toHaveValue("17:30");
    fireEvent.click(screen.getByRole("checkbox", { name: "日" }));
    expect(screen.getAllByRole("cell", { name: "休診" })).toHaveLength(3);
  });

  it("休診曜日checkboxのfocusable targetを44px以上に保つ", () => {
    render(<StandardClosingTimeSection settings={makeSettings()} canEdit={true} />);

    const sunday = screen.getByRole("checkbox", { name: "日" });
    expect(sunday).toHaveClass("absolute", "inset-0", "size-full");
    expect(sunday.closest("label")).toHaveClass("relative", "min-h-11", "min-w-11");
  });

  it("keeps focus on the same weekday checkbox after its controlled value changes", async () => {
    const user = userEvent.setup();
    render(<StandardClosingTimeSection settings={makeSettings()} canEdit={true} />);

    const monday = screen.getByRole("checkbox", { name: "月" });
    monday.focus();
    expect(monday).toHaveFocus();

    await user.click(monday);
    const updatedMonday = screen.getByRole("checkbox", { name: "月" });
    expect(updatedMonday).toBe(monday);
    expect(updatedMonday).toBeChecked();
    expect(updatedMonday).toHaveFocus();
  });

  it("toggles with Space, activates the next weekday label, and submits checked weekdays", async () => {
    const user = userEvent.setup();
    render(<StandardClosingTimeSection settings={makeSettings()} canEdit={true} />);

    const monday = screen.getByRole("checkbox", { name: "月" });
    const form = monday.closest("form");
    expect(form).not.toBeNull();
    monday.focus();
    await user.tab({ shift: true });
    expect(monday).not.toHaveFocus();
    await user.tab();
    expect(monday).toHaveFocus();
    await user.keyboard(" ");

    const updatedMonday = screen.getByRole("checkbox", { name: "月" });
    expect(updatedMonday).toBe(monday);
    expect(updatedMonday).toBeChecked();
    expect(updatedMonday).toHaveFocus();

    await user.click(screen.getByText("火"));
    expect(screen.getByRole("checkbox", { name: "火" })).toBeChecked();

    await act(async () => form?.requestSubmit());

    await waitFor(() =>
      expect(mockUpdateClosingSettings).toHaveBeenCalledWith(
        expect.objectContaining({ closed_weekdays: [1, 2] }),
      ),
    );
  });

  it("初期設定から平日・日曜の AM/PM/EMG レンジを表示する", () => {
    render(<StandardClosingTimeSection settings={makeSettings()} canEdit={true} />);

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
    render(<StandardClosingTimeSection settings={makeSettings()} canEdit={true} />);

    // 単一の時刻ではなく開始〜終了レンジが存在することを固定（回帰防止）。
    // 同日内: AM(2) + PM(2) = 4 セル、越日: EMG(2) = 2 セル（#215）。
    const sameDayRanges = screen.getAllByText(/^\d{2}:\d{2}:\d{2}～\d{2}:\d{2}:\d{2}$/);
    expect(sameDayRanges).toHaveLength(4);
    const crossDayRanges = screen.getAllByText(/^\d{2}:\d{2}:\d{2}～翌\d{2}:\d{2}:\d{2}$/);
    expect(crossDayRanges).toHaveLength(2);
  });

  it("境界時刻の編集でレンジが即時に再計算される（ライブプレビュー）", () => {
    render(<StandardClosingTimeSection settings={makeSettings()} canEdit={true} />);

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
    render(<StandardClosingTimeSection settings={makeSettings()} canEdit={true} />);

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

  it("canEdit=false のとき formAction は mutate せず toast.error する", async () => {
    const { rerender } = render(
      <StandardClosingTimeSection settings={makeSettings()} canEdit={true} />,
    );
    const form = screen.getByLabelText("午前・午後 区切り時間").closest("form");
    expect(form).not.toBeNull();

    rerender(<StandardClosingTimeSection settings={makeSettings()} canEdit={false} />);

    await act(async () => {
      form?.requestSubmit();
    });

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(mockUpdateClosingSettings).not.toHaveBeenCalled();
  });

  // EMR-319: 必須時刻が空で invalid になると赤枠+エラー文+aria-invalid が出る。
  describe("必須エラー表示 (EMR-319)", () => {
    it("各必須時刻が invalid になると aria-invalid・赤枠・エラー文が出る", () => {
      render(<StandardClosingTimeSection settings={makeSettings()} canEdit={true} />);

      const cases = [
        {
          label: "午前・午後 区切り時間",
          errorId: "closing-am-pm-boundary-error",
          message: "午前・午後 区切り時間を入力してください",
        },
        {
          label: "平日 終了時間",
          errorId: "closing-weekday-end-error",
          message: "平日 終了時間を入力してください",
        },
        {
          label: "日曜 終了時間",
          errorId: "closing-sunday-end-error",
          message: "日曜 終了時間を入力してください",
        },
      ];

      for (const { label } of cases) {
        fireEvent.invalid(screen.getByLabelText(label));
      }

      for (const { label, errorId, message } of cases) {
        const input = screen.getByLabelText(label);
        expect(input).toHaveAttribute("aria-invalid", "true");
        expect(input).toHaveAttribute("aria-describedby", errorId);
        expect(input.className).toContain("ring-2");
        expect(document.getElementById(errorId)).toHaveTextContent(message);
      }
    });

    it("平日 終了時間だけ invalid なら他項目は通常表示のまま、入力で解除される", () => {
      render(<StandardClosingTimeSection settings={makeSettings()} canEdit={true} />);

      fireEvent.invalid(screen.getByLabelText("平日 終了時間"));

      expect(screen.getByLabelText("平日 終了時間")).toHaveAttribute("aria-invalid", "true");
      expect(screen.getByLabelText("午前・午後 区切り時間")).not.toHaveAttribute("aria-invalid");
      expect(screen.getByLabelText("日曜 終了時間")).not.toHaveAttribute("aria-invalid");

      fireEvent.change(screen.getByLabelText("平日 終了時間"), {
        target: { value: "19:00" },
      });

      expect(screen.getByLabelText("平日 終了時間")).not.toHaveAttribute("aria-invalid");
      expect(document.getElementById("closing-weekday-end-error")).toBeNull();
    });
  });
});
