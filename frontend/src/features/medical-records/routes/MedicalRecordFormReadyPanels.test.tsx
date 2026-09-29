import type { ComponentProps, ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Pet } from "@/types";
import type { PetResponse } from "@/types/generated/pet-responses";

import { MedicalRecordFormReadyPanels } from "./MedicalRecordFormReadyPanels";

// BUG-MR-DOCTOR-HEADER-STALE: 取得済みカルテの doctor とセッションユーザーを
// 差し替え可能にし、ヘッダー表示の出所を直接検証する。
const { mockRecord, mockHandleChangeDoctor, mockCopyTreatments, mockCopyHookArgs } = vi.hoisted(
  () => ({
    mockRecord: {
      current: undefined as { doctor: string; status?: string; clinicId?: string } | undefined,
    },
    mockHandleChangeDoctor: vi.fn(),
    mockCopyTreatments: vi.fn(),
    mockCopyHookArgs: { current: [] as [string | undefined, string | undefined][] },
  }),
);

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: { displayName: "ログイン医師", clinic: undefined } }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: false, canEdit: true }),
}));

vi.mock("../api/get-medical-record", () => ({
  useGetMedicalRecord: () => ({ data: mockRecord.current }),
}));

vi.mock("../api/get-medical-records", () => ({
  useGetPetMedicalHistory: () => ({ historyItems: [] }),
}));

vi.mock("../api/clinical-plan", () => ({
  useGetClinicalPlan: () => ({ data: undefined }),
}));

vi.mock("../api/treatments", () => ({
  useGetTreatments: () => ({ data: [] }),
}));

// EMR-219: QueryClientProvider 無しで描くこのテストでは hook 実体を差し替え、
// (recordId, recordClinicId) での初期化と TabsArea への callback 配線のみ検証する。
// 明細複写の内部挙動（fetch/逐次POST/価格再解決）は use-copy-treatment-details.test.ts が担当。
vi.mock("../hooks/use-copy-treatment-details", () => ({
  useCopyTreatmentDetails: (recordId: string | undefined, recordClinicId?: string) => {
    mockCopyHookArgs.current.push([recordId, recordClinicId]);
    return { copyTreatmentsFromRecord: mockCopyTreatments, isPending: false };
  },
}));

vi.mock("../api/billing-confirmation", () => ({
  useGetBillingConfirmation: () => ({ data: undefined, isLoading: false, isError: false }),
}));

vi.mock("@/hooks/use-owner-line-tags", () => ({
  useGetOwnerLineTags: () => ({ data: undefined }),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({ isDirty: false, markDirty: vi.fn(), markClean: vi.fn() }),
}));

vi.mock("../hooks/use-medical-record-post-save", () => ({
  useMedicalRecordPostSave: () => ({ handleRegisterEstimateSave: vi.fn() }),
}));

// dirty-fields の各ハンドラはスタブ化した TabsArea にしか渡らないため空でよい。
vi.mock("../hooks/use-medical-record-dirty-fields", () => ({
  useMedicalRecordDirtyFields: () => ({}),
}));

vi.mock("../hooks/use-medical-record-form-modals", () => ({
  useMedicalRecordFormModals: () => ({
    isDeleteConfirmOpen: false,
    setIsDeleteConfirmOpen: vi.fn(),
    isFinalizeConfirmOpen: false,
    setIsFinalizeConfirmOpen: vi.fn(),
    isVitalsOpen: false,
    setIsVitalsOpen: vi.fn(),
    isPrinting: false,
    isStaffModalOpen: false,
    isOwnerSearchOpen: false,
    setIsOwnerSearchOpen: vi.fn(),
    handleStaffModalOpenChange: vi.fn(),
    handleOpenStaffModal: vi.fn(),
    handleOpenOwnerSearch: vi.fn(),
    handlePrintClick: vi.fn(),
  }),
}));

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/components/shared/UnifiedTabs", () => ({
  UnifiedTabsRoot: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  UnifiedTabsList: () => <div data-testid="medical-record-tabs" />,
  UnifiedTabsContent: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

vi.mock("../components/MedicalRecordAddenda/MedicalRecordAddenda", () => ({
  MedicalRecordAddenda: () => null,
}));

vi.mock("../components/MedicalRecordAutoCreateFailure", () => ({
  MedicalRecordAutoCreateFailure: () => null,
}));

vi.mock("../components/MedicalRecordFormActions", () => ({
  MedicalRecordFinalizeDialog: () => null,
  MedicalRecordFloatingActions: () => null,
  MedicalRecordPrintArea: () => null,
}));

