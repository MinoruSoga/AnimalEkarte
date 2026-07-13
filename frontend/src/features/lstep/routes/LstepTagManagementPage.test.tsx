import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse, delay } from "msw";
import { server } from "@/testing/mocks/node";
import { AuthContext } from "@/contexts/auth-context";
import { LstepTagManagementPage } from "./LstepTagManagementPage";
import type { LstepTagSummaryResponse } from "../api/get-lstep-tag-summary";

const mockAuthContext = {
  user: null,
  currentClinicId: "clinic-test-1",
  isAuthenticated: true,
  isLoading: false,
  login: async () => {},
  logout: async () => {},
  switchClinic: () => {},
  hasPermission: () => true as boolean,
  refreshPermissions: async () => {},
};

const CLINIC_ID = "clinic-test-1";

const mockSummary: LstepTagSummaryResponse = {
  tags: [
    { tag_name: "HLTH_健診あり", owner_count: 5, category: "auto" },
    { tag_name: "HLTH_健診未受診", owner_count: 3, category: "auto" },
    { tag_name: "HLTH_年4回候補", owner_count: 2, category: "auto" },
    { tag_name: "PREV_ワクチン期限", owner_count: 8, category: "auto" },
    { tag_name: "PREV_フィラリア未完了", owner_count: 4, category: "auto" },
    { tag_name: "PREV_ノミダニ対象", owner_count: 6, category: "auto" },
    { tag_name: "LTV_フード購入あり", owner_count: 10, category: "auto" },
  ],
  total_owners_with_lstep: 50,
  as_of: "2026-05-01T10:00:00Z",
};

function setupSummaryHandler(data: LstepTagSummaryResponse) {
  server.use(
    http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/tag-summary`, () =>
      HttpResponse.json(data)
    )
  );
}

function setupOwnersHandler() {
  server.use(
    http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/owners`, () =>
      HttpResponse.json({ owners: [], total: 0 })
    )
  );
}

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <AuthContext.Provider value={mockAuthContext}>
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>{children}</MemoryRouter>
      </QueryClientProvider>
    </AuthContext.Provider>
  );
}

async function renderAndWait(data: LstepTagSummaryResponse = mockSummary) {
  setupSummaryHandler(data);
  render(<LstepTagManagementPage />, { wrapper: createWrapper() });
  await waitFor(() => {
    expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
  });
}

beforeEach(() => {
  localStorage.setItem("auth_current_clinic:v1", CLINIC_ID);
});

afterEach(() => {
  localStorage.removeItem("auth_current_clinic:v1");
});

// ─────────────────────────────────────────────────────────────
// A: 健診タグ一覧（FEAT-380）
// ─────────────────────────────────────────────────────────────

