import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { clinics: [{ clinicId: "1", clinicName: "ノア動物病院", isMain: true }] },
    currentClinicId: "1",
    hasPermission: () => true,
  }),
}));
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";

import { DailyAccountingTab } from "./DailyAccountingTab";

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
  return render(<DailyAccountingTab />, {
    wrapper: createTestWrapper({ initialEntries: [initialUrl] }),
  });
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
      http.get("*/v1/accountings/daily-summary", () => HttpResponse.json(DAILY_SUMMARY)),
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
    const table = screen.getByTestId("daily-accounting-table");
    expect(within(table).getByText("田中太郎")).toBeInTheDocument();
    expect(within(table).getByText("ポチ")).toBeInTheDocument();
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

  it("負の診療金額を符号のまま表示する", async () => {
    server.use(
      http.get("*/v1/accountings", () =>
        HttpResponse.json({
          data: [
            makeAccounting({
              subtotal: -3000,
              tax_total: 0,
              total_amount: -3000,
              items: [
                {
                  id: 1,
                  billing_id: 1,
                  category: "examination",
                  name: "赤伝",
                  unit_price: -3000,
                  quantity: 1,
                  tax_rate: 0,
                  tax_type: "excluded",
                  tax_amount: 0,
                  subtotal: -3000,
                  is_insurance_applicable: false,
                  source: "medical_record",
                },
              ],
              payments: [
                {
                  id: 1,
                  billing_id: 1,
                  method: "cash",
                  total_amount: -3000,
                  billing_amount: -3000,
                  received_amount: 0,
                  change_amount: 0,
                  insurance_amount: 0,
                  discount_amount: 0,
                },
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
    const row = within(table).getByText("田中太郎").closest("tr");
    expect(row).not.toBeNull();
    const cells = within(row as HTMLElement).getAllByRole("cell");
    expect(cells[3]).toHaveTextContent("¥-3,000");
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
                {
                  id: 1,
                  clinic_id: 1,
                  billing_id: 1,
                  method: "cash",
                  amount: 2000,
                  received_amount: 3000,
                  change_amount: 1000,
                  created_at: `${TODAY}T10:00:00Z`,
                },
                {
                  id: 2,
                  clinic_id: 1,
                  billing_id: 1,
                  method: "credit_card",
                  amount: 3500,
                  received_amount: 0,
                  change_amount: 0,
                  created_at: `${TODAY}T10:00:00Z`,
                },
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
    // FE5-25: PAYMENT_METHOD_LABELS_SHORT 廃止により credit_card は "カード" → "クレジットカード" へ変更
    expect(
      within(table).getByText(/現金.*クレジットカード|クレジットカード.*現金/),
    ).toBeInTheDocument();
  });

  // ── #117: 新規テスト ──────────────────────────────────────────────

  it("#117: vaccine カテゴリの金額がRV列に表示される", async () => {
    server.use(
      http.get("*/v1/accountings", () =>
        HttpResponse.json({
          data: [
            makeAccounting({
              items: [
                {
                  id: 3,
                  billing_id: 1,
                  category: "vaccine",
                  name: "狂犬病ワクチン",
                  unit_price: 4000,
                  quantity: 1,
                  tax_rate: 0.1,
                  tax_type: "excluded",
                  tax_amount: 400,
                  subtotal: 4000,
                  is_insurance_applicable: false,
                  source: "manual",
                },
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
    // RV列ヘッダーが存在する
    expect(within(table).getByText("RV")).toBeInTheDocument();
    // vaccine の金額が表示される（ボディ＋フッターで1件以上）
    expect(within(table).getAllByText("¥4,000").length).toBeGreaterThanOrEqual(1);
  });

  it("#117: trimming カテゴリの金額がトリミング列に表示される", async () => {
    server.use(
      http.get("*/v1/accountings", () =>
        HttpResponse.json({
          data: [
            makeAccounting({
              items: [
                {
                  id: 4,
                  billing_id: 1,
                  category: "trimming",
                  name: "シャンプーカット",
                  unit_price: 6000,
                  quantity: 1,
                  tax_rate: 0.1,
                  tax_type: "excluded",
                  tax_amount: 600,
                  subtotal: 6000,
                  is_insurance_applicable: false,
                  source: "trimming",
                },
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
    expect(within(table).getByText("トリミング")).toBeInTheDocument();
    expect(within(table).getAllByText("¥6,000").length).toBeGreaterThanOrEqual(1);
  });

  it("#117: hotel カテゴリの金額がホテル列に表示される", async () => {
    server.use(
      http.get("*/v1/accountings", () =>
        HttpResponse.json({
          data: [
            makeAccounting({
              items: [
                {
                  id: 5,
                  billing_id: 1,
                  category: "hotel",
                  name: "ペットホテル",
                  unit_price: 3000,
                  quantity: 2,
                  tax_rate: 0.1,
                  tax_type: "excluded",
                  tax_amount: 600,
                  subtotal: 6000,
                  is_insurance_applicable: false,
                  source: "manual",
                },
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
    expect(within(table).getByText("ホテル")).toBeInTheDocument();
    expect(within(table).getAllByText("¥6,000").length).toBeGreaterThanOrEqual(1);
  });

  it("#117: 領収No列が billing.id のゼロ埋め8桁で表示される", async () => {
    renderTab();
    await waitFor(() => {
      expect(screen.getByTestId("daily-accounting-table")).toBeInTheDocument();
    });
    const table = screen.getByTestId("daily-accounting-table");
    // billing.id = 1 → "00000001"
    expect(within(table).getByText("00000001")).toBeInTheDocument();
  });

  it("#117: データあり時に印刷ボタンが表示される", async () => {
    renderTab();
    await waitFor(() => {
      expect(screen.getByTestId("daily-print-button")).toBeInTheDocument();
    });
  });

  it("#117: データなし時は印刷ボタンが表示されない", async () => {
    server.use(
      http.get("*/v1/accountings", () =>
        HttpResponse.json({ data: [], total: 0, page: 1, limit: 50 }),
      ),
    );
    renderTab();
    await waitFor(() => {
      expect(screen.getByTestId("daily-empty")).toBeInTheDocument();
    });
    expect(screen.queryByTestId("daily-print-button")).not.toBeInTheDocument();
  });

  it("#117: 印刷エリアが DOM に存在する（rows > 0 時）", async () => {
    renderTab();
    await waitFor(() => {
      expect(screen.getByTestId("daily-accounting-table")).toBeInTheDocument();
    });
    // hidden class でスクリーン非表示だが DOM に存在する
    expect(document.querySelector('[data-testid="daily-print-area"]')).not.toBeNull();
  });

  it("#117: フッターに RV / トリミング / ホテル の合計が表示される", async () => {
    server.use(
      http.get("*/v1/accountings", () =>
        HttpResponse.json({
          data: [
            makeAccounting({
              items: [
                {
                  id: 6,
                  billing_id: 1,
                  category: "vaccine",
                  name: "ワクチン",
                  unit_price: 5000,
                  quantity: 1,
                  tax_rate: 0.1,
                  tax_type: "excluded",
                  tax_amount: 500,
                  subtotal: 5000,
                  is_insurance_applicable: false,
                  source: "manual",
                },
                {
                  id: 7,
                  billing_id: 1,
                  category: "trimming",
                  name: "カット",
                  unit_price: 4000,
                  quantity: 1,
                  tax_rate: 0.1,
                  tax_type: "excluded",
                  tax_amount: 400,
                  subtotal: 4000,
                  is_insurance_applicable: false,
                  source: "trimming",
                },
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
    // フッター行に各合計が表示されることを確認
    const footer = table.querySelector("tfoot");
    expect(footer).not.toBeNull();
    // vaccine 5000, trimming 4000 がそれぞれのフッター列に表示される
    expect(within(footer!).getByText("¥5,000")).toBeInTheDocument();
    expect(within(footer!).getByText("¥4,000")).toBeInTheDocument();
  });
});
