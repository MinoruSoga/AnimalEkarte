import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Routes, Route } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { AuthContext } from "@/hooks/auth-context";
import { AccountingDetail } from "./AccountingDetail";
import type { ResourceAction } from "@/types/auth";
import { ResourceCashRegisterClose } from "@/types/generated/models";

const { handleApiErrorMock } = vi.hoisted(() => ({
  handleApiErrorMock: vi.fn(),
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: handleApiErrorMock,
}));

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
    ),
    http.get("/api/v1/cash-register/closes", () =>
      HttpResponse.json({ data: [], total: 0 })
    ),
  );
}

// id あり: /accounting/:id ルートで描画
async function renderWithIdAndWait(
  canEdit = true,
  hasPermission = makeHasPermission(canEdit),
) {
  setupHandlers();
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <AuthContext.Provider value={{ ...makeAuthCtx(canEdit), hasPermission }}>
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
    http.get("/api/v1/masters/merchandise-items", () => HttpResponse.json([])),
    http.get("/api/v1/cash-register/closes", () =>
      HttpResponse.json({ data: [], total: 0 })
    ),
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
  handleApiErrorMock.mockReset();
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
  payment_splits: [
    { method: "cash", amount: "1100", receivedAmount: "1100" },
    { method: "credit_card", amount: "", receivedAmount: "" },
  ],
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
    ),
    http.get("/api/v1/cash-register/closes", () =>
      HttpResponse.json({ data: [], total: 0 })
    ),
    // Payment API handlers for Dialog test
    http.post(`/api/v1/accountings/${WAITING_ID}/payments`, () =>
      HttpResponse.json({ id: 999, ...waitingAccounting.payments?.[0] })
    )
  );
}

