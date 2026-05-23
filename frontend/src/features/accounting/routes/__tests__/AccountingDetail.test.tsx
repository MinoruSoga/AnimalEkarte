import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Routes, Route } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { AuthContext } from "@/contexts/auth-context";
import { AccountingDetail } from "../AccountingDetail";
import type { ResourceAction } from "@/types/auth";

const CLINIC_ID = "clinic-test-1";
const ACCOUNTING_ID = "123";

// hasPermission factory: canEdit を制御する
function makeHasPermission(canEdit: boolean) {
  return (_resource: string, action: ResourceAction): boolean => {
    if (action === "edit") return canEdit;
    return true;
  };
}

function makeAuthCtx(canEdit: boolean) {
  return {
    user: null,
    currentClinicId: CLINIC_ID,
    isAuthenticated: true,
    isLoading: false,
    isSwitchingClinic: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: makeHasPermission(canEdit),
    refreshPermissions: async () => {},
  };
}

// status=completed + payment あり (印刷ボタン表示に必要)
const completedAccounting = {
  id: 123,
  clinic_id: 1,
  status: "completed",
  scheduled_date: "2026-05-01",
  subtotal: 1000,
  tax_total: 100,
  total_amount: 1100,
  has_insurance: false,
  memo: "",
  created_at: "2026-05-01T00:00:00Z",
  updated_at: "2026-05-01T00:00:00Z",
  total_refunded_amount: 0,
  owner: { name: "テスト飼い主" },
  pet: { name: "テストペット" },
  items: [],
  payments: [
    {
      id: 1,
      billing_id: 123,
      subtotal: 1000,
      tax_total: 100,
      total_amount: 1100,
      insurance_name: "",
      insurance_ratio: 0,
      insurance_amount: 0,
      discount_amount: 0,
      billing_amount: 1100,
      received_amount: 1100,
      change_amount: 0,
      method: "cash",
      created_at: "2026-05-01T00:00:00Z",
      updated_at: "2026-05-01T00:00:00Z",
    },
  ],
};

function setupHandlers() {
  server.use(
    http.get(`/api/v1/accountings/${ACCOUNTING_ID}`, () =>
      HttpResponse.json(completedAccounting)
    ),
    http.get(`/api/v1/accountings/${ACCOUNTING_ID}/refunds`, () =>
      HttpResponse.json([])
    ),
    http.get("/api/v1/masters/merchandise-items", () =>
      HttpResponse.json([])
    )
  );
}

// id あり: /accounting/:id ルートで描画
async function renderWithIdAndWait(canEdit = true) {
  setupHandlers();
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <AuthContext.Provider value={makeAuthCtx(canEdit)}>
      <QueryClientProvider client={qc}>
        <MemoryRouter initialEntries={[`/accounting/${ACCOUNTING_ID}`]}>
          <Routes>
            <Route path="/accounting/:id" element={<AccountingDetail />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    </AuthContext.Provider>
  );
  await waitFor(() => {
    expect(screen.getByRole("heading", { name: "会計精算" })).toBeInTheDocument();
  });
}

// 新規作成モード: id なし (/accounting/new はパラメータなし)
async function renderNewModeAndWait(canEdit = false) {
  server.use(
    http.get("/api/v1/masters/merchandise-items", () => HttpResponse.json([]))
  );
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <AuthContext.Provider value={makeAuthCtx(canEdit)}>
      <QueryClientProvider client={qc}>
        <MemoryRouter initialEntries={["/accounting/new"]}>
          <Routes>
            <Route path="/accounting/new" element={<AccountingDetail />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    </AuthContext.Provider>
  );
  await waitFor(() => {
    expect(screen.getByRole("heading", { name: "会計精算" })).toBeInTheDocument();
  });
}

beforeEach(() => {
  localStorage.setItem("auth_current_clinic:v1", CLINIC_ID);
});

