import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { C } from "@/lib/design-tokens";
import { TableCell, TableHead, TableHeader } from "./table";

describe("Table cell primitives", () => {
  it("TableHeader は DESIGN.md の canvas-soft background を使う", () => {
    render(
      <table>
        <TableHeader data-testid="table-header" />
      </table>,
    );

    expect(screen.getByTestId("table-header")).toHaveClass(C.bgPage);
    expect(screen.getByTestId("table-header")).not.toHaveClass("bg-muted/30");
  });

  it("TableHead は DESIGN.md の eyebrow と cell padding に一致する", () => {
    render(
      <table>
        <thead>
          <tr>
            <TableHead>項目</TableHead>
          </tr>
        </thead>
      </table>,
    );

    const header = screen.getByRole("columnheader", { name: "項目" });
    // text-2xs は globals.css で 12px / 1.33 / +0.125px に固定されている。
    expect(header).toHaveClass("text-2xs", "font-semibold", "px-4", "py-3");
    expect(header).not.toHaveClass("px-2");
    expect(header).not.toHaveClass("p-2");
  });

  it("TableCell は DESIGN.md の body-sm と cell padding に一致する", () => {
    render(
      <table>
        <tbody>
          <tr>
            <TableCell>内容</TableCell>
          </tr>
        </tbody>
      </table>,
    );

    const cell = screen.getByRole("cell", { name: "内容" });
    // text-sm は globals.css で 15px / 1.33 に固定されている。
    expect(cell).toHaveClass("text-sm", "font-normal", "px-4", "py-3");
    expect(cell).not.toHaveClass("text-base");
    expect(cell).not.toHaveClass("p-2");
    expect(cell).not.toHaveClass("py-2.5");
  });
});
