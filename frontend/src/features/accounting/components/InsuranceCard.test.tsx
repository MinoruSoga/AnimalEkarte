import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { InsuranceCard } from "./InsuranceCard";

function renderInsuranceCard(
  overrides: Partial<{
    useInsurance: boolean;
    insuranceRatio: string;
    insuranceAmount: number;
    onUseInsuranceChange: (v: boolean) => void;
    onInsuranceRatioChange: (v: string) => void;
  }> = {},
) {
  const onUseInsuranceChange = overrides.onUseInsuranceChange ?? vi.fn();
  const onInsuranceRatioChange = overrides.onInsuranceRatioChange ?? vi.fn();
  return {
    onUseInsuranceChange,
    onInsuranceRatioChange,
    ...render(
      <InsuranceCard
        useInsurance={overrides.useInsurance ?? true}
        onUseInsuranceChange={onUseInsuranceChange}
        insuranceRatio={overrides.insuranceRatio ?? "0.5"}
        onInsuranceRatioChange={onInsuranceRatioChange}
        insuranceAmount={overrides.insuranceAmount ?? 0}
      />,
    ),
  };
}

describe("InsuranceCard accessibility", () => {
  it("保険利用switchに操作内容を表すaccessible nameがある", () => {
    renderInsuranceCard({ useInsurance: false });

    expect(screen.getByRole("switch", { name: "ペット保険を利用" })).toBeInTheDocument();
  });
});

describe("InsuranceCard insurance ratio options", () => {
  it("新規会計の負担割合は50%と70%のみで90%と100%は選択肢に出ない", async () => {
    const user = userEvent.setup();
    renderInsuranceCard({ insuranceRatio: "0.5" });

    expect(screen.getByRole("combobox", { name: "負担割合" })).toHaveTextContent("50%");

    await user.click(screen.getByRole("combobox", { name: "負担割合" }));

    expect(screen.getByRole("option", { name: "50%" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "70%" })).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "90%" })).not.toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "100%" })).not.toBeInTheDocument();
  });

  it("保存済み90%は50%へ書き換えず表示する", () => {
    const { onInsuranceRatioChange } = renderInsuranceCard({ insuranceRatio: "0.9" });

    expect(screen.getByRole("combobox", { name: "負担割合" })).toHaveTextContent("90%");
    expect(onInsuranceRatioChange).not.toHaveBeenCalled();
  });

  it("保存済み100%は50%へ書き換えず表示する", () => {
    const { onInsuranceRatioChange } = renderInsuranceCard({ insuranceRatio: "1.0" });

    expect(screen.getByRole("combobox", { name: "負担割合" })).toHaveTextContent("100%");
    expect(onInsuranceRatioChange).not.toHaveBeenCalled();
  });
});
