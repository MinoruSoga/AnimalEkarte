import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Routes, Route, useSearchParams } from "react-router";

import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import type { Resource, ResourceAction } from "@/types/generated/models";
import { buildJSTWallDateTime, todayJSTISO } from "@/lib/jst-date";
import {
  ResourceExaminations,
  ResourceOwners,
  ResourceVaccinations,
} from "@/types/generated/models";

import type { OwnerReportPet } from "../api/get-owner-report-pets";
import { toPet } from "../lib/owner-report-pet";
import { OwnerReport } from "./OwnerReport";

// ---- data hooks をモックしてセクションデータを決定的に注入する ----
const hooks = vi.hoisted(() => ({
  useGetOwner: vi.fn(),
  useGetOwnerReportPets: vi.fn(),
  useGetPetVaccinations: vi.fn(),
  useGetPetExaminations: vi.fn(),
  useGetPetTreatmentHistory: vi.fn(),
  useGetPetTrimmingHistory: vi.fn(),
  useGetPetFirstVisit: vi.fn(),
  useGetPetCheckupResults: vi.fn(),
  useGetMedicalRecords: vi.fn(),
  useGetReservations: vi.fn(),
}));

vi.mock("@/hooks/use-owner", () => ({ useGetOwner: hooks.useGetOwner }));
vi.mock("../api/get-owner-report-pets", () => ({
  useGetOwnerReportPets: hooks.useGetOwnerReportPets,
}));
vi.mock("@/hooks/use-pet-vaccinations", () => ({
  useGetPetVaccinations: hooks.useGetPetVaccinations,
}));
vi.mock("../api/get-pet-examinations", () => ({
  useGetPetExaminations: hooks.useGetPetExaminations,
}));
vi.mock("../api/get-pet-treatment-history", () => ({
  useGetPetTreatmentHistory: hooks.useGetPetTreatmentHistory,
}));
vi.mock("../api/get-pet-trimming-history", () => ({
  useGetPetTrimmingHistory: hooks.useGetPetTrimmingHistory,
}));
vi.mock("../api/get-pet-first-visit", () => ({
  useGetPetFirstVisit: hooks.useGetPetFirstVisit,
}));
vi.mock("../hooks/use-pet-checkup-results", () => ({
  useGetPetCheckupResults: hooks.useGetPetCheckupResults,
}));
vi.mock("@/hooks/use-medical-records", () => ({
  useGetMedicalRecords: hooks.useGetMedicalRecords,
}));
vi.mock("@/hooks/use-get-reservations", () => ({
  useGetReservations: hooks.useGetReservations,
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
    breed: "柴犬",
    color: "赤",
    bloodType: "DEA1.1陽性",
    microchipNumber: "392140000123456",
    neuteredDate: "2016-05-20",
    acquisitionType: "購入",
    dangerLevel: "中",
    food: "療法食",
    environment: "室内",
    insuranceName: "アニコム",
    insuranceDetails: "70%補償",
    remarks: "咬傷注意",
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

function makeAuth(
  hasPermission: (r: Resource, a: ResourceAction) => boolean,
): AuthContextValue {
  return {
    user: null,
    currentClinicId: "1",
    isAuthenticated: true,
    isLoading: false,
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
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
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
  hooks.useGetOwnerReportPets.mockReturnValue(ok(pets));
  hooks.useGetPetVaccinations.mockReturnValue(
    ok([
      {
        id: 1,
        name: "狂犬病ワクチン",
        date: "25/5/1",
        next: "27/5/1",
        nextDate: "2027-05-01",
      },
    ]),
  );
  hooks.useGetPetExaminations.mockReturnValue(
    ok({
      items: [
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
              isAbnormal: false,
              status: "normal",
            },
          ],
        },
      ],
      isTruncated: false,
    }),
  );
  hooks.useGetPetFirstVisit.mockReturnValue(ok("2022-01-10"));
  hooks.useGetPetCheckupResults.mockReturnValue(ok([]));
  hooks.useGetMedicalRecords.mockReturnValue(
    ok({
      data: [
        {
          id: "mr2",
          date: "2026/05/10",
          chiefComplaint: "健康診断",
          doctor: "佐藤",
          visitType: "再診",
          nextVisitRecommendedDate: "",
        },
        {
          id: "mr1",
          date: "2026/03/01",
          chiefComplaint: "外耳炎",
          doctor: "田中",
          visitType: "再診",
          nextVisitRecommendedDate: "2026-04-01",
        },
      ],
      total: 2,
      page: 1,
      limit: 5,
    }),
  );
  hooks.useGetReservations.mockReturnValue(
    ok([
      {
        id: "reservation-today",
        start: buildJSTWallDateTime(todayJSTISO(), "10:30"),
        end: buildJSTWallDateTime(todayJSTISO(), "11:00"),
        ownerName: "山田太郎",
        petName: "ポチ",
        visitType: "revisit",
        type: "健康診断",
        doctor: "佐藤",
        status: "checked_in",
        petId: "7",
        ownerId: "42",
      },
    ]),
  );
  hooks.useGetPetTrimmingHistory.mockReturnValue(
    ok({
      items: [
        {
          id: "t1",
          date: "2026-02-01",
          status: "完了",
          courseName: "シャンプー＆カット",
          staff: "鈴木",
        },
      ],
      isTruncated: false,
    }),
  );
  hooks.useGetPetTreatmentHistory.mockImplementation(
    (_petId: string | undefined, filter: string) => {
      if (filter === "medicine") {
        return ok({
          items: [
            {
              id: "m1",
              date: "25/5/1",
              itemType: "medicine",
              name: "アモキシシリン",
              adminRoute: "経口",
              quantity: 1,
              medicalRecordId: "9",
            },
          ],
          isTruncated: false,
        });
      }
      if (filter === "procedure") {
        return ok({
          items: [
            {
              id: "p1",
              date: "25/5/2",
              itemType: "procedure",
              name: "避妊手術",
              adminRoute: "",
              quantity: 1,
              anesthesia: "全身麻酔",
              medicalRecordId: "9",
            },
          ],
          isTruncated: false,
        });
      }
      return ok({
        items: [
          {
            id: "m1",
            date: "25/5/1",
            itemType: "medicine",
            name: "アモキシシリン",
            adminRoute: "経口",
            quantity: 1,
            medicalRecordId: "9",
          },
          {
            id: "p1",
            date: "25/5/2",
            itemType: "procedure",
            name: "避妊手術",
            adminRoute: "",
            quantity: 1,
            anesthesia: "全身麻酔",
            medicalRecordId: "9",
          },
        ],
        isTruncated: false,
      });
    },
  );
});

