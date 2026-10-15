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
