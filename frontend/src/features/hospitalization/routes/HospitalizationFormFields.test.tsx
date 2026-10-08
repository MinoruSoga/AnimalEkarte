import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { HospitalizationFormData } from "../types";
import { HospitalizationFormFields } from "./HospitalizationFormFields";

// 本テストは PatientInfoCard 経路の誕生月バッジのみ検証する。
// 配下のフォーム部品・履歴パネルはスタブ化する。
vi.mock("../components/HospitalizationBasicInfo", () => ({
  HospitalizationBasicInfo: () => null,
}));
vi.mock("../components/HospitalizationNoteCard", () => ({
  HospitalizationNoteCard: () => null,
}));
vi.mock("../components/HospitalizationTreatmentTable", () => ({
  HospitalizationTreatmentTable: () => null,
}));
vi.mock("../components/HospitalizationCostSummary", () => ({
  HospitalizationCostSummary: () => null,
}));
vi.mock("@/components/shared/PastRecordHistoryPanel", () => ({
  PastRecordHistoryPanel: () => null,
}));
vi.mock("@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png", () => ({ default: "" }));

const baseFormData: HospitalizationFormData = {
  hospitalizationType: "入院",
  ownerName: "山田 太郎",
  species: "犬",
  petName: "ポチ",
  petNumber: "0001",
  petInsurance: "",
  petDetails: "",
  visit: "",
  nextVisit: "",
  weight: "5.0",
  displayDate: "",
  endDate: "",
  memo: "",
  ownerRequest: "",
  staffNotes: "",
  doctorId: "",
  doctorName: "",
  cageId: "",
  isInsurance: false,
  insuranceCompanyName: "",
  insuranceNumber: "",
};

const basePet = {
  id: "pet-1",
  ownerId: "owner-1",
  ownerName: "山田 太郎",
  name: "ポチ",
  status: "生存",
} satisfies { ownerName: string; name: string };

function renderFields(selectedPet: Parameters<typeof HospitalizationFormFields>[0]["selectedPet"]) {
  render(
    <MemoryRouter>
      <HospitalizationFormFields
        selectedPet={selectedPet}
        formData={baseFormData}
        fieldErrors={undefined}
        cageItems={[]}
        isEdit={false}
        canDelete={false}
        hasChildTreatmentPlans={false}
        treatmentPlans={[]}
        totals={{ subtotalBeforeDiscount: 0, consumptionTax: 0, total: 0 }}
        historyItems={[]}
        isHistoryLoading={false}
        canSubmit={true}
        onFormChange={vi.fn()}
        onOpenStaffModal={vi.fn()}
        onAddTreatmentPlan={vi.fn()}
        onUpdateTreatmentPlan={vi.fn()}
        onRemoveTreatmentPlan={vi.fn()}
      />
    </MemoryRouter>,
  );
}

describe("HospitalizationFormFields 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月の生存ペットは PatientInfoCard 上に Cake バッジを表示する", () => {
    renderFields({ ...basePet, birthDate: "2018-10-05" });

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05" as const],
    ["birthDate 未登録", undefined],
  ])("%s なら Cake バッジを表示しない", (_label, birthDate) => {
    renderFields({ ...basePet, birthDate });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（status=死亡）には誕生月でも Cake バッジを表示しない", () => {
    renderFields({ ...basePet, birthDate: "2018-10-05", status: "死亡" });

    expect(screen.getByText("【死亡】")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});
