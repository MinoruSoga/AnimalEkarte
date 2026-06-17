import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Routes, Route, useSearchParams } from "react-router";

import { AuthContext } from "@/contexts/auth-context";
import type { AuthContextValue } from "@/types/auth";
import type { Resource, ResourceAction } from "@/types/generated/models";
import { ResourceVaccinations, ResourceMedicalRecords } from "@/types/generated/models";

import { OwnerReport } from "../routes/OwnerReport";

// ---- data hooks をモックしてセクションデータを決定的に注入する ----
const hooks = vi.hoisted(() => ({
  useGetOwner: vi.fn(),
  useGetPets: vi.fn(),
  useGetPetVaccinations: vi.fn(),
  useGetPetExaminations: vi.fn(),
  useGetPetTreatmentHistory: vi.fn(),
}));

vi.mock("@/features/owners", () => ({ useGetOwner: hooks.useGetOwner }));
vi.mock("@/features/pets", () => ({ useGetPets: hooks.useGetPets }));
vi.mock("@/features/medical-records", () => ({
  useGetPetVaccinations: hooks.useGetPetVaccinations,
}));
vi.mock("../api/get-pet-examinations", () => ({
  useGetPetExaminations: hooks.useGetPetExaminations,
}));
vi.mock("../api/get-pet-treatment-history", () => ({
  useGetPetTreatmentHistory: hooks.useGetPetTreatmentHistory,
}));

const owner = {
  id: "42",
  ownerName: "山田太郎",
  ownerNameKana: "ヤマダタロウ",
  phone: "090-1111-2222",
  membershipType: "会員",
  email: "",
};

const pets = [
  { id: "7", name: "ポチ", species: "犬", ownerId: "42" },
  { id: "8", name: "タマ", species: "猫", ownerId: "42" },
];

function ok<T>(data: T) {
  return { data, isLoading: false, isError: false };
}

function SearchParamProbe() {
  const [params] = useSearchParams();
  return <span data-testid="petid-probe">{params.get("petId") ?? ""}</span>;
}

function makeAuth(hasPermission: (r: Resource, a: ResourceAction) => boolean): AuthContextValue {
  return {
    user: null,
    currentClinicId: "1",
    isAuthenticated: true,
    isLoading: false,
    isSwitchingClinic: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission,
    refreshPermissions: async () => {},
  };
}

function renderReport(
  auth: AuthContextValue,
  initialPath = "/owners/42/report?petId=7",
) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <AuthContext.Provider value={auth}>
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={[initialPath]}>
          <Routes>
            <Route
              path="/owners/:id/report"
              element={
                <>
                  <OwnerReport />
                  <SearchParamProbe />
                </>
              }
            />
            <Route path="/login" element={<div>ログイン画面</div>} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    </AuthContext.Provider>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  hooks.useGetOwner.mockReturnValue(ok(owner));
  hooks.useGetPets.mockReturnValue(ok(pets));
  hooks.useGetPetVaccinations.mockReturnValue(
    ok([{ id: 1, name: "狂犬病ワクチン", date: "25/5/1", next: "26/5/1" }]),
  );
  hooks.useGetPetExaminations.mockReturnValue(
    ok([
      {
        id: "1",
        testType: "血液検査",
        date: "2026-05-10",
        status: "確定",
        items: [
          {
            id: "1",
            name: "WBC",
            inspectionValue: "120",
            result: "",
            unit: "",
            referenceValue: "60-120",
            status: "normal",
          },
        ],
      },
    ]),
  );
  hooks.useGetPetTreatmentHistory.mockImplementation(
    (_petId: string | undefined, filter: string) => {
      if (filter === "medicine") {
        return ok([
          { id: "m1", date: "25/5/1", itemType: "medicine", name: "アモキシシリン", adminRoute: "経口", quantity: 1, medicalRecordId: "9" },
        ]);
      }
      if (filter === "procedure") {
        return ok([
          { id: "p1", date: "25/5/2", itemType: "procedure", name: "避妊手術", adminRoute: "", quantity: 1, anesthesia: "全身麻酔", medicalRecordId: "9" },
        ]);
      }
      return ok([
        { id: "m1", date: "25/5/1", itemType: "medicine", name: "アモキシシリン", adminRoute: "経口", quantity: 1, medicalRecordId: "9" },
        { id: "p1", date: "25/5/2", itemType: "procedure", name: "避妊手術", adminRoute: "", quantity: 1, anesthesia: "全身麻酔", medicalRecordId: "9" },
      ]);
    },
  );
});

