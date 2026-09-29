import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, useLocation } from "react-router";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { ResourceCheckups, ResourceMedicalRecords } from "@/types/generated/models";
import { CheckupsList } from "./CheckupsList";
import type { CheckupFilters } from "../types";

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

const allowAllPermissions: AuthContextValue["hasPermission"] = () => true;

function makeAuthCtx(
  hasPermission: AuthContextValue["hasPermission"] = allowAllPermissions,
): AuthContextValue {
  return {
    user: null,
    currentClinicId: "clinic-test-1",
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission,
    refreshPermissions: async () => {},
  };
}

function makeCheckupRecord(
  overrides: Partial<{
    id: string;
    medicalRecordId: string;
    petId: string | undefined;
    date: string;
    ownerName: string;
    petName: string;
    checkupTypeName: string;
    result: string;
    nextDate: string | undefined;
    doctorName: string;
  }> = {},
) {
  return {
    id: "chk-1",
    medicalRecordId: "mr-1",
    petId: undefined,
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

function LocationProbe() {
  const { pathname, search } = useLocation();
  return <output data-testid="location">{`${pathname}${search}`}</output>;
}

function createWrapper(hasPermission: AuthContextValue["hasPermission"] = allowAllPermissions) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <AuthContext.Provider value={makeAuthCtx(hasPermission)}>
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={["/checkups"]}>
          {children}
          <LocationProbe />
        </MemoryRouter>
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
      const lastCall = vi.mocked(useGetCheckups).mock.calls.at(-1)?.[0] as
        CheckupFilters | undefined;
      expect(lastCall?.page).toBe(2);
    });
  });
});

describe("CheckupsList — row navigation accessibility", () => {
  beforeEach(() => {
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult([makeCheckupRecord({ date: "2026-07-13" })]),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);
  });

  it("ペット名・実施日を含む44px以上のnative medical record linkを行内に表示する", () => {
    render(<CheckupsList />, { wrapper: createWrapper() });

    const detailLink = screen.getByRole("link", { name: /ポチ/ });
    expect(detailLink).toHaveAttribute(
      "href",
      "/medical-records/mr-1?tab=%E5%AE%9A%E6%9C%9F%E5%81%A5%E8%A8%BA&checkupId=chk-1",
    );
    expect(detailLink).toHaveAccessibleName(/ポチ/);
    expect(detailLink).toHaveAccessibleName(/2026-07-13/);
    expect(detailLink).toHaveAccessibleName(/chk-1/);
    expect(detailLink).toHaveClass("min-h-11", "min-w-11");
  });

  it("detail link以外のセルclickでは行遷移しない", () => {
    render(<CheckupsList />, { wrapper: createWrapper() });

    fireEvent.click(screen.getByText("山田 太郎"));

    expect(screen.getByTestId("location")).toHaveTextContent(/^\/checkups$/);
  });

  it("編集ボタンは定期健診タブ付きカルテへ遷移する (BUG-022)", async () => {
    const user = userEvent.setup();
    render(<CheckupsList />, { wrapper: createWrapper() });
    await user.click(screen.getByRole("button", { name: /健診操作/ }));
    expect(screen.getByTestId("location")).toHaveTextContent(
      /\/medical-records\/mr-1\?tab=.*checkupId=chk-1/,
    );
  });
});

describe("CheckupsList — medical record permission boundary", () => {
  it.each([
    { canCreate: true, canEdit: false },
    { canCreate: false, canEdit: true },
  ])(
    "新規登録CTAはmedical-records:create/editの片方だけでは表示しない ($canCreate/$canEdit)",
    ({ canCreate, canEdit }) => {
      const hasPermission: AuthContextValue["hasPermission"] = vi.fn(
        (resource, action) =>
          (resource === ResourceCheckups && action === "view") ||
          (resource === ResourceMedicalRecords && action === "view") ||
          (resource === ResourceMedicalRecords && action === "create" && canCreate) ||
          (resource === ResourceMedicalRecords && action === "edit" && canEdit),
      );

      render(<CheckupsList />, { wrapper: createWrapper(hasPermission) });

      expect(screen.queryByRole("button", { name: /新規登録/ })).not.toBeInTheDocument();
      expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "create");
      expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "edit");
    },
  );

  it("新規登録CTAはmedical-records:create/editの両方がある場合だけ表示する", () => {
    const hasPermission: AuthContextValue["hasPermission"] = vi.fn(
      (resource, action) =>
        (resource === ResourceCheckups && action === "view") ||
        (resource === ResourceMedicalRecords && action === "view") ||
        (resource === ResourceMedicalRecords && (action === "create" || action === "edit")),
    );

    render(<CheckupsList />, { wrapper: createWrapper(hasPermission) });

    expect(screen.getByRole("button", { name: /新規登録/ })).toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "create");
    expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "edit");
  });

  it("row edit actionはmedical-records:editがあってもviewがなければ表示しない", () => {
    const hasPermission: AuthContextValue["hasPermission"] = vi.fn(
      (resource, action) =>
        (resource === ResourceCheckups && action === "view") ||
        (resource === ResourceMedicalRecords && action === "edit"),
    );
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult([makeCheckupRecord()]),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);

    render(<CheckupsList />, { wrapper: createWrapper(hasPermission) });

    expect(screen.queryByRole("button", { name: "操作" })).not.toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "view");
    expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "edit");
  });

  it("row edit actionはmedical-records:view/editの両方がある場合だけ表示する", () => {
    const hasPermission: AuthContextValue["hasPermission"] = vi.fn(
      (resource, action) =>
        (resource === ResourceCheckups && action === "view") ||
        (resource === ResourceMedicalRecords && (action === "view" || action === "edit")),
    );
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult([makeCheckupRecord()]),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);

    render(<CheckupsList />, { wrapper: createWrapper(hasPermission) });

    expect(screen.getByRole("button", { name: /chk-1/ })).toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "view");
    expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "edit");
  });
});

