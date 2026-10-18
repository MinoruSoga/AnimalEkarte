import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Pet } from "@/types";
import { ExaminationFormHeader } from "./ExaminationFormHeader";

// 本テストは PatientInfoCard 経路の誕生月バッジのみ検証する。
// react-query 依存のバナー・ダイアログ・印刷領域はスタブ化する。
vi.mock("@/components/shared/LabDeviceUnlinkedBanner/LabDeviceUnlinkedBanner", () => ({
  LabDeviceUnlinkedBanner: () => null,
}));
vi.mock("./ExaminationPatientChangeDialog", () => ({
  ExaminationPatientChangeDialog: () => null,
}));
vi.mock("./ExaminationUnconfirmDialog", () => ({
  ExaminationUnconfirmDialog: () => null,
}));
vi.mock("./ExaminationPrintArea", () => ({
  ExaminationPrintArea: () => null,
}));
vi.mock("@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png", () => ({ default: "" }));

const basePet = {
  id: "pet-1",
  ownerId: "owner-1",
  ownerName: "山田 太郎",
  name: "ポチ",
  species: "犬",
  petNumber: "0001",
  status: "生存",
} as Pet;

function renderHeader({
  selectedPet = basePet,
  isPetDeceased = false,
}: {
  selectedPet?: Pet;
  isPetDeceased?: boolean;
} = {}) {
  render(
    <MemoryRouter>
      <ExaminationFormHeader
        selectedPet={selectedPet}
        selectedDoctorName="佐藤医師"
        isPetDeceased={isPetDeceased}
        isEdit={false}
        isPatientChangeLocked={false}
        onPatientSelect={vi.fn()}
        isPersistedConfirmed={false}
        canUnconfirm={false}
        examinationId={undefined}
        onUnconfirm={vi.fn().mockResolvedValue(true)}
        printModel={null}
      />
    </MemoryRouter>,
  );
}

describe("ExaminationFormHeader 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月の生存ペットは PatientInfoCard 上に Cake バッジを表示する", () => {
    renderHeader({ selectedPet: { ...basePet, birthDate: "2018-10-05" } });

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05"],
    ["birthDate 未登録", undefined],
  ])("%s なら Cake バッジを表示しない", (_label, birthDate) => {
    renderHeader({ selectedPet: { ...basePet, birthDate } });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（isPetDeceased=true）には誕生月でも Cake バッジを表示しない", () => {
    renderHeader({
      selectedPet: { ...basePet, birthDate: "2018-10-05", status: "死亡" },
      isPetDeceased: true,
    });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});

describe("ExaminationFormHeader 患者属性 (EMR-415 E3)", () => {
  it("名前に種を含めず、petDetails に年齢・性別・去勢避妊を出し、id 代用しない", () => {
    renderHeader({
      selectedPet: {
        ...basePet,
        birthDate: "2018-04-01",
        gender: "メス",
        neuteredDate: "2020-01-01",
        weight: "5.2",
      } as Pet,
    });

    expect(screen.getAllByText("ポチ").length).toBeGreaterThan(0);
    expect(screen.queryByText(/ポチ\(/)).not.toBeInTheDocument();
    expect(screen.getByText(/犬 \/ .+/)).toBeInTheDocument();
    expect(screen.queryByText("pet-1")).not.toBeInTheDocument();
  });
});