describe("LstepTagManagementPage — A: 健診タグ一覧 (FEAT-380)", () => {
  it("健診タグ 3 種がテーブルに表示される", async () => {
    await renderAndWait();
    expect(screen.getByText("HLTH_健診あり")).toBeInTheDocument();
    expect(screen.getByText("HLTH_健診未受診")).toBeInTheDocument();
    expect(screen.getByText("HLTH_年4回候補")).toBeInTheDocument();
  });

  it("HLTH_健診あり の件数 5 件がテーブル行に表示される", async () => {
    await renderAndWait();
    const row = screen.getByText("HLTH_健診あり").closest("tr");
    expect(row).not.toBeNull();
    expect(within(row!).getByText("5")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// B: 予防処置タグ一覧（FEAT-380）
// ─────────────────────────────────────────────────────────────

describe("LstepTagManagementPage — B: 予防処置タグ一覧 (FEAT-380)", () => {
  it("予防タグ 3 種がテーブルに表示される", async () => {
    await renderAndWait();
    expect(screen.getByText("PREV_ワクチン期限")).toBeInTheDocument();
    expect(screen.getByText("PREV_フィラリア未完了")).toBeInTheDocument();
    expect(screen.getByText("PREV_ノミダニ対象")).toBeInTheDocument();
  });

  it("PREV_ノミダニ対象 の件数 6 件がテーブル行に表示される", async () => {
    await renderAndWait();
    const row = screen.getByText("PREV_ノミダニ対象").closest("tr");
    expect(row).not.toBeNull();
    expect(within(row!).getByText("6")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// C: LTV タグ一覧（FEAT-380）
// ─────────────────────────────────────────────────────────────

describe("LstepTagManagementPage — C: LTV タグ一覧 (FEAT-380)", () => {
  it("LTV_フード購入あり がテーブルに表示される", async () => {
    await renderAndWait();
    expect(screen.getByText("LTV_フード購入あり")).toBeInTheDocument();
  });

  it("LTV_フード購入あり の件数 10 件がテーブル行に表示される", async () => {
    await renderAndWait();
    const row = screen.getByText("LTV_フード購入あり").closest("tr");
    expect(row).not.toBeNull();
    expect(within(row!).getByText("10")).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// D: タグ別飼い主一覧ドロワー（FEAT-380）
// ─────────────────────────────────────────────────────────────

describe("LstepTagManagementPage — D: タグ別飼い主一覧ドロワー (FEAT-380)", () => {
  it("HLTH_健診あり 行の「対象者一覧」クリックでドロワーが開く", async () => {
    setupOwnersHandler();
    await renderAndWait();

    const row = screen.getByText("HLTH_健診あり").closest("tr");
    expect(row).not.toBeNull();

    const user = userEvent.setup();
    await user.click(within(row!).getByRole("button", { name: /対象者一覧/ }));

    await waitFor(() => {
      expect(
        screen.getByText("タグ「HLTH_健診あり」の対象者一覧")
      ).toBeInTheDocument();
    });
  });
});

// ─────────────────────────────────────────────────────────────
// E: ローディング・エラー状態（FEAT-380）
// ─────────────────────────────────────────────────────────────

describe("LstepTagManagementPage — E: ローディング・エラー状態 (FEAT-380)", () => {
  it("ローディング中は「読み込み中...」がテーブルに表示される", () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/tag-summary`, async () => {
        await delay("infinite");
        return HttpResponse.json(mockSummary);
      })
    );
    render(<LstepTagManagementPage />, { wrapper: createWrapper() });
    expect(screen.getByText("読み込み中...")).toBeInTheDocument();
  });

  it("APIエラー時はテーブルに「タグが見つかりません」が表示される", async () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/tag-summary`, () =>
        HttpResponse.json({ message: "Internal Server Error" }, { status: 500 })
      )
    );
    render(<LstepTagManagementPage />, { wrapper: createWrapper() });
    await waitFor(() => {
      expect(screen.getByText("タグが見つかりません")).toBeInTheDocument();
    });
  });
});

// ─────────────────────────────────────────────────────────────
// F: 判定理由表示（FEAT-379-supplement）
// ─────────────────────────────────────────────────────────────

describe("LstepTagManagementPage — F: 判定理由表示 (FEAT-379-supplement)", () => {
  it("reason が存在する場合「判定理由: ...」が描画される", async () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/owners`, () =>
        HttpResponse.json({
          owners: [
            {
              owner_id: "1",
              owner_name: "田中 太郎",
              last_visit_date: null,
              reason: "最終健診: 2025-12-01",
            },
          ],
          total: 1,
        })
      )
    );
    await renderAndWait();

    const row = screen.getByText("HLTH_健診あり").closest("tr");
    expect(row).not.toBeNull();
    const user = userEvent.setup();
    await user.click(within(row!).getByRole("button", { name: /対象者一覧/ }));

    await waitFor(() => {
      expect(
        screen.getByText("判定理由: 最終健診: 2025-12-01")
      ).toBeInTheDocument();
    });
  });

  it("reason が undefined の場合「判定理由」テキストは描画されない", async () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/owners`, () =>
        HttpResponse.json({
          owners: [
            {
              owner_id: "2",
              owner_name: "鈴木 花子",
              last_visit_date: null,
            },
          ],
          total: 1,
        })
      )
    );
    await renderAndWait();

    const row = screen.getByText("HLTH_健診あり").closest("tr");
    expect(row).not.toBeNull();
    const user = userEvent.setup();
    await user.click(within(row!).getByRole("button", { name: /対象者一覧/ }));

    await waitFor(() => {
      expect(screen.getByText("鈴木 花子")).toBeInTheDocument();
    });
    expect(screen.queryByText(/判定理由/)).not.toBeInTheDocument();
  });
});
