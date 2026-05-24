import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { NotionDatePicker } from "./NotionDatePicker";

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
describe("NotionDatePicker — fixedWeeks layout (Issue #48)", () => {
  it.each([
    ["2026-02-15", "うるう年でない 2 月（28 日 / 通常 4 週）"],
    ["2026-04-15", "30 日（通常 5 週）"],
    ["2026-08-15", "31 日 + 月初土曜（通常 6 週）"],
    ["2027-01-15", "31 日 + 月初金曜（通常 5 週）"],
  ])(
    "value=%s (%s) を開いたとき、常に 42 個の日付セル（6 週 × 7 曜日）を表示する",
    async (value) => {
      const user = userEvent.setup();
      render(<NotionDatePicker value={value} onChange={() => {}} />);

      await user.click(screen.getByLabelText("カレンダーを開く"));

      const cells = screen.getAllByRole("gridcell");
      expect(cells).toHaveLength(42);
    },
  );

  it("前月/翌月ナビゲーションで月を切り替えても 42 セルを維持する", async () => {
    const user = userEvent.setup();
    render(<NotionDatePicker value="2026-04-15" onChange={() => {}} />);

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
    render(<NotionDatePicker mode="range" value="" onChange={() => {}} />);

    await user.click(screen.getByRole("button", { name: /期間を選択/ }));

    const cells = screen.getAllByRole("gridcell");
    expect(cells).toHaveLength(84);
  });
});
