import { useState, type ReactNode } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { MedicalRecordForm } from "./MedicalRecordForm";

const mockDeleteRecord = vi.hoisted(() => vi.fn());
const mockUsePermission = vi.hoisted(() => vi.fn());
const mockBoundaryState = vi.hoisted(() => ({
  selectedPetStatus: "生存" as "生存" | "死亡",
  isFinalized: false,
  capturedDeleteConfirm: undefined as (() => void) | undefined,
  setPermissions: undefined as
    | ((permissions: { canEdit: boolean; canCreate: boolean; canDelete: boolean }) => void)
    | undefined,
}));

vi.mock("react-router", () => ({
  useParams: () => ({ id: "record-1" }),
  useNavigate: () => vi.fn(),
}));

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: null }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: mockUsePermission,
}));

vi.mock("../hooks/use-medical-record-form", () => ({
  useMedicalRecordForm: () => ({
    isNewRecord: false,
    activeTab: "問診",
    setActiveTab: vi.fn(),
    selectedPet: {
      id: "pet-1",
      ownerId: "owner-1",
      ownerName: "飼主",
      name: "ペット",
      petNumber: "P-1",
      species: "犬",
      status: mockBoundaryState.selectedPetStatus,
    },
    isPetLoading: false,
    shouldRedirectToSelectPet: false,
    notFound: false,
    isReadLoading: false,
    isReadNotFound: false,
    isReadError: false,
    retryRead: undefined,
    handleBack: vi.fn(),
    formAction: vi.fn(),
    formState: {},
    isSaving: false,
    isFinalized: mockBoundaryState.isFinalized,
    isCreating: false,
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
    visitCount: 1,
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
    recordDate: "2026-08-01",
    handleChangeDate: vi.fn(),
    handleFinalize: vi.fn(),
    isFinalizeSaving: false,
  }),
}));

