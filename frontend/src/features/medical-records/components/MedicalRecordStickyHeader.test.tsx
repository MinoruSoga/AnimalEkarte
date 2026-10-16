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
import { useGetVitals } from "../api/vitals";
import type { Vital } from "../types";

// EMR-164 / 改善依頼 No.28③④ の固定。カルテ画面のヘッダーで
// PatientInfoCard の実描画（年齢・性別・避妊去勢・特記バッジ）を検証するため
// 共通カードは mock しない（props 伝搬だけの検証は FormPanels 側で済み）。
vi.mock("@/components/shared/UnifiedTabs", () => ({
  UnifiedTabsList: () => <div data-testid="medical-record-tabs" />,
  UnifiedTabsContent: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: false }),
}));

// PR679-REVIEW-R4: 最新バイタル要約の配置テストで値を差し替えるため、
// 型付きの可変状態と vi.fn のフックにする（既定は従来どおり空配列）。
const vitalsMock = vi.hoisted(() => ({
  data: [] as Vital[],
}));

vi.mock("../api/vitals", () => ({
  useGetVitals: vi.fn(() => ({ data: vitalsMock.data })),
}));

vi.mock("./VisitTypeSelect", () => ({
  VisitTypeSelect: () => null,
}));

vi.mock("./NextVisitButton", () => ({
  NextVisitButton: vi.fn(() => <div data-testid="next-visit-button" />),
}));

// テストごとにフックの戻り値を新しい配列へ戻し、呼び出し履歴も初期化する。
beforeEach(() => {
  vitalsMock.data = [];
  vi.mocked(useGetVitals).mockClear();
});

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
  recordClinicId,
  recordDate = "2026/10/02",
  onDateChange,
}: {
  selectedPet?: Pet;
  recordStatus?: string;
  isNewRecord?: boolean;
  medicalRecordId?: string;
  recordClinicId?: string;
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
        recordClinicId={recordClinicId}
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
    expect(screen.getAllByText(`2018-05-10生（${years}歳${months}ヶ月）`).length).toBeGreaterThan(
      0,
    );

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

// ─────────────────────────────────────────────────────────────
// PR679 F3: 1歳未満の年齢表記。「0歳5ヶ月」ではなく「5ヶ月」を出す旧仕様を復元。
// 固定時計で実ヘッダ描画を検証する（buildPetInfoText は非公開のため経由検証）。
// ─────────────────────────────────────────────────────────────

describe("MedicalRecordStickyHeader 幼齢ペットの年齢表記 (PR679 F3)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-16T12:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("1歳未満は「0歳Nヶ月」ではなく「Nヶ月」だけを表示する", () => {
    // 2026-05-10 → 2026-10-16 = 0歳5ヶ月
    renderHeader({ selectedPet: makePet({ birthDate: "2026-05-10" }) });

    // Tooltip は content を portal にも描画するため複数ヒットする
    expect(screen.getAllByText("2026-05-10生（5ヶ月）").length).toBeGreaterThan(0);
    expect(screen.queryByText(/0歳/)).not.toBeInTheDocument();
  });

  it("1歳以上は従来どおり「N歳Mヶ月」を表示する", () => {
    // 2024-05-10 → 2026-10-16 = 2歳5ヶ月
    renderHeader({ selectedPet: makePet({ birthDate: "2024-05-10" }) });

    expect(screen.getAllByText("2024-05-10生（2歳5ヶ月）").length).toBeGreaterThan(0);
  });
});

// ─────────────────────────────────────────────────────────────
// PR679 F4: hasPetAttributes は !== undefined ではなく実値（非空・trim）で判定する。
// backend breed: "" が transform を素通りして届くケースを再現する。
// ─────────────────────────────────────────────────────────────

describe("MedicalRecordStickyHeader 属性ガード (PR679 F4)", () => {
  it.each(["", "   "])(
    "breed が空/空白のみ（%s）で他属性も未記録なら属性3項目を出さない",
    (breed) => {
      // transformBackendPetToFrontend は breed を無加工で通すため空文字がそのまま届く
      renderHeader({ selectedPet: makePet({ breed }) });

      expect(screen.queryByText("性別")).not.toBeInTheDocument();
      expect(screen.queryByText("避妊去勢")).not.toBeInTheDocument();
      expect(screen.queryByText("品種")).not.toBeInTheDocument();
    },
  );

  it("breed だけ記録されていれば属性行を出し、未記録項目は推測せず「—」を表示する", () => {
    renderHeader({ selectedPet: makePet({ breed: "柴犬" }) });

    expect(screen.getByText("品種")).toBeInTheDocument();
    expect(screen.getByText("柴犬")).toBeInTheDocument();
    // 未記録属性は値を推測しない（既存契約の保持）
    expect(screen.getByText("性別").parentElement).toHaveTextContent("—");
    expect(screen.getByText("避妊去勢").parentElement).toHaveTextContent("—");
  });
});

// ─────────────────────────────────────────────────────────────
// PR679-REVIEW-R4: 最新バイタル要約は PatientInfoCard の患者属性行に置く。
// 旧 PatientContextHeader と同じく患者名の直下の属性行に表示し、操作欄の下に
// 独立した帯を出さない（medical-record-form.md の既存仕様）。
// ─────────────────────────────────────────────────────────────

// Tooltip は本文を document.body の portal にも複製する。portal を role で除外し、
// 既定の script / style 除外は維持する。
const IGNORE_TOOLTIP_PORTAL = 'script, style, [role="tooltip"]';

