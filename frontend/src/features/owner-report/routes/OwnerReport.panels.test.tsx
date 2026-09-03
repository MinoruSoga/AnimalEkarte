import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, Routes, Route, useSearchParams } from "react-router";

import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import type { Resource, ResourceAction } from "@/types/generated/models";
import { buildJSTWallDateTime, todayJSTISO } from "@/lib/jst-date";
import {
  ResourceExaminations,
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

// FE-RC-045: OwnerReport.test.tsx (820行) を分割した1ファイル目（基本パネル表示）。

describe("toPet status mapping (fail-closed)", () => {
  it("既知 status「生存」「死亡」はそのまま写像する", () => {
    expect(toPet(baseReportPet({ status: "生存" }), "42").status).toBe("生存");
    expect(toPet(baseReportPet({ status: "死亡" }), "42").status).toBe("死亡");
  });

  it("未知・欠損 status は「不明」へ fail-closed 写像する", () => {
    expect(toPet(baseReportPet({ status: "pending" }), "42").status).toBe(
      "不明",
    );
    expect(toPet(baseReportPet({ status: "" }), "42").status).toBe("不明");
    expect(
      toPet(
        baseReportPet({ status: undefined as unknown as string }),
        "42",
      ).status,
    ).toBe("不明");
  });
});

describe("OwnerReport", () => {
  it("飼主パネルを常時表示し、タブを使わず6パネルを同時表示する", () => {
    renderReport(makeAuth(allowAll));

    expect(hooks.useGetOwnerReportPets).toHaveBeenCalledWith("42");
    expect(screen.getByText("山田太郎")).toBeInTheDocument();
    const main = screen.getByRole("main");
    const panelNames = [
      "診療前の確認",
      "今日の来院",
      "次の行動",
      "前回診療",
      "基本情報",
      "種類別履歴",
    ];
    for (const panelName of panelNames) {
      expect(
        within(main).getByRole("region", { name: panelName }),
      ).toBeInTheDocument();
      expect(
        within(main).getByRole("heading", { name: panelName }),
      ).toBeInTheDocument();
    }
    expect(within(main).getAllByRole("region")).toHaveLength(6);
    expect(screen.getByTestId("owner-report-viewport")).toHaveClass(
      "fixed",
      "overflow-hidden",
    );
    expect(
      document.querySelectorAll("[data-owner-report-scroll]"),
    ).toHaveLength(6);
    expect(screen.queryByRole("tablist")).not.toBeInTheDocument();
    expect(screen.queryByRole("tab")).not.toBeInTheDocument();
    expect(screen.queryByRole("tabpanel")).not.toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "ペット切替" })).toHaveValue(
      "7",
    );
  });

  it("診療前確認・今日の来院・次の行動・前回診療を優先表示する", () => {
    renderReport(makeAuth(allowAll));

    expect(screen.getByText("飼主ペットレポート")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { name: "ポチ", level: 2 }),
    ).toBeInTheDocument();
    expect(screen.getAllByText(/犬・柴犬/)).not.toHaveLength(0);

    const attention = screen.getByRole("region", { name: "診療前の確認" });
    expect(within(attention).getByText("最新検査")).toBeInTheDocument();
    expect(within(attention).getByText("予防予定")).toBeInTheDocument();

    const today = screen.getByRole("region", { name: "今日の来院" });
    expect(within(today).getByText("10:30")).toBeInTheDocument();
    expect(within(today).getByText("再診")).toBeInTheDocument();
    expect(within(today).getByText("受付済")).toBeInTheDocument();

    const nextAction = screen.getByRole("region", { name: "次の行動" });
    expect(within(nextAction).getByText("次回予約")).toBeInTheDocument();
    expect(within(nextAction).getByText("来院推奨")).toBeInTheDocument();
    expect(within(nextAction).getByText("予防予定")).toBeInTheDocument();

    const previousVisit = screen.getByRole("region", { name: "前回診療" });
    expect(within(previousVisit).getByText("健康診断")).toBeInTheDocument();
  });

  it("レガシー EMR 準拠の飼主・ペット項目（住所/勤務先/ふりがな/年齢/前回来院）を表示する", () => {
    renderReport(makeAuth(allowAll));

    const basicInfo = screen.getByRole("region", { name: "基本情報" });

    // 飼主: 郵便番号 + 住所 を 1 フィールドに連結表示
    expect(
      within(basicInfo).getByText(/〒192-0916 東京都八王子市みなみ野2-8-13/),
    ).toBeInTheDocument();
    // 飼主: 勤務先 / 勤務先電話
    expect(within(basicInfo).getByText("ノア商事")).toBeInTheDocument();
    expect(within(basicInfo).getByText("042-000-0000")).toBeInTheDocument();

    // ペット詳細: ふりがな
    expect(within(basicInfo).getByText("ふりがな")).toBeInTheDocument();
    expect(within(basicInfo).getByText("ぽち")).toBeInTheDocument();
    // ペット詳細: 年齢（birthDate からの導出値。値は基準日依存なので形式で検証）
    expect(within(basicInfo).getByText("年齢")).toBeInTheDocument();
    expect(within(basicInfo).getByText(/^\d+歳\d+ヶ月$/)).toBeInTheDocument();
    // ペット詳細: 前回来院（共有 formatDate で JST 整合の YYYY/MM/DD 表示）
    expect(within(basicInfo).getByText("前回来院")).toBeInTheDocument();
    expect(within(basicInfo).getByText("2024/08/25")).toBeInTheDocument();

    // ペット詳細: 血液型 / マイクロチップ（pets API 由来。データがある時のみ値表示）
    expect(within(basicInfo).getByText("血液型")).toBeInTheDocument();
    expect(within(basicInfo).getByText("DEA1.1陽性")).toBeInTheDocument();
    expect(within(basicInfo).getByText("マイクロチップ")).toBeInTheDocument();
    expect(within(basicInfo).getByText("392140000123456")).toBeInTheDocument();
    // ペット詳細: owner-report 専用 response 由来の既存詳細項目。curated response で落とさない。
    expect(within(basicInfo).getByText(/柴犬/)).toBeInTheDocument();
    expect(within(basicInfo).getByText("赤")).toBeInTheDocument();
    expect(within(basicInfo).getByText("購入")).toBeInTheDocument();
    expect(within(basicInfo).getByText("療法食")).toBeInTheDocument();
    expect(within(basicInfo).getByText(/アニコム/)).toBeInTheDocument();
    expect(within(basicInfo).getByText(/70%補償/)).toBeInTheDocument();
    expect(screen.getByText("咬傷注意")).toBeInTheDocument();
    // #229: 飼主向けサーフェスに危険度を平文表示しない（fixture に dangerLevel があっても非描画）
    expect(screen.queryByText("危険度")).not.toBeInTheDocument();
    expect(screen.queryByText("中")).not.toBeInTheDocument();

    // ペット詳細: 初診日（useGetPetFirstVisit 由来の派生値。formatDate で YYYY/MM/DD 表示）
    expect(within(basicInfo).getByText("初診日")).toBeInTheDocument();
    expect(within(basicInfo).getByText("2022/01/10")).toBeInTheDocument();

    // 飼主: DM 区分（dmPreference=true → 必要）
    expect(within(basicInfo).getByText("DM")).toBeInTheDocument();
    expect(within(basicInfo).getByText("必要")).toBeInTheDocument();
  });

  it("データが無い飼主・ペット項目は行を出さず空状態を壊さない", () => {
    hooks.useGetOwner.mockReturnValue(
      ok({
        id: "42",
        ownerName: "山田太郎",
        phone: "090-1111-2222",
        membershipType: "会員",
        email: "",
        dmPreference: null,
      }),
    );
    hooks.useGetOwnerReportPets.mockReturnValue(
      ok([{ id: "7", name: "ポチ", species: "犬", ownerId: "42" }]),
    );
    // 受診歴なし: 初診日は null（捏造せず "-" 表示）。
    hooks.useGetPetFirstVisit.mockReturnValue(ok(null));
    renderReport(makeAuth(allowAll));

    const basicInfo = screen.getByRole("region", { name: "基本情報" });

    // 住所/勤務先フィールドは出ない（"-" で潰さず非表示）
    expect(within(basicInfo).queryByText("勤務先")).not.toBeInTheDocument();
    expect(within(basicInfo).queryByText("勤務先TEL")).not.toBeInTheDocument();
    // DM 区分は未設定（null）のため行ごと出さない（"不要" と誤表示しない）
    expect(within(basicInfo).queryByText("DM")).not.toBeInTheDocument();
    // ふりがな/年齢/前回来院/血液型/初診日 のラベルは行として残り、値は "-"
    expect(within(basicInfo).getByText("ふりがな")).toBeInTheDocument();
    expect(within(basicInfo).getByText("年齢")).toBeInTheDocument();
    expect(within(basicInfo).getByText("前回来院")).toBeInTheDocument();
    expect(within(basicInfo).getByText("血液型")).toBeInTheDocument();
    expect(within(basicInfo).getByText("初診日")).toBeInTheDocument();
  });

  it("種類別履歴は薬・予防接種・処置を別の縦行にして横を日付順にする", () => {
    renderReport(makeAuth(allowAll));

    const table = screen.getByRole("table", {
      name: "診療履歴を種類別に分け、日付の新しい順に左から表示",
    });
    expect(
      within(table).getByRole("rowheader", { name: /診療/ }),
    ).toBeInTheDocument();
    expect(
      within(table).getByRole("rowheader", { name: /検査/ }),
    ).toBeInTheDocument();
    expect(
      within(table).getByRole("rowheader", { name: /薬・処方/ }),
    ).toBeInTheDocument();
    expect(
      within(table).getByRole("rowheader", { name: /予防接種/ }),
    ).toBeInTheDocument();
    expect(
      within(table).getByRole("rowheader", { name: /処置/ }),
    ).toBeInTheDocument();
    expect(
      within(table).getByRole("rowheader", { name: /ケア/ }),
    ).toBeInTheDocument();
    expect(within(table).getAllByRole("columnheader")[0]).toHaveTextContent(
      "種類",
    );
  });

  it("ペット切替でページ遷移せず ?petId= が同期し、飼主パネルは消えない", async () => {
    const user = userEvent.setup();
    renderReport(makeAuth(allowAll));

    expect(screen.getByTestId("petid-probe")).toHaveTextContent("7");

    await user.selectOptions(
      screen.getByRole("combobox", { name: "ペット切替" }),
      "8",
    );

    // URL ?petId= が 8 に同期
    expect(screen.getByTestId("petid-probe")).toHaveTextContent("8");
    expect(screen.getByRole("combobox", { name: "ペット切替" })).toHaveValue(
      "8",
    );
    // 飼主パネルは残る
    expect(screen.getByText("山田太郎")).toBeInTheDocument();
    expect(hooks.useGetMedicalRecords).toHaveBeenLastCalledWith(
      expect.objectContaining({ petId: "8", status: "finalized" }),
    );
  });

  it("選択中ペットが死亡なら非色依存で表示し、生存ペットへの切替後は表示しない", async () => {
    const user = userEvent.setup();
    hooks.useGetOwnerReportPets.mockReturnValue(
      ok([
        { ...pets[0], status: "死亡" },
        { ...pets[1], status: "生存" },
      ]),
    );
    renderReport(makeAuth(allowAll));

    expect(screen.getByText("死亡")).toBeInTheDocument();

    await user.selectOptions(
      screen.getByRole("combobox", { name: "ペット切替" }),
      "8",
    );

    expect(screen.queryByText("死亡")).not.toBeInTheDocument();

    await user.selectOptions(
      screen.getByRole("combobox", { name: "ペット切替" }),
      "7",
    );

    expect(screen.getByText("死亡")).toBeInTheDocument();
  });

  it("別飼主を含む不正なpetIdは先頭ペットへフォールバックし、問い合わせへ渡さない", () => {
    renderReport(makeAuth(allowAll), "/owners/42/report?petId=999999");

    expect(screen.getByRole("combobox", { name: "ペット切替" })).toHaveValue(
      "7",
    );
    expect(hooks.useGetMedicalRecords).toHaveBeenLastCalledWith(
      expect.objectContaining({ petId: "7", status: "finalized" }),
    );
    expect(hooks.useGetPetTreatmentHistory).toHaveBeenCalledWith("7", "all");
    expect(hooks.useGetPetExaminations).not.toHaveBeenCalledWith("999999");
  });

  it("セクション個別の権限不足はページ全体を落とさず縮退表示する", () => {
    // vaccinations だけ拒否（medical-records と examinations は許可）
    const auth = makeAuth((resource) => resource !== ResourceVaccinations);
    renderReport(auth);

    // 予防接種だけ縮退し、検査は表示する
    const attention = screen.getByRole("region", { name: "診療前の確認" });
    expect(within(attention).getByText("閲覧権限なし")).toBeInTheDocument();
    expect(screen.queryByText("狂犬病ワクチン")).not.toBeInTheDocument();
    expect(screen.getByText("血液検査")).toBeInTheDocument();
    expect(hooks.useGetPetVaccinations).toHaveBeenCalledWith(undefined);
  });

  it("検査の閲覧権限が無い場合は要確認件数をサマリーへ露出しない", () => {
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
                name: "ALT",
                inspectionValue: "180",
                result: "",
                unit: "U/L",
                referenceValue: "17-78",
                isAbnormal: true,
                status: "high",
              },
            ],
          },
        ],
        isTruncated: false,
      }),
    );
    const auth = makeAuth((resource) => resource !== ResourceExaminations);

    renderReport(auth);

    const attention = screen.getByRole("region", { name: "診療前の確認" });
    expect(within(attention).getByText("閲覧権限なし")).toBeInTheDocument();
    expect(within(attention).queryByText("基準外 1件")).not.toBeInTheDocument();
    expect(screen.getByText("一部閲覧権限なし")).toBeInTheDocument();
    expect(hooks.useGetPetExaminations).toHaveBeenCalledWith(undefined);
  });

  it("サマリー取得失敗を0件や予定なしとして誤表示しない", () => {
    const failed = { data: undefined, isLoading: false, isError: true };
    hooks.useGetPetExaminations.mockReturnValue(failed);
    hooks.useGetPetVaccinations.mockReturnValue(failed);
    hooks.useGetMedicalRecords.mockReturnValue(failed);
    hooks.useGetPetTreatmentHistory.mockReturnValue(failed);
    hooks.useGetReservations.mockReturnValue(failed);

    renderReport(makeAuth(allowAll));

    const attention = screen.getByRole("region", { name: "診療前の確認" });
    expect(within(attention).getAllByText("取得失敗")).toHaveLength(3);
    expect(within(attention).queryByText("予定なし")).not.toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "前回診療" })).getByText(
        "取得失敗",
      ),
    ).toBeInTheDocument();
    const nextAction = screen.getByRole("region", { name: "次の行動" });
    expect(within(nextAction).getAllByText("取得失敗")).toHaveLength(3);
    expect(within(nextAction).queryByText("予約なし")).not.toBeInTheDocument();
    expect(within(nextAction).queryByText("記録なし")).not.toBeInTheDocument();
    expect(within(nextAction).queryByText("予定なし")).not.toBeInTheDocument();
  });
});
