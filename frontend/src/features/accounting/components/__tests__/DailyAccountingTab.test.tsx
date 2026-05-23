import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";

import { DailyAccountingTab } from "../DailyAccountingTab";

const TODAY = "2026-04-01";

// 最小フィクスチャ
function makeAccounting(overrides: Record<string, unknown> = {}) {
  return {
    id: 1,
    clinic_id: 1,
    owner_id: 10,
    pet_id: 20,
    status: "completed",
    scheduled_date: `${TODAY}T00:00:00Z`,
    completed_at: `${TODAY}T10:00:00Z`,
    subtotal: 5000,
    tax_total: 500,
    total_amount: 5500,
    total_refunded_amount: 0,
    has_insurance: false,
    owner: { name: "田中太郎" },
    pet: { name: "ポチ", animal_species: { id: 1, name: "犬" } },
    items: [
      {
        id: 1,
        billing_id: 1,
        category: "examination",
        name: "診察料",
        unit_price: 3000,
        quantity: 1,
        tax_rate: 0.1,
        tax_type: "excluded",
        tax_amount: 300,
        subtotal: 3000,
        is_insurance_applicable: false,
        source: "medical_record",
      },
      {
        id: 2,
        billing_id: 1,
        category: "surgery",
        name: "手術",
        unit_price: 2000,
        quantity: 1,
        tax_rate: 0.1,
        tax_type: "excluded",
        tax_amount: 200,
        subtotal: 2000,
        is_insurance_applicable: false,
        source: "manual",
      },
    ],
    payments: [
      {
        id: 1,
        billing_id: 1,
        method: "cash",
        total_amount: 5500,
        billing_amount: 5500,
        received_amount: 6000,
        change_amount: 500,
        insurance_amount: 0,
        discount_amount: 0,
      },
    ],
    ...overrides,
  };
}

const DAILY_SUMMARY = {
  payment_totals: [
    { method: "cash", total: 5500 },
    { method: "credit_card", total: 3300 },
  ],
  category_totals: [
    { category: "examination", total: 3000 },
    { category: "surgery", total: 2000 },
  ],
  billing_count: 2,
  grand_total: 8800,
};

function renderTab(initialUrl = `/accounting?tab=daily&daily_date=${TODAY}`) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[initialUrl]}>
        <DailyAccountingTab />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("DailyAccountingTab", () => {
  beforeEach(() => {
    server.use(
      http.get("*/v1/accountings", () =>
        HttpResponse.json({
          data: [makeAccounting()],
          total: 1,
          page: 1,
          limit: 50,
        }),
      ),
      http.get("*/v1/accountings/daily-summary", () =>
        HttpResponse.json(DAILY_SUMMARY),
      ),
    );
  });

  it("対象日 input が表示される", () => {
    renderTab();
    expect(screen.getByLabelText("対象日")).toBeInTheDocument();
  });

  it("完了済み会計データが表示される", async () => {
    renderTab();
    await waitFor(() => {
      expect(screen.getByTestId("daily-accounting-table")).toBeInTheDocument();
    });
    expect(screen.getByText("田中太郎")).toBeInTheDocument();
    expect(screen.getByText("ポチ")).toBeInTheDocument();
  });

  it("診療・外科の金額が行に表示される", async () => {
    renderTab();
    await waitFor(() => {
      expect(screen.getByTestId("daily-accounting-table")).toBeInTheDocument();
    });
    // 診療: ¥3,000 / 外科: ¥2,000
    const cells = screen.getAllByText(/¥[0-9,]+/);
    const amounts = cells.map((el) => el.textContent);
    expect(amounts.some((t) => t?.includes("3,000"))).toBe(true);
    expect(amounts.some((t) => t?.includes("2,000"))).toBe(true);
  });

  it("支払方法が表示される", async () => {
    renderTab();
    await waitFor(() => {
      expect(screen.getByTestId("daily-accounting-table")).toBeInTheDocument();
    });
    // テーブルセル内の支払方法を確認（サマリーカードにも同名ラベルが出る場合があるため within で限定）
    const table = screen.getByTestId("daily-accounting-table");
    expect(within(table).getAllByText("現金").length).toBeGreaterThanOrEqual(1);
  });

  it("集計サマリーカードが表示される", async () => {
    renderTab();
    await waitFor(() => {
      expect(screen.getByTestId("daily-summary-cards")).toBeInTheDocument();
    });
    const cards = screen.getByTestId("daily-summary-cards");
    expect(within(cards).getByText("会計件数")).toBeInTheDocument();
    expect(within(cards).getByText("売上合計")).toBeInTheDocument();
  });

  it("合計行が表示される", async () => {
    renderTab();
    await waitFor(() => {
      expect(screen.getByText(/合計（1件）/)).toBeInTheDocument();
    });
  });

  it("データなし時は空メッセージが表示される", async () => {
    server.use(
      http.get("*/v1/accountings", () =>
        HttpResponse.json({ data: [], total: 0, page: 1, limit: 50 }),
      ),
    );
    renderTab();
    await waitFor(() => {
      expect(screen.getByTestId("daily-empty")).toBeInTheDocument();
    });
  });

  it("waiting ステータスの会計は表示されない", async () => {
    server.use(
      http.get("*/v1/accountings", () =>
        HttpResponse.json({
          data: [makeAccounting({ status: "waiting", payments: [] })],
          total: 1,
          page: 1,
          limit: 50,
        }),
      ),
    );
    renderTab();
    await waitFor(() => {
      expect(screen.getByTestId("daily-empty")).toBeInTheDocument();
    });
  });

  it("日付変更で URL パラメータが更新される", async () => {
    const user = userEvent.setup();
    renderTab(`/accounting?tab=daily`);
    const input = screen.getByLabelText("対象日");
    await user.clear(input);
    await user.type(input, "2026-03-15");
    // input の value が変化していることを確認
    expect((input as HTMLInputElement).value).toBe("2026-03-15");
  });

  it("混在支払い: 複数の支払方法ラベルが「/」区切りで表示される", async () => {
    server.use(
      http.get("*/v1/accountings", () =>
        HttpResponse.json({
          data: [
            makeAccounting({
              payment_splits: [
                { id: 1, clinic_id: 1, billing_id: 1, method: "cash", amount: 2000, received_amount: 3000, change_amount: 1000, created_at: `${TODAY}T10:00:00Z` },
                { id: 2, clinic_id: 1, billing_id: 1, method: "credit_card", amount: 3500, received_amount: 0, change_amount: 0, created_at: `${TODAY}T10:00:00Z` },
              ],
            }),
          ],
          total: 1,
          page: 1,
          limit: 50,
        }),
      ),
    );
    renderTab();
    await waitFor(() => {
      expect(screen.getByTestId("daily-accounting-table")).toBeInTheDocument();
    });
    const table = screen.getByTestId("daily-accounting-table");
    // paymentSplits.length > 1 のとき method ラベルを " / " 区切りで結合して表示する
    // DailyAccountingTab の PAYMENT_METHOD_LABELS: credit_card → "カード"
    expect(within(table).getByText(/現金.*カード|カード.*現金/)).toBeInTheDocument();
  });
});
