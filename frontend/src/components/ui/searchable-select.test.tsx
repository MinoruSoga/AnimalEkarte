import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
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

    expect(screen.getByRole("combobox", { name: "合成監査選択" })).toHaveClass(
      "h-11",
      "min-w-11",
    );
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
});