afterEach(() => {
  localStorage.removeItem("auth_current_clinic:v1");
  vi.restoreAllMocks();
  // Radix UI Dialog sets these on body imperatively; React cleanup() doesn't undo them
  document.body.removeAttribute("data-scroll-locked");
  document.body.style.removeProperty("pointer-events");
});

// ─────────────────────────────────────────────────────────────
// A: 印刷ボタン（Print Performance）
// ─────────────────────────────────────────────────────────────

describe("AccountingDetail — A: 印刷ボタン (Print Performance)", () => {
  it("status=completed → 「明細兼領収書」ボタンが表示される", async () => {
    await renderWithIdAndWait();
    expect(
      screen.getByRole("button", { name: /明細兼領収書/ })
    ).toBeInTheDocument();
  });

  it("「明細兼領収書」クリック → プレビューダイアログが開く", async () => {
    await renderWithIdAndWait();
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /明細兼領収書/ }));
    await waitFor(() => {
      expect(screen.getByText("明細兼領収書プレビュー")).toBeInTheDocument();
    });
    // afterEach removes Radix UI body side-effects (pointer-events:none / data-scroll-locked)
  });

  it("ダイアログ内「印刷する」クリック → window.print() が呼ばれる", async () => {
    const printSpy = vi.spyOn(window, "print").mockImplementation(() => {});
    await renderWithIdAndWait();
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /明細兼領収書/ }));
    await screen.findByText("明細兼領収書プレビュー");
    // Radix UI sets pointer-events:none on body while dialog is open;
    // fireEvent bypasses pointer-events and dispatches the click directly.
    const printBtn = await screen.findByRole("button", { name: /印刷する/, hidden: true });
    printBtn.click();
    expect(printSpy).toHaveBeenCalledOnce();
  });
});

// ─────────────────────────────────────────────────────────────
// C: 混在支払い UI / payment_splits — 共通フィクスチャ
// ─────────────────────────────────────────────────────────────

const WAITING_ID = "456";

const waitingAccounting = {
  id: 456,
  clinic_id: 1,
  status: "waiting",
  scheduled_date: "2026-05-01",
  subtotal: 1000,
  tax_total: 100,
  total_amount: 1100,
  has_insurance: false,
  memo: "",
  created_at: "2026-05-01T00:00:00Z",
  updated_at: "2026-05-01T00:00:00Z",
  total_refunded_amount: 0,
  owner: { name: "テスト飼い主" },
  pet: { name: "テストペット" },
  items: [
    {
      id: 1,
      billing_id: 456,
      unit_price: 1000,
      quantity: 1,
      tax_rate: 0.1,
      tax_amount: 100,
      subtotal: 1000,
      name: "テスト商品",
      category: "other",
      tax_type: "excluded",
      is_insurance_applicable: false,
      source: "manual",
    },
  ],
  payments: [],
  payment_splits: [],
};

function setupWaitingHandlers() {
  server.use(
    http.get(`/api/v1/accountings/${WAITING_ID}`, () =>
      HttpResponse.json(waitingAccounting)
    ),
    http.get(`/api/v1/accountings/${WAITING_ID}/refunds`, () =>
      HttpResponse.json([])
    ),
    http.get("/api/v1/masters/merchandise-items", () =>
      HttpResponse.json([])
    )
  );
}

async function renderWaitingAndWait() {
  setupWaitingHandlers();
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <AuthContext.Provider value={makeAuthCtx(true)}>
      <QueryClientProvider client={qc}>
        <MemoryRouter initialEntries={[`/accounting/${WAITING_ID}`]}>
          <Routes>
            <Route path="/accounting/:id" element={<AccountingDetail />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    </AuthContext.Provider>
  );
  await waitFor(() => {
    expect(screen.getByRole("heading", { name: "会計精算" })).toBeInTheDocument();
  });
}

// ─────────────────────────────────────────────────────────────
// B: 閲覧専用バナー（ReadOnly banner）
// ─────────────────────────────────────────────────────────────

