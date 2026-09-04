import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { VaccinationForm } from "./VaccinationForm";

const mocks = vi.hoisted(() => ({
  id: undefined as string | undefined,
  canCreate: false,
  canEdit: true,
  canDelete: false,
  selectedPetStatus: undefined as string | undefined,
  useVaccinationForm: vi.fn(),
}));

vi.mock("react-router", () => ({
  useNavigate: () => vi.fn(),
  useParams: () => ({ id: mocks.id }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canCreate: mocks.canCreate,
    canEdit: mocks.canEdit,
    canDelete: mocks.canDelete,
  }),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({
    isDirty: false,
    markDirty: vi.fn(),
    markClean: vi.fn(),
  }),
}));

vi.mock("../hooks/use-vaccination-form", () => ({
  useVaccinationForm: mocks.useVaccinationForm,
}));

vi.mock("../api/get-vaccinations", () => ({
  useGetVaccinations: () => ({ data: [] }),
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

vi.mock("@/components/shared/ConfirmDialog/ConfirmDialog", () => ({
  ConfirmDialog: () => null,
}));

vi.mock("@/components/shared/PatientInfoCard", () => ({
  PatientInfoCard: () => null,
  formatPatientPetDetails: () => "不明 / 不明 / 不明",
}));

vi.mock("../components/VaccinationFormPanels", () => ({
  VaccinationFieldsPanel: () => null,
  VaccinationHistoryPanel: () => null,
}));

beforeEach(() => {
  mocks.id = undefined;
  mocks.canCreate = false;
  mocks.canEdit = true;
  mocks.canDelete = false;
  mocks.selectedPetStatus = undefined;
  mocks.useVaccinationForm.mockReset();
  mocks.useVaccinationForm.mockImplementation(() => ({
    isEdit: mocks.id !== undefined,
    isReadLoading: false,
    isEditPetReady: true,
    petSelection: {
      selectedPets: [
        {
          id: "pet-1",
          ownerId: "owner-1",
          name: "ポチ",
          status: mocks.selectedPetStatus,
        },
      ],
    },
    form: {
      doctorName: "",
      date: "",
      setDate: vi.fn(),
      vaccineId: "",
      setVaccineId: vi.fn(),
      vaccineOptions: [],
      nextScheduleType: "1year",
      setNextScheduleType: vi.fn(),
      nextDate: "",
      setNextDate: vi.fn(),
      supplemental: "",
      setSupplemental: vi.fn(),
      lot1: "",
      setLot1: vi.fn(),
      lot2: "",
      setLot2: vi.fn(),
      lot3: "",
      setLot3: vi.fn(),
      lot4: "",
      setLot4: vi.fn(),
      remarks: "",
      setRemarks: vi.fn(),
    },
    formAction: vi.fn(),
    formState: { success: false, timestamp: 0 },
    isSaving: false,
    fieldErrors: {},
    handleDelete: vi.fn(),
    isDeleting: false,
    historyFilter: {
      filterStartDate: "",
      setFilterStartDate: vi.fn(),
      filterEndDate: "",
      setFilterEndDate: vi.fn(),
      historySearchTerm: "",
      setHistorySearchTerm: vi.fn(),
      sortOrder: "desc",
      setSortOrder: vi.fn(),
      handleClearHistoryFilter: vi.fn(),
    },
  }));
});

describe("VaccinationForm — mutation permission wiring", () => {
  it("create/edit/delete の現在値を hook へ渡す", () => {
    const view = render(<VaccinationForm />);

    expect(mocks.useVaccinationForm).toHaveBeenLastCalledWith(undefined, {
      canCreate: false,
      canEdit: true,
      canDelete: false,
    });

    mocks.id = "vaccination-1";
    mocks.canCreate = true;
    mocks.canEdit = false;
    mocks.canDelete = true;
    view.unmount();
    render(<VaccinationForm />);

    expect(mocks.useVaccinationForm).toHaveBeenLastCalledWith("vaccination-1", {
      canCreate: true,
      canEdit: false,
      canDelete: true,
    });
  });
});

describe("VaccinationForm — 死亡ペット render側二重防壁 (FE-RC-103/108)", () => {
  it("死亡ペットでは権限があっても保存・削除を出さず、理由バナーを出す", () => {
    mocks.id = "vaccination-1";
    mocks.canCreate = true;
    mocks.canEdit = true;
    mocks.canDelete = true;
    mocks.selectedPetStatus = "死亡";

    render(<VaccinationForm />);

    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
    expect(screen.getByText("死亡したペットの予防接種記録は保存できません")).toBeInTheDocument();
    expect(screen.getByRole("status", { name: "死亡ペットのため保存不可" })).toBeInTheDocument();
  });

  it("生存ペットでは作成権限があれば保存を表示し、死亡理由は出さない", () => {
    mocks.canCreate = true;
    mocks.selectedPetStatus = "生存";

    render(<VaccinationForm />);

    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
    expect(
      screen.queryByText("死亡したペットの予防接種記録は保存できません"),
    ).not.toBeInTheDocument();
  });

  it("編集かつ削除権限ありの生存ペットでは削除を表示する", () => {
    mocks.id = "vaccination-1";
    mocks.canEdit = true;
    mocks.canDelete = true;
    mocks.selectedPetStatus = "生存";

    render(<VaccinationForm />);

    expect(screen.getByRole("button", { name: "削除" })).toBeInTheDocument();
  });
});
