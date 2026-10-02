import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { calcAgePartsAt } from "@/lib/calc-age";
import { C } from "@/lib/design-tokens";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { PetResponse } from "@/types/generated/pet-responses";

import { MedicalRecordStickyHeader } from "./MedicalRecordStickyHeader";

// EMR-164 / 改善依頼 No.28③④ の固定。カルテ画面のヘッダーで
// PatientContextHeader の実描画（年齢・性別・避妊去勢・特記バッジ）を検証するため
// PatientContextHeader は mock しない（props 伝搬だけの検証は FormPanels 側で済み）。
vi.mock("@/components/shared/UnifiedTabs", () => ({
  UnifiedTabsList: () => <div data-testid="medical-record-tabs" />,
  UnifiedTabsContent: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: false }),
}));

vi.mock("../api/vitals", () => ({
  useGetVitals: () => ({ data: [] }),
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

function renderHeader({ selectedPet = makePet() }: { selectedPet?: Pet } = {}) {
  return render(
    <MemoryRouter>
      <MedicalRecordStickyHeader
        selectedPet={selectedPet}
        cohabitingPets={[]}
        staffName="佐藤医師"
        visitType="再診"
        visitCount={3}
        canEdit={true}
        isNewRecord={false}
        tabs={[{ value: "問診", label: "問診" }]}
        recordDate="2026/10/02"
        recordStatus="draft"
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

describe("MedicalRecordStickyHeader 患者属性ヘッダー (EMR-164 No.28③)", () => {
  it("生年月日から計算した年齢・性別・避妊去勢・品種などをヘッダーに表示する", () => {
    renderHeader({
      selectedPet: makePet({
        birthDate: "2018-05-10",
        gender: "雄",
        neuteredDate: "2020-06-01",
        breed: "柴犬",
        weight: "8.4kg",
        petNumber: "P-0001",
        insuranceName: "アニコム",
        insuranceDetails: "70%補償",
      }),
    });

    // Tooltip は content を document.body の portal にも描画するため複数ヒットする。
    const { years, months } = calcAgePartsAt("2018-05-10", new Date());
    expect(
      screen.getAllByText(`2018-05-10生（${years}歳${months}ヶ月） / 犬`).length,
    ).toBeGreaterThan(0);

    expect(screen.getByText("性別")).toBeInTheDocument();
    expect(screen.getByText("雄")).toBeInTheDocument();
    expect(screen.getByText("避妊去勢")).toBeInTheDocument();
    expect(screen.getByText("去勢済")).toBeInTheDocument();
    expect(screen.getByText("品種")).toBeInTheDocument();
    expect(screen.getByText("柴犬")).toBeInTheDocument();
    expect(screen.getByText("8.4kg")).toBeInTheDocument();
    expect(screen.getByText("#P-0001")).toBeInTheDocument();
    expect(screen.getByText("来院 3 回")).toBeInTheDocument();
    expect(screen.getByText("アニコム")).toBeInTheDocument();
    expect(screen.getByText("70%補償")).toBeInTheDocument();
  });

  it("雌の避妊済表示と、避妊去勢日未記録時の「—」推測しない表示を出し分ける", () => {
    const { unmount } = renderHeader({
      selectedPet: makePet({ gender: "雌", neuteredDate: "2019-01-15" }),
    });
    expect(screen.getByText("避妊済")).toBeInTheDocument();
    unmount();

    renderHeader({ selectedPet: makePet({ gender: "雌" }) });
    const neuteredLabel = screen.getByText("避妊去勢");
    expect(neuteredLabel.parentElement).toHaveTextContent("—");
    expect(screen.queryByText("避妊済")).not.toBeInTheDocument();
  });
});

describe("MedicalRecordStickyHeader スタッフ向け特記バッジ (EMR-164 No.28④)", () => {
  it("ペット特記 高 は赤いアイコンバッジを出し、クリックで理由メモを開く", async () => {
    const user = userEvent.setup();
    renderHeader({
      selectedPet: makePet({ dangerLevel: "高", dangerReason: "保定時に噛む" }),
    });

    const trigger = screen.getByRole("button", { name: "モカの詳細を表示" });
    expect(trigger).toHaveClass(C.bgDanger10, C.danger, C.borderDanger20);

    await user.click(trigger);
    expect(await screen.findByText(/特記レベル:\s*高/)).toBeInTheDocument();
    expect(await screen.findByText("保定時に噛む")).toBeInTheDocument();
  });

  it("飼主 is_dangerous は名前横のアイコンマークで出し、特記レベル低・未設定はバッジを出さない", () => {
    renderHeader({ selectedPet: makePet({ ownerIsDangerous: true, dangerLevel: "低" }) });

    expect(screen.getByRole("img", { name: "特記" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "モカの詳細を表示" })).not.toBeInTheDocument();
  });
});
