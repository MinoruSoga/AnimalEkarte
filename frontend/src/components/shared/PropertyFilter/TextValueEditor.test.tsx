import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { C } from "@/lib/design-tokens";

import { TextValueEditor } from "./TextValueEditor";

/**
 * EMR-411 C-1b: 適用（確定）ボタンのカラーを C.text60 へ引き上げる。
 * 既存の Enter/ボタン確定挙動は変えない。
 */
describe("TextValueEditor 適用ボタン (EMR-411 C-1b)", () => {
  it("適用ボタンが C.text60 を持ち、確定動作は維持される", async () => {
    const user = userEvent.setup();
    const onApply = vi.fn();
    render(<TextValueEditor label="飼主名" onApply={onApply} />);

    const apply = screen.getByRole("button", { name: "飼主名フィルタを適用" });
    expect(apply).toHaveClass(C.text60);
    expect(apply).toHaveClass("min-h-11", "min-w-11");

    await user.type(screen.getByRole("textbox", { name: "飼主名を入力" }), "山田");
    await user.click(apply);
    expect(onApply).toHaveBeenCalledWith("山田", "山田");
  });
});