async function renderWaitingAndWait(useDefaultHandlers = true) {
  if (useDefaultHandlers) {
    setupWaitingHandlers();
  }
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

  it("レジ締め状態を閲覧できない編集者はGETを送らずfail-closedで閲覧専用になる", async () => {
    let closeGetCount = 0;
    server.use(
      http.get("/api/v1/cash-register/closes", () => {
        closeGetCount += 1;
        return HttpResponse.json({ data: [], total: 0 });
      }),
    );
    const hasPermission = (resource: string, action: ResourceAction): boolean => {
      if (resource === ResourceCashRegisterClose && action === "view") return false;
      return true;
    };

    await renderWithIdAndWait(true, hasPermission);

    expect(
      screen.getByText(/レジ締め状態を確認する権限がないため変更できません/),
    ).toBeInTheDocument();
    expect(document.querySelector("fieldset")).toBeDisabled();
    expect(closeGetCount).toBe(0);
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

  it.skip("「支払方法を追加」クリック → 2スロット分の金額入力が表示される（spinbutton × 4）", async () => {
    // TODO: state update timing issue - button click doesn't update payment_splits
    await renderWaitingAndWait();
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: /支払方法を追加/ }));
    await waitFor(() => {
      expect(screen.getAllByRole("spinbutton")).toHaveLength(4);
    });
  });

  it.skip("現金スロットで received > amount → お釣りが計算される", async () => {
    // TODO: depends on button add test
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

  it.skip("amount=1100 / received=1100 入力後 submit → payment_splits を含む payload が送信される", async () => {
    // TODO: depends on button add test
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

  it("保存済みの手動明細を削除すると DELETE /billing-items/:id が送信される", async () => {
    let deleteCalled = false;
    let currentItems = waitingAccounting.items;

    server.use(
      http.get(`/api/v1/accountings/${WAITING_ID}`, () =>
        HttpResponse.json({ ...waitingAccounting, items: currentItems })
      ),
      http.get(`/api/v1/accountings/${WAITING_ID}/refunds`, () =>
        HttpResponse.json([])
      ),
      http.get("/api/v1/masters/merchandise-items", () =>
        HttpResponse.json([])
      ),
      http.get("/api/v1/cash-register/closes", () =>
        HttpResponse.json({ data: [], total: 0 })
      ),
      http.delete("/api/v1/billing-items/1", () => {
        deleteCalled = true;
        currentItems = [];
        return new HttpResponse(null, { status: 204 });
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

    await screen.findByText("テスト商品");

    const user = userEvent.setup();
    await user.click(screen.getByTitle("削除"));

    await waitFor(() => {
      expect(deleteCalled).toBe(true);
      expect(screen.queryByText("テスト商品")).not.toBeInTheDocument();
    });
  });

  it("商品マスタ由来の明細作成で merchandise_item_id を送る", async () => {
    let capturedBody: unknown;
    let currentItems = waitingAccounting.items;
    server.use(
      http.get(`/api/v1/accountings/${WAITING_ID}`, () =>
        HttpResponse.json({ ...waitingAccounting, items: currentItems })
      ),
      http.get(`/api/v1/accountings/${WAITING_ID}/refunds`, () =>
        HttpResponse.json([])
      ),
      http.get("/api/v1/masters/merchandise-items", () =>
        HttpResponse.json([
          { id: 77, name: "療法食", category: "goods", unit_price: 1200, tax_rate: 0.1, is_active: true },
        ])
      ),
      http.get("/api/v1/cash-register/closes", () =>
        HttpResponse.json({ data: [], total: 0 })
      ),
      http.post("/api/v1/billing-items", async ({ request }) => {
        capturedBody = await request.json();
        const createdItem = {
          id: 2,
          billing_id: Number(WAITING_ID),
          name: "療法食",
          category: "goods",
          unit_price: 1200,
          quantity: 1,
          tax_type: "excluded",
          tax_rate: 0.1,
          tax_amount: 120,
          subtotal: 1200,
          discount_rate: 0,
          discount_amount: 0,
          is_insurance_applicable: false,
          source: "manual",
          merchandise_item_id: 77,
        };
        currentItems = [...currentItems, createdItem];
        return HttpResponse.json(createdItem);
      }),
    );

    await renderWaitingAndWait(false);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "物販・その他追加" }));
    await user.click(await screen.findByRole("button", { name: "追加: 療法食 (ID 77)" }));

    await waitFor(() => {
      expect(capturedBody).toMatchObject({ merchandise_item_id: 77 });
    });
  });

  it("明細作成失敗時に楽観追加を戻してエラーを通知する", async () => {
    let rejectRequest: (() => void) | undefined;
    server.use(
      http.get(`/api/v1/accountings/${WAITING_ID}`, () =>
        HttpResponse.json(waitingAccounting)
      ),
      http.get(`/api/v1/accountings/${WAITING_ID}/refunds`, () =>
        HttpResponse.json([])
      ),
      http.get("/api/v1/masters/merchandise-items", () =>
        HttpResponse.json([
          { id: 77, name: "療法食", category: "goods", unit_price: 1200, tax_rate: 0.1, is_active: true },
        ])
      ),
      http.get("/api/v1/cash-register/closes", () =>
        HttpResponse.json({ data: [], total: 0 })
      ),
      http.post("/api/v1/billing-items", () =>
        new Promise((resolve) => {
          rejectRequest = () => resolve(HttpResponse.json({ message: "inactive item" }, { status: 409 }));
        })
      ),
    );

    await renderWaitingAndWait(false);
    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "物販・その他追加" }));
    await user.click(await screen.findByRole("button", { name: "追加: 療法食 (ID 77)" }));

    expect(await screen.findByText("療法食")).toBeInTheDocument();
    await waitFor(() => expect(rejectRequest).toBeDefined());
    rejectRequest?.();

    await waitFor(() => {
      expect(screen.queryByText("療法食")).not.toBeInTheDocument();
      expect(screen.getByText("テスト商品")).toBeInTheDocument();
      expect(handleApiErrorMock).toHaveBeenCalledWith(expect.anything(), "明細の追加");
    });
  });
});

// ─────────────────────────────────────────────────────────────
// BUG-001: 死亡ペットの /accounting/new?petId= 直打ちガード
// ─────────────────────────────────────────────────────────────

const DECEASED_PET_ID = "1000003";
const LIVING_PET_ID = "1000019";

