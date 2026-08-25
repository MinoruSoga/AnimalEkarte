import { type ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { MedicalRecordForm } from "./MedicalRecordForm";

const mockFormState = vi.hoisted(() => ({
  current: {
    isNewRecord: false,
    notFound: false,
    isReadLoading: false,
    isReadNotFound: false,
    isReadError: false,
    retryRead: undefined as (() => void) | undefined,
    selectedPet: null as null | {
      id: string;
      ownerId: string;
      ownerName: string;
      name: string;
      petNumber: string;
      species: string;
      status: "生存" | "死亡";
    },
    isPetLoading: false,
  },
}));

vi.mock("react-router", () => ({
  useParams: () => ({ id: "999999999" }),
  useNavigate: () => vi.fn(),
}));

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: ({ children, title }: { children: ReactNode; title?: string }) => (
    <div data-testid="page-layout" data-title={title ?? ""}>
      {children}
    </div>
  ),
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: null }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: true, canEdit: true, canDelete: true }),
}));

vi.mock("@/hooks/use-title", () => ({
  useTitle: vi.fn(),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({
    isDirty: false,
    markDirty: vi.fn(),
    markClean: vi.fn(),
  }),
}));

vi.mock("@/hooks/use-owner-line-tags", () => ({
  useGetOwnerLineTags: () => ({ data: undefined }),
}));

vi.mock("../hooks/use-medical-record-form", () => ({
  useMedicalRecordForm: () => ({
    isNewRecord: mockFormState.current.isNewRecord,
    activeTab: "問診",
    setActiveTab: vi.fn(),
    selectedPet: mockFormState.current.selectedPet,
    cohabitingPets: [],
    isPetLoading: mockFormState.current.isPetLoading,
    shouldRedirectToSelectPet: false,
    notFound: mockFormState.current.notFound,
    isReadLoading: mockFormState.current.isReadLoading,
    isReadNotFound: mockFormState.current.isReadNotFound,
    isReadError: mockFormState.current.isReadError,
    retryRead: mockFormState.current.retryRead,
    handleBack: vi.fn(),
    formAction: vi.fn(),
    formState: {},
    isSaving: false,
    isFinalized: false,
    isCreating: false,
    autoCreateFailurePhase: null,
    retryAutoCreate: vi.fn(),
    treatmentPlanItems: [],
    setTreatmentPlanItems: vi.fn(),
    chiefComplaint: "",
    setChiefComplaint: vi.fn(),
    chiefComplaintTypeId: null,
    setChiefComplaintTypeId: vi.fn(),
    treatmentPolicy: "",
    setTreatmentPolicy: vi.fn(),
    physicalExam: "",
    setPhysicalExam: vi.fn(),
    plan: "",
    setPlan: vi.fn(),
    assessment: "",
    setAssessment: vi.fn(),
    diagnosis1CategoryId: null,
    setDiagnosis1CategoryId: vi.fn(),
    diagnosis1NameId: null,
    setDiagnosis1NameId: vi.fn(),
    diagnosis2CategoryId: null,
    setDiagnosis2CategoryId: vi.fn(),
    diagnosis2NameId: null,
    setDiagnosis2NameId: vi.fn(),
    ownerDiscountRate: 0,
    visitType: "再診",
    visitCount: 0,
    handleChangeDoctor: vi.fn(),
    handleVisitTypeChange: vi.fn(),
    pendingOwnerChange: null,
    requestOwnerChange: vi.fn(),
    confirmOwnerChange: vi.fn(),
    cancelOwnerChange: vi.fn(),
    nextVisitDate: "",
    handleNextVisitDateChange: vi.fn(),
    handleNextVisitDatePatch: vi.fn(),
    isNextVisitDateValid: true,
    handleNextVisitDateValidChange: vi.fn(),
    recommendationReason: null,
    setRecommendationReason: vi.fn(),
    recordDate: "",
    handleChangeDate: vi.fn(),
    handleFinalize: vi.fn(),
    isFinalizeSaving: false,
    fieldErrors: {},
  }),
}));

