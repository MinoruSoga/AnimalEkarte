import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { HospitalizationCostSummary } from "./HospitalizationCostSummary";

describe("HospitalizationCostSummary", () => {
  it("authoritativeな保険条件がない段階では負担額を確定表示しない", () => {
    render(
      <HospitalizationCostSummary
        totals={{
          subtotalBeforeDiscount: 10_000,
          subtotalAfterDiscount: 9_000,
          consumptionTax: 900,
          total: 9_900,
        }}
        globalDiscount={10}
        setGlobalDiscount={vi.fn()}
        globalDiscountAmount={0}
        setGlobalDiscountAmount={vi.fn()}
        readOnly
      />,
    );

    expect(screen.queryByText("保険請求額")).not.toBeInTheDocument();
    expect(screen.queryByText("飼主請求額")).not.toBeInTheDocument();
    expect(screen.getByText(/会計時に確定します/)).toBeVisible();
    expect(screen.getByLabelText("割引適用額")).toBeDisabled();
    expect(screen.getByLabelText("値引適用額")).toBeDisabled();
  });
});
