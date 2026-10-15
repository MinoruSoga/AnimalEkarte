import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
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
    {
      category: "house_call",
      count: 1,
      byMethod: { 現金: 600, クレジット: 400 },
      rowTotal: 1000,
    },
  ],
  totals: {
    count: 4,
    byMethod: { 現金: 4100, クレジット: 7400 },
    grandTotal: 11500,
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
    expect(screen.getByText("4件")).toBeInTheDocument();
  });

  it.each([false, true])("compact=%s でも往診と既存カテゴリを独立表示する", (compact) => {
    render(<CategoryPaymentMatrixTable matrix={MATRIX} compact={compact} />);

    const table = screen.getByRole("table");
    const houseCallRow = within(table).getByRole("row", { name: /往診/ });
    expect(within(houseCallRow).getByText("1件")).toBeInTheDocument();
    expect(within(houseCallRow).getByText("¥600")).toBeInTheDocument();
    expect(within(houseCallRow).getByText("¥400")).toBeInTheDocument();
    expect(within(houseCallRow).getByText("¥1,000")).toBeInTheDocument();

    const goodsRow = within(table).getByRole("row", { name: /用品/ });
    expect(within(goodsRow).getAllByText("¥500")).toHaveLength(2);
    const footerRow = within(table).getAllByRole("row").at(-1)!;
    expect(within(footerRow).getByText("¥11,500")).toBeInTheDocument();
  });

  it("house_call のみの matrix でも EmptyState ではなく往診行を表示する", () => {
    const houseCallOnly: CategoryPaymentMatrix = {
      ...MATRIX,
      rows: [
        {
          category: "house_call",
          count: 1,
          byMethod: { 現金: 600, クレジット: 400 },
          rowTotal: 1000,
        },
      ],
      totals: {
        count: 1,
        byMethod: { 現金: 600, クレジット: 400 },
        grandTotal: 1000,
      },
    };
    render(<CategoryPaymentMatrixTable matrix={houseCallOnly} />);

    expect(screen.getByRole("table")).toBeInTheDocument();
    expect(screen.getByRole("row", { name: /往診/ })).toBeInTheDocument();
    expect(screen.queryByText("対象期間の会計データがありません")).not.toBeInTheDocument();
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
