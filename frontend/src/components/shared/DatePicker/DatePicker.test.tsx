import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { DatePicker } from "./DatePicker";
import { CalendarNav, ClearButton, MonthGrid, YearNav } from "./DatePickerParts";

describe("DatePicker — 44px touch targets", () => {
  it("CalendarNavの前月・タイトル・次月を44x44px以上に保つ", () => {
    render(
      <CalendarNav
        displayMonth={new Date(2026, 3, 1)}
        onPrev={() => {}}
        onNext={() => {}}
        onTitleClick={() => {}}
      />,
    );

    const prev = screen.getByRole("button", { name: "前の月" });
    const title = screen.getByRole("button", { name: "2026年 4月" });
    const next = screen.getByRole("button", { name: "次の月" });

    for (const button of [prev, title, next]) {
      expect(button).toHaveClass("min-h-11", "min-w-11");
    }
    expect(prev.querySelector("svg")).toHaveClass("size-5");
    expect(title).toHaveClass("text-sm");
    expect(next.querySelector("svg")).toHaveClass("size-5");
  });

  it("YearNavの前年・次年を44x44px以上に保つ", () => {
    render(<YearNav year={2026} onPrevYear={() => {}} onNextYear={() => {}} />);

    const prev = screen.getByRole("button", { name: "前の年" });
    const next = screen.getByRole("button", { name: "次の年" });

    for (const button of [prev, next]) {
      expect(button).toHaveClass("min-h-11", "min-w-11");
      expect(button.querySelector("svg")).toHaveClass("size-5");
    }
  });

  it("MonthGridの各月を44x44px以上にし、文字サイズを維持する", () => {
    render(<MonthGrid currentMonth={3} onSelect={() => {}} />);

    const monthButtons = screen.getAllByRole("button");
    expect(monthButtons).toHaveLength(12);
    for (const button of monthButtons) {
      expect(button).toHaveClass("min-h-11", "min-w-11", "text-sm");
    }
  });

  it("ClearButtonを44x44px以上にし、glyphサイズを維持する", () => {
    render(<ClearButton onClick={() => {}} />);

    const clear = screen.getByRole("button", { name: "日付をクリア" });
    expect(clear).toHaveClass("min-h-11", "min-w-11", "-my-px");
    expect(clear.querySelector("svg")).toHaveClass("size-5");
  });

  it("single modeのcalendar triggerを44x44px以上に保つ", () => {
    render(<DatePicker value="" onChange={() => {}} />);

    expect(screen.getByRole("button", { name: "カレンダーを開く" })).toHaveClass(
      "min-h-11",
      "min-w-11",
      "-my-px",
    );
    const input = screen.getByPlaceholderText("日付を選択…");
    expect(input).toHaveClass("min-h-11");
    expect(input.parentElement).toHaveClass("focus-within:ring-1", "focus-within:ring-ring");
  });

  it("single modeのToday buttonを44x44px以上に保つ", async () => {
    const user = userEvent.setup();
    render(<DatePicker value="" onChange={() => {}} />);

    await user.click(screen.getByRole("button", { name: "カレンダーを開く" }));

    expect(screen.getByRole("button", { name: "Today" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
  });

  it("flex内で複数配置しても親幅を押し広げない", () => {
    render(<DatePicker value="" onChange={() => {}} placeholder="開始日" />);

    expect(screen.getByPlaceholderText("開始日").parentElement).toHaveClass("min-w-0");
  });

  it("range modeのpopover triggerとclear操作を別buttonにしてinteractive要素をネストしない", () => {
    const { container } = render(
      <DatePicker mode="range" value="2026-04-01~2026-04-30" onChange={() => {}} />,
    );

    expect(screen.getByRole("button", { name: /2026\/4\/1.*2026\/4\/30/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "日付をクリア" })).toBeInTheDocument();
    expect(container.querySelector("button button")).toBeNull();
  });
});

/**
 * Issue #48 回帰テスト
 *
 * 飼主・ペット登録フォームで使用される日付選択 UI は、月によって
 * グリッドの行数（4〜6 週）が変わると Popover 全体の高さが変動し、
 * 上部の前月/翌月ナビゲーション (◀ ▶) の位置が見かけ上ずれてしまう。
 *
 * `react-day-picker` の `fixedWeeks` を有効化することで、
 * 常に 6 週固定（42 セル）でレンダリングされることを保証する。
 *
 * NOTE: `role="gridcell"` のセル数を検証することで、
 * fixedWeeks プロパティが実際に反映されているかを確認する。
 */
describe("DatePicker — fixedWeeks layout (Issue #48)", () => {
  it.each([
    ["2026-02-15", "うるう年でない 2 月（28 日 / 通常 4 週）"],
    ["2026-04-15", "30 日（通常 5 週）"],
    ["2026-08-15", "31 日 + 月初土曜（通常 6 週）"],
    ["2027-01-15", "31 日 + 月初金曜（通常 5 週）"],
  ])(
    "value=%s (%s) を開いたとき、常に 42 個の日付セル（6 週 × 7 曜日）を表示する",
    async (value) => {
      const user = userEvent.setup();
      render(<DatePicker value={value} onChange={() => {}} />);

      await user.click(screen.getByLabelText("カレンダーを開く"));

      const cells = screen.getAllByRole("gridcell");
      expect(cells).toHaveLength(42);
    },
  );

  it("前月/翌月ナビゲーションで月を切り替えても 42 セルを維持する", async () => {
    const user = userEvent.setup();
    render(<DatePicker value="2026-04-15" onChange={() => {}} />);

    await user.click(screen.getByLabelText("カレンダーを開く"));
    expect(screen.getAllByRole("gridcell")).toHaveLength(42);

    await user.click(screen.getByLabelText("前の月"));
    expect(screen.getAllByRole("gridcell")).toHaveLength(42);

    await user.click(screen.getByLabelText("次の月"));
    await user.click(screen.getByLabelText("次の月"));
    expect(screen.getAllByRole("gridcell")).toHaveLength(42);
  });

  it("range mode でも 2 ヶ月分とも fixedWeeks が適用される（合計 84 セル）", async () => {
    const user = userEvent.setup();
    render(<DatePicker mode="range" value="" onChange={() => {}} />);

    await user.click(screen.getByRole("button", { name: /期間を選択/ }));

    const cells = screen.getAllByRole("gridcell");
    expect(cells).toHaveLength(84);
  });
});