const allowAll = () => true;

function baseReportPet(
  overrides: Partial<OwnerReportPet> = {},
): OwnerReportPet {
  return {
    id: "7",
    name: "ポチ",
    petNameKana: "",
    gender: "",
    status: "生存",
    breed: "",
    color: "",
    food: "",
    environment: "",
    remarks: "",
    species: "犬",
    ...overrides,
  };
}

// FE-RC-045: OwnerReport.test.tsx (820行) を分割した2ファイル目（サマリー/権限/認証境界）。

describe("OwnerReport — サマリー/権限/認証境界", () => {
  it("サマリー読み込み中は空状態へ置き換えず各領域に進行状態を表示する", () => {
    const loading = { data: undefined, isLoading: true, isError: false };
    hooks.useGetPetExaminations.mockReturnValue(loading);
    hooks.useGetPetVaccinations.mockReturnValue(loading);
    hooks.useGetMedicalRecords.mockReturnValue(loading);
    hooks.useGetPetTreatmentHistory.mockReturnValue(loading);
    hooks.useGetReservations.mockReturnValue(loading);

    renderReport(makeAuth(allowAll));

    expect(
      within(screen.getByRole("region", { name: "診療前の確認" })).getAllByText(
        "読み込み中...",
      ),
    ).toHaveLength(3);
    expect(
      within(screen.getByRole("region", { name: "今日の来院" })).getByText(
        "読み込み中...",
      ),
    ).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "次の行動" })).getAllByText(
        "読み込み中...",
      ),
    ).toHaveLength(3);
    expect(
      within(screen.getByRole("region", { name: "前回診療" })).getByText(
        "読み込み中...",
      ),
    ).toBeInTheDocument();
  });

  it("選択ペットごとの履歴を既存条件で一度ずつ取得する", () => {
    renderReport(makeAuth(allowAll));

    expect(hooks.useGetMedicalRecords).toHaveBeenCalledTimes(1);
    expect(hooks.useGetMedicalRecords).toHaveBeenCalledWith({
      petId: "7",
      status: "finalized",
      page: 1,
      limit: HISTORY_FETCH_LIMIT,
      sort: "date",
      order: "desc",
    });
    expect(hooks.useGetPetTreatmentHistory).toHaveBeenCalledTimes(1);
    expect(hooks.useGetPetTreatmentHistory).toHaveBeenCalledWith("7", "all");
    expect(hooks.useGetReservations).toHaveBeenCalledTimes(1);
    expect(hooks.useGetReservations).toHaveBeenCalledWith(
      expect.objectContaining({
        petId: "7",
        enabled: true,
        startDate: todayJSTISO(),
      }),
    );
  });

  it("初診日の取得失敗を未登録として表示しない", () => {
    hooks.useGetPetFirstVisit.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    });
    renderReport(makeAuth(allowAll));

    const basicInfo = screen.getByRole("region", { name: "基本情報" });
    expect(within(basicInfo).getByText("初診日")).toBeInTheDocument();
    expect(within(basicInfo).getByText("取得失敗")).toBeInTheDocument();
  });

  it("診療記録が取得上限を超える場合は履歴件数を確定値にしない", () => {
    hooks.useGetMedicalRecords.mockReturnValue(
      ok({
        data: [{ id: "mr2", date: "2026/05/10", chiefComplaint: "健康診断" }],
        total: 101,
        page: 1,
        limit: 100,
      }),
    );
    renderReport(makeAuth(allowAll));

    expect(
      within(screen.getByRole("region", { name: "種類別履歴" })).getByText(
        /\d+件\+/,
      ),
    ).toBeInTheDocument();
  });

  it("ペット一覧取得失敗を登録ペットなしとして表示しない", () => {
    hooks.useGetOwnerReportPets.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
    });
    renderReport(makeAuth(allowAll));

    expect(screen.getByRole("alert")).toHaveTextContent(
      "飼主・ペット情報の取得に失敗しました",
    );
    expect(
      screen.queryByText("この飼主に登録されたペットがありません"),
    ).not.toBeInTheDocument();
  });

  it("履歴ゼロのセクションは空状態を表示する", () => {
    hooks.useGetPetTreatmentHistory.mockReturnValue(
      ok({ items: [], isTruncated: false }),
    );
    renderReport(makeAuth(allowAll));

    const table = screen.getByRole("table");
    expect(
      within(table).getByRole("rowheader", { name: /薬・処方0件/ }),
    ).toBeInTheDocument();
    expect(
      within(table).getByRole("rowheader", { name: /処置0件/ }),
    ).toBeInTheDocument();
  });

  it("medical-records:view が無ければアクセス拒否を表示する", () => {
    const auth = makeAuth(() => false);
    renderReport(auth);

    expect(screen.getByText("アクセス権限がありません")).toBeInTheDocument();
    expect(screen.queryByText("山田太郎")).not.toBeInTheDocument();
  });

  it("owners:view が無ければキャッシュ済みの飼主・ペット情報も表示しない", () => {
    const auth = makeAuth((resource) => resource !== ResourceOwners);
    renderReport(auth);

    expect(screen.getByText("アクセス権限がありません")).toBeInTheDocument();
    expect(screen.queryByText("山田太郎")).not.toBeInTheDocument();
    expect(hooks.useGetOwner).not.toHaveBeenCalled();
    expect(hooks.useGetOwnerReportPets).not.toHaveBeenCalled();
  });

  it("未認証ならログインへリダイレクトする", () => {
    const auth = makeAuth(allowAll);
    auth.isAuthenticated = false;
    renderReport(auth);

    expect(screen.getByText("ログイン画面")).toBeInTheDocument();
  });
});
