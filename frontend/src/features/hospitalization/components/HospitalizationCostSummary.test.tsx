import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { HospitalizationCostSummary } from "./HospitalizationCostSummary";

describe("HospitalizationCostSummary responsive layout", () => {
  it("保険・飼主請求額はmobileで全幅1列、sm以上で既存の2列に戻る", () => {
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
      />,
    );

    const claimGrid = screen.getByText("保険請求額").closest('[class*="grid-cols"]');
    expect(claimGrid).toHaveClass("w-full", "grid-cols-1", "sm:grid-cols-2");
    expect(claimGrid).not.toHaveClass("grid-cols-2");
  });
});
