import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { AuthContext } from "@/contexts/auth-context";
import type { ResourceAction } from "@/types/auth";
import { CheckupsList } from "./CheckupsList";
import type { CheckupFilters } from "../api/types";

// ---- mock useGetCheckups ----

vi.mock("../api/get-checkups", () => ({
  useGetCheckups: vi.fn(),
}));

import { useGetCheckups } from "../api/get-checkups";

// ---- helpers ----

function getISODate(dayOffset: number): string {
  const d = new Date();
  d.setDate(d.getDate() + dayOffset);
  return d.toISOString().slice(0, 10);
}

function makeAuthCtx() {
  return {
    user: null,
    currentClinicId: "clinic-test-1",
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: (_resource: string, _action: ResourceAction): boolean => true,
    refreshPermissions: async () => {},
  };
}

function makeCheckupRecord(overrides: Partial<{
  id: string;
  medicalRecordId: string;
  date: string;
  ownerName: string;
  petName: string;
  checkupTypeName: string;
  result: string;
  nextDate: string | undefined;
  doctorName: string;
}> = {}) {
  return {
    id: "chk-1",
    medicalRecordId: "mr-1",
    date: "2026-01-01",
    ownerName: "山田 太郎",
    petName: "ポチ",
    checkupTypeName: "定期健診",
    result: "正常",
    nextDate: undefined,
    doctorName: "鈴木 医師",
    ...overrides,
  };
}

