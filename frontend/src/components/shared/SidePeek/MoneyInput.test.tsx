import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { MoneyInput } from "./MoneyInput";

describe("MoneyInput", () => {
  it("optional minimum reaches the native input and clearing reports zero", () => {
    const onChange = vi.fn();
    render(<MoneyInput value={5000} onChange={onChange} min={0} />);
    const input = screen.getByRole("spinbutton", { name: "単価(税込)" });

    expect(input).toHaveAttribute("min", "0");
    fireEvent.change(input, { target: { value: "" } });
    expect(onChange).toHaveBeenCalledWith(0);
  });

  it("string-backed callers preserve empty and typed price values", () => {
    const onStringChange = vi.fn();
    const { rerender } = render(
      <MoneyInput stringValue="" onStringChange={onStringChange} min={0} />,
    );
    const input = screen.getByRole("spinbutton", { name: "単価(税込)" });

    expect(input).toHaveValue(null);
    fireEvent.change(input, { target: { value: "1250" } });
    expect(onStringChange).toHaveBeenCalledWith("1250");

    rerender(<MoneyInput stringValue="0" onStringChange={onStringChange} min={0} />);
    expect(input).toHaveValue(0);
    fireEvent.change(input, { target: { value: "" } });
    expect(onStringChange).toHaveBeenLastCalledWith("");
  });
});

describe("MoneyInput variant=inline", () => {
  it("外部 Label と id で紐付き、ネイティブ props をそのまま通す", () => {
    const onStringChange = vi.fn();
    render(
      <>
        <label htmlFor="amount-field">金額</label>
        <MoneyInput
          variant="inline"
          id="amount-field"
          name="amount"
          stringValue="1200"
          onStringChange={onStringChange}
          min={0}
          step={1}
          disabled
        />
      </>,
    );
    const input = screen.getByLabelText("金額");
    expect(input).toHaveAttribute("id", "amount-field");
    expect(input).toHaveAttribute("name", "amount");
    expect(input).toHaveAttribute("min", "0");
    expect(input).toHaveAttribute("step", "1");
    expect(input).toBeDisabled();
  });

  it("PropertyRow と先頭 ¥ を出さず、w-32/text-sm/font-bold を強制しない", () => {
    render(
      <MoneyInput
        variant="inline"
        stringValue="500"
        onStringChange={vi.fn()}
        suffix="円"
        align="right"
        className="h-12"
      />,
    );
    const input = screen.getByRole("spinbutton");
    expect(input.className).toContain("h-12");
    expect(input.className).toContain("text-base");
    expect(input.className).toContain("font-normal");
    expect(input.className).toContain("text-right");
    expect(input).not.toHaveClass("w-32");
    expect(input).not.toHaveClass("text-sm");
    expect(input).not.toHaveClass("font-bold");
    expect(screen.queryByText("¥")).not.toBeInTheDocument();
    expect(screen.getByText("円")).toBeInTheDocument();
  });

  it("文字列入力と空文字をそのまま通知し、error を描画する", () => {
    const onStringChange = vi.fn();
    render(
      <MoneyInput
        variant="inline"
        aria-label="返金額"
        stringValue=""
        onStringChange={onStringChange}
        error="0以上の金額を入力してください"
      />,
    );
    const input = screen.getByRole("spinbutton", { name: "返金額" });
    fireEvent.change(input, { target: { value: "300" } });
    expect(onStringChange).toHaveBeenCalledWith("300");
    expect(screen.getByText("0以上の金額を入力してください")).toBeInTheDocument();
    expect(input).toHaveAttribute("aria-invalid", "true");
  });

  it("既定の property variant は従来どおり PropertyRow + 先頭 ¥ を保つ", () => {
    render(<MoneyInput value={800} onChange={vi.fn()} />);
    const input = screen.getByRole("spinbutton", { name: "単価(税込)" });
    expect(input).toHaveClass("w-32");
    expect(input).toHaveClass("text-sm");
    expect(screen.getByText("¥")).toBeInTheDocument();
  });
});

describe("MoneyInput 追加回帰", () => {
  it("inline string モードは 0 を保持しクリアで空文字を通知する", () => {
    const onStringChange = vi.fn();
    const { rerender } = render(
      <MoneyInput variant="inline" stringValue="0" onStringChange={onStringChange} />,
    );
    const input = screen.getByRole("spinbutton");
    expect(input).toHaveValue(0);
    fireEvent.change(input, { target: { value: "" } });
    expect(onStringChange).toHaveBeenLastCalledWith("");
    rerender(<MoneyInput variant="inline" stringValue="450" onStringChange={onStringChange} />);
    expect(input).toHaveValue(450);
  });

  it("inline は error を ${id}-error で関連付け外部 describedby を残す", () => {
    render(
      <MoneyInput
        variant="inline"
        id="amount-x"
        stringValue=""
        onStringChange={vi.fn()}
        error="必須です"
        aria-describedby="external-hint"
      />,
    );
    const input = screen.getByRole("spinbutton");
    expect(input).toHaveAttribute("aria-describedby", "external-hint amount-x-error");
    expect(document.getElementById("amount-x-error")).toHaveTextContent("必須です");
  });

  it("property variant はネイティブ属性をフォワードする", () => {
    render(
      <MoneyInput
        value={100}
        onChange={vi.fn()}
        id="prop-field"
        name="unitPrice"
        min={0}
        max={9999}
        disabled
        required
        readOnly
        aria-required={true}
        aria-describedby="prop-hint"
      />,
    );
    const input = screen.getByRole("spinbutton", { name: "単価(税込)" });
    expect(input).toHaveAttribute("id", "prop-field");
    expect(input).toHaveAttribute("name", "unitPrice");
    expect(input).toHaveAttribute("min", "0");
    expect(input).toHaveAttribute("max", "9999");
    expect(input).toBeDisabled();
    expect(input).toBeRequired();
    expect(input).toHaveAttribute("readonly", "");
    expect(input).toHaveAttribute("aria-describedby", "prop-hint");
  });

  it("numeric mode のクリアは 0 を通知する（inline でも同じ）", () => {
    const onChange = vi.fn();
    render(<MoneyInput variant="inline" value={300} onChange={onChange} />);
    fireEvent.change(screen.getByRole("spinbutton"), { target: { value: "" } });
    expect(onChange).toHaveBeenCalledWith(0);
  });
});