// スタッフモーダルを差し替え、実配線の onSelectStaff → handleSelectStaff を直接駆動する。
vi.mock("../components/MedicalRecordFormModals", () => ({
  MedicalRecordFormModals: ({
    onSelectStaff,
  }: {
    onSelectStaff: (staffId: string, staffName: string) => void;
  }) => (
    <button
      type="button"
      data-testid="select-doctor"
      onClick={() => onSelectStaff("9", "鈴木医師")}
    >
      担当医を選択
    </button>
  ),
}));

// MedicalRecordStickyHeader は実物を使い 担当医 aria-label を検証する。
// TabsArea のみスタブ化（配下タブの深い依存を切る）。
vi.mock("../components/MedicalRecordFormPanels", async (importOriginal) => {
  const mod = await importOriginal<typeof import("../components/MedicalRecordFormPanels")>();
  return {
    ...mod,
    MedicalRecordTabsArea: (props: ComponentProps<typeof mod.MedicalRecordTabsArea>) => (
      <div data-testid="tabs-area">
        <button
          type="button"
          data-testid="copy-record-treatments"
          onClick={() => props.onCopyRecordTreatments?.("999")}
        />
      </div>
    ),
  };
});

vi.mock("../api/vitals", () => ({
  useGetVitals: () => ({ data: [] }),
}));

// contextControls（担当医ボタン等）は実配線のまま描画する。
vi.mock("@/components/shared/PatientContextHeader", () => ({
  PatientContextHeader: ({ contextControls }: { contextControls?: ReactNode }) => (
    <div data-testid="patient-context-header">{contextControls}</div>
  ),
}));

vi.mock("../components/VisitTypeSelect", () => ({
  VisitTypeSelect: () => null,
}));

vi.mock("../components/NextVisitButton", () => ({
  NextVisitButton: () => null,
}));

type ReadyPanelsProps = ComponentProps<typeof MedicalRecordFormReadyPanels>;

function makePet(): Pet {
  return {
    ...transformBackendPetToFrontend({} as PetResponse),
    id: "10",
    ownerId: "20",
    ownerName: "山田太郎",
    name: "モカ",
    species: "犬",
    status: "生存",
  };
}

function makeForm(): ReadyPanelsProps["form"] {
  return {
    formAction: vi.fn(),
    handleBack: vi.fn(),
    activeTab: "問診",
    setActiveTab: vi.fn(),
    formState: null,
    isFinalized: false,
    isNewRecord: false,
    isCreating: false,
    isSaving: false,
    isFinalizeSaving: false,
    autoCreateFailurePhase: null,
    retryAutoCreate: vi.fn(),
    cohabitingPets: [],
    recordDate: "2026-09-24",
    visitType: "再診",
    visitCount: 2,
    nextVisitDate: "",
    chiefComplaint: "",
    chiefComplaintTypeId: null,
    treatmentPolicy: "",
    physicalExam: "",
    plan: "",
    assessment: "",
    diagnosis1CategoryId: null,
    diagnosis1NameId: null,
    diagnosis2CategoryId: null,
    diagnosis2NameId: null,
    ownerDiscountRate: 0,
    recommendationReason: "",
    setRecommendationReason: vi.fn(),
    setAssessment: vi.fn(),
    setChiefComplaint: vi.fn(),
    setChiefComplaintTypeId: vi.fn(),
    setDiagnosis1CategoryId: vi.fn(),
    setDiagnosis1NameId: vi.fn(),
    setDiagnosis2CategoryId: vi.fn(),
    setDiagnosis2NameId: vi.fn(),
    setPhysicalExam: vi.fn(),
    setPlan: vi.fn(),
    setTreatmentPolicy: vi.fn(),
    handleChangeDoctor: mockHandleChangeDoctor,
    handleFinalize: vi.fn(),
    handleVisitTypeChange: vi.fn(),
    handleChangeDate: vi.fn(),
    handleNextVisitDateChange: vi.fn(),
    handleNextVisitDatePatch: vi.fn(),
    handleNextVisitDateValidChange: vi.fn(),
    requestOwnerChange: vi.fn(),
    pendingOwnerChange: null,
    cancelOwnerChange: vi.fn(),
    confirmOwnerChange: vi.fn(),
  } as unknown as ReadyPanelsProps["form"];
}

