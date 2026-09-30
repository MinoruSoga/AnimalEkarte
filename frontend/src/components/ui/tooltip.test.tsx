import { describe, it, expect } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Tooltip } from "./tooltip";

describe("Tooltip", () => {
  it("初期状態ではツールチップが非表示", () => {
    render(
      <Tooltip content="説明文">
        <button type="button">対象</button>
      </Tooltip>,
    );
    const tip = screen.getByRole("tooltip", { hidden: true });
    expect(tip).toHaveTextContent("説明文");
    expect(tip).not.toBeVisible();
  });

  it("ホバーで表示し、マウスアウトで閉じる", () => {
    render(
      <Tooltip content="説明文">
        <button type="button">対象</button>
      </Tooltip>,
    );
    const tip = screen.getByRole("tooltip", { hidden: true });
    fireEvent.mouseEnter(screen.getByText("対象"));
    expect(tip).toBeVisible();
    fireEvent.mouseLeave(screen.getByText("対象"));
    expect(tip).not.toBeVisible();
  });

  it("キーボードフォーカスで表示し、ブラーで閉じる", async () => {
    const user = userEvent.setup();
    render(
      <>
        <Tooltip content="説明文">
          <button type="button">対象</button>
        </Tooltip>
        <button type="button">次</button>
      </>,
    );
    const tip = screen.getByRole("tooltip", { hidden: true });
    await user.tab();
    expect(screen.getByText("対象")).toHaveFocus();
    expect(tip).toBeVisible();
    await user.tab();
    expect(tip).not.toBeVisible();
  });

  it("Escape キーで閉じる", async () => {
    const user = userEvent.setup();
    render(
      <Tooltip content="説明文">
        <button type="button">対象</button>
      </Tooltip>,
    );
    const tip = screen.getByRole("tooltip", { hidden: true });
    await user.tab();
    expect(tip).toBeVisible();
    await user.keyboard("{Escape}");
    expect(tip).not.toBeVisible();
  });

  it("複数行の説明文を保持する（whitespace-pre-line）", () => {
    render(
      <Tooltip content={"1行目\n2行目"}>
        <button type="button">対象</button>
      </Tooltip>,
    );
    const tip = screen.getByRole("tooltip", { hidden: true });
    expect(tip).toHaveClass("whitespace-pre-line");
    expect(tip.textContent).toBe("1行目\n2行目");
  });

  it("w-max で shrink-to-fit による縦長化を防ぐ（画面右端のトリガー対策）", () => {
    render(
      <Tooltip content="説明文">
        <button type="button">対象</button>
      </Tooltip>,
    );
    expect(screen.getByRole("tooltip", { hidden: true })).toHaveClass("w-max");
  });
});
