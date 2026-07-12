import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { MonthlyReportPrintArea } from "./MonthlyReportPrintArea";
import type { MonthlyReportResponse } from "../api/get-monthly-report";

const SUMMARY: MonthlyReportResponse["summary"] = {
  workingDays: 20,
  totalBillings: 50,
  totalAmount: 1234567,
  totalRefund: 3000,
  netAmount: 1231567,
  byPaymentMethod: { 現金: 500000, クレジットカード: 734567 },
  byCategory: {},
  taxBreakdown: {
    standard: { taxableAmount: 1000000, taxAmount: 90000 },
    reduced: { taxableAmount: 200000, taxAmount: 16000 },
  },
};

const DAILY_DETAILS: MonthlyReportResponse["dailyDetails"] = [
  {
    date: "2026-07-01",
    weekday: "水",
    amCount: 2,
    amNet: 10000,
    pmCount: 1,
    pmNet: 0,
    dayNet: 10500,
    refund: 0,
    amClosed: false,
    pmClosed: false,
    isHoliday: false,
  },
  {
    date: "2026-07-02",
    weekday: "木",
    amCount: 1,
    amNet: 5000,
    pmCount: 1,
    pmNet: 2000,
    dayNet: 7500,
    refund: 1500,
    amClosed: false,
    pmClosed: false,
    isHoliday: false,
  },
];

function renderPrint() {
  return render(
    <MonthlyReportPrintArea
      periodLabel="2026年7月"
      clinicName="テスト動物病院"
      summary={SUMMARY}
      dailyDetails={DAILY_DETAILS}
      standardTaxRate={0.1}
      reducedTaxRate={0.08}
    />,
  );
}

describe("MonthlyReportPrintArea: 金額セルの印字が固定されている (FE5-14)", () => {
  it("サマリーKPI（売上合計・返金・純売上）を ¥ 区切りで出力する", () => {
    renderPrint();
    const area = screen.getByTestId("monthly-report-print-area");
    expect(within(area).getByText("¥1,234,567")).toBeInTheDocument();
    expect(within(area).getByText("-¥3,000")).toBeInTheDocument();
    expect(within(area).getByText("¥1,231,567")).toBeInTheDocument();
  });

  it("消費税内訳を ¥ 区切りで出力する", () => {
    renderPrint();
    const area = screen.getByTestId("monthly-report-print-area");
    expect(within(area).getByText("¥1,000,000")).toBeInTheDocument();
    expect(within(area).getByText("¥90,000")).toBeInTheDocument();
    expect(within(area).getByText("¥200,000")).toBeInTheDocument();
    expect(within(area).getByText("¥16,000")).toBeInTheDocument();
  });

  it("日次明細の0円行は ¥0、返金なし行は — を出力する", () => {
    renderPrint();
    const area = screen.getByTestId("monthly-report-print-area");
    expect(within(area).getByText("¥0")).toBeInTheDocument();
    expect(within(area).getAllByText("—").length).toBeGreaterThan(0);
  });

  it("返金がある日次明細行は -¥ 表記で出力する", () => {
    renderPrint();
    const area = screen.getByTestId("monthly-report-print-area");
    expect(within(area).getByText("-¥1,500")).toBeInTheDocument();
  });
});
