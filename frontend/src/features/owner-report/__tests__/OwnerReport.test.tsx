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
  useGetPetFirstVisit: vi.fn(),
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
vi.mock("../api/get-pet-first-visit", () => ({
  useGetPetFirstVisit: hooks.useGetPetFirstVisit,
}));

const owner = {
  id: "42",
  ownerName: "山田太郎",
  ownerNameKana: "ヤマダタロウ",
  phone: "090-1111-2222",
  membershipType: "会員",
  email: "",
  postalCode: "192-0916",
  address1: "東京都八王子市みなみ野2-8-13",
  address2: "",
  company: "ノア商事",
  companyPhone: "042-000-0000",
  dmPreference: true,
};

const pets = [
  {
    id: "7",
    name: "ポチ",
    species: "犬",
    ownerId: "42",
    petNameKana: "ぽち",
    birthDate: "2015-04-14",
    bloodType: "DEA1.1陽性",
    microchipNumber: "392140000123456",
    // transformBackendPetToFrontend 正規化後の形（日付のみ）を模す。
    lastVisit: "2024-08-25",
  },
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
  hooks.useGetPetFirstVisit.mockReturnValue(ok("2022-01-10"));
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

  it("レガシー EMR 準拠の飼主・ペット項目（住所/勤務先/ふりがな/年齢/前回来院）を表示する", () => {
    renderReport(makeAuth(allowAll));

    // 飼主: 郵便番号 + 住所 を 1 フィールドに連結表示
    expect(screen.getByText(/〒192-0916 東京都八王子市みなみ野2-8-13/)).toBeInTheDocument();
    // 飼主: 勤務先 / 勤務先電話
    expect(screen.getByText("ノア商事")).toBeInTheDocument();
    expect(screen.getByText("042-000-0000")).toBeInTheDocument();

    // ペット詳細: ふりがな
    expect(screen.getByText("ふりがな")).toBeInTheDocument();
    expect(screen.getByText("ぽち")).toBeInTheDocument();
    // ペット詳細: 年齢（birthDate からの導出値。値は基準日依存なので形式で検証）
    expect(screen.getByText("年齢")).toBeInTheDocument();
    expect(screen.getByText(/^\d+歳\d+ヶ月$/)).toBeInTheDocument();
    // ペット詳細: 前回来院（共有 formatDate で JST 整合の YYYY/MM/DD 表示）
    expect(screen.getByText("前回来院")).toBeInTheDocument();
    expect(screen.getByText("2024/08/25")).toBeInTheDocument();

    // ペット詳細: 血液型 / マイクロチップ（pets API 由来。データがある時のみ値表示）
    expect(screen.getByText("血液型")).toBeInTheDocument();
    expect(screen.getByText("DEA1.1陽性")).toBeInTheDocument();
    expect(screen.getByText("マイクロチップ")).toBeInTheDocument();
    expect(screen.getByText("392140000123456")).toBeInTheDocument();

    // ペット詳細: 初診日（useGetPetFirstVisit 由来の派生値。formatDate で YYYY/MM/DD 表示）
    expect(screen.getByText("初診日")).toBeInTheDocument();
    expect(screen.getByText("2022/01/10")).toBeInTheDocument();

    // 飼主: DM 区分（dmPreference=true → 必要）
    expect(screen.getByText("DM")).toBeInTheDocument();
    expect(screen.getByText("必要")).toBeInTheDocument();
  });

  it("データが無い飼主・ペット項目は行を出さず空状態を壊さない", () => {
    hooks.useGetOwner.mockReturnValue(
      ok({ id: "42", ownerName: "山田太郎", phone: "090-1111-2222", membershipType: "会員", email: "" }),
    );
    hooks.useGetPets.mockReturnValue(ok([{ id: "7", name: "ポチ", species: "犬", ownerId: "42" }]));
    // 受診歴なし: 初診日は null（捏造せず "-" 表示）。
    hooks.useGetPetFirstVisit.mockReturnValue(ok(null));
    renderReport(makeAuth(allowAll));

    // 住所/勤務先フィールドは出ない（"-" で潰さず非表示）
    expect(screen.queryByText("勤務先")).not.toBeInTheDocument();
    expect(screen.queryByText("勤務先TEL")).not.toBeInTheDocument();
    // DM 区分は未設定（undefined）のため行ごと出さない（"不要" と誤表示しない）
    expect(screen.queryByText("DM")).not.toBeInTheDocument();
    // ふりがな/年齢/前回来院/血液型/初診日 のラベルは行として残り、値は "-"
    expect(screen.getByText("ふりがな")).toBeInTheDocument();
    expect(screen.getByText("年齢")).toBeInTheDocument();
    expect(screen.getByText("前回来院")).toBeInTheDocument();
    expect(screen.getByText("血液型")).toBeInTheDocument();
    expect(screen.getByText("初診日")).toBeInTheDocument();
  });

  it("6 セクションは見出しを名前に持つ region ランドマークとして密集ワークスペースに並ぶ", () => {
    renderReport(makeAuth(allowAll));

    const regions = screen.getAllByRole("region");
    const names = regions.map((r) => r.getAttribute("aria-label") ?? r.textContent ?? "");
    // 6 セクションがそれぞれ独立した region になっている（タブ等の背後に隠れず一画面に同時提示）。
    expect(regions).toHaveLength(6);
    for (const title of [
      "ペット詳細",
      "予防接種履歴",
      "健康診断（検査）履歴",
      "投薬履歴",
      "手術・処置履歴",
      "治療履歴",
    ]) {
      expect(screen.getByRole("region", { name: title })).toBeInTheDocument();
    }
    expect(names.length).toBe(6);
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
