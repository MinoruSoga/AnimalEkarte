import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { CashReconciliationCard } from "./CashReconciliationCard";

describe("CashReconciliationCard 差額", () => {
  it("正の差額はプラスを付けない", () => {
    render(<CashReconciliationCard theoreticalCash={1000} actualCash={2234} />);
    expect(screen.getByText("¥1,234")).toBeInTheDocument();
    expect(screen.queryByText("+¥1,234")).not.toBeInTheDocument();
    expect(screen.queryByText("+1,234")).not.toBeInTheDocument();
  });

  it("負の差額は -¥ 順で表示する", () => {
    render(<CashReconciliationCard theoreticalCash={2234} actualCash={1000} />);
    expect(screen.getByText("-¥1,234")).toBeInTheDocument();
  });

  it("0 の差額は ¥0 と表示する", () => {
    render(<CashReconciliationCard theoreticalCash={1000} actualCash={1000} />);
    expect(screen.getByText("¥0")).toBeInTheDocument();
  });
});
