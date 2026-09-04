import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { TrimmingForm } from "./TrimmingForm";
import type { TrimmingFormData } from "@/types/trimming";

const mocks = vi.hoisted(() => ({
  id: undefined as string | undefined,
  canCreate: true,
  canEdit: true,
  canDelete: true,
  petStatus: "生存" as string,
  useTrimmingForm: vi.fn(),
}));

const emptyFormData: TrimmingFormData = {
  reservationTypeId: "",
  startTime: "",
  endTime: "",
  styleRequest: "",
  styleImage: null,
  bw: "",
  bwUnit: "Kg",
  bt: "",
  usedShampoo: "",
  usedRibbon: "",
  remarks: "",
  completedImage: null,
  courseId: "",
  optionIds: [],
  staffId: "",
  staffName: "",
  initialStatus: "in_consultation",
  nextScheduleType: "4weeks",
  nextDate: "",
};

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useNavigate: () => vi.fn(),
    useLocation: () => ({ state: undefined }),
    useParams: () => ({ id: mocks.id }),
    useSearchParams: () => [new URLSearchParams({ petId: "10" }), vi.fn()],
  };
});

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: mocks.canCreate,
    canEdit: mocks.canEdit,
    canDelete: mocks.canDelete,
  }),
}));

vi.mock("../hooks/use-trimming-form", () => ({
  useTrimmingForm: mocks.useTrimmingForm,
}));

vi.mock("../hooks/use-trimming-form-chrome", () => ({
  useTrimmingFormChrome: () => ({
    courses: [],
    options: [],
    isDirty: false,
    courseModalOpen: false,
    setCourseModalOpen: vi.fn(),
    staffModalOpen: false,
    setStaffModalOpen: vi.fn(),
    deleteConfirmOpen: false,
    setDeleteConfirmOpen: vi.fn(),
    history: {
      sortedHistory: [],
      isHistoryLoading: false,
      historySearchTerm: "",
      historySortOrder: "desc" as const,
      historyDateRange: { from: "", to: "" },
      setHistorySearchTerm: vi.fn(),
      setHistorySortOrder: vi.fn(),
      handleHistoryClear: vi.fn(),
      setHistoryDateRangeFrom: vi.fn(),
      setHistoryDateRangeTo: vi.fn(),
      handleHistoryClick: vi.fn(),
    },
    handleFormChange: vi.fn(),
    handleDeleteClick: vi.fn(),
    handleBack: vi.fn(),
    handleHistoryClick: vi.fn(),
    activeStaffItems: [],
  }),
}));

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: ({ children, headerAction }: { children: ReactNode; headerAction?: ReactNode }) => (
    <div>
      {headerAction}
      {children}
    </div>
  ),
}));

vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

vi.mock("@/components/shared/PatientInfoCard", () => ({
  PatientInfoCard: () => null,
  formatPatientPetDetails: () => "不明 / 不明 / 不明",
}));

function renderForm() {
  return render(
    <MemoryRouter>
      <TrimmingForm />
    </MemoryRouter>,
  );
}

vi.mock("./TrimmingLazyModals", () => ({
  ConfirmDialog: () => null,
  MasterSelectModal: () => null,
}));

function hookReturn() {
  return {
    mode: mocks.id ? ("edit" as const) : ("new" as const),
    formData: emptyFormData,
    setFormData: vi.fn(),
    styleImagePreview: null,
    completedImagePreview: null,
    petSelection: {
      selectedPets: [
        {
          id: "10",
          ownerId: "20",
          name: "ポチ",
          ownerName: "山田",
          petNumber: "0001",
          status: mocks.petStatus,
        },
      ],
    },
    handleStyleImageChange: vi.fn(),
    handleCompletedImageChange: vi.fn(),
    removeStyleImage: vi.fn(),
    removeCompletedImage: vi.fn(),
    formAction: vi.fn(),
    formState: { success: false, timestamp: 0 },
    handleDelete: vi.fn(),
    isSaving: false,
    isDeleting: false,
    fieldErrors: {},
    isLoading: false,
    isEditPetReady: true,
    notFound: false,
    hasExistingAppointment: false,
  };
}

beforeEach(() => {
  mocks.id = undefined;
  mocks.canCreate = true;
  mocks.canEdit = true;
  mocks.canDelete = true;
  mocks.petStatus = "生存";
  mocks.useTrimmingForm.mockReset();
  mocks.useTrimmingForm.mockImplementation(hookReturn);
});

describe("TrimmingForm — mutation permission wiring (FE-RC-101)", () => {
  it("create/edit/delete の現在値を hook へ渡す", () => {
    const view = renderForm();

    expect(mocks.useTrimmingForm).toHaveBeenLastCalledWith(undefined, {
      canCreate: true,
      canEdit: true,
      canDelete: true,
    });

    mocks.id = "trim-1";
    mocks.canCreate = false;
    mocks.canEdit = false;
    mocks.canDelete = true;
    view.unmount();
    renderForm();

    expect(mocks.useTrimmingForm).toHaveBeenLastCalledWith("trim-1", {
      canCreate: false,
      canEdit: false,
      canDelete: true,
    });
  });
});

describe("TrimmingForm 死亡ペット render側二重防壁 (FE-RC-102)", () => {
  it("死亡ペットでは SubmitButton を非表示にし、理由を表示する", () => {
    mocks.petStatus = "死亡";

    renderForm();

    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.getByText("死亡したペットのトリミング記録は保存できません")).toBeInTheDocument();
    expect(screen.getByRole("status", { name: "死亡ペットのため保存不可" })).toBeInTheDocument();
  });

  it("生存ペットでは SubmitButton を表示し、死亡理由は表示しない", () => {
    renderForm();

    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
    expect(
      screen.queryByText("死亡したペットのトリミング記録は保存できません"),
    ).not.toBeInTheDocument();
  });

  it("編集かつ canDelete でも死亡ペットでは削除ボタンを出さない", () => {
    mocks.id = "trim-1";
    mocks.canDelete = true;
    mocks.petStatus = "死亡";

    renderForm();

    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
  });

  it("編集かつ生存ペットで canDelete なら削除ボタンを出す", () => {
    mocks.id = "trim-1";
    mocks.canDelete = true;
    mocks.petStatus = "生存";

    renderForm();

    expect(screen.getByRole("button", { name: "削除" })).toBeInTheDocument();
  });
});
