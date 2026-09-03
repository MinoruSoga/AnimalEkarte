import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { PropertyFilter } from "./PropertyFilter";

describe("PropertyFilter accessibility", () => {
  it("検索クリア操作のhit areaを44px以上に保つ", () => {
    render(
      <PropertyFilter
        properties={[]}
        activeFilters={[]}
        onFilterChange={vi.fn()}
        searchTerm="ポチ"
        onSearchChange={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "検索" }));

    expect(screen.getByRole("button", { name: "検索をクリア" })).toHaveClass(
      "min-h-11",
      "min-w-11",
    );
  });

  it("検索クリア操作の表示中はinputに44px分の右余白を確保する", () => {
    render(
      <PropertyFilter
        properties={[]}
        activeFilters={[]}
        onFilterChange={vi.fn()}
        searchTerm="ポチ"
        onSearchChange={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "検索" }));

    expect(screen.getByRole("textbox", { name: "検索..." })).toHaveClass("pr-11");
  });
});
