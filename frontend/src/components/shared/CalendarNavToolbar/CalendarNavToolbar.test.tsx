import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CalendarNavToolbar } from "./CalendarNavToolbar";

describe("CalendarNavToolbar", () => {
  it.each(["sm", "lg", "auto"] as const)(
    "%s サイズでも全ボタンが44px以上のタッチターゲットを持つ",
    (size) => {
      render(
        <CalendarNavToolbar
          size={size}
          label={<span>2026年7月</span>}
          onPrev={() => {}}
          onToday={() => {}}
          onNext={() => {}}
          prevAriaLabel="前の月"
          nextAriaLabel="次の月"
        />,
      );

      expect(screen.getByRole("button", { name: "前の月" })).toHaveClass("min-h-11", "min-w-11");
      expect(screen.getByRole("button", { name: "今日" })).toHaveClass("h-11");
      expect(screen.getByRole("button", { name: "次の月" })).toHaveClass("min-h-11", "min-w-11");
    },
  );

  it("clustered レイアウトでは today ボタンが表示され、クリックで onToday が呼ばれる", async () => {
    const user = userEvent.setup();
    const onToday = vi.fn();
    render(
      <CalendarNavToolbar
        label={<h2>2026年 7月</h2>}
        onPrev={() => {}}
        onNext={() => {}}
        onToday={onToday}
        prevAriaLabel="前の月"
        nextAriaLabel="次の月"
      />,
    );

    expect(screen.getByText("2026年 7月")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "今日" }));
    expect(onToday).toHaveBeenCalledTimes(1);
  });

  it("onToday を渡さない場合 today ボタンは表示されない", () => {
    render(
      <CalendarNavToolbar
        layout="inline"
        label={<span>2026年7月</span>}
        onPrev={() => {}}
        onNext={() => {}}
        prevAriaLabel="前の月"
        nextAriaLabel="次の月"
      />,
    );
    expect(screen.queryByRole("button", { name: "今日" })).not.toBeInTheDocument();
  });

  it("prev/next クリックでそれぞれ onPrev/onNext が呼ばれる", async () => {
    const user = userEvent.setup();
    const onPrev = vi.fn();
    const onNext = vi.fn();
    render(
      <CalendarNavToolbar
        layout="spread"
        label={<span>2026-07-04</span>}
        onPrev={onPrev}
        onNext={onNext}
        prevAriaLabel="前日"
        nextAriaLabel="翌日"
      />,
    );

    await user.click(screen.getByRole("button", { name: "前日" }));
    await user.click(screen.getByRole("button", { name: "翌日" }));

    expect(onPrev).toHaveBeenCalledTimes(1);
    expect(onNext).toHaveBeenCalledTimes(1);
  });

  it("prevDisabled/nextDisabled を渡すとボタンが disabled になる", () => {
    render(
      <CalendarNavToolbar
        layout="spread"
        label={<span>label</span>}
        onPrev={() => {}}
        onNext={() => {}}
        prevAriaLabel="前日"
        nextAriaLabel="翌日"
        prevDisabled
      />,
    );

    expect(screen.getByRole("button", { name: "前日" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "翌日" })).not.toBeDisabled();
  });

  it("spread レイアウトでは prev - label - next の順に並ぶ", () => {
    render(
      <CalendarNavToolbar
        layout="spread"
        label={<span data-testid="label">中央ラベル</span>}
        onPrev={() => {}}
        onNext={() => {}}
        prevAriaLabel="前へ"
        nextAriaLabel="次へ"
      />,
    );

    const container = screen.getByTestId("label").parentElement;
    const children = container ? Array.from(container.children) : [];
    expect(children[0]).toHaveAccessibleName("前へ");
    expect(children[2]).toHaveAccessibleName("次へ");
  });
});
