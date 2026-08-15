import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { PetResponse } from "@/types/generated/pet-responses";

import { selectCohabitingPets } from "../hooks/use-medical-record-form";
import { MedicalRecordStickyHeader } from "./MedicalRecordFormPanels";

vi.mock("@/components/shared/PatientContextHeader", () => ({
  PatientContextHeader: () => <div data-testid="patient-context-header" />,
}));

vi.mock("@/components/shared/UnifiedTabs", () => ({
  UnifiedTabsList: () => <div data-testid="medical-record-tabs" />,
  UnifiedTabsContent: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: false }),
}));

vi.mock("./VisitTypeSelect", () => ({
  VisitTypeSelect: () => null,
}));

vi.mock("./NextVisitButton", () => ({
  NextVisitButton: () => null,
}));

function makePet(overrides: Partial<Pet> = {}): Pet {
  return {
    ...transformBackendPetToFrontend({} as PetResponse),
    id: "10",
    ownerId: "20",
    ownerName: "山田太郎",
    name: "モカ",
    species: "犬",
    status: "生存",
    ...overrides,
  };
}

function renderHeader({
  selectedPet = makePet(),
  cohabitingPets = [],
  isNewRecord = false,
}: {
  selectedPet?: Pet;
  cohabitingPets?: Pet[];
  isNewRecord?: boolean;
} = {}) {
  return render(
    <MemoryRouter>
      <MedicalRecordStickyHeader
        selectedPet={selectedPet}
        cohabitingPets={cohabitingPets}
        staffName="佐藤医師"
        visitType="再診"
        visitCount={3}
        canEdit={true}
        isNewRecord={isNewRecord}
        tabs={[{ value: "問診", label: "問診" }]}
        nextVisitDate=""
        onVisitTypeChange={vi.fn()}
        onStaffClick={vi.fn()}
        onOwnerClick={vi.fn()}
        onNextVisitDatePatch={vi.fn()}
        onNextVisitDateValidChange={vi.fn()}
      />
    </MemoryRouter>,
  );
}

describe("MedicalRecordStickyHeader cohabiting pets", () => {
  it("同居ペット2匹を名前（種別）でヘッダ直下・タブ直上に表示する", () => {
    renderHeader({
      cohabitingPets: [
        makePet({ id: "11", name: "ソラ", species: "猫" }),
        makePet({ id: "12", name: "ハル", species: "犬" }),
      ],
    });

    const row = screen.getByRole("region", { name: "同居ペット" });
    const patientHeader = screen.getByTestId("patient-context-header");
    const tabs = screen.getByTestId("medical-record-tabs");

    expect(screen.getByText("ソラ（猫）")).toBeInTheDocument();
    expect(screen.getByText("ハル（犬）")).toBeInTheDocument();
    expect(row.previousElementSibling).toBe(patientHeader);
    expect(row.nextElementSibling).toContainElement(tabs);
  });

  it("自分自身・死亡ペット・別飼主のペットを除外する", () => {
    const selectedPet = makePet();
    const cohabitingPets = selectCohabitingPets(
      [
        selectedPet,
        makePet({ id: "11", name: "ソラ" }),
        makePet({ id: "12", name: "ユキ", status: "死亡" }),
        makePet({ id: "13", ownerId: "99", name: "他院の子" }),
      ],
      selectedPet,
    );

    renderHeader({ selectedPet, cohabitingPets });

    expect(screen.getByRole("link", { name: "ソラ（犬）" })).toBeInTheDocument();
    expect(screen.queryByText("モカ（犬）")).not.toBeInTheDocument();
    expect(screen.queryByText("ユキ（犬）")).not.toBeInTheDocument();
    expect(screen.queryByText("他院の子（犬）")).not.toBeInTheDocument();
  });

  it("同居ペットが0匹なら行ごと表示しない", () => {
    renderHeader();

    expect(screen.queryByRole("region", { name: "同居ペット" })).not.toBeInTheDocument();
  });

  it("新規カルテ作成モードでは行ごと表示しない", () => {
    renderHeader({
      isNewRecord: true,
      cohabitingPets: [makePet({ id: "11", name: "ソラ" })],
    });

    expect(screen.queryByRole("region", { name: "同居ペット" })).not.toBeInTheDocument();
  });

  it("native link の href に対象ペットの pet_id を含める", () => {
    renderHeader({
      cohabitingPets: [makePet({ id: "11", name: "ソラ", species: "猫" })],
    });

    expect(screen.getByRole("link", { name: "ソラ（猫）" })).toHaveAttribute(
      "href",
      "/medical-records?pet_id=11",
    );
    expect(screen.queryByRole("button", { name: "ソラ（猫）" })).not.toBeInTheDocument();
  });

  it("種別が空なら空括弧を付けずペット名だけを表示する", () => {
    renderHeader({
      cohabitingPets: [makePet({ id: "11", name: "ノア", species: "" })],
    });

    expect(screen.getByRole("link", { name: "ノア" })).toBeInTheDocument();
    expect(screen.queryByText("ノア（）")).not.toBeInTheDocument();
  });
});
