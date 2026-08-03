import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { AuthContext } from "@/hooks/auth-context";
import type { ResourceAction } from "@/types/auth";
import { HospitalizationList } from "./HospitalizationList";

// ---- mocks ----

vi.mock("../api/get-hospitalizations", () => ({
  useGetHospitalizations: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({ canCreate: true, canEdit: true })),
}));

vi.mock("@/hooks/use-master-items", () => ({
  useMasterItems: vi.fn(() => ({
    data: [{ id: "cage-1", name: "ケージ1", category: "犬舎", price: 0, status: "active" as const }],
    isLoading: false,
    error: null,
  })),
}));

import { useGetHospitalizations } from "../api/get-hospitalizations";

// ---- helpers ----

function makeAuthCtx() {
  return {
    user: null,
    currentClinicId: "clinic-1",
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: (_resource: string, _action: ResourceAction): boolean => true,
    refreshPermissions: async () => {},
  };
}

function makeHosp(overrides: Partial<{
  id: string;
  petName: string;
  ownerName: string;
  hospitalizationNo: string;
  status: string;
  hospitalizationType: string;
  cageId: string;
  startDate: string;
  endDate: string | undefined;
  species: string;
  petIsDeceased: boolean;
}> = {}) {
  return {
    id: "h-1",
    petName: "ポチ",
    ownerName: "ヤマダ",
    hospitalizationNo: "H-001",
    status: "入院中",
    hospitalizationType: "short_term",
    cageId: "cage-1",
    startDate: "2026-01-01",
    endDate: undefined,
    species: "犬",
    petIsDeceased: false,
    ...overrides,
  };
}

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <AuthContext.Provider value={makeAuthCtx()}>
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>{children}</MemoryRouter>
      </QueryClientProvider>
    </AuthContext.Provider>
  );
}

function mockHospitalizationsPage(
  items: ReturnType<typeof makeHosp>[],
  total = items.length,
) {
  vi.mocked(useGetHospitalizations).mockReturnValue({
    data: {
      data: items,
      total,
      page: 1,
      limit: 20,
    },
    isLoading: false,
    isError: false,
  } as ReturnType<typeof useGetHospitalizations>);
}

beforeEach(() => {
  localStorage.setItem("auth_current_clinic:v1", "clinic-1");
  mockHospitalizationsPage([]);
});

describe("HospitalizationList — view切替の操作領域", () => {
  it("Board/List view buttonを44x44px以上に保つ", () => {
    render(<HospitalizationList />, { wrapper: createWrapper() });

    for (const name of ["Board View", "List View"]) {
      expect(screen.getByRole("radio", { name })).toHaveClass("h-11", "min-w-11");
    }
  });
});

// ─────────────────────────────────────────────────────────────
// かな正規化テキスト検索 (list view で検証)
// ─────────────────────────────────────────────────────────────

