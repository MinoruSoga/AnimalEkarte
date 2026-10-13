import type { ComponentProps } from "react";
import { MemoryRouter } from "react-router";
import { act, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { UnifiedTabsRoot } from "@/components/shared/UnifiedTabs";
import type { Pet } from "@/types";
import { MedicalRecordClinicalTabs } from "./MedicalRecordClinicalTabs";
import type { InterviewHistoryItem } from "../types";

// EMR-356 repair1: 体重パネルの親合成 contract を実コンポーネントで検証する
// bounded 合成 test。重い隣接タブ/セレクトのみ stub にし、
// 問診タブ配下（MedicalRecordInterview → InterviewHistory → InterviewWeightPanel）
// は実物を通す。

const { createVitalMock } = vi.hoisted(() => ({ createVitalMock: vi.fn() }));

vi.mock("../api/vitals", () => ({
  useGetVitals: () => ({ data: [], isLoading: false, isError: false }),
  useCreateVital: () => ({ mutateAsync: createVitalMock, isPending: false }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: true, canEdit: true, canDelete: true }),
}));

vi.mock("./MedicalRecordDiagnosisPlan", () => ({
  MedicalRecordDiagnosisPlan: () => null,
}));
vi.mock("./MedicalRecordTreatment", () => ({ MedicalRecordTreatment: () => null }));
vi.mock("./RecommendationReasonSelect", () => ({ RecommendationReasonSelect: () => null }));
vi.mock("../api/get-chief-complaint-types", () => ({
  useGetChiefComplaintTypes: () => ({ data: [] }),
}));

const ALIVE_PET: Pet = {
  id: "pet-1",
  clinicId: "7",
  ownerId: "owner-1",
  ownerNumber: 1,
  ownerName: "飼主",
  ownerNameKana: undefined,
  address: undefined,
  phone: "000-0000-0000",
  petNumber: "P-1",
  name: "ポチ",
  petNameKana: undefined,
  species: "犬",
  animalSpeciesId: undefined,
  breed: "柴犬",
  color: "茶",
  bloodType: undefined,
  microchipNumber: undefined,
  gender: "雄",
  status: "生存",
  birthDate: undefined,
  neuteredDate: undefined,
  neuteredStatus: undefined,
  weight: undefined,
  food: "",
  environment: "",
  acquisitionType: undefined,
  dangerLevel: undefined,
  dangerReason: undefined,
  lastVisit: undefined,
  insuranceId: undefined,
  insuranceName: undefined,
  insuranceDetails: undefined,
  remarks: "",
  nameOrigin: undefined,
  meetingStory: undefined,
  deceasedAt: undefined,
  deceasedReason: undefined,
};

const historyItemBase = {
  date: "2026-03-01",
  type: "診察",
  title: "過去カルテ",
  content: "本文",
  author: "田中",
};

type TabsProps = ComponentProps<typeof MedicalRecordClinicalTabs>;

const baseProps: TabsProps = {
  activeTab: "問診",
  mountedTabs: new Set(["問診"]),
  isNewRecord: false,
  recordId: "10",
  selectedPet: ALIVE_PET,
  chiefComplaint: "",
  chiefComplaintTypeId: null,
  treatmentPolicy: "",
  historyItems: [],
  physicalExam: "",
  plan: "",
  assessment: "",
  diagnosis1CategoryId: null,
  diagnosis1NameId: null,
  diagnosis2CategoryId: null,
  diagnosis2NameId: null,
  ownerDiscountRate: 0,
  hasLineIntegration: false,
  recommendationReason: null,
  lstepStatus: undefined,
  recordStatus: "診察中",
  diagnosis1NameIdError: undefined,
  isLocked: false,
  isFinalized: false,
  recordClinicId: "7",
  recordPetId: "pet-1",
  // EMR356-1: 既存 fixture の 2026-03-01 測定が引き続き「過去」になるよう翌日を境界にする
  recordBoundary: new Date("2026-03-02T00:00:00+09:00"),
  onChiefComplaintChange: () => {},
  onChiefComplaintTypeIdChange: () => {},
  onTreatmentPolicyChange: () => {},
  onPhysicalExamChange: () => {},
  onPlanChange: () => {},
  onAssessmentChange: () => {},
  onDiagnosis1CategoryIdChange: () => {},
  onDiagnosis1NameIdChange: () => {},
  onDiagnosis2CategoryIdChange: () => {},
  onDiagnosis2NameIdChange: () => {},
  onRecommendationReasonChange: () => {},
  onRegisterEstimateSave: () => {},
};

