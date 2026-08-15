import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import { MonthlySummaryCards } from "./MonthlySummaryCards";
import type { MonthlyReportResponse } from "../api/get-monthly-report";

// 画面表示の transform 後（camelCase）shape。税率ラベルの動的化検証に集中する。
const SUMMARY: MonthlyReportResponse["summary"] = {
  workingDays: 20,
  totalBillings: 150,
  totalAmount: 1234567,
  totalRefund: 5000,
  netAmount: 1229567,
  byPaymentMethod: { 現金: 800000 },
  byCategory: { 診療: 600000 },
  taxBreakdown: {
    standard: { taxableAmount: 1100000, taxAmount: 110000 },
    reduced: { taxableAmount: 100000, taxAmount: 8000 },
  },
};

describe("MonthlySummaryCards 消費税率ラベル動的化 (#179 ②)", () => {
  it("既定税率(0.1 / 0.08)では 標準税率（10%）/ 軽減税率（8%）を表示する", () => {
    render(<MonthlySummaryCards summary={SUMMARY} standardTaxRate={0.1} reducedTaxRate={0.08} />);
    expect(screen.getByText("標準税率（10%）")).toBeInTheDocument();
    expect(screen.getByText("軽減税率（8%）")).toBeInTheDocument();
  });

  it("病院マスタ設定値(0.12 / 0.05)が税率ラベルへ反映され、固定の10%/8%表記が出ない", () => {
    render(<MonthlySummaryCards summary={SUMMARY} standardTaxRate={0.12} reducedTaxRate={0.05} />);
    // 設定値が動的にラベルへ反映される
    expect(screen.getByText("標準税率（12%）")).toBeInTheDocument();
    expect(screen.getByText("軽減税率（5%）")).toBeInTheDocument();
    // 旧ハードコード（10% / 8%）が残っていないこと（撤廃の回帰防止）
    expect(screen.queryByText("標準税率（10%）")).not.toBeInTheDocument();
    expect(screen.queryByText("軽減税率（8%）")).not.toBeInTheDocument();
  });

  it("taxSettingsLink を渡すと消費税内訳ヘッダに表示する", () => {
    render(
      <MonthlySummaryCards
        summary={SUMMARY}
        standardTaxRate={0.1}
        reducedTaxRate={0.08}
        taxSettingsLink={<a href="/settings/clinic">税率設定を変更</a>}
      />,
    );
    expect(screen.getByRole("link", { name: "税率設定を変更" })).toHaveAttribute(
      "href",
      "/settings/clinic",
    );
  });

  it("taxSettingsLink 未指定では税率設定リンクを出さない", () => {
    render(<MonthlySummaryCards summary={SUMMARY} standardTaxRate={0.1} reducedTaxRate={0.08} />);
    expect(screen.queryByRole("link", { name: /税率設定を変更/ })).not.toBeInTheDocument();
  });
});
