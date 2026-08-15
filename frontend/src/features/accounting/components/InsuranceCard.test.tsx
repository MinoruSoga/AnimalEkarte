import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { InsuranceCard } from "./InsuranceCard";

describe("InsuranceCard accessibility", () => {
  it("保険利用switchに操作内容を表すaccessible nameがある", () => {
    render(
      <InsuranceCard
        useInsurance={false}
        onUseInsuranceChange={vi.fn()}
        insuranceRatio="0.5"
        onInsuranceRatioChange={vi.fn()}
        insuranceAmount={0}
      />,
    );

    expect(screen.getByRole("switch", { name: "ペット保険を利用" })).toBeInTheDocument();
  });
});
