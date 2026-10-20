import type { ComponentProps } from "react";
import { MemoryRouter } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { server } from "@/testing/mocks/node";

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

// PR659-RR-1: 親の historyIncomplete を MedicalRecordInterview 経由で
// InterviewHistory まで伝え、部分履歴の空を「該当なし」と断定させない。
describe("MedicalRecordClinicalTabs — 部分履歴の問診履歴への伝播（PR659-RR-1）", () => {
  it("historyIncomplete では InterviewHistory が確定の空を出さず部分履歴を示す", () => {
    renderTabs({ historyItems: [], historyIncomplete: true });

    expect(screen.queryByText("該当する履歴はありません")).not.toBeInTheDocument();
    expect(screen.getByText("一部の履歴のみを表示しています")).toBeInTheDocument();
  });

  it("履歴未完でも読み込み済み行は残り、部分履歴の注意を添える", () => {
    renderTabs({
      historyItems: [{ ...historyItemBase, id: "1" }],
      historyIncomplete: true,
    });

    expect(screen.getByRole("link", { name: /過去カルテ/ })).toBeInTheDocument();
    expect(screen.getByText("一部の履歴のみを表示しています")).toBeInTheDocument();
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

// ─────────────────────────────────────────────────────────────
// EMR-477: 飼主事前問診（読み取り専用）の親合成 contract
// ─────────────────────────────────────────────────────────────

const PLAN_TAB = "診察/治療プラン";
const INTAKE_ROUTE = "/api/v1/medical-records/:recordId/revisit-intake";

function intakeBody(petLabel: string) {
  const ref = {
    target_id: `t-${petLabel}`,
    appointment_id: "301",
    pet_id: petLabel,
    target_generation: "1",
  };
  const answer = {
    answer_id: `a-${petLabel}`,
    revision: "1",
    form_definition_version: 1,
    received_at: "2026-10-18T01:30:00Z",
    provenance: "owner_submission",
    answers: {
      visit_purpose: `合成の来院目的-${petLabel}`,
      onset: null,
      energy: null,
      appetite: null,
      current_symptoms: null,
      improvement: null,
      examination_preference: null,
      main_concern: null,
    },
  };
  return {
    binding: "bound",
    target: {
      ref,
      head_revision: "1",
      answer_state: "answered",
      latest_answer: answer,
      content_available: true,
    },
    history: [],
    next_cursor: null,
  };
}

function makeClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } });
}

