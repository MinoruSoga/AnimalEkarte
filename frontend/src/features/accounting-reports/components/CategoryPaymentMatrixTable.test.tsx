import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { CategoryPaymentMatrixTable } from "./CategoryPaymentMatrixTable";
import type { CategoryPaymentMatrix } from "../api/get-monthly-report";

const MATRIX: CategoryPaymentMatrix = {
  paymentMethods: [
    { id: 1, name: "現金", isActive: true },
    { id: 2, name: "クレジット", isActive: true },
  ],
  rows: [
    {
      category: "examination",
      count: 2,
      byMethod: { 現金: 3000, クレジット: 7000 },
      rowTotal: 10000,
    },
    {
      category: "goods",
      count: 1,
      byMethod: { 現金: 500 },
      rowTotal: 500,
    },
  ],
  totals: {
    count: 3,
    byMethod: { 現金: 3500, クレジット: 7000 },
    grandTotal: 10500,
  },
};

describe("CategoryPaymentMatrixTable (#247)", () => {
  it("部門ラベル集約・件数・支払列・合計を表示する", () => {
    render(<CategoryPaymentMatrixTable matrix={MATRIX} />);
    expect(screen.getByText("診療")).toBeInTheDocument();
    expect(screen.getByText("用品")).toBeInTheDocument();
    expect(screen.getByText("現金")).toBeInTheDocument();
    expect(screen.getByText("クレジット")).toBeInTheDocument();
    // フッタ件数 = 会計 distinct 総数
    expect(screen.getByText("3件")).toBeInTheDocument();
  });

  it("空 rows は EmptyState", () => {
    render(
      <CategoryPaymentMatrixTable
        matrix={{ paymentMethods: [], rows: [], totals: { count: 0, byMethod: {}, grandTotal: 0 } }}
      />,
    );
    expect(screen.getByText("対象期間の会計データがありません")).toBeInTheDocument();
  });
});
