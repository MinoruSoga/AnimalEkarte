import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { NextVisitButton } from "./NextVisitButton";

describe("NextVisitButton", () => {
  it("value が空のとき aria-label='次回予定を追加' のボタンが表示される", () => {
    render(<NextVisitButton value="" onChange={vi.fn()} onValidationChange={vi.fn()} />);
    expect(screen.getByRole("button", { name: "次回予定を追加" })).toBeInTheDocument();
  });

  it("value が設定されているとき aria-label='次回予定を変更' で日付が表示される", () => {
    render(<NextVisitButton value="2025-01-15" onChange={vi.fn()} onValidationChange={vi.fn()} />);
    expect(screen.getByRole("button", { name: "次回予定を変更" })).toBeInTheDocument();
    expect(screen.getByText("2025-01-15")).toBeInTheDocument();
  });

  it("ボタンをクリックで Popover が開き NextVisitDateField が表示される", async () => {
    const user = userEvent.setup();
    render(<NextVisitButton value="" onChange={vi.fn()} onValidationChange={vi.fn()} />);
    await user.click(screen.getByRole("button", { name: "次回予定を追加" }));
    // NextVisitDateField のラベルが表示される
    expect(screen.getByText("次回来院推奨日")).toBeInTheDocument();
  });

  it("tooltip が DOM に存在し value なしで「次回予定を追加」テキストを持つ", () => {
    render(<NextVisitButton value="" onChange={vi.fn()} onValidationChange={vi.fn()} />);
    // portal で常時 DOM に存在、非表示時は aria-hidden=true のため { hidden: true } で取得
    expect(screen.getByRole("tooltip", { hidden: true })).toHaveTextContent("次回予定を追加");
  });

  it("value がある場合 tooltip は「次回予定を変更」テキストを持つ", () => {
    render(<NextVisitButton value="2025-01-15" onChange={vi.fn()} onValidationChange={vi.fn()} />);
    expect(screen.getByRole("tooltip", { hidden: true })).toHaveTextContent("次回予定を変更");
  });
});
