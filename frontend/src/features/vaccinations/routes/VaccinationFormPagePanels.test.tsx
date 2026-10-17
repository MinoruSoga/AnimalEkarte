import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { VaccinationFormBody } from "./VaccinationFormPagePanels";

// 本テストは PatientInfoCard 経路の誕生月バッジのみ検証する。
// ページ枠・ナビゲーション制御・配下パネル・削除確認はスタブ化する。
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
vi.mock("../components/VaccinationFormPanels", () => ({
  VaccinationFieldsPanel: () => null,
  VaccinationHistoryPanel: () => null,
}));
vi.mock("@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png", () => ({ default: "" }));

const basePet = {
  id: "pet-1",
  ownerId: "owner-1",
  ownerName: "山田 太郎",
  name: "ポチ",
  petNumber: "0001",
  status: "生存",
};

const baseForm = {
  doctorName: "佐藤医師",
  date: "",
  vaccineId: "",
  vaccineOptions: [],
  supplemental: "",
  lot1: "",
  lot2: "",
  lot3: "",
  lot4: "",
  nextScheduleType: "",
  nextDate: "",
  remarks: "",
  setDate: vi.fn(),
  setVaccineId: vi.fn(),
  setSupplemental: vi.fn(),
  setLot1: vi.fn(),
  setLot2: vi.fn(),
  setLot3: vi.fn(),
  setLot4: vi.fn(),
  setNextScheduleType: vi.fn(),
  setNextDate: vi.fn(),
  setRemarks: vi.fn(),
};

const baseHistoryFilter = {
  filterStartDate: "",
  filterEndDate: "",
  historySearchTerm: "",
  sortOrder: "desc" as const,
  setFilterStartDate: vi.fn(),
  setFilterEndDate: vi.fn(),
  setHistorySearchTerm: vi.fn(),
  setSortOrder: vi.fn(),
  handleClearHistoryFilter: vi.fn(),
};

function renderBody({
  selectedPet = basePet,
  isPetDeceased = false,
}: {
  selectedPet?: typeof basePet & { birthDate?: string };
  isPetDeceased?: boolean;
} = {}) {
  render(
    <MemoryRouter>
      <VaccinationFormBody
        isEdit={false}
        canSubmit={true}
        canDelete={false}
        isPetDeceased={isPetDeceased}
        isDeleting={false}
        isDirty={false}
        isSaving={false}
        selectedPet={selectedPet}
        form={baseForm}
        fieldErrors={{}}
        petHistory={[]}
        historyFilter={baseHistoryFilter}
        deleteConfirmOpen={false}
        formAction={vi.fn()}
        onBack={vi.fn()}
        onMarkDirty={vi.fn()}
        onOpenDeleteConfirm={vi.fn()}
        onCloseDeleteConfirm={vi.fn()}
        onConfirmDelete={vi.fn()}
      />
    </MemoryRouter>,
  );
}

describe("VaccinationFormPagePanels 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月の生存ペットは PatientInfoCard 上に Cake バッジを表示する", () => {
    renderBody({ selectedPet: { ...basePet, birthDate: "2018-10-05" } });

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05" as const],
    ["birthDate 未登録", undefined],
  ])("%s なら Cake バッジを表示しない", (_label, birthDate) => {
    renderBody({ selectedPet: { ...basePet, birthDate } });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（status=死亡）は誕生月でも Cake バッジを表示しない", () => {
    renderBody({
      selectedPet: { ...basePet, birthDate: "2018-10-05", status: "死亡" },
      isPetDeceased: true,
    });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});

describe("VaccinationFormPagePanels レイアウト (EMR-415 V3)", () => {
  it("患者カードとフォームの間に上余白 16px（mt-4）がある", () => {
    renderBody({ selectedPet: { ...basePet } });

    const formGrid = screen.getByTestId("vaccination-form-grid");
    expect(formGrid).toHaveClass("mt-4");
  });
});
