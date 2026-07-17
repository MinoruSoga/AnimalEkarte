import { useState } from "react";
import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ClearableSearchInput } from "./ClearableSearchInput";

function ControlledClearableSearchInput({ placeholder }: { placeholder?: string }) {
  const [value, setValue] = useState("");
  return <ClearableSearchInput value={value} onChange={setValue} placeholder={placeholder} />;
}

describe("ClearableSearchInput", () => {
  it("value が空のときクリアボタンは表示されない", () => {
    render(<ClearableSearchInput value="" onChange={() => {}} placeholder="検索..." />);
    expect(screen.queryByRole("button", { name: "検索をクリア" })).not.toBeInTheDocument();
  });

  it("value があるときクリアボタンが表示される", () => {
    render(<ClearableSearchInput value="再診" onChange={() => {}} />);
    expect(screen.getByRole("button", { name: "検索をクリア" })).toBeInTheDocument();
  });

  it("入力すると値が反映され、確定した文字列で onChange 相当の状態になる", async () => {
    const user = userEvent.setup();
    render(<ControlledClearableSearchInput placeholder="検索..." />);

    await user.type(screen.getByPlaceholderText("検索..."), "abc");

    expect(screen.getByPlaceholderText("検索...")).toHaveValue("abc");
    expect(screen.getByRole("button", { name: "検索をクリア" })).toBeInTheDocument();
  });

  it("クリアボタンをクリックすると onChange が空文字で呼ばれる", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<ClearableSearchInput value="再診" onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: "検索をクリア" }));

    expect(onChange).toHaveBeenCalledWith("");
  });

  it("clearAriaLabel でクリアボタンのラベルをカスタマイズできる", () => {
    render(<ClearableSearchInput value="a" onChange={() => {}} clearAriaLabel="クリア" />);
    expect(screen.getByRole("button", { name: "クリア" })).toBeInTheDocument();
  });
});
