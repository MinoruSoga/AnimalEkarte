import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { http, HttpResponse, delay } from "msw";
import { server } from "@/testing/mocks/node";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import {
  ResourceLstepAnalytics,
  ResourceOwners,
} from "@/types/generated/models";
import { LstepTagManagementPage } from "./LstepTagManagementPage";
import type { LstepTagSummaryResponse } from "../api/get-lstep-tag-summary";
import { fetchLstepTagOwnersCsv } from "../api/get-lstep-tag-owners";

const mockHasPermission = vi.fn<AuthContextValue["hasPermission"]>(() => true);

const mockAuthContext: AuthContextValue = {
  user: null,
  currentClinicId: "clinic-test-1",
  isAuthenticated: true,
  isLoading: false,
  login: async () => {},
  logout: async () => {},
  switchClinic: () => {},
  hasPermission: mockHasPermission,
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
  mockHasPermission.mockReset();
  mockHasPermission.mockReturnValue(true);
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
    const drawer = screen.getByRole("dialog");
    expect(drawer).toHaveAccessibleDescription("5名");
    expect(drawer).toHaveClass("w-full", "max-w-full", "sm:max-w-[480px]");
    expect(drawer).not.toHaveClass("w-[480px]");
    expect(
      within(drawer)
        .getByText("タグ「HLTH_健診あり」の対象者一覧")
        .closest('[data-slot="sheet-header"]')
    ).toHaveClass("pr-16");
    expect(within(drawer).getByRole("button", { name: "閉じる" })).toHaveClass(
      "min-h-11",
      "min-w-11"
    );
    expect(within(drawer).getByRole("button", { name: "CSV" })).toHaveClass(
      "min-h-11"
    );
  });

  it("対象者が100名を超える場合は先頭100名に打ち切り、一覧APIも上限内で呼ぶ", async () => {
    let requestedPerPage: string | null = null;
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/owners`, ({ request }) => {
        requestedPerPage = new URL(request.url).searchParams.get("per_page");
        return HttpResponse.json({
          owners: Array.from({ length: 100 }, (_, index) => ({
            owner_id: String(index + 1),
            owner_name: `飼主${index + 1}`,
            line_user_id: null,
            last_visit_date: null,
          })),
          total: 250,
        });
      }),
    );
    await renderAndWait({
      ...mockSummary,
      tags: [{ tag_name: "HLTH_健診あり", owner_count: 250, category: "auto" }],
    });

    const row = screen.getByText("HLTH_健診あり").closest("tr");
    expect(row).not.toBeNull();
    await userEvent.setup().click(within(row!).getByRole("button", { name: /対象者一覧/ }));

    expect(await screen.findByText("先頭100名を表示しています")).toBeInTheDocument();
    expect(requestedPerPage).toBe("100");
  });

  it("手動タグの対象者が打ち切られた場合は部分的な一括解除を許可しない", async () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/owners`, () =>
        HttpResponse.json({
          owners: Array.from({ length: 100 }, (_, index) => ({
            owner_id: String(index + 1),
            owner_name: `飼主${index + 1}`,
            line_user_id: null,
            last_visit_date: null,
          })),
          total: 250,
        }),
      ),
    );
    await renderAndWait({
      ...mockSummary,
      tags: [{ tag_name: "手動フォロー", owner_count: 250, category: "manual" }],
    });

    const row = screen.getByText("手動フォロー").closest("tr");
    expect(row).not.toBeNull();
    await userEvent.setup().click(within(row!).getByRole("button", { name: /対象者一覧/ }));

    const drawer = await screen.findByRole("dialog");
    expect(within(drawer).queryByRole("button", { name: "一括解除" })).not.toBeInTheDocument();
  });

  it("CSV出力はpageクロールせずformat=csvの1リクエストに委譲する", async () => {
    const requests: URL[] = [];
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/owners`, ({ request }) => {
        requests.push(new URL(request.url));
        return new HttpResponse("owner_id,owner_name\n1,山田", {
          headers: { "Content-Type": "text/csv; charset=utf-8" },
        });
      }),
    );

    const csv = await fetchLstepTagOwnersCsv("HLTH_健診あり");

    expect(csv).toBeInstanceOf(Blob);
    expect(requests).toHaveLength(1);
    expect(requests[0].searchParams.get("tag")).toBe("HLTH_健診あり");
    expect(requests[0].searchParams.get("format")).toBe("csv");
    expect(requests[0].searchParams.has("page")).toBe(false);
  });

  it("対象者が5000名を超える場合は不完全なCSV出力を無効化して上限を明示する", async () => {
    setupOwnersHandler();
    await renderAndWait({
      ...mockSummary,
      tags: [{ tag_name: "HLTH_健診あり", owner_count: 5_001, category: "auto" }],
    });

    const row = screen.getByText("HLTH_健診あり").closest("tr");
    expect(row).not.toBeNull();
    await userEvent.setup().click(within(row!).getByRole("button", { name: /対象者一覧/ }));

    const drawer = await screen.findByRole("dialog");
    expect(within(drawer).getByText("CSV出力は5000名までです")).toBeInTheDocument();
    expect(within(drawer).getByRole("button", { name: "CSV" })).toBeDisabled();
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

  it("403エラー時は空一覧ではなく権限不足を表示する", async () => {
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/tag-summary`, () =>
        HttpResponse.json({ error: "forbidden" }, { status: 403 })
      )
    );
    render(<LstepTagManagementPage />, { wrapper: createWrapper() });
    await waitFor(() => {
      expect(screen.getByText(/アクセス権限がありません/)).toBeInTheDocument();
    });
    expect(screen.queryByText("タグが見つかりません")).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// G: RBAC 契約