vi.mock("../hooks/use-medical-record-dirty-fields", () => ({
  useMedicalRecordDirtyFields: () => ({
    handleSetAssessment: vi.fn(),
    handleSetChiefComplaint: vi.fn(),
    handleSetChiefComplaintTypeId: vi.fn(),
    handleSetDiagnosis1CategoryId: vi.fn(),
    handleSetDiagnosis1NameId: vi.fn(),
    handleSetDiagnosis2CategoryId: vi.fn(),
    handleSetDiagnosis2NameId: vi.fn(),
    handleSetPhysicalExam: vi.fn(),
    handleSetPlan: vi.fn(),
    handleSetTreatmentPolicy: vi.fn(),
  }),
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

vi.mock("../hooks/use-medical-record-post-save", () => ({
  useMedicalRecordPostSave: () => ({
    handleRegisterEstimateSave: vi.fn(),
  }),
}));

vi.mock("../api/get-medical-record", () => ({
  useGetMedicalRecord: () => ({ data: undefined }),
}));
vi.mock("../api/billing-confirmation", () => ({
  useGetBillingConfirmation: () => ({ data: { status: "confirmed" }, isLoading: false, isError: false }),
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

vi.mock("../api/delete-medical-record", () => ({
  useDeleteMedicalRecord: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock("../components/MedicalRecordAddenda", () => ({
  MedicalRecordAddenda: () => null,
}));

vi.mock("../components/MedicalRecordAutoCreateFailure", () => ({
  MedicalRecordAutoCreateFailure: () => null,
}));

vi.mock("../components/MedicalRecordFormPanels", () => ({
  MedicalRecordStickyHeader: () => <div>sticky-header</div>,
  MedicalRecordTabsArea: () => <div>tabs-area</div>,
}));

vi.mock("../components/MedicalRecordFormActions", () => ({
  MedicalRecordFinalizeDialog: () => null,
  MedicalRecordFloatingActions: () => <button type="button">保存</button>,
  MedicalRecordPrintArea: () => null,
}));

vi.mock("../components/MedicalRecordFormModals", () => ({
  MedicalRecordFormModals: () => null,
}));

vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

vi.mock("@/components/shared/UnifiedTabs", () => ({
  UnifiedTabsRoot: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

describe("MedicalRecordForm BUG-017 not-found / network gate", () => {
  beforeEach(() => {
    mockFormState.current = {
      isNewRecord: false,
      notFound: false,
      isReadLoading: false,
      isReadNotFound: false,
      isReadError: false,
      retryRead: undefined,
      selectedPet: null,
      isPetLoading: false,
    };
  });

  it("loading: blank ではなく LoadingFallback を出す", () => {
    mockFormState.current.isReadLoading = true;
    render(<MedicalRecordForm />);
    expect(screen.queryByText("カルテが見つかりません")).not.toBeInTheDocument();
    expect(screen.queryByText("sticky-header")).not.toBeInTheDocument();
    expect(document.querySelector(".animate-spin")).toBeTruthy();
  });

  it("404 相当: カルテが見つかりません を表示し、保存 UI を出さない", () => {
    mockFormState.current.notFound = true;
    mockFormState.current.isReadNotFound = true;
    render(<MedicalRecordForm />);
    expect(screen.getByText("カルテが見つかりません")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.queryByText("sticky-header")).not.toBeInTheDocument();
  });

  it("403 相当: 404 と同一の非開示メッセージ", () => {
    mockFormState.current.notFound = true;
    mockFormState.current.isReadNotFound = true;
    render(<MedicalRecordForm />);
    expect(screen.getByText("カルテが見つかりません")).toBeInTheDocument();
  });

  it("network error: 取得失敗 + 再試行、blank form ではない", () => {
    const retry = vi.fn();
    mockFormState.current.isReadError = true;
    mockFormState.current.retryRead = retry;
    render(<MedicalRecordForm />);
    expect(screen.getByText("カルテの取得に失敗しました")).toBeInTheDocument();
    const btn = screen.getByRole("button", { name: "再試行" });
    btn.click();
    expect(retry).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
  });

  it("edit で selectedPet 未解決でも return null の白紙にしない", () => {
    mockFormState.current.selectedPet = null;
    mockFormState.current.isPetLoading = false;
    mockFormState.current.isNewRecord = false;
    render(<MedicalRecordForm />);
    expect(screen.getByText("カルテが見つかりません")).toBeInTheDocument();
    expect(screen.queryByText("sticky-header")).not.toBeInTheDocument();
  });

  it("正常 edit: pet 解決済みならフォーム本体を出す", () => {
    mockFormState.current.selectedPet = {
      id: "pet-1",
      ownerId: "owner-1",
      ownerName: "飼主",
      name: "ポチ",
      petNumber: "P-1",
      species: "犬",
      status: "生存",
    };
    render(<MedicalRecordForm />);
    expect(screen.getByText("sticky-header")).toBeInTheDocument();
    expect(screen.queryByText("カルテが見つかりません")).not.toBeInTheDocument();
  });
});

describe("MedicalRecordForm BUG-002 deceased new hard stop", () => {
  const livingPet = {
    id: "pet-1",
    ownerId: "owner-1",
    ownerName: "飼主",
    name: "ポチ",
    petNumber: "P-1",
    species: "犬",
    status: "生存" as const,
  };

  const deceasedPet = {
    ...livingPet,
    id: "1000003",
    name: "クロ",
    status: "死亡" as const,
  };

  beforeEach(() => {
    mockFormState.current = {
      isNewRecord: false,
      notFound: false,
      isReadLoading: false,
      isReadNotFound: false,
      isReadError: false,
      retryRead: undefined,
      selectedPet: null,
      isPetLoading: false,
    };
  });

  it("新規 + 死亡 pet: フルフォームを出さず BE 同文言で hard stop", () => {
    mockFormState.current.isNewRecord = true;
    mockFormState.current.selectedPet = deceasedPet;
    render(<MedicalRecordForm />);
    expect(
      screen.getByText("死亡したペットは新規カルテを作成できません"),
    ).toBeInTheDocument();
    expect(screen.queryByText("sticky-header")).not.toBeInTheDocument();
    expect(screen.queryByText("tabs-area")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
  });

  it("新規 + 生存 pet: フォーム本体を出す", () => {
    mockFormState.current.isNewRecord = true;
    mockFormState.current.selectedPet = livingPet;
    render(<MedicalRecordForm />);
    expect(screen.getByText("sticky-header")).toBeInTheDocument();
    expect(screen.getByText("tabs-area")).toBeInTheDocument();
    expect(
      screen.queryByText("死亡したペットは新規カルテを作成できません"),
    ).not.toBeInTheDocument();
  });

  it("編集 + 死亡 pet: 既存カルテ閲覧のため hard stop しない（新規のみ対象）", () => {
    mockFormState.current.isNewRecord = false;
    mockFormState.current.selectedPet = deceasedPet;
    render(<MedicalRecordForm />);
    expect(screen.getByText("sticky-header")).toBeInTheDocument();
    expect(
      screen.queryByText("死亡したペットは新規カルテを作成できません"),
    ).not.toBeInTheDocument();
  });
});
