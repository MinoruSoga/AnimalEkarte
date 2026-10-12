import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import {
  isDialogPortaledOverlayTarget,
  shouldPreventDialogOutsideInteraction,
} from "./dialog-portaled-overlay";
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

  // EMR-316: clearable の項目ラベルを呼び出し側で「選択なし」に差し替えられる（既定は「選択をクリア」）。
  it("clearLabel 指定時は選択中のみその名前のクリア項目を出し、選ぶと空文字を返す", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();
    render(
      <SearchableSelect
        value="2"
        onValueChange={onValueChange}
        options={[
          { value: "1", label: "A-1" },
          { value: "2", label: "A-2" },
        ]}
        placeholder="選択なし"
        ariaLabel="ケージ"
        clearable
        clearLabel="選択なし"
      />,
    );

    await user.click(screen.getByRole("combobox", { name: "ケージ" }));

    expect(screen.queryByRole("option", { name: "選択をクリア" })).toBeNull();
    await user.click(screen.getByRole("option", { name: "選択なし" }));
    expect(onValueChange).toHaveBeenCalledWith("");
  });

  it("clearLabel 未指定の clearable は従来どおり「選択をクリア」を出す", async () => {
    const user = userEvent.setup();
    render(
      <SearchableSelect
        value="2"
        onValueChange={vi.fn()}
        options={[{ value: "2", label: "A-2" }]}
        ariaLabel="ケージ"
        clearable
      />,
    );

    await user.click(screen.getByRole("combobox", { name: "ケージ" }));

    expect(screen.getByRole("option", { name: "選択をクリア" })).toBeInTheDocument();
  });

  it("値が空のときは clearable でもクリア項目を出さない", async () => {
    const user = userEvent.setup();
    render(
      <SearchableSelect
        value=""
        onValueChange={vi.fn()}
        options={[{ value: "2", label: "A-2" }]}
        ariaLabel="ケージ"
        clearable
        clearLabel="選択なし"
      />,
    );

    await user.click(screen.getByRole("combobox", { name: "ケージ" }));

    expect(screen.queryByRole("option", { name: "選択なし" })).toBeNull();
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
    expect(
      shouldPreventDialogOutsideInteraction({
        target: document.createElement("div"),
        detail: { originalEvent: { target: option } as unknown as Event },
      }),
    ).toBe(true);
    expect(
      shouldPreventDialogOutsideInteraction({
        target: document.createElement("div"),
        detail: { originalEvent: { target: document.body } as unknown as Event },
      }),
    ).toBe(false);
  });

  // iPad/touch: SearchableSelect must open as a modal Popover so the dismissable
  // layer nests under Dialog and options remain hittable.
  it("担当者候補を開き選択できる（modal Popover）", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();
    render(
      <SearchableSelect
        value=""
        onValueChange={onValueChange}
        options={[
          { value: "7", label: "鈴木 諒平" },
          { value: "37", label: "三井隆行" },
        ]}
        placeholder="選択してください"
        ariaLabel="担当者"
        triggerTestId="res-staff-trigger"
      />,
    );

    await user.click(screen.getByTestId("res-staff-trigger"));
    expect(document.querySelector('[data-slot="popover-content"]')).not.toBeNull();
    await user.click(screen.getByRole("option", { name: "三井隆行" }));
    expect(onValueChange).toHaveBeenCalledWith("37");
  });
});