const allowAll = () => true;

describe("OwnerReport", () => {
  it("飼主パネルを常時表示し、初期 petId のペットで6セクションを描画する", () => {
    renderReport(makeAuth(allowAll));

    // R4: 飼主情報
    expect(screen.getByText("山田太郎")).toBeInTheDocument();

    // 6 セクション
    expect(screen.getByText("ペット詳細")).toBeInTheDocument();
    expect(screen.getByText("予防接種履歴")).toBeInTheDocument();
    expect(screen.getByText("健康診断（検査）履歴")).toBeInTheDocument();
    expect(screen.getByText("投薬履歴")).toBeInTheDocument();
    expect(screen.getByText("手術・処置履歴")).toBeInTheDocument();
    expect(screen.getByText("治療履歴")).toBeInTheDocument();

    // 各セクションのデータ
    expect(screen.getByText("狂犬病ワクチン")).toBeInTheDocument();
    expect(screen.getByText("血液検査")).toBeInTheDocument();
    expect(screen.getAllByText("アモキシシリン").length).toBeGreaterThan(0);
    expect(screen.getAllByText("避妊手術").length).toBeGreaterThan(0);

    // 初期選択 = petId=7 (ポチ)
    const pochiTab = screen.getByRole("tab", { name: /ポチ/ });
    expect(pochiTab).toHaveAttribute("aria-selected", "true");
  });

  it("ペット切替でページ遷移せず ?petId= が同期し、飼主パネルは消えない", async () => {
    const user = userEvent.setup();
    renderReport(makeAuth(allowAll));

    expect(screen.getByTestId("petid-probe")).toHaveTextContent("7");

    await user.click(screen.getByRole("tab", { name: /タマ/ }));

    // URL ?petId= が 8 に同期
    expect(screen.getByTestId("petid-probe")).toHaveTextContent("8");
    // タマがアクティブ
    expect(screen.getByRole("tab", { name: /タマ/ })).toHaveAttribute("aria-selected", "true");
    // 飼主パネルは残る
    expect(screen.getByText("山田太郎")).toBeInTheDocument();
  });

  it("セクション個別の権限不足はページ全体を落とさず縮退表示する", () => {
    // vaccinations だけ拒否（medical-records と examinations は許可）
    const auth = makeAuth((resource) => resource !== ResourceVaccinations);
    renderReport(auth);

    // 予防接種セクションは縮退、検査セクションは通常表示
    expect(screen.getByText("閲覧権限がありません")).toBeInTheDocument();
    expect(screen.queryByText("狂犬病ワクチン")).not.toBeInTheDocument();
    expect(screen.getByText("血液検査")).toBeInTheDocument();
  });

  it("履歴ゼロのセクションは空状態を表示する", () => {
    hooks.useGetPetTreatmentHistory.mockReturnValue(ok([]));
    renderReport(makeAuth(allowAll));

    expect(screen.getByText("投薬の履歴はありません")).toBeInTheDocument();
    expect(screen.getByText("治療の履歴はありません")).toBeInTheDocument();
  });

  it("medical-records:view が無ければアクセス拒否を表示する", () => {
    const auth = makeAuth(() => false);
    renderReport(auth);

    expect(screen.getByText("アクセス権限がありません")).toBeInTheDocument();
    expect(screen.queryByText("山田太郎")).not.toBeInTheDocument();
  });

  it("未認証ならログインへリダイレクトする", () => {
    const auth = makeAuth(allowAll);
    auth.isAuthenticated = false;
    renderReport(auth);

    expect(screen.getByText("ログイン画面")).toBeInTheDocument();
  });
});
