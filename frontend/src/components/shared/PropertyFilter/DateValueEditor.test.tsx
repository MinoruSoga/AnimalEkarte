import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { DateValueEditor } from "./DateValueEditor";

type OnApply = (value: { from?: string; to?: string }, displayValue: string) => void;

// RDP の日ボタンは aria-label "2026年10月20日火曜日"（今日は "Today, " 接頭辞）なので日付部分で一致させる
function dayButton(isoDay: string) {
  const [year, month, day] = isoDay.split("-").map(Number);
  return screen.getByRole("button", { name: new RegExp(`${year}年${month}月${day}日`) });
}

describe("DateValueEditor", () => {
  it("適用済みレンジで開くと終了日が編集対象になり、日クリックで終了日だけ更新される", async () => {
    const onApply = vi.fn<OnApply>();
    render(
      <DateValueEditor currentValue={{ from: "2026-10-01", to: "2026-10-05" }} onApply={onApply} />,
    );
    const user = userEvent.setup();

    expect(screen.getByRole("button", { name: "終了日 2026/10/5" })).toHaveAttribute(
      "aria-pressed",
      "true",
    );

    await user.click(dayButton("2026-10-20"));

    expect(onApply).toHaveBeenCalledTimes(1);
    expect(onApply).toHaveBeenLastCalledWith(
      { from: "2026-10-01", to: "2026-10-20" },
      "2026/10/1〜2026/10/20",
    );
  });

  it("開始日を切り替えて選び直せる。逆転する既存終了日は破棄され次の終了日クリックで確定する", async () => {
    const onApply = vi.fn<OnApply>();
    render(
      <DateValueEditor currentValue={{ from: "2026-10-01", to: "2026-10-05" }} onApply={onApply} />,
    );
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /開始日/ }));
    await user.click(dayButton("2026-10-10"));

    expect(onApply).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "終了日" })).toHaveAttribute("aria-pressed", "true");

    await user.click(dayButton("2026-10-20"));

    expect(onApply).toHaveBeenLastCalledWith(
      { from: "2026-10-10", to: "2026-10-20" },
      "2026/10/10〜2026/10/20",
    );
  });

  it("開始日の選び直しで既存終了日が逆転しなければ保持して即適用する", async () => {
    const onApply = vi.fn<OnApply>();
    render(
      <DateValueEditor currentValue={{ from: "2026-10-01", to: "2026-10-20" }} onApply={onApply} />,
    );
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /開始日/ }));
    await user.click(dayButton("2026-10-10"));

    expect(onApply).toHaveBeenLastCalledWith(
      { from: "2026-10-10", to: "2026-10-20" },
      "2026/10/10〜2026/10/20",
    );
  });

  it("同一日を2回選ぶと単日レンジとして適用される", async () => {
    const onApply = vi.fn<OnApply>();
    render(
      <DateValueEditor currentValue={{ from: "2026-10-01", to: "2026-10-05" }} onApply={onApply} />,
    );
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: /開始日/ }));
    await user.click(dayButton("2026-10-10"));
    await user.click(dayButton("2026-10-10"));

    expect(onApply).toHaveBeenLastCalledWith(
      { from: "2026-10-10", to: "2026-10-10" },
      "2026/10/10",
    );
  });

  it("終了日編集中に開始日より前の日を選ぶと端点を入れ替えて適用する", async () => {
    const onApply = vi.fn<OnApply>();
    render(
      <DateValueEditor currentValue={{ from: "2026-10-10", to: "2026-10-20" }} onApply={onApply} />,
    );
    const user = userEvent.setup();

    await user.click(dayButton("2026-10-08"));

    expect(onApply).toHaveBeenLastCalledWith(
      { from: "2026-10-08", to: "2026-10-10" },
      "2026/10/8〜2026/10/10",
    );
  });

  it("プリセットクリックはラベル表示で適用する", async () => {
    const onApply = vi.fn<OnApply>();
    render(<DateValueEditor onApply={onApply} />);
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "今月" }));

    expect(onApply).toHaveBeenCalledTimes(1);
    const [value, displayValue] = onApply.mock.calls[0];
    expect(value.from).toMatch(/^\d{4}-\d{2}-01$/);
    expect(value.to).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(displayValue).toBe("今月");
  });

  it("タイトルクリックで月グリッドに切り替わり、年ナビと月選択で表示月を移動する", async () => {
    render(
      <DateValueEditor currentValue={{ from: "2026-10-01", to: "2026-10-05" }} onApply={vi.fn()} />,
    );
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "2026年 10月" }));
    expect(screen.getByRole("button", { name: "前の年" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "前の年" }));
    await user.click(screen.getByRole("button", { name: "3月" }));

    expect(screen.getByRole("button", { name: "2025年 3月" })).toBeInTheDocument();
  });
});