// X-16②: BE は {data,total,page,limit} の実ページング封筒を返す。
function makeCheckupsResult(
  data: ReturnType<typeof makeCheckupRecord>[],
  overrides: Partial<{ total: number; page: number; limit: number }> = {},
) {
  return {
    data,
    total: overrides.total ?? data.length,
    page: overrides.page ?? 1,
    limit: overrides.limit ?? 20,
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

beforeEach(() => {
  localStorage.setItem("auth_current_clinic:v1", "clinic-test-1");
  vi.mocked(useGetCheckups).mockReturnValue({
    data: makeCheckupsResult([]),
    isLoading: false,
    error: null,
  } as ReturnType<typeof useGetCheckups>);
});

// ─────────────────────────────────────────────────────────────
// A: 初期レンダリング時の API 呼び出しパラメータ
// ─────────────────────────────────────────────────────────────

describe("CheckupsList — A: 初期 API 呼び出しパラメータ (FEAT-372)", () => {
  it("フィルタ未適用時は nextStartDate / nextEndDate が undefined で呼ばれる", async () => {
    render(<CheckupsList />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(vi.mocked(useGetCheckups)).toHaveBeenCalled();
    });

    const calledWith = vi.mocked(useGetCheckups).mock.calls[0][0] as CheckupFilters | undefined;
    expect(calledWith?.nextStartDate).toBeUndefined();
    expect(calledWith?.nextEndDate).toBeUndefined();
  });
});

// ─────────────────────────────────────────────────────────────
// B: 期限切れバッジ
// ─────────────────────────────────────────────────────────────

describe("CheckupsList — B: 期限切れバッジ表示 (FEAT-372)", () => {
  it("nextDate が過去の場合、「期限切れ」バッジが表示される", async () => {
    const overdueDate = getISODate(-10);
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult([makeCheckupRecord({ nextDate: overdueDate })]),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);

    render(<CheckupsList />, { wrapper: createWrapper() });

    expect(await screen.findByText("期限切れ")).toBeInTheDocument();
  });

  it("nextDate が未来の場合、「期限切れ」バッジは表示されない", async () => {
    const futureDate = getISODate(60);
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult([makeCheckupRecord({ nextDate: futureDate })]),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);

    render(<CheckupsList />, { wrapper: createWrapper() });

    await screen.findByText("ポチ");
    expect(screen.queryByText("期限切れ")).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// C: 期限間近バッジ
// ─────────────────────────────────────────────────────────────

describe("CheckupsList — C: 期限間近バッジ表示 (FEAT-372)", () => {
  it("nextDate が今後30日以内の場合、「期限間近」バッジが表示される", async () => {
    const upcomingDate = getISODate(15);
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult([makeCheckupRecord({ nextDate: upcomingDate })]),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);

    render(<CheckupsList />, { wrapper: createWrapper() });

    expect(await screen.findByText("期限間近")).toBeInTheDocument();
  });

  it("nextDate が30日より先の場合、「期限間近」バッジは表示されない", async () => {
    const farFutureDate = getISODate(60);
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult([makeCheckupRecord({ nextDate: farFutureDate })]),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);

    render(<CheckupsList />, { wrapper: createWrapper() });

    await screen.findByText("ポチ");
    expect(screen.queryByText("期限間近")).not.toBeInTheDocument();
  });

  it("nextDate が undefined の場合、バッジは一切表示されない", async () => {
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult([makeCheckupRecord({ nextDate: undefined })]),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);

    render(<CheckupsList />, { wrapper: createWrapper() });

    await screen.findByText("ポチ");
    expect(screen.queryByText("期限切れ")).not.toBeInTheDocument();
    expect(screen.queryByText("期限間近")).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// D: かな正規化テキスト検索
// ─────────────────────────────────────────────────────────────

describe("CheckupsList — D: かな正規化テキスト検索", () => {
  it("ひらがな入力でカタカナ petName がヒットする", async () => {
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult([
        makeCheckupRecord({ id: "1", petName: "ポチ", ownerName: "ヤマダ" }),
        makeCheckupRecord({ id: "2", petName: "たろう", ownerName: "さとう" }),
      ]),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);

    const user = userEvent.setup();
    render(<CheckupsList />, { wrapper: createWrapper() });

    await user.click(screen.getByRole("button", { name: "検索" }));
    await user.type(screen.getByPlaceholderText("ペット名・飼主名・種別で検索..."), "ぽち");

    expect(await screen.findByText("ポチ")).toBeInTheDocument();
    expect(screen.queryByText("たろう")).not.toBeInTheDocument();
  });

  it("カタカナ入力でひらがな ownerName がヒットする", async () => {
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult([
        makeCheckupRecord({ id: "1", petName: "ポチ", ownerName: "ヤマダ" }),
        makeCheckupRecord({ id: "2", petName: "たろう", ownerName: "さとう" }),
      ]),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);

    const user = userEvent.setup();
    render(<CheckupsList />, { wrapper: createWrapper() });

    await user.click(screen.getByRole("button", { name: "検索" }));
    await user.type(screen.getByPlaceholderText("ペット名・飼主名・種別で検索..."), "サトウ");

    expect(await screen.findByText("たろう")).toBeInTheDocument();
    expect(screen.queryByText("ポチ")).not.toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// E: サーバページング (X-16②)
// ─────────────────────────────────────────────────────────────

describe("CheckupsList — E: サーバページング (X-16②)", () => {
  it("初期表示時は page=1, limit=20 で useGetCheckups が呼ばれる", async () => {
    render(<CheckupsList />, { wrapper: createWrapper() });

    await waitFor(() => {
      expect(vi.mocked(useGetCheckups)).toHaveBeenCalled();
    });

    const calledWith = vi.mocked(useGetCheckups).mock.calls[0][0] as CheckupFilters | undefined;
    expect(calledWith?.page).toBe(1);
    expect(calledWith?.limit).toBe(20);
  });

  it("サーバの total 件数がページネーション表示に反映される（取得件数はページ分のみでも total は全体件数を表示）", async () => {
    const pageRecords = Array.from({ length: 20 }, (_, i) =>
      makeCheckupRecord({ id: `chk-${i}`, petName: `ペット${i}` }),
    );
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult(pageRecords, { total: 45, page: 1, limit: 20 }),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);

    render(<CheckupsList />, { wrapper: createWrapper() });

    expect(await screen.findByText(/45件中/)).toBeInTheDocument();
  });

  it("次のページボタン押下で page=2 として useGetCheckups が呼ばれる", async () => {
    const pageRecords = Array.from({ length: 20 }, (_, i) =>
      makeCheckupRecord({ id: `chk-${i}`, petName: `ペット${i}` }),
    );
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult(pageRecords, { total: 45, page: 1, limit: 20 }),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);

    const user = userEvent.setup();
    render(<CheckupsList />, { wrapper: createWrapper() });

    await screen.findByText(/45件中/);
    await user.click(screen.getByRole("button", { name: "次のページ" }));

    await waitFor(() => {
      const lastCall = vi.mocked(useGetCheckups).mock.calls.at(-1)?.[0] as CheckupFilters | undefined;
      expect(lastCall?.page).toBe(2);
    });
  });
});