describe("HospitalizationList — かな正規化テキスト検索", () => {
  it("ひらがな入力でカタカナ petName がヒットする", async () => {
    mockHospitalizationsPage([
      makeHosp({ id: "1", petName: "ポチ", ownerName: "ヤマダ", hospitalizationNo: "H-001" }),
      makeHosp({ id: "2", petName: "たろう", ownerName: "さとう", hospitalizationNo: "H-002" }),
    ]);

    const user = userEvent.setup();
    render(<HospitalizationList />, { wrapper: createWrapper() });

    // リストビューに切り替え (ToggleGroup type="single" は role="radio")
    await user.click(screen.getByRole("radio", { name: "List View" }));

    // 検索入力を開いて入力
    await user.click(screen.getByRole("button", { name: "検索" }));
    await user.type(screen.getByPlaceholderText("飼主名、ペット名、入院No..."), "ぽち");

    expect(await screen.findByText("ポチ")).toBeInTheDocument();
    expect(screen.queryByText("たろう")).not.toBeInTheDocument();
  });

  it("カタカナ入力でひらがな ownerName がヒットする", async () => {
    mockHospitalizationsPage([
      makeHosp({ id: "1", petName: "ポチ", ownerName: "ヤマダ", hospitalizationNo: "H-001" }),
      makeHosp({ id: "2", petName: "たろう", ownerName: "さとう", hospitalizationNo: "H-002" }),
    ]);

    const user = userEvent.setup();
    render(<HospitalizationList />, { wrapper: createWrapper() });

    await user.click(screen.getByRole("radio", { name: "List View" }));
    await user.click(screen.getByRole("button", { name: "検索" }));
    await user.type(screen.getByPlaceholderText("飼主名、ペット名、入院No..."), "サトウ");

    expect(await screen.findByText("たろう")).toBeInTheDocument();
    expect(screen.queryByText("ポチ")).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// BUG-009: タブ status → server query 連動 / page-window 件数 / 二重 filter 除去
// ─────────────────────────────────────────────────────────────

describe("HospitalizationList — BUG-009 status tab → server filter", () => {
  it("タブ切替で useGetHospitalizations に statusFilter が渡り、reserved 行が表示対象になる", async () => {
    // 実装は server が status で絞った集合をそのまま描画する（client status 二重 filter なし）
    mockHospitalizationsPage(
      [
        makeHosp({
          id: "r-1",
          petName: "予約ワン",
          status: "予約",
          hospitalizationNo: "H-R1",
          cageId: "cage-1",
        }),
      ],
      12,
    );

    const user = userEvent.setup();
    render(<HospitalizationList />, { wrapper: createWrapper() });

    await user.click(screen.getByRole("tab", { name: "予約" }));

    const calls = vi.mocked(useGetHospitalizations).mock.calls;
    const lastFilters = calls[calls.length - 1]?.[0];
    expect(lastFilters).toEqual(
      expect.objectContaining({
        statusFilter: "reserved",
        page: 1,
        limit: 20,
      }),
    );

    // list view で reserved 行が残る（旧失敗: client が admitted のみ残す）
    await user.click(screen.getByRole("radio", { name: "List View" }));
    expect(await screen.findByText("予約ワン")).toBeInTheDocument();
    expect(screen.getByText("H-R1")).toBeInTheDocument();
  });

  it("board と list が同一 data source を参照する", async () => {
    mockHospitalizationsPage([
      makeHosp({
        id: "1",
        petName: "ボード兼リスト",
        status: "入院中",
        cageId: "cage-1",
      }),
    ]);

    const user = userEvent.setup();
    render(<HospitalizationList />, { wrapper: createWrapper() });

    // board: cage に紐づく petName
    expect(await screen.findByText("ボード兼リスト")).toBeInTheDocument();

    await user.click(screen.getByRole("radio", { name: "List View" }));
    expect(await screen.findByText("ボード兼リスト")).toBeInTheDocument();
  });

  it("件数表示の正本が server total である（page 内 length ではない）", async () => {
    mockHospitalizationsPage(
      [makeHosp({ id: "1", petName: "1件だけ返却", status: "入院中" })],
      42,
    );

    render(<HospitalizationList />, { wrapper: createWrapper() });

    expect(await screen.findByText(/42/)).toBeInTheDocument();
  });

  it("server total が limit を超えると list view に Pagination が出る（client window ではない）", async () => {
    mockHospitalizationsPage(
      [makeHosp({ id: "1", petName: "ページ1行", status: "入院中" })],
      42,
    );

    const user = userEvent.setup();
    render(<HospitalizationList />, { wrapper: createWrapper() });
    await user.click(screen.getByRole("radio", { name: "List View" }));

    // total=42, limit=20 → totalPages=3 → Pagination（server 正本）が表示される
    expect(await screen.findByText(/42件中/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "次のページ" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "最後のページ" })).toBeEnabled();
  });
});