vi.mock("../api/get-medical-records", () => ({
  useGetPetMedicalHistory: () => ({ historyItems: [] }),
}));
vi.mock("../api/get-medical-record", () => ({
  useGetMedicalRecord: () => ({ data: { clinicId: "clinic-1" } }),
}));
vi.mock("../api/billing-confirmation", () => ({
  useGetBillingConfirmation: () => ({
    data: { status: "confirmed" },
    isLoading: false,
    isError: false,
  }),
}));
vi.mock("../api/clinical-plan", () => ({
  useGetClinicalPlan: () => ({ data: undefined }),
}));
vi.mock("../api/treatments", () => ({
  useGetTreatments: () => ({ data: [] }),
}));
vi.mock("../api/delete-medical-record", () => ({
  useDeleteMedicalRecord: () => ({ mutate: mockDeleteRecord, isPending: false }),
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
    handleSetPlan: vi.fn(),
    handleSetTreatmentPolicy: vi.fn(),
  }),
}));
vi.mock("../hooks/use-medical-record-form-modals", () => ({
  useMedicalRecordFormModals: () => ({
    isDeleteConfirmOpen: true,
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
vi.mock("@/hooks/use-owner-line-tags", () => ({
  useGetOwnerLineTags: () => ({ data: undefined }),
}));
vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({ isDirty: false, markDirty: vi.fn(), markClean: vi.fn() }),
}));
vi.mock("@/hooks/use-title", () => ({ useTitle: vi.fn() }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));
vi.mock("sonner", () => ({ toast: { success: vi.fn() } }));

vi.mock("../components/MedicalRecordFormPanels", () => ({
  MedicalRecordStickyHeader: () => null,
  MedicalRecordTabsArea: () => <input aria-label="clinical field" />,
}));
vi.mock("../components/MedicalRecordFormActions", () => ({
  MedicalRecordFloatingActions: ({
    isFinalized,
    canSubmit,
  }: {
    isFinalized: boolean;
    canSubmit: boolean;
  }) => (canSubmit && !isFinalized ? <button type="submit">保存</button> : null),
  MedicalRecordFinalizeDialog: () => null,
  MedicalRecordPrintArea: () => null,
}));
vi.mock("../components/MedicalRecordFormModals", () => ({
  MedicalRecordFormModals: ({ onConfirmDelete }: { onConfirmDelete: () => void }) => {
    mockBoundaryState.capturedDeleteConfirm ??= onConfirmDelete;
    return (
      <button type="button" onClick={onConfirmDelete}>
        confirm delete
      </button>
    );
  },
}));
vi.mock("../components/MedicalRecordAddenda/MedicalRecordAddenda", () => ({
  MedicalRecordAddenda: () => null,
}));
vi.mock("@/components/shared/NavigationBlocker", () => ({ NavigationBlocker: () => null }));
vi.mock("@/components/shared/DataStates", () => ({
  LoadingFallback: () => null,
  ErrorFallback: () => null,
}));
vi.mock("@/components/shared/UnifiedTabs", () => ({
  UnifiedTabsRoot: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

beforeEach(() => {
  vi.clearAllMocks();
  mockBoundaryState.selectedPetStatus = "生存";
  mockBoundaryState.isFinalized = false;
  mockBoundaryState.capturedDeleteConfirm = undefined;
  mockBoundaryState.setPermissions = undefined;
  mockUsePermission.mockReturnValue({ canEdit: false, canCreate: true, canDelete: false });
});

describe("MedicalRecordForm — mutation permission boundary", () => {
  function installStatefulPermissionMock() {
    mockUsePermission.mockImplementation(() => {
      const [permissions, setPermissions] = useState({
        canEdit: true,
        canCreate: true,
        canDelete: true,
      });
      mockBoundaryState.setPermissions = setPermissions;
      return permissions;
    });
  }

  it("canDelete=falseではdelete mutationを発行せず、canEdit=falseではfieldsetを無効化する", () => {
    render(<MedicalRecordForm />);

    expect(screen.getByRole("group")).toBeDisabled();
    expect(screen.getByTestId("medical-record-edit-lock")).toBeDisabled();
    expect(screen.getByRole("textbox", { name: "clinical field" })).toBeDisabled();

    fireEvent.click(screen.getByRole("button", { name: "confirm delete" }));

    expect(mockDeleteRecord).not.toHaveBeenCalled();
  });

  it("isFinalized=true では clinical 入力を disabled にし保存ボタンを出さない", () => {
    mockBoundaryState.isFinalized = true;
    mockUsePermission.mockReturnValue({ canEdit: true, canCreate: true, canDelete: true });

    render(<MedicalRecordForm />);

    expect(screen.getByTestId("medical-record-edit-lock")).toBeDisabled();
    expect(screen.getByRole("textbox", { name: "clinical field" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.getByText(/このカルテは確定済みのため編集できません/)).toBeInTheDocument();
  });

  it("delete権限を失ったcommit直後は取得済み削除callbackがmutationを発行しない", () => {
    installStatefulPermissionMock();
    render(<MedicalRecordForm />);

    act(() =>
      mockBoundaryState.setPermissions?.({
        canEdit: true,
        canCreate: true,
        canDelete: false,
      }),
    );
    act(() => mockBoundaryState.capturedDeleteConfirm?.());

    expect(mockDeleteRecord).not.toHaveBeenCalled();
  });

  it("選択ペットが死亡したcommit直後は取得済み削除callbackがmutationを発行しない", () => {
    installStatefulPermissionMock();
    render(<MedicalRecordForm />);

    mockBoundaryState.selectedPetStatus = "死亡";
    act(() =>
      mockBoundaryState.setPermissions?.({
        canEdit: true,
        canCreate: true,
        canDelete: true,
      }),
    );
    act(() => mockBoundaryState.capturedDeleteConfirm?.());

    expect(mockDeleteRecord).not.toHaveBeenCalled();
  });
});