describe("AccountingDetail — B: 閲覧専用バナー (ReadOnly banner)", () => {
  it("id あり + canEdit=false → role=status のバナーが表示される", async () => {
    await renderWithIdAndWait(false);
    const banner = screen.getByRole("status");
    expect(banner).toBeInTheDocument();
  });

  it("id あり + canEdit=false → 「閲覧専用」テキストがバナーに表示される", async () => {
    await renderWithIdAndWait(false);
    expect(
      screen.getByText(/閲覧専用 — 編集権限がないため変更できません/)
    ).toBeInTheDocument();
  });

  it("新規作成モード (id なし) + canEdit=false → バナーが表示されない", async () => {
    await renderNewModeAndWait(false);
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// C: 混在支払い UI / payment_splits
// ─────────────────────────────────────────────────────────────

describe("AccountingDetail — C: 混在支払い UI / payment_splits", () => {
  it("canEdit=true + status=waiting → 「支払方法を追加」ボタンが表示される", async () => {
    await renderWaitingAndWait();
    expect(
      screen.getByRole("button", { name: /支払方法を追加/ })
    ).toBeInTheDocument();
  });

  it("「支払方法を追加」クリック → 2スロット分の金額入力が表示される（spinbutton × 4）", async () => {
    await renderWaitingAndWait();
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /支払方法を追加/ }));
    await waitFor(() => {
      expect(screen.getAllByRole("spinbutton")).toHaveLength(4);
    });
  });

  it("現金スロットで received > amount → お釣りが計算される", async () => {
    await renderWaitingAndWait();
    const user = userEvent.setup();
    const [amountInput, receivedInput] = screen.getAllByRole("spinbutton");
    await user.clear(amountInput);
    await user.type(amountInput, "1100");
    await user.clear(receivedInput);
    await user.type(receivedInput, "1200");
    await waitFor(() => {
      const changeLabel = screen.getByText("お釣り");
      expect(changeLabel.nextElementSibling).toHaveTextContent("¥100");
    });
  });

  it("デフォルト状態（未入力）→ 「会計を確定する」ボタンが disabled", async () => {
    await renderWaitingAndWait();
    expect(
      screen.getByRole("button", { name: /会計を確定する/ })
    ).toBeDisabled();
  });

  it("amount=1100 / received=1100 入力後 submit → payment_splits を含む payload が送信される", async () => {
    let capturedBody: unknown;
    setupWaitingHandlers();
    server.use(
      http.patch(`/api/v1/accountings/${WAITING_ID}`, async ({ request }) => {
        capturedBody = await request.json();
        return HttpResponse.json({ ...waitingAccounting, status: "completed" });
      })
    );
    const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <AuthContext.Provider value={makeAuthCtx(true)}>
        <QueryClientProvider client={qc}>
          <MemoryRouter initialEntries={[`/accounting/${WAITING_ID}`]}>
            <Routes>
              <Route path="/accounting/:id" element={<AccountingDetail />} />
            </Routes>
          </MemoryRouter>
        </QueryClientProvider>
      </AuthContext.Provider>
    );
    await waitFor(() =>
      expect(screen.getByRole("heading", { name: "会計精算" })).toBeInTheDocument()
    );

    const user = userEvent.setup();
    const [amountInput, receivedInput] = screen.getAllByRole("spinbutton");
    await user.clear(amountInput);
    await user.type(amountInput, "1100");
    await user.clear(receivedInput);
    await user.type(receivedInput, "1100");

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: /会計を確定する/ })
      ).not.toBeDisabled();
    });

    await user.click(screen.getByRole("button", { name: /会計を確定する/ }));

    await waitFor(() => expect(capturedBody).toBeDefined());

    expect(capturedBody).toMatchObject({
      payment_splits: [
        {
          method: "cash",
          amount: 1100,
          received_amount: 1100,
          change_amount: 0,
        },
      ],
    });
  });
});