// ─────────────────────────────────────────────────────────────

describe("LstepTagManagementPage — G: RBAC 契約", () => {
  it("集計・対象者一覧の閲覧権限を lstep-analytics で判定する", async () => {
    await renderAndWait();

    expect(mockHasPermission).toHaveBeenCalledWith(ResourceLstepAnalytics, "view");
  });

  it("手動タグの解除操作を owners:delete で判定し、削除は編集パネルからのみ提供する", async () => {
    mockHasPermission.mockImplementation(
      (resource, action) =>
        (resource === ResourceLstepAnalytics && action === "view") ||
        (resource === ResourceOwners && action === "delete")
    );
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/owners`, () =>
        HttpResponse.json({
          owners: [
            {
              owner_id: "owner-1",
              owner_name: "山田 太郎",
              line_user_id: null,
              last_visit_date: null,
            },
            {
              owner_id: "owner-2",
              owner_name: "佐藤 花子",
              line_user_id: null,
              last_visit_date: null,
            },
          ],
          total: 2,
        })
      )
    );
    await renderAndWait({
      ...mockSummary,
      tags: [{ tag_name: "campaign_summer", owner_count: 2, category: "manual" }],
    });

    const row = screen.getByText("campaign_summer").closest("tr");
    expect(row).not.toBeNull();
    // BUG-024: リスト行に赤「削除」は置かない（他マスタ同様、詳細/編集面からのみ）
    expect(within(row!).queryByRole("button", { name: /削除/ })).not.toBeInTheDocument();
    expect(within(row!).getByRole("button", { name: /対象者一覧/ })).toBeInTheDocument();

    const user = userEvent.setup();
    await user.click(within(row!).getByRole("button", { name: /対象者一覧/ }));

    const drawer = await screen.findByRole("dialog");
    expect(within(drawer).getByRole("button", { name: "一括解除" })).toBeInTheDocument();
    expect(mockHasPermission).toHaveBeenCalledWith(ResourceOwners, "delete");
  });
});

// ─────────────────────────────────────────────────────────────
// H: BUG-024 削除 UI は編集パネルのみ
// ─────────────────────────────────────────────────────────────

describe("LstepTagManagementPage — H: BUG-024 削除 UI パターン", () => {
  it("リスト行にインラインの破壊的「削除」ボタンを出さない", async () => {
    await renderAndWait({
      ...mockSummary,
      tags: [
        { tag_name: "campaign_summer", owner_count: 2, category: "manual" },
        { tag_name: "HLTH_健診あり", owner_count: 5, category: "auto" },
      ],
    });

    const manualRow = screen.getByText("campaign_summer").closest("tr");
    const autoRow = screen.getByText("HLTH_健診あり").closest("tr");
    expect(manualRow).not.toBeNull();
    expect(autoRow).not.toBeNull();

    expect(within(manualRow!).queryByRole("button", { name: /削除/ })).not.toBeInTheDocument();
    expect(within(autoRow!).queryByRole("button", { name: /削除/ })).not.toBeInTheDocument();
    expect(within(manualRow!).getByRole("button", { name: /対象者一覧/ })).toBeInTheDocument();
  });

  it("対象者一覧パネルから一括解除を開き、確認後に既存削除 API を呼ぶ", async () => {
    const deleted: Array<{ ownerId: string; tagName: string }> = [];
    const summaryData: LstepTagSummaryResponse = {
      ...mockSummary,
      tags: [{ tag_name: "campaign_summer", owner_count: 1, category: "manual" }],
    };
    server.use(
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/tag-summary`, () =>
        HttpResponse.json(summaryData)
      ),
      http.get(`/api/v1/clinics/${CLINIC_ID}/lstep/owners`, () =>
        HttpResponse.json({
          owners: [
            {
              owner_id: "owner-1",
              owner_name: "山田 太郎",
              line_user_id: null,
              last_visit_date: null,
            },
          ],
          total: 1,
        })
      ),
      http.delete(
        `/api/v1/clinics/${CLINIC_ID}/owners/:ownerId/lstep/tags/:tagName`,
        ({ params }) => {
          deleted.push({
            ownerId: String(params.ownerId),
            tagName: decodeURIComponent(String(params.tagName)),
          });
          return new HttpResponse(null, { status: 204 });
        }
      )
    );
    render(<LstepTagManagementPage />, { wrapper: createWrapper() });
    expect(await screen.findByText("campaign_summer")).toBeInTheDocument();

    const user = userEvent.setup();
    const row = screen.getByText("campaign_summer").closest("tr");
    expect(row).not.toBeNull();
    await user.click(within(row!).getByRole("button", { name: /対象者一覧/ }));

    const drawer = await screen.findByRole("dialog");
    await user.click(within(drawer).getByRole("button", { name: "一括解除" }));

    const confirm = await screen.findByRole("alertdialog");
    expect(within(confirm).getByText("タグを一括解除します")).toBeInTheDocument();
    await user.click(within(confirm).getByRole("button", { name: "一括解除" }));

    await waitFor(() => {
      expect(deleted).toEqual([{ ownerId: "owner-1", tagName: "campaign_summer" }]);
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
    expect(screen.getByRole("link", { name: "カルテを開く" })).toHaveClass(
      "min-h-11"
    );
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
