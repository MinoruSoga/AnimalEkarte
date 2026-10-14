import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { EstimateLineItem } from "../../types";
import { EstimateLineItems } from "./EstimateLineItems";

const item: EstimateLineItem = {
  id: "1",
  estimateId: "1",
  name: "診察料",
  category: "consultation",
  unitPrice: 1000,
  quantity: 1,
  taxType: "excluded",
  taxRate: 0.1,
  discountRate: 0,
  discountAmount: 0,
  isInsuranceApplicable: false,
  sortOrder: 0,
  createdAt: "2026-07-21T00:00:00Z",
};

describe("EstimateLineItems DESIGN.md table contract", () => {
  it.each([
    ["excluded", 1100, "外税 10%"],
    ["included", 1000, "内税 10%"],
    ["exempt", 1000, "非課税"],
  ] as const)("保存後の税区分 %s と税込行金額を表示する", (taxType, total, label) => {
    render(
      <EstimateLineItems
        items={[{ ...item, taxType }]}
        subtotal={1000}
        taxTotal={taxType === "exempt" ? 0 : 100}
        insuranceAmount={0}
        discountAmount={0}
        totalAmount={total}
      />,
    );
    const row = screen.getByText("診察料").closest("tr");
    if (!row) throw new Error("missing row");
    const cells = within(row).getAllByRole("cell");
    expect(cells[5]).toHaveTextContent(label);
    expect(cells[7]).toHaveTextContent(`¥${total.toLocaleString("ja-JP")}`);
  });

  it("明細 body cell は body-sm と 12px 16px padding を維持する", () => {
    render(
      <EstimateLineItems
        items={[item]}
        subtotal={1000}
        taxTotal={100}
        insuranceAmount={0}
        discountAmount={0}
        totalAmount={1100}
      />,
    );

    const row = screen.getByText("診察料").closest("tr");
    if (row === null) throw new Error("estimate line item row was not rendered");
    for (const cell of within(row).getAllByRole("cell")) {
      expect(cell).toHaveClass("text-sm", "px-4", "py-3");
      expect(cell).not.toHaveClass("py-2");
    }
  });
});