function renderPanelsWithRecordId(recordId: string | undefined) {
  return render(
    <MemoryRouter>
      <MedicalRecordFormReadyPanels
        recordId={recordId}
        selectedPet={makePet()}
        form={makeForm()}
        canEdit={true}
        canSubmit={true}
        canDelete={false}
        isDeleting={false}
        onDeleteConfirm={vi.fn()}
      />
    </MemoryRouter>,
  );
}

function renderPanels() {
  return renderPanelsWithRecordId("77");
}

describe("MedicalRecordFormReadyPanels BUG-MR-DOCTOR-HEADER-STALE", () => {
  beforeEach(() => {
    mockRecord.current = undefined;
    mockHandleChangeDoctor.mockClear();
    mockCopyTreatments.mockClear();
    mockCopyHookArgs.current = [];
  });

  it("取得済みカルテの担当医名をヘッダーに表示し、ログインユーザー名で上書きしない", () => {
    mockRecord.current = { doctor: "山田医師", status: "draft", clinicId: "1" };
    renderPanels();

    expect(screen.getByRole("button", { name: "担当医: 山田医師" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "担当医: ログイン医師" })).not.toBeInTheDocument();
  });

  it("カルテに担当医が未割当ならログインユーザー表示名へフォールバックする", () => {
    mockRecord.current = { doctor: "", status: "draft", clinicId: "1" };
    renderPanels();

    expect(screen.getByRole("button", { name: "担当医: ログイン医師" })).toBeInTheDocument();
  });

  it("担当医選択は楽観表示し、record doctor が変わらない再取得では保持し、変化すれば取得値に戻る", async () => {
    mockRecord.current = { doctor: "山田医師", status: "draft", clinicId: "1" };
    const view = renderPanels();
    const user = userEvent.setup();

    await user.click(screen.getByTestId("select-doctor"));
    expect(mockHandleChangeDoctor).toHaveBeenCalledWith("9", "鈴木医師");
    expect(screen.getByRole("button", { name: "担当医: 鈴木医師" })).toBeInTheDocument();

    // 未変化データの再取得（同じ doctor 値のまま再描画）では選択表示を維持する
    mockRecord.current = { doctor: "山田医師", status: "draft", clinicId: "1" };
    view.rerender(
      <MemoryRouter>
        <MedicalRecordFormReadyPanels
          recordId="77"
          selectedPet={makePet()}
          form={makeForm()}
          canEdit={true}
          canSubmit={true}
          canDelete={false}
          isDeleting={false}
          onDeleteConfirm={vi.fn()}
        />
      </MemoryRouter>,
    );
    expect(screen.getByRole("button", { name: "担当医: 鈴木医師" })).toBeInTheDocument();

    // 取得値が実際に変わったらオーバーライドを破棄して新しい担当医を表示する
    mockRecord.current = { doctor: "田中医師", status: "draft", clinicId: "1" };
    view.rerender(
      <MemoryRouter>
        <MedicalRecordFormReadyPanels
          recordId="77"
          selectedPet={makePet()}
          form={makeForm()}
          canEdit={true}
          canSubmit={true}
          canDelete={false}
          isDeleting={false}
          onDeleteConfirm={vi.fn()}
        />
      </MemoryRouter>,
    );
    expect(screen.getByRole("button", { name: "担当医: 田中医師" })).toBeInTheDocument();
  });
});

describe("MedicalRecordFormReadyPanels EMR-219 治療明細複写", () => {
  beforeEach(() => {
    mockRecord.current = undefined;
    mockHandleChangeDoctor.mockClear();
    mockCopyTreatments.mockClear();
    mockCopyHookArgs.current = [];
  });

  it("useCopyTreatmentDetails を (recordId, recordClinicId) で初期化し、callback を TabsArea へ配線する", async () => {
    mockRecord.current = { doctor: "山田医師", status: "draft", clinicId: "3" };
    const user = userEvent.setup();
    renderPanelsWithRecordId("77");

    expect(mockCopyHookArgs.current).toContainEqual(["77", "3"]);

    await user.click(screen.getByTestId("copy-record-treatments"));
    expect(mockCopyTreatments).toHaveBeenCalledWith("999");
  });

  it("recordId 未定でも hook は undefined で初期化される（明細複写は hook 内でスキップ）", () => {
    mockRecord.current = undefined;
    renderPanelsWithRecordId(undefined);

    expect(mockCopyHookArgs.current.length).toBeGreaterThan(0);
    expect(mockCopyHookArgs.current[0][0]).toBeUndefined();
  });
});
