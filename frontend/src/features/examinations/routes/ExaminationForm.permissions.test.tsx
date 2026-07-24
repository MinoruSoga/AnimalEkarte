import type { ReactNode } from "react";
import { render } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ExaminationForm } from "./ExaminationForm";

const mocks = vi.hoisted(() => ({
  id: undefined as string | undefined,
  canCreate: false,
  canEdit: true,
  canDelete: false,
  useExaminationForm: vi.fn(),
}));

vi.mock("react-router", () => ({
  useNavigate: () => vi.fn(),
  useLocation: () => ({ state: undefined }),
  useParams: () => ({ id: mocks.id }),
  useSearchParams: () => [new URLSearchParams(), vi.fn()],
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canCreate: mocks.canCreate,
    canEdit: mocks.canEdit,
    canDelete: mocks.canDelete,
  }),
}));

vi.mock("@/hooks/use-master-items", () => ({
  useMasterItems: () => ({ data: [], isLoading: false }),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({
    isDirty: false,
    markDirty: vi.fn(),
    markClean: vi.fn(),
  }),
}));

vi.mock("../hooks/use-examination-form", () => ({
  useExaminationForm: mocks.useExaminationForm,
}));

vi.mock("../api/get-examinations", () => ({
  useGetExaminations: () => ({ data: [] }),
}));

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

vi.mock("@/components/shared/ConfirmDialog", () => ({
  ConfirmDialog: () => null,
}));

vi.mock("@/components/shared/PatientInfoCard", () => ({
  PatientInfoCard: () => null,
}));

vi.mock("../components/ExamItemsTable", () => ({
  ExamItemsTable: () => null,
}));

vi.mock("../components/ExaminationFormFields", () => ({
  ExaminationFormFields: () => null,
}));

vi.mock("../components/ExaminationHistoryPanel", () => ({
  ExaminationHistoryPanel: () => null,
}));

beforeEach(() => {
  mocks.id = undefined;
  mocks.canCreate = false;
  mocks.canEdit = true;
  mocks.canDelete = false;
  mocks.useExaminationForm.mockReset();
  mocks.useExaminationForm.mockImplementation(() => ({
    formData: { status: "依頼中", petId: "pet-1" },
    setFormData: vi.fn(),
    petSelection: {
      selectedPets: [{
        id: "pet-1",
        ownerId: "owner-1",
        ownerName: "飼主",
        name: "ポチ",
        species: "犬",
      }],
    },
    formAction: vi.fn(),
    formState: { success: false, timestamp: 0 },
    handleDelete: vi.fn(),
    isEdit: mocks.id !== undefined,
    isSaving: false,
    isDeleting: false,
    formItems: [],
    setInspectionValue: vi.fn(),
  }));
});

describe("ExaminationForm — mutation permission wiring", () => {
  it("create/edit/delete の現在値を hook へ渡す", () => {
    const view = render(<ExaminationForm />);

    expect(mocks.useExaminationForm).toHaveBeenLastCalledWith(
      undefined,
      undefined,
      {
        canCreate: false,
        canEdit: true,
        canDelete: false,
      },
    );

    mocks.id = "examination-1";
    mocks.canCreate = true;
    mocks.canEdit = false;
    mocks.canDelete = true;
    view.rerender(<ExaminationForm />);

    expect(mocks.useExaminationForm).toHaveBeenLastCalledWith(
      "examination-1",
      undefined,
      {
        canCreate: true,
        canEdit: false,
        canDelete: true,
      },
    );
  });
});
