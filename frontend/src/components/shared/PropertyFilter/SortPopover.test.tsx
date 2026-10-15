import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { C } from "@/lib/design-tokens";

import { SortPopover } from "./SortPopover";

/**
 * EMR-411 C-1b: hover に頼って初めて現れる削除ボタンは、タッチ端末
 * (hover:none) では常時表示にする。opacity-0 + group-hover 系の既存クラスは
 * デスクトップのホバー表示契約として維持しつつ、
 * [@media(hover:none)]:opacity-100 を併記する。カラーも C.text60 へ引き上げる。
 */
describe("SortPopover ソート削除ボタンのタッチ可視性 (EMR-411 C-1b)", () => {
  it("削除ボタンが hover:none 常時表示・マウス hover/focus 表示・text60 を併せ持つ", async () => {
    const user = userEvent.setup();
    render(
      <SortPopover
        sortProperties={[{ key: "name", label: "名前" }]}
        activeSorts={[{ key: "name", direction: "asc" }]}
        onSortChange={vi.fn()}
      />,
    );

    await user.click(screen.getByRole("button", { name: "並べ替え" }));
    const remove = await screen.findByRole("button", { name: "名前 ソートを削除" });

    expect(remove).toHaveClass("[@media(hover:none)]:opacity-100");
    expect(remove).toHaveClass(
      "opacity-0",
      "group-hover:opacity-100",
      "focus-visible:opacity-100",
      "group-focus-within:opacity-100",
    );
    expect(remove).toHaveClass(C.text60);
    expect(remove).toHaveClass("min-h-11", "min-w-11");
  });
});
