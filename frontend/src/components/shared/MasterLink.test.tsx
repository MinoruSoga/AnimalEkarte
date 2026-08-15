import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it } from "vitest";
import { MasterLink } from "./MasterLink";

describe("MasterLink", () => {
  it("マスタ編集導線は44px以上の操作領域を持つ", () => {
    render(
      <MemoryRouter>
        <MasterLink category="examination" label="編集" />
      </MemoryRouter>,
    );

    const link = screen.getByRole("link", { name: "編集" });
    expect(link).toHaveClass("min-h-11", "min-w-11");
    expect(link).toHaveAttribute("href", "/settings/treatment-items?tab=examination");
  });
});