// ─────────────────────────────────────────────────────────────
// F: 動物種フィルタ (EMR-223)
// ─────────────────────────────────────────────────────────────

const PET_ID_DOG = "101";
const PET_ID_CAT = "102";
const PET_ID_RABBIT = "103";

function makePetResponse(id: string, speciesName: string, speciesId: number) {
  return {
    id: Number(id),
    version: 1,
    clinic_id: 1,
    owner_id: 10,
    animal_species_id: speciesId,
    pet_number: id,
    name: "テストペット",
    pet_name_kana: "",
    gender: "unknown",
    status: "alive",
    breed: "",
    color: "",
    danger_level: "none",
    food: "",
    environment: "",
    phone: "",
    remarks: "",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    animal_species: { id: speciesId, name: speciesName, sort_order: speciesId },
  };
}

function stubPetSpeciesApis() {
  server.use(
    http.get(`/api/v1/pets/${PET_ID_DOG}`, () =>
      HttpResponse.json(makePetResponse(PET_ID_DOG, "犬", 1)),
    ),
    http.get(`/api/v1/pets/${PET_ID_CAT}`, () =>
      HttpResponse.json(makePetResponse(PET_ID_CAT, "猫", 2)),
    ),
    http.get(`/api/v1/pets/${PET_ID_RABBIT}`, () =>
      HttpResponse.json(makePetResponse(PET_ID_RABBIT, "うさぎ", 4)),
    ),
  );
}

describe("CheckupsList — F: 動物種フィルタ (EMR-223)", () => {
  function setupSpeciesRows() {
    vi.mocked(useGetCheckups).mockReturnValue({
      data: makeCheckupsResult([
        makeCheckupRecord({ id: "1", petId: PET_ID_DOG, petName: "ポチ" }),
        makeCheckupRecord({ id: "2", petId: PET_ID_CAT, petName: "たろう" }),
        makeCheckupRecord({ id: "3", petId: PET_ID_RABBIT, petName: "モモ" }),
        makeCheckupRecord({ id: "4", petName: "名無し" }),
      ]),
      isLoading: false,
      error: null,
    } as ReturnType<typeof useGetCheckups>);
    stubPetSpeciesApis();
  }

  async function selectSpeciesFilter(user: ReturnType<typeof userEvent.setup>, label: string) {
    await user.click(screen.getByRole("button", { name: "フィルタを追加" }));
    await user.click(await screen.findByText("動物種"));
    await user.click(await screen.findByText("次と一致"));
    await user.click(await screen.findByText(label));
  }

  it("犬 を選ぶと犬のペットの行だけが残り、全期間検索のAPIパラメータは汚染されない", async () => {
    setupSpeciesRows();
    const user = userEvent.setup();
    render(<CheckupsList />, { wrapper: createWrapper() });

    await screen.findByText("ポチ");
    await selectSpeciesFilter(user, "犬");

    // ペット種の解決後: 犬の行だけ残り、猫・その他・petId なし行は消える
    await waitFor(() => {
      expect(screen.getByText("ポチ")).toBeInTheDocument();
      expect(screen.queryByText("たろう")).not.toBeInTheDocument();
      expect(screen.queryByText("モモ")).not.toBeInTheDocument();
      expect(screen.queryByText("名無し")).not.toBeInTheDocument();
    });

    // species は buildCheckupListFilters の返り値（= GET /v1/checkups パラメータ）に混入しない
    const lastCall = vi.mocked(useGetCheckups).mock.calls.at(-1)?.[0] as CheckupFilters | undefined;
    expect(lastCall).not.toHaveProperty("species");
    expect(lastCall).toEqual({ page: 1, limit: 20 });
  });

  it("その他 を選ぶと犬・猫以外の種の行だけが残る", async () => {
    setupSpeciesRows();
    const user = userEvent.setup();
    render(<CheckupsList />, { wrapper: createWrapper() });

    await screen.findByText("ポチ");
    await selectSpeciesFilter(user, "その他");

    await waitFor(() => {
      expect(screen.getByText("モモ")).toBeInTheDocument();
      expect(screen.queryByText("ポチ")).not.toBeInTheDocument();
      expect(screen.queryByText("たろう")).not.toBeInTheDocument();
    });
  });
});
