import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { isDialogPortaledOverlayTarget } from "./dialog-portaled-overlay";
import { Popover, PopoverContent, PopoverTrigger } from "./popover";
import { SearchableSelect } from "./searchable-select";

describe("SearchableSelect", () => {
  it("選択トリガーは44px以上の操作領域を持つ", () => {
    render(
      <SearchableSelect
        value=""
        onValueChange={vi.fn()}
        options={[]}
        placeholder="選択してください"
        ariaLabel="合成監査選択"
      />,
    );

    expect(screen.getByRole("combobox", { name: "合成監査選択" })).toHaveClass("h-11", "min-w-11");
  });

  it("BUG-017: ariaInvalid と ariaDescribedBy をトリガーへ伝播する", () => {
    render(
      <SearchableSelect
        value=""
        onValueChange={vi.fn()}
        options={[]}
        placeholder="選択してください"
        ariaLabel="検査種別"
        ariaInvalid
        ariaDescribedBy="testTypeId-error"
      />,
    );

    const trigger = screen.getByRole("combobox", { name: "検査種別" });
    expect(trigger).toHaveAttribute("aria-invalid", "true");
    expect(trigger).toHaveAttribute("aria-describedby", "testTypeId-error");
  });

  it("options に無い value は fallbackLabel を表示し、未指定時は placeholder", () => {
    const { rerender } = render(
      <SearchableSelect
        value="10"
        onValueChange={vi.fn()}
        options={[{ value: "11", label: "鈴木" }]}
        placeholder="選択してください"
        fallbackLabel="三井"
        ariaLabel="担当者"
      />,
    );

    expect(screen.getByRole("combobox", { name: "担当者" })).toHaveTextContent("三井");

    rerender(
      <SearchableSelect
        value="10"
        onValueChange={vi.fn()}
        options={[{ value: "11", label: "鈴木" }]}
        placeholder="選択してください"
        ariaLabel="担当者"
      />,
    );
    expect(screen.getByRole("combobox", { name: "担当者" })).toHaveTextContent("選択してください");
  });
});

describe("Dialog × portaled Popover (担当者セレクト)", () => {
  // Modal Dialog sets body { pointer-events: none }. Portaled Popover must opt back in,
  // otherwise ReservationFormModal staff options receive no mouse clicks.
  it("PopoverContent は pointer-events-auto を持つ", () => {
    render(
      <Popover open>
        <PopoverTrigger type="button">担当者</PopoverTrigger>
        <PopoverContent>候補</PopoverContent>
      </Popover>,
    );

    expect(document.querySelector('[data-slot="popover-content"]')).toHaveClass(
      "pointer-events-auto",
    );
  });

  it("portaled popover/select を Dialog 外側操作として扱わない", () => {
    const popper = document.createElement("div");
    popper.setAttribute("data-radix-popper-content-wrapper", "");
    const option = document.createElement("div");
    option.setAttribute("role", "option");
    popper.appendChild(option);

    expect(isDialogPortaledOverlayTarget(option)).toBe(true);
    expect(isDialogPortaledOverlayTarget(document.createElement("div"))).toBe(false);
  });
});