function renderTabs(overrides: Partial<TabsProps> = {}) {
  const props = { ...baseProps, ...overrides };
  return render(
    <MemoryRouter>
      <UnifiedTabsRoot value="問診" onValueChange={() => {}}>
        <MedicalRecordClinicalTabs {...props} />
      </UnifiedTabsRoot>
    </MemoryRouter>,
  );
}

beforeEach(() => {
  createVitalMock.mockReset();
  createVitalMock.mockResolvedValue({});
});

describe("MedicalRecordClinicalTabs — 体重パネルへの投影集約（C2, EMR-356 repair1）", () => {
  it("全履歴行が weightMeasurements 未提供（旧API投影なし）でも「体重の記録はありません」と断定しない", () => {
    const items: InterviewHistoryItem[] = [
      { ...historyItemBase, id: "1" },
      { ...historyItemBase, id: "2" },
    ];

    renderTabs({ historyItems: items });

    expect(screen.queryByText("体重の記録はありません")).toBeNull();
  });

  it("医療行がすべて明示 [] のとき「体重の記録はありません」を表示する", () => {
    const items: InterviewHistoryItem[] = [
      { ...historyItemBase, id: "1", weightMeasurements: [] },
      { ...historyItemBase, id: "2", weightMeasurements: [] },
    ];

    renderTabs({ historyItems: items });

    expect(screen.getByText("体重の記録はありません")).toBeInTheDocument();
  });

  it("行ローカル測定はパネルの「過去」一覧へ届く", () => {
    const items: InterviewHistoryItem[] = [
      {
        ...historyItemBase,
        id: "1",
        weightMeasurements: [
          { id: "11", value: 4.2, unit: "Kg", measuredAt: "2026-03-01T09:30:00+09:00" },
        ],
      },
    ];

    renderTabs({ historyItems: items });

    expect(screen.getByText("過去")).toBeInTheDocument();
    expect(screen.getAllByText(/4\.2/)).toHaveLength(2);
  });

  it("既知の空と未提供の医療行の混在を測定なしと断定しない", () => {
    renderTabs({
      historyItems: [
        { ...historyItemBase, id: "1" },
        { ...historyItemBase, id: "2", weightMeasurements: [] },
      ],
    });
    expect(screen.queryByText("体重の記録はありません")).toBeNull();
    expect(screen.getAllByText(/体重.*確認でき/).length).toBeGreaterThan(0);
  });

  it("既知の測定は表示を保ち、未提供の医療行があることも知らせる", () => {
    renderTabs({
      historyItems: [
        {
          ...historyItemBase,
          id: "1",
          weightMeasurements: [
            { id: "11", value: 4.2, unit: "Kg", measuredAt: "2026-03-01T09:30:00+09:00" },
          ],
        },
        { ...historyItemBase, id: "2" },
      ],
    });
    expect(screen.getAllByText(/4\.2/)).toHaveLength(2);
    expect(screen.getAllByText(/体重.*確認でき/).length).toBeGreaterThan(0);
  });

  it("表示中カルテより後の測定は「過去」に出ない（EMR356-1: 古いカルテの後日体重）", () => {
    const items: InterviewHistoryItem[] = [
      {
        ...historyItemBase,
        id: "1",
        weightMeasurements: [
          { id: "11", value: 4.2, unit: "Kg", measuredAt: "2026-03-05T09:30:00+09:00" },
        ],
      },
    ];

    renderTabs({
      historyItems: items,
      recordBoundary: new Date("2026-03-01T00:00:00+09:00"),
    });

    // 履歴行には残るがパネル「過去」には届かない → 4.2 は1箇所のみ
    expect(screen.getAllByText(/4\.2/)).toHaveLength(1);
  });

  it("選択カルテ当日の測定を「過去」に含めない", () => {
    const items: InterviewHistoryItem[] = [
      {
        ...historyItemBase,
        id: "1",
        weightMeasurements: [
          { id: "11", value: 4.2, unit: "Kg", measuredAt: "2026-03-01T09:30:00+09:00" },
        ],
      },
    ];

    renderTabs({
      historyItems: items,
      recordBoundary: new Date("2026-03-01T00:00:00+09:00"),
    });

    // 行には残るが、訪問時刻不明の同日測定を「過去」に配置しない。
    expect(screen.getAllByText(/4\.2/)).toHaveLength(1);
  });

  it("境界未確定（recordBoundary なし）では測定を「過去」と断定しない", () => {
    const items: InterviewHistoryItem[] = [
      {
        ...historyItemBase,
        id: "1",
        weightMeasurements: [
          { id: "11", value: 4.2, unit: "Kg", measuredAt: "2026-03-01T09:30:00+09:00" },
        ],
      },
    ];

    renderTabs({ historyItems: items, recordBoundary: undefined });

    expect(screen.getAllByText(/4\.2/)).toHaveLength(1);
    expect(screen.queryByText("体重の記録はありません")).toBeNull();
    expect(screen.getAllByText(/体重.*確認でき/).length).toBeGreaterThan(0);
  });

  it("履歴未完（historyIncomplete）では「体重の記録はありません」と断定しない（EMR356-2）", () => {
    const items: InterviewHistoryItem[] = [{ ...historyItemBase, id: "1", weightMeasurements: [] }];

    renderTabs({ historyItems: items, historyIncomplete: true });

    expect(screen.queryByText("体重の記録はありません")).toBeNull();
    expect(screen.getAllByText(/体重.*確認でき/).length).toBeGreaterThan(0);
  });

  it("履歴未完でも既知の過去測定は表示を保ち、未完であることも知らせる", () => {
    const items: InterviewHistoryItem[] = [
      {
        ...historyItemBase,
        id: "1",
        weightMeasurements: [
          { id: "11", value: 4.2, unit: "Kg", measuredAt: "2026-03-01T09:30:00+09:00" },
        ],
      },
    ];

    renderTabs({ historyItems: items, historyIncomplete: true });

    expect(screen.getAllByText(/4\.2/)).toHaveLength(2);
    expect(screen.getAllByText(/体重.*確認でき/).length).toBeGreaterThan(0);
  });
});

