import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: {
      clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }],
      clinic: { name: "テスト動物病院" },
    },
    currentClinicId: "1",
    hasPermission: () => true,
  }),
}));

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

import { AccountingReportsPage } from "./AccountingReportsPage";

// Backend raw shape（transform 前 / snake_case）。distinctive な金額で
// 「画面値 == 印刷値」を検証しやすくする。
const MONTHLY = {
  year: 2026,
  month: 5,
  summary: {
    working_days: 20,
    total_billings: 150,
    total_amount: 1234567,
    total_refund: 5000,
    net_amount: 1229567,
    by_payment_method: { 現金: 800000, クレジットカード: 434567 },
    by_category: { 診療: 600000, 外科: 300000 },
    tax_breakdown: {
      standard: { taxable_amount: 1100000, tax_amount: 110000 },
      reduced: { taxable_amount: 100000, tax_amount: 8000 },
    },
  },
  daily_details: [
    {
      date: "2026-05-01",
      weekday: "金",
      am_count: 5,
      am_net: 50000,
      pm_count: 3,
      pm_net: 30000,
      day_net: 80000,
      refund: 0,
      am_closed: true,
      pm_closed: false,
      is_holiday: false,
    },
  ],
};

function renderPage() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={["/accounting/reports"]}>
        <AccountingReportsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("AccountingReportsPage 印刷 / PDF出力 (#184) + 操作UI除外 (#179 ④-a)", () => {
  beforeEach(() => {
    server.use(http.get("*/v1/reports/monthly", () => HttpResponse.json(MONTHLY)));
  });

  it("#184: データあり時に「印刷 / PDF出力」ボタンが表示される", async () => {
    renderPage();
    expect(await screen.findByTestId("monthly-report-print-button")).toBeInTheDocument();
  });

  it("#179 ④-a: 操作UI(ヘッダ)に print:hidden が付与され印刷時に出力されない", async () => {
    renderPage();
    const actions = await screen.findByTestId("report-actions");
    // JSDOM は @media print を評価できないため、印刷除外を class ベースで固定化する
    expect(actions.classList.contains("print:hidden")).toBe(true);
    // CSV出力ボタンが除外対象コンテナの内側にあること（= 印刷面から除外される）
    expect(within(actions).getByText("CSV出力")).toBeInTheDocument();
  });

  it("#184: 印刷エリアに 病院名・対象年月・日次明細 が含まれる", async () => {
    renderPage();
    const printArea = await screen.findByTestId("monthly-report-print-area");
    expect(within(printArea).getByText("テスト動物病院")).toBeInTheDocument();
    expect(printArea.textContent).toContain("2026年5月");
    expect(within(printArea).getByText("2026-05-01")).toBeInTheDocument();
  });

  it("#184: 印刷値が画面サマリー値と一致する（売上合計）", async () => {
    renderPage();
    const printArea = await screen.findByTestId("monthly-report-print-area");
    // 印刷エリアに売上合計が出る
    expect(within(printArea).getByText("¥1,234,567")).toBeInTheDocument();
    // 画面サマリーカードにも同一表記で出る（同一データ源 = 差分0）
    const allOccurrences = screen.getAllByText("¥1,234,567");
    expect(allOccurrences.length).toBeGreaterThanOrEqual(2);
  });

  it("#184: データ取得失敗時は印刷ボタンが表示されない", async () => {
    server.use(
      http.get("*/v1/reports/monthly", () => new HttpResponse(null, { status: 500 })),
    );
    renderPage();
    await waitFor(() => {
      expect(screen.getByText("データの取得に失敗しました")).toBeInTheDocument();
    });
    expect(screen.queryByTestId("monthly-report-print-button")).not.toBeInTheDocument();
    expect(screen.queryByTestId("monthly-report-print-area")).not.toBeInTheDocument();
  });
});
