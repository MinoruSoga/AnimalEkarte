import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { FieldHelp } from "./FieldHelp";
import { C } from "@/lib/design-tokens";

describe("FieldHelp", () => {
  it("小さい説明アイコンを保ちながら44pxの操作領域と読みやすい色を使う", () => {
    render(<FieldHelp label="単価" content="税込価格です" />);
    const button = screen.getByRole("button", { name: "単価の説明" });
    expect(button).toHaveClass("min-h-11", "min-w-11", C.text60);
    expect(button.querySelector("svg")).toHaveClass("size-4");
  });
  it("「{label}の説明」の aria-label を持つボタンを描画する", () => {
    render(<FieldHelp label="単価" content="税込価格です" />);
    expect(screen.getByRole("button", { name: "単価の説明" })).toBeInTheDocument();
  });

  it("フォーカスで説明文のツールチップを表示する", async () => {
    const user = userEvent.setup();
    render(<FieldHelp label="単価" content="税込価格です" />);
    const tip = screen.getByRole("tooltip", { hidden: true });
    expect(tip).not.toBeVisible();
    await user.tab();
    expect(screen.getByRole("button", { name: "単価の説明" })).toHaveFocus();
    expect(tip).toBeVisible();
    expect(tip).toHaveTextContent("税込価格です");
  });

  it("type=button で描画される（フォーム送信を起こさない）", () => {
    render(<FieldHelp label="単価" content="税込価格です" />);
    expect(screen.getByRole("button", { name: "単価の説明" })).toHaveAttribute("type", "button");
  });
});