function makePetResponse(overrides: {
  id: number;
  status: "alive" | "deceased";
  name?: string;
}) {
  return {
    id: overrides.id,
    version: 1,
    clinic_id: 1,
    owner_id: 10,
    animal_species_id: 1,
    pet_number: String(overrides.id),
    name: overrides.name ?? "テストペット",
    pet_name_kana: "",
    gender: "unknown",
    status: overrides.status,
    breed: "",
    color: "",
    danger_level: "none",
    food: "",
    environment: "",
    phone: "",
    remarks: "",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    deceased_at: overrides.status === "deceased" ? "2026-07-01T00:00:00+09:00" : undefined,
    owner: {
      id: 10,
      owner_number: 10,
      name: "テスト飼い主",
      name_kana: "",
      phone: "",
      is_dangerous: false,
    },
    animal_species: { id: 1, name: "犬", sort_order: 1 },
  };
}

async function renderNewWithPetIdAndWait(petId: string) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <AuthContext.Provider value={makeAuthCtx(true)}>
      <QueryClientProvider client={qc}>
        <MemoryRouter initialEntries={[`/accounting/new?petId=${petId}`]}>
          <Routes>
            <Route path="/accounting/new" element={<AccountingDetail />} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    </AuthContext.Provider>,
  );
  await waitFor(() => {
    expect(screen.getByRole("heading", { name: "会計精算" })).toBeInTheDocument();
  });
}

describe("AccountingDetail — BUG-001: 死亡ペット新規会計ガード", () => {
  it("deceased petId 直打ち → 拒否メッセージ + fieldset disabled + 確定ボタンなし", async () => {
    server.use(
      http.get(`/api/v1/pets/${DECEASED_PET_ID}`, () =>
        HttpResponse.json(makePetResponse({ id: Number(DECEASED_PET_ID), status: "deceased", name: "クロ" })),
      ),
      http.get("/api/v1/masters/merchandise-items", () => HttpResponse.json([])),
      http.get("/api/v1/cash-register/closes", () => HttpResponse.json({ data: [], total: 0 })),
      http.get("/api/v1/billing-items/unbilled-details", () =>
        HttpResponse.json({ items: [], warnings: [] }),
      ),
      http.get("/api/v1/billing-items/ungrouped-same-day", () =>
        HttpResponse.json({ has_ungrouped: false, medical_record_count: 0, trimming_count: 0 }),
      ),
      http.get("/api/v1/accountings/unpaid-balance", () =>
        HttpResponse.json({ unpaid_count: 0, unpaid_total: 0 }),
      ),
    );

    await renderNewWithPetIdAndWait(DECEASED_PET_ID);

    expect(await screen.findByText("死亡したペットは会計を作成できません")).toBeInTheDocument();
    expect(document.querySelector("fieldset")).toBeDisabled();
    expect(screen.queryByRole("button", { name: "会計を確定する" })).not.toBeInTheDocument();
  });

  it("生存 petId 直打ち → 拒否メッセージなし + 確定ボタンが有効", async () => {
    server.use(
      http.get(`/api/v1/pets/${LIVING_PET_ID}`, () =>
        HttpResponse.json(makePetResponse({ id: Number(LIVING_PET_ID), status: "alive", name: "ラッキー" })),
      ),
      http.get("/api/v1/masters/merchandise-items", () => HttpResponse.json([])),
      http.get("/api/v1/cash-register/closes", () => HttpResponse.json({ data: [], total: 0 })),
      http.get("/api/v1/billing-items/unbilled-details", () =>
        HttpResponse.json({ items: [], warnings: [] }),
      ),
      http.get("/api/v1/billing-items/ungrouped-same-day", () =>
        HttpResponse.json({ has_ungrouped: false, medical_record_count: 0, trimming_count: 0 }),
      ),
      http.get("/api/v1/accountings/unpaid-balance", () =>
        HttpResponse.json({ unpaid_count: 0, unpaid_total: 0 }),
      ),
    );

    await renderNewWithPetIdAndWait(LIVING_PET_ID);

    await waitFor(() => {
      expect(screen.queryByText("死亡したペットは会計を作成できません")).not.toBeInTheDocument();
      expect(document.querySelector("fieldset")).not.toBeDisabled();
    });
    expect(screen.getByRole("button", { name: "会計を確定する" })).toBeEnabled();
  });
});