function makeVital(overrides: Partial<Vital> = {}): Vital {
  return {
    id: "v-1",
    medical_record_id: "99",
    recorded_at: "2026-10-02T11:00:00+09:00",
    weight_unit: "Kg",
    version: 1,
    created_at: "2026-10-02T11:00:00+09:00",
    updated_at: "2026-10-02T11:00:00+09:00",
    ...overrides,
  };
}

// DOM の前後関係を CSS クラスに依存せず検証する。
function precedesInDocument(before: Element, after: Element): boolean {
  return Boolean(before.compareDocumentPosition(after) & Node.DOCUMENT_POSITION_FOLLOWING);
}

describe("MedicalRecordStickyHeader 最新バイタル要約の配置 (PR679-REVIEW-R4)", () => {
  it("最新バイタルの全項目を1回だけ、ペット番号・体重と同じ属性行に表示する", () => {
    vitalsMock.data = [
      makeVital({
        id: "v-latest",
        temperature: 36.7,
        heart_rate: 120,
        respiration_rate: 30,
        weight: 12.3,
        weight_unit: "Kg",
      }),
    ];
    renderHeader({
      medicalRecordId: "99",
      selectedPet: makePet({ petNumber: "P-0001", weight: "8.4kg" }),
    });

    const vitals = screen.getByLabelText("今回のバイタル");
    expect(screen.getAllByText("T 36.7")).toHaveLength(1);
    expect(screen.getAllByText("HR 120")).toHaveLength(1);
    expect(screen.getAllByText("RR 30")).toHaveLength(1);
    // 測定体重は単位まで含めて1回だけ（ペット属性の体重とは別の行内項目）
    expect(screen.getAllByText("測定体重 12.3Kg")).toHaveLength(1);

    const vitalsRow = vitals.closest("div");
    expect(vitalsRow).not.toBeNull();
    expect(vitalsRow).toBe(screen.getByText("#P-0001").closest("div"));
    expect(vitalsRow).toBe(screen.getByText("8.4kg").closest("div"));
  });

  it("ペット名の行と属性行は同じ親に並び、要約は操作欄より前にある", () => {
    vitalsMock.data = [makeVital({ temperature: 36.7 })];
    renderHeader({ medicalRecordId: "99", selectedPet: makePet({ petNumber: "P-0001" }) });

    const nameLeaf = screen.getByText("モカ", { ignore: IGNORE_TOOLTIP_PORTAL });
    const nameRow = nameLeaf.closest("div");
    const attributesRow = screen.getByText("#P-0001").closest("div");
    expect(nameRow).not.toBeNull();
    expect(attributesRow).not.toBeNull();
    expect(nameRow?.parentElement).toBe(attributesRow?.parentElement);

    // 要約は属性行の中にあり、操作欄（担当医など）より前に出る
    const vitals = screen.getByLabelText("今回のバイタル");
    expect(attributesRow).toContainElement(vitals);
    const controlsAnchor = screen.getByRole("button", { name: "担当医: 佐藤医師" });
    expect(precedesInDocument(vitals, controlsAnchor)).toBe(true);
  });

  it("古いバイタルの値・計測時刻は表示しない（最新行のみ）", () => {
    vitalsMock.data = [
      makeVital({
        id: "v-old",
        temperature: 39.9,
        heart_rate: 140,
        respiration_rate: 28,
        weight: 4.5,
      }),
      makeVital({ id: "v-new", recorded_at: "2026-10-03T09:00:00+09:00", temperature: 36.7 }),
    ];
    renderHeader({ medicalRecordId: "99" });

    expect(screen.getByText("T 36.7")).toBeInTheDocument();
    expect(screen.queryByText("T 39.9")).not.toBeInTheDocument();
    expect(screen.queryByText("HR 140")).not.toBeInTheDocument();
    expect(screen.queryByText("RR 28")).not.toBeInTheDocument();
    expect(screen.queryByText("測定体重 4.5Kg")).not.toBeInTheDocument();
    expect(screen.queryByText(/2026-10-0[23]T/)).not.toBeInTheDocument();
  });

  it("記録IDなし・バイタルなし・最新が全項目nullなら要約を出さない", () => {
    vitalsMock.data = [makeVital({ temperature: 36.7 })];
    const withoutRecord = renderHeader();
    expect(screen.queryByLabelText("今回のバイタル")).not.toBeInTheDocument();
    withoutRecord.unmount();

    vitalsMock.data = [];
    const withoutData = renderHeader({ medicalRecordId: "99" });
    expect(screen.queryByLabelText("今回のバイタル")).not.toBeInTheDocument();
    withoutData.unmount();

    vitalsMock.data = [
      makeVital({
        id: "v-old-full",
        recorded_at: "2026-10-01T09:00:00+09:00",
        temperature: 38.5,
        heart_rate: 120,
        respiration_rate: 24,
        weight: 4.2,
      }),
      makeVital({
        notes: "問診のみ",
        temperature: null,
        heart_rate: null,
        respiration_rate: null,
        weight: null,
      }),
    ];
    renderHeader({ medicalRecordId: "99" });
    expect(screen.queryByLabelText("今回のバイタル")).not.toBeInTheDocument();
  });

  it("0 の測定値は欠測扱いにせず表示する", () => {
    vitalsMock.data = [makeVital({ temperature: 0, weight: 0, weight_unit: "Kg" })];
    renderHeader({ medicalRecordId: "99" });

    expect(screen.getByText("T 0")).toBeInTheDocument();
    expect(screen.getByText("測定体重 0Kg")).toBeInTheDocument();
  });

  it("useGetVitals には記録IDとその記録の clinicId を渡す", () => {
    renderHeader({ medicalRecordId: "99", recordClinicId: "clinic-7" });

    expect(vi.mocked(useGetVitals).mock.calls[0]).toEqual(["99", "clinic-7"]);
  });
});
