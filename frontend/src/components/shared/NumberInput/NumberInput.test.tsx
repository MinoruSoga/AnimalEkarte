import { useState } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { NumberInput } from "./NumberInput";

function ControlledNumberInput() {
  const [value, setValue] = useState("1.5");
  return (
    <NumberInput
      name="dose"
      value={value}
      onChange={setValue}
      step="any"
      min={0}
      required
      aria-required="true"
      aria-label="投与量"
    />
  );
}

describe("NumberInput native number contract", () => {
  it("FormData に名前付き値を保持し、小数 step と入力制約を渡す", () => {
    const { container } = render(
      <form>
        <NumberInput
          name="dose"
          value="1.5"
          onChange={vi.fn()}
          step="any"
          min={0}
          max={10}
          required
          aria-required="true"
          aria-label="投与量"
        />
      </form>,
    );
    const input = screen.getByRole("spinbutton", { name: "投与量" });
    const form = container.querySelector("form");

    expect(form).not.toBeNull();
    expect(new FormData(form!).getAll("dose")).toEqual(["1.5"]);
    expect(input).toHaveAttribute("step", "any");
    expect(input).toHaveAttribute("min", "0");
    expect(input).toHaveAttribute("max", "10");
    expect(input).toBeRequired();
    expect(input).toHaveAttribute("aria-required", "true");
  });

  it("blank value を空文字のまま保持する", () => {
    render(<ControlledNumberInput />);
    const input = screen.getByRole("spinbutton", { name: "投与量" });

    fireEvent.change(input, { target: { value: "" } });

    expect(input).toHaveValue(null);
  });

  it("readOnly と onInvalid をネイティブ入力へ渡す", () => {
    const onInvalid = vi.fn();
    render(
      <NumberInput
        value={1}
        onChange={vi.fn()}
        readOnly
        onInvalid={onInvalid}
        aria-label="優先順位"
      />,
    );
    const input = screen.getByRole("spinbutton", { name: "優先順位" });

    fireEvent.invalid(input);

    expect(input).toHaveAttribute("readonly");
    expect(onInvalid).toHaveBeenCalledOnce();
  });
});
