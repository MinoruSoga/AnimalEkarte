import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { HospitalizationCostSummary } from "./HospitalizationCostSummary";

describe("HospitalizationCostSummary", () => {
  it("authoritativeな保険条件がない段階では負担額を確定表示しない", () => {
    render(
      <HospitalizationCostSummary
        totals={{
          subtotalBeforeDiscount: 7_700,
          consumptionTax: 700,
          total: 7_700,
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
    expect(screen.getByText("診療費 小計（明細合計）")).toBeVisible();
    expect(screen.getAllByText("¥7,700")).toHaveLength(2);
    expect(screen.getByText("消費税（内税分を含む）")).toBeVisible();
    expect(screen.getByText("¥700")).toBeVisible();
    expect(screen.getByText("診療費の目安（税込）")).toBeVisible();
    expect(screen.queryByText("請求額")).not.toBeInTheDocument();
  });

  it("内税・外税混在時も税を二重加算せず渡された合計をそのまま表示する", () => {
    render(
      <HospitalizationCostSummary
        totals={{
          subtotalBeforeDiscount: 8_700,
          consumptionTax: 800,
          total: 8_800,
        }}
      />,
    );

    expect(screen.getByText("¥8,700")).toBeVisible();
    expect(screen.getByText("¥800")).toBeVisible();
    expect(screen.getByText("¥8,800")).toBeVisible();
    expect(screen.queryByText("¥9,500")).not.toBeInTheDocument();
  });
});
