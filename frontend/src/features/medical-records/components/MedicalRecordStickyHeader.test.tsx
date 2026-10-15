import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { calcAgePartsAt } from "@/lib/calc-age";
import { C } from "@/lib/design-tokens";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { PetResponse } from "@/types/generated/pet-responses";

import { MedicalRecordStickyHeader } from "./MedicalRecordStickyHeader";
import { NextVisitButton } from "./NextVisitButton";

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
  NextVisitButton: vi.fn(() => <div data-testid="next-visit-button" />),
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
  recordStatus = "draft",
  isNewRecord = false,
  medicalRecordId,
  recordDate = "2026/10/02",
  onDateChange,
}: {
  selectedPet?: Pet;
  recordStatus?: string;
  isNewRecord?: boolean;
  medicalRecordId?: string;
  recordDate?: string;
  onDateChange?: (date: string) => void;
} = {}) {
  return render(
    <MemoryRouter>
      <MedicalRecordStickyHeader
        selectedPet={selectedPet}
        cohabitingPets={[]}
        staffName="佐藤医師"
        visitType="再診"
        visitCount={3}
        canEdit={true}
        isNewRecord={isNewRecord}
        tabs={[{ value: "問診", label: "問診" }]}
        recordDate={recordDate}
        recordStatus={recordStatus}
        medicalRecordId={medicalRecordId}
        onVisitTypeChange={vi.fn()}
        onStaffClick={vi.fn()}
        onOwnerClick={vi.fn()}
        onDateChange={onDateChange}
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

describe("MedicalRecordStickyHeader 次回予定コントロール (EMR-308)", () => {
  beforeEach(() => {
    vi.mocked(NextVisitButton).mockClear();
  });

  it("既存カルテでは NextVisitButton に petId / medicalRecordId / isPetDeceased を渡す", () => {
    renderHeader({ medicalRecordId: "99" });

    expect(screen.getByTestId("next-visit-button")).toBeInTheDocument();
    expect(vi.mocked(NextVisitButton).mock.calls[0]?.[0]).toEqual(
      expect.objectContaining({
        petId: "10",
        medicalRecordId: "99",
        isPetDeceased: false,
      }),
    );
  });

  it("確定済みカルテでも次回予定コントロールは無効化・非表示にしない（ペット属性の tentative は別権限）", () => {
    renderHeader({ recordStatus: "finalized", medicalRecordId: "99" });

    expect(screen.getByTestId("next-visit-button")).toBeInTheDocument();
  });

  it("死亡ペットでは isPetDeceased=true を渡す", () => {
    renderHeader({ selectedPet: makePet({ status: "死亡" }), medicalRecordId: "99" });

    expect(vi.mocked(NextVisitButton).mock.calls[0]?.[0]).toEqual(
      expect.objectContaining({ isPetDeceased: true }),
    );
  });

  it("新規作成時はコントロールを出さない（recordId 未確定のため）", () => {
    renderHeader({ isNewRecord: true });

    expect(screen.queryByTestId("next-visit-button")).not.toBeInTheDocument();
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

describe("MedicalRecordStickyHeader 担当医コントロール (EMR-380)", () => {
  it("編集可能時の担当医ボタンは横スクロール領域の外に固定され、ヘッダ右端で見切れない", () => {
    renderHeader();

    const doctorButton = screen.getByRole("button", { name: "担当医: 佐藤医師" });
    expect(doctorButton).toBeInTheDocument();
    // コントロール列（overflow-x-auto）内にあると狭幅時にスクロールで隠れてしまう。
    // 担当医は必ず視認できる固定領域に置く。
    expect(doctorButton.closest(".overflow-x-auto")).toBeNull();
  });

  it("確定済み（read-only）でも担当医表示は横スクロール領域の外に残る", () => {
    renderHeader({ recordStatus: "finalized" });

    const doctorLabel = screen.getByText("担当医");
    expect(doctorLabel.closest(".overflow-x-auto")).toBeNull();
    expect(screen.getByText("佐藤医師")).toBeInTheDocument();
  });
});

describe("MedicalRecordStickyHeader 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月の生存ペットは患者名横に Cake バッジを表示する", () => {
    renderHeader({ selectedPet: makePet({ birthDate: "2018-10-05" }) });

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05" as const],
    ["birthDate 未登録", undefined],
  ])("%s のペットには Cake バッジを表示しない", (_label, birthDate) => {
    renderHeader({ selectedPet: makePet({ birthDate }) });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（status=死亡）には誕生月でも Cake バッジを表示しない", () => {
    renderHeader({ selectedPet: makePet({ birthDate: "2018-10-05", status: "死亡" }) });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});

describe("MedicalRecordStickyHeader 診療日コールバック", () => {
  it("フォーカスして変更せず blur しても onDateChange を呼ばない（重複 PATCH 抑止）", async () => {
    const user = userEvent.setup();
    const onDateChange = vi.fn();
    renderHeader({ recordDate: "2026/10/13", medicalRecordId: "record1", onDateChange });

    const input = screen.getByRole("textbox", { name: "診療日" });
    await user.click(input);
    await user.tab();
    expect(onDateChange).not.toHaveBeenCalled();
  });

  it("変更せず Enter→blur しても onDateChange を呼ばない", async () => {
    const user = userEvent.setup();
    const onDateChange = vi.fn();
    renderHeader({ recordDate: "2026/10/13", medicalRecordId: "record1", onDateChange });

    const input = screen.getByRole("textbox", { name: "診療日" });
    await user.click(input);
    await user.keyboard("{Enter}");
    await user.tab();
    expect(onDateChange).not.toHaveBeenCalled();
  });

  it("カレンダーで現在と同じ日を選んでも onDateChange を呼ばない", async () => {
    const user = userEvent.setup();
    const onDateChange = vi.fn();
    renderHeader({ recordDate: "2026/10/13", medicalRecordId: "record1", onDateChange });

    await user.click(screen.getByRole("button", { name: "診療日を選択" }));
    await user.click(screen.getByRole("button", { name: /2026年10月13日/ }));
    expect(onDateChange).not.toHaveBeenCalled();
  });

  it("実際に別の日へ変更すると onDateChange を ISO 値で1回だけ呼ぶ", async () => {
    const user = userEvent.setup();
    const onDateChange = vi.fn();
    renderHeader({ recordDate: "2026/10/13", medicalRecordId: "record1", onDateChange });

    const input = screen.getByRole<HTMLInputElement>("textbox", { name: "診療日" });
    await user.click(input);
    await user.clear(input);
    await user.type(input, "2026/10/14");
    await user.tab();
    expect(onDateChange).toHaveBeenCalledTimes(1);
    expect(onDateChange).toHaveBeenCalledWith("2026-10-14");
  });

  it("Enterで確定してからblurしても onDateChange は同じ値で1回だけ呼ばれる", async () => {
    const user = userEvent.setup();
    const onDateChange = vi.fn();
    renderHeader({ recordDate: "2026/10/13", medicalRecordId: "record1", onDateChange });

    const input = screen.getByRole<HTMLInputElement>("textbox", { name: "診療日" });
    await user.click(input);
    await user.clear(input);
    await user.type(input, "2026/10/14");
    await user.keyboard("{Enter}");
    await user.tab();
    expect(onDateChange).toHaveBeenCalledTimes(1);
    expect(onDateChange).toHaveBeenCalledWith("2026-10-14");
  });

  it("診療日は内蔵クリアボタンを出さない（空値は API が受け付けない）", () => {
    renderHeader({ recordDate: "2026/10/13", medicalRecordId: "record1", onDateChange: vi.fn() });

    expect(screen.queryByRole("button", { name: "日付をクリア" })).not.toBeInTheDocument();
  });
});
