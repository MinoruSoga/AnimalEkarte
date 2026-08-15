import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { HospitalizationCostSummary } from "./HospitalizationCostSummary";

describe("HospitalizationCostSummary", () => {
  it("authoritativeな保険条件がない段階では負担額を確定表示しない", () => {
    render(
      <HospitalizationCostSummary
        totals={{
          subtotalBeforeDiscount: 10_000,
          subtotalAfterDiscount: 10_000,
          consumptionTax: 1_000,
          total: 11_000,
        }}
      />,
    );

    expect(screen.queryByText("保険請求額")).not.toBeInTheDocument();
    expect(screen.queryByText("飼主請求額")).not.toBeInTheDocument();
    expect(screen.getByText(/会計時に確定します/)).toBeVisible();
    // W-003: bulk discount % / 円 inputs removed entirely
    expect(screen.queryByLabelText("割引適用額")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("値引適用額")).not.toBeInTheDocument();
    expect(screen.queryByRole("spinbutton")).not.toBeInTheDocument();
    // RO totals remain
    expect(screen.getByText("診療費 小計")).toBeVisible();
    expect(screen.getByText("￥10,000")).toBeVisible();
    expect(screen.getByText("消費税")).toBeVisible();
    expect(screen.getByText("￥1,000")).toBeVisible();
    expect(screen.getByText("請求額")).toBeVisible();
    expect(screen.getByText("￥11,000")).toBeVisible();
  });
});
