import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { C } from "@/lib/design-tokens";
import { DaysRangeToggle } from "./DaysRangeToggle";

describe("DaysRangeToggle", () => {
  it("5日ボタンと7日ボタンが表示される", () => {
    render(<DaysRangeToggle days={5} onChange={vi.fn()} />);
    expect(screen.getByTestId("days-toggle-5")).toBeInTheDocument();
    expect(screen.getByTestId("days-toggle-7")).toBeInTheDocument();
  });

  it("5日・7日ボタンは44px以上のタッチターゲットを持つ", () => {
    render(<DaysRangeToggle days={5} onChange={vi.fn()} />);

    expect(screen.getByTestId("days-toggle-5")).toHaveClass("h-11", "min-w-11");
    expect(screen.getByTestId("days-toggle-7")).toHaveClass("h-11", "min-w-11");
  });

  it("days=5 のとき 5日ボタンが aria-pressed=true", () => {
    render(<DaysRangeToggle days={5} onChange={vi.fn()} />);
    expect(screen.getByTestId("days-toggle-5")).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByTestId("days-toggle-7")).toHaveAttribute("aria-pressed", "false");
  });

  it("days=7 のとき 7日ボタンが aria-pressed=true", () => {
    render(<DaysRangeToggle days={7} onChange={vi.fn()} />);
    expect(screen.getByTestId("days-toggle-7")).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByTestId("days-toggle-5")).toHaveAttribute("aria-pressed", "false");
  });

  it("active ボタンは brand と同じ primary teal を使う", () => {
    render(<DaysRangeToggle days={5} onChange={vi.fn()} />);
    const active = screen.getByTestId("days-toggle-5");
    expect(active.className).toContain(C.bgActionPrimary);
  });

  it("7日ボタンをクリックすると onChange(7) が呼ばれる", async () => {
    const onChange = vi.fn();
    render(<DaysRangeToggle days={5} onChange={onChange} />);
    await userEvent.click(screen.getByTestId("days-toggle-7"));
    expect(onChange).toHaveBeenCalledWith(7);
  });

  it("5日ボタンをクリックすると onChange(5) が呼ばれる", async () => {
    const onChange = vi.fn();
    render(<DaysRangeToggle days={7} onChange={onChange} />);
    await userEvent.click(screen.getByTestId("days-toggle-5"));
    expect(onChange).toHaveBeenCalledWith(5);
  });
});
