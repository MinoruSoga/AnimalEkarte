import { describe, it, expect, beforeEach, vi } from "vitest";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";

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

import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";

import { AccountingReportsPage } from "./AccountingReportsPage";

// Backend raw shape（transform 前 / snake_case）。distinctive な金額で
// 「画面値 == 印刷値」を検証しやすくする。
const MONTHLY = {
  year: 2026,
  month: 5,
  start_date: "2026-05-01",
  end_date: "2026-05-31",
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
  return render(<AccountingReportsPage />, {
    wrapper: createTestWrapper({ initialEntries: ["/accounting/reports"] }),
  });
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

  it("画面上の操作と集計条件を44px以上に保つ", async () => {
    const { container } = renderPage();
    const actions = await screen.findByTestId("report-actions");

    for (const button of within(actions).getAllByRole("button")) {
      expect(button).toHaveClass("min-h-11");
    }
    for (const select of Array.from(container.querySelectorAll("select"))) {
      expect(select).toHaveClass("min-h-11");
    }
    expect(await screen.findByRole("link", { name: /税率設定を変更/ })).toHaveClass("min-h-11");
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
    server.use(http.get("*/v1/reports/monthly", () => new HttpResponse(null, { status: 500 })));
    renderPage();
    await waitFor(() => {
      expect(screen.getByText("データの取得に失敗しました")).toBeInTheDocument();
    });
    expect(screen.queryByTestId("monthly-report-print-button")).not.toBeInTheDocument();
    expect(screen.queryByTestId("monthly-report-print-area")).not.toBeInTheDocument();
  });
});

describe("AccountingReportsPage カード配置 (#179 ④-b) + 税率設定導線 (#179 ②)", () => {
  beforeEach(() => {
    server.use(http.get("*/v1/reports/monthly", () => HttpResponse.json(MONTHLY)));
  });

  it("#179 ④-b: 集計数字カード(KPI)が日次明細セクションより後ろ(ページ下部)に配置される", async () => {
    renderPage();
    const printArea = await screen.findByTestId("monthly-report-print-area");
    // 印刷ポータル(body直下)にも同名テキストが出るため、画面側(ポータル外)のみを対象にする
    const screenKpi = screen.getAllByText("診療日数").find((el) => !printArea.contains(el));
    const screenDailyHeading = screen
      .getAllByText(/日次明細/)
      .find((el) => !printArea.contains(el));
    expect(screenKpi).toBeDefined();
    expect(screenDailyHeading).toBeDefined();
    // KPI(診療日数) が 日次明細見出し より DOM 上で後ろにある = 下部配置
    const relativePosition = screenDailyHeading!.compareDocumentPosition(screenKpi!);
    expect(relativePosition & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it("#179 ②: 税率設定(/settings/clinic)への導線が表示される", async () => {
    renderPage();
    const link = await screen.findByRole("link", { name: /税率設定を変更/ });
    expect(link).toHaveAttribute("href", "/settings/clinic");
  });

  it("#179 ④-c: 期間指定では start_date/end_date で月次レポートAPIを再取得する", async () => {
    const requestUrls: string[] = [];
    server.use(
      http.get("*/v1/reports/monthly", ({ request }) => {
        requestUrls.push(request.url);
        return HttpResponse.json(MONTHLY);
      }),
    );
    renderPage();
    await screen.findByTestId("monthly-report-print-area");

    fireEvent.change(screen.getByLabelText("集計単位"), { target: { value: "period" } });
    fireEvent.change(screen.getByLabelText("開始日"), { target: { value: "2026-04-15" } });
    fireEvent.change(screen.getByLabelText("終了日"), { target: { value: "2026-05-20" } });

    await waitFor(() => {
      expect(
        requestUrls.some((url) => {
          const parsed = new URL(url);
          return (
            parsed.searchParams.get("start_date") === "2026-04-15" &&
            parsed.searchParams.get("end_date") === "2026-05-20"
          );
        }),
      ).toBe(true);
    });
  });
});