function treeFor(client: QueryClient, tab: string, overrides: Partial<TabsProps> = {}) {
  const props = {
    ...baseProps,
    mountedTabs: new Set(["問診", PLAN_TAB]),
    ...overrides,
  };
  return (
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <UnifiedTabsRoot value={tab} onValueChange={() => {}}>
          <MedicalRecordClinicalTabs {...props} />
        </UnifiedTabsRoot>
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe("MedicalRecordClinicalTabs — 飼主事前問診 reader（EMR-477）", () => {
  afterEach(() => {
    server.resetHandlers();
  });

  it("新規カルテでは表示せず取得もしない", async () => {
    let called = 0;
    server.use(
      http.get(INTAKE_ROUTE, () => {
        called += 1;
        return HttpResponse.json(intakeBody("pet-1"));
      }),
    );
    render(
      treeFor(makeClient(), PLAN_TAB, {
        isNewRecord: true,
        recordId: undefined,
        recordPetId: undefined,
      }),
    );

    expect(screen.queryByTestId("owner-intake-card")).toBeNull();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(called).toBe(0);
  });

  it("カルテの petId が選択患者と不一致（切替途中）なら取得も表示もしない", async () => {
    let called = 0;
    server.use(
      http.get(INTAKE_ROUTE, () => {
        called += 1;
        return HttpResponse.json(intakeBody("pet-1"));
      }),
    );
    render(treeFor(makeClient(), PLAN_TAB, { recordPetId: "pet-999" }));

    expect(screen.queryByTestId("owner-intake-card")).toBeNull();
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(called).toBe(0);
  });

  it("患者切替で前患者の本文を残さない（遷移中は旧本文が消え、新患者の本文だけが出る）", async () => {
    let releaseSecond: (() => void) | undefined;
    const gate = new Promise<void>((resolve) => {
      releaseSecond = resolve;
    });
    server.use(
      http.get(INTAKE_ROUTE, async ({ params }) => {
        if (params.recordId === "20") {
          await gate;
          return HttpResponse.json(intakeBody("pet-2"));
        }
        return HttpResponse.json(intakeBody("pet-1"));
      }),
    );
    const client = makeClient();
    const { rerender } = render(treeFor(client, PLAN_TAB));

    expect(await screen.findByText("合成の来院目的-pet-1")).toBeInTheDocument();

    rerender(
      treeFor(client, PLAN_TAB, {
        recordId: "20",
        recordPetId: "pet-2",
        selectedPet: { ...ALIVE_PET, id: "pet-2" },
      }),
    );

    // 新患者の応答待ち: 旧患者の本文は残らない。
    expect(screen.queryByText("合成の来院目的-pet-1")).toBeNull();
    releaseSecond?.();
    expect(await screen.findByText("合成の来院目的-pet-2")).toBeInTheDocument();
    expect(screen.queryByText("合成の来院目的-pet-1")).toBeNull();
  });

  it("問診履歴のコピー操作を起こしても原文の表示・DTO は不変で、原文 route へ書き込まない", async () => {
    const writes: string[] = [];
    let gets = 0;
    server.use(
      http.get(INTAKE_ROUTE, () => {
        gets += 1;
        return HttpResponse.json(intakeBody("pet-1"));
      }),
      http.post(INTAKE_ROUTE, () => {
        writes.push("POST");
        return HttpResponse.json({}, { status: 500 });
      }),
      http.put(INTAKE_ROUTE, () => {
        writes.push("PUT");
        return HttpResponse.json({}, { status: 500 });
      }),
      http.patch(INTAKE_ROUTE, () => {
        writes.push("PATCH");
        return HttpResponse.json({}, { status: 500 });
      }),
      http.delete(INTAKE_ROUTE, () => {
        writes.push("DELETE");
        return HttpResponse.json({}, { status: 500 });
      }),
    );
    const onCopyRecordTreatments = vi.fn();
    const onChiefComplaintChange = vi.fn();
    const onTreatmentPolicyChange = vi.fn();
    const copyProps: Partial<TabsProps> = {
      historyItems: [
        {
          ...historyItemBase,
          id: "9",
          copySource: {
            recordId: "9",
            chiefComplaint: "前回の主訴詳細",
            treatmentPolicy: "前回の治療方針",
          },
        },
      ],
      onCopyRecordTreatments,
      onChiefComplaintChange,
      onTreatmentPolicyChange,
    };
    const client = makeClient();
    const { rerender } = render(treeFor(client, PLAN_TAB, copyProps));
    expect(await screen.findByText("合成の来院目的-pet-1")).toBeInTheDocument();
    const before = screen.getByTestId("owner-intake-card").textContent;

    // 問診タブでコピーを実行
    rerender(treeFor(client, "問診", copyProps));
    const user = userEvent.setup();
    await user.click(await screen.findByRole("button", { name: "コピー" }));
    await waitFor(() => expect(onCopyRecordTreatments).toHaveBeenCalledWith("9"));
    expect(onChiefComplaintChange).toHaveBeenCalledWith("前回の主訴詳細");

    // 診察/治療プランへ戻る: 原文は同一、複写由来の書き込みは発生しない
    rerender(treeFor(client, PLAN_TAB, copyProps));
    expect(await screen.findByText("合成の来院目的-pet-1")).toBeInTheDocument();
    expect(screen.getByTestId("owner-intake-card").textContent).toBe(before);
    expect(writes).toEqual([]);
    expect(gets).toBeGreaterThanOrEqual(1);
  });

  it("isLocked でも再試行ボタンは操作可能（disabled fieldset の外）", async () => {
    server.use(
      http.get(INTAKE_ROUTE, () =>
        HttpResponse.json({ error: "x", code: "INTERNAL", details: null }, { status: 500 }),
      ),
    );
    render(treeFor(makeClient(), PLAN_TAB, { isLocked: true }));

    const retry = await screen.findByRole("button", { name: "再試行" });
    expect(retry).toBeEnabled();
  });
});
