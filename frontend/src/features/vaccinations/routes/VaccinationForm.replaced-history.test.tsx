import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { VaccinationForm } from "./VaccinationForm";

/**
 * EMR-428: 履歴カードの期限警告は「ペット×ワクチンの最新の接種記録」だけに出す。
 * 置き換わり判定は絞り込み前の配列（useGetVaccinations の結果）から計算するため、
 * 検索語で新しい記録が隠れても古い記録に警告が戻らない（受入条件 2）。
 */

const mocks = vi.hoisted(() => ({
  canEdit: true,
  historySearchTerm: "",
  useVaccinationForm: vi.fn(),
  vaccinations: [] as Array<Record<string, unknown>>,
}));

vi.mock("react-router", async (importOriginal) => ({
  ...(await importOriginal<typeof import("react-router")>()),
  useNavigate: () => vi.fn(),
  useParams: () => ({ id: undefined }),
  useLocation: () => ({ state: null }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canCreate: false, canEdit: mocks.canEdit, canDelete: false }),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({ isDirty: false, markDirty: vi.fn(), markClean: vi.fn() }),
}));

vi.mock("../hooks/use-vaccination-form", () => ({
  useVaccinationForm: mocks.useVaccinationForm,
}));

vi.mock("../api/get-vaccinations", () => ({
  useGetVaccinations: () => ({ data: mocks.vaccinations }),
}));

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: ({ children, headerAction }: { children: ReactNode; headerAction?: ReactNode }) => (
    <div>
      {headerAction}
      {children}
    </div>
  ),
}));

vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

vi.mock("@/components/shared/ConfirmDialog/ConfirmDialog", () => ({
  ConfirmDialog: () => null,
}));

vi.mock("@/components/shared/PatientInfoCard", () => ({
  PatientInfoCard: () => null,
  formatPatientPetDetails: () => "不明 / 不明 / 不明",
}));

function record(overrides: Record<string, unknown>) {
  return {
    id: "v-old",
    petId: "pet-1",
    ownerName: "山田 太郎",
    petName: "ポチ",
    vaccineId: "10",
    vaccineName: "混合ワクチン",
    doctor: null,
    date: "2026-01-01",
    nextDate: "2026-07-23",
    nextScheduleType: "1year",
    lot1: null,
    lot2: null,
    lot3: null,
    lot4: null,
    supplemental: null,
    remarks: null,
    status: "pending",
    version: 1,
    ...overrides,
  };
}

beforeEach(() => {
  mocks.canEdit = true;
  mocks.historySearchTerm = "";
  // 接種日の新しい順（get-vaccinations の並び）
  mocks.vaccinations = [
    record({
      id: "v-new",
      date: "2026-02-01",
      nextDate: "2026-07-01",
      vaccineName: "フィラリア予防",
    }),
    record({ id: "v-old", date: "2026-01-01", nextDate: "2026-07-23" }),
  ];
  mocks.useVaccinationForm.mockImplementation(() => ({
    isEdit: false,
    isReadLoading: false,
    isEditPetReady: true,
    petSelection: {
      selectedPets: [{ id: "pet-1", ownerId: "owner-1", name: "ポチ", status: "生存" }],
    },
    form: {
      doctorName: "",
      date: "",
      setDate: vi.fn(),
      vaccineId: "",
      setVaccineId: vi.fn(),
      vaccineOptions: [],
      nextScheduleType: "1year",
      setNextScheduleType: vi.fn(),
      nextDate: "",
      setNextDate: vi.fn(),
      supplemental: "",
      setSupplemental: vi.fn(),
      lot1: "",
      setLot1: vi.fn(),
      lot2: "",
      setLot2: vi.fn(),
      lot3: "",
      setLot3: vi.fn(),
      lot4: "",
      setLot4: vi.fn(),
      remarks: "",
      setRemarks: vi.fn(),
    },
    formAction: vi.fn(),
    formState: { success: false, timestamp: 0 },
    isSaving: false,
    fieldErrors: {},
    handleDelete: vi.fn(),
    isDeleting: false,
    historyFilter: {
      filterStartDate: "",
      setFilterStartDate: vi.fn(),
      filterEndDate: "",
      setFilterEndDate: vi.fn(),
      historySearchTerm: mocks.historySearchTerm,
      setHistorySearchTerm: vi.fn(),
      sortOrder: "desc",
      setSortOrder: vi.fn(),
      handleClearHistoryFilter: vi.fn(),
    },
  }));
});

describe("VaccinationForm 履歴カードの期限警告は最新の記録だけ (EMR-428)", () => {
  it("同じワクチンのより新しい接種がある古い記録は警告を出さず、最新には出す", () => {
    render(
      <MemoryRouter>
        <VaccinationForm />
      </MemoryRouter>,
    );

    const oldCard = screen.getByText("次回: 2026/07/23").closest("div.flex");
    expect(oldCard).not.toBeNull();
    expect(oldCard).not.toHaveTextContent("期限切れ");

    expect(screen.getByText("次回: 2026/07/01").closest("div.flex")).toHaveTextContent("期限切れ");
  });

  it("検索語で新しい記録が隠れても、古い記録に警告が戻らない（受入条件 2）", () => {
    mocks.historySearchTerm = "混合";
    render(
      <MemoryRouter>
        <VaccinationForm />
      </MemoryRouter>,
    );

    // 新しい記録（フィラリア予防）は検索で隠れる
    expect(screen.queryByText("次回: 2026/07/01")).not.toBeInTheDocument();
    // それでも古い記録に警告は戻らない
    const oldCard = screen.getByText("次回: 2026/07/23").closest("div.flex");
    expect(oldCard).not.toBeNull();
    expect(oldCard).not.toHaveTextContent("期限切れ");
  });
});
