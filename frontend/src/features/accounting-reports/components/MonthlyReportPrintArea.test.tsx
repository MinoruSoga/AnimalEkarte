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
    amNet: 0,
    pmCount: 1,
    pmNet: 0,
    dayNet: 0,
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
  it("prints daily ISO dates with slashes without changing zero currency", () => {
    renderPrint();
    const area = screen.getByTestId("monthly-report-print-area");
    expect(within(area).getByText("2026/07/01")).toHaveClass("whitespace-nowrap");
    expect(within(area).getByText("¥0")).toBeInTheDocument();
  });
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

describe("MonthlyReportPrintArea 会計件数列と見出しレベル (EMR-401)", () => {
  it("日次明細は午前/午後の4列を出さず 会計件数 に合算し、日計を保持する", () => {
    render(
      <MonthlyReportPrintArea
        periodLabel="2026年7月"
        clinicName="テスト動物病院"
        summary={SUMMARY}
        dailyDetails={[
          {
            date: "2026-07-10",
            weekday: "金",
            amCount: 2,
            amNet: 2200,
            pmCount: 3,
            pmNet: 3300,
            dayNet: 5500,
            refund: 0,
            amClosed: true,
            pmClosed: false,
            isHoliday: false,
          },
        ]}
        standardTaxRate={0.1}
        reducedTaxRate={0.08}
      />,
    );
    const area = screen.getByTestId("monthly-report-print-area");
    const detailsTable = within(area).getByText("2026/07/10").closest("table");
    expect(detailsTable).not.toBeNull();

    for (const heading of ["午前件数", "午前売上", "午後件数", "午後売上"]) {
      expect(within(detailsTable!).queryByText(heading)).not.toBeInTheDocument();
    }
    expect(within(detailsTable!).getByText("会計件数")).toBeInTheDocument();
    const row = within(area).getByText("2026/07/10").closest("tr");
    expect(within(row!).getByText("5")).toBeInTheDocument();
    expect(within(row!).getByText("¥5,500")).toBeInTheDocument();
  });

  it("印刷見出しは文書内 h1 重複を避けるため h2 を使う", () => {
    renderPrint();
    const area = screen.getByTestId("monthly-report-print-area");

    expect(within(area).queryByRole("heading", { level: 1 })).not.toBeInTheDocument();
    expect(
      within(area).getByRole("heading", { level: 2, name: "月次集計レポート" }),
    ).toBeInTheDocument();
  });
});