describe("MedicalRecordClinicalTabs — 親フラグの最新値が書き込みゲートに入る（R2）", () => {
  it("isLocked の間は体重記録 control が inert で、捕捉済み click/Enter も mutation しない", async () => {
    renderTabs({ isLocked: true });

    const input = screen.getByLabelText("体重");
    const saveButton = screen.getByRole("button", { name: "体重を記録" });
    expect(input).toBeDisabled();
    expect(saveButton).toBeDisabled();

    // fieldset の disabled は DOM 上の操作を止めるが、React ハンドラは
    // fireEvent で到達し得る —— 最新ゲートの再検査が mutation 境界。
    await act(async () => {
      fireEvent.change(input, { target: { value: "4.5" } });
      fireEvent.keyDown(input, { key: "Enter" });
      fireEvent.click(saveButton);
    });

    expect(createVitalMock).not.toHaveBeenCalled();
  });

  it("isFinalized の間も記録 control が inert で mutation しない", async () => {
    renderTabs({ isFinalized: true });

    const input = screen.getByLabelText("体重");
    const saveButton = screen.getByRole("button", { name: "体重を記録" });
    expect(input).toBeDisabled();
    expect(saveButton).toBeDisabled();

    await act(async () => {
      fireEvent.change(input, { target: { value: "4.5" } });
      fireEvent.keyDown(input, { key: "Enter" });
      fireEvent.click(saveButton);
    });

    expect(createVitalMock).not.toHaveBeenCalled();
  });

  it("選択中カルテの petId が selectedPet と不一致なら記録 control を出さない", () => {
    renderTabs({ recordPetId: "pet-999" });

    expect(screen.queryByRole("button", { name: "体重を記録" })).toBeNull();
    expect(screen.queryByLabelText("体重")).toBeNull();
  });

  it("record 切替の遷移途中（recordPetId 未到達）では書き込まない", async () => {
    const { rerender } = renderTabs();
    const user = userEvent.setup();
    await user.type(screen.getByLabelText("体重"), "4.5");

    // 新しいカルテ選択へ進行中: recordPetId がまだ旧値/未到達
    const nextProps = { ...baseProps, recordId: "20", recordPetId: undefined };
    rerender(
      <MemoryRouter>
        <UnifiedTabsRoot value="問診" onValueChange={() => {}}>
          <MedicalRecordClinicalTabs {...nextProps} />
        </UnifiedTabsRoot>
      </MemoryRouter>,
    );

    expect(screen.queryByLabelText("体重")).toBeNull();
    fireEvent.keyDown(document.activeElement ?? document.body, { key: "Enter" });
    expect(createVitalMock).not.toHaveBeenCalled();
  });
});
