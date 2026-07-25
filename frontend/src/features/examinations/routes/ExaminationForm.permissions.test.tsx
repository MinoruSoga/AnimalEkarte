import type { ReactNode } from "react";
import { act, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ExaminationForm } from "./ExaminationForm";

const mocks = vi.hoisted(() => ({
  id: undefined as string | undefined,
  canCreate: false,
  canEdit: true,
  canDelete: false,
  useExaminationForm: vi.fn(),
  useGetExaminations: vi.fn(),
  historyPanel: vi.fn(),
  searchParams: "",
  setSearchParams: vi.fn(),
}));

vi.mock("react-router", () => ({
  useNavigate: () => vi.fn(),
  useLocation: () => ({ state: undefined }),
  useParams: () => ({ id: mocks.id }),
  useSearchParams: () => [
    new URLSearchParams(mocks.searchParams),
    mocks.setSearchParams,
  ],
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
  useGetExaminations: mocks.useGetExaminations,
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
  ExaminationHistoryPanel: (props: unknown) => {
    mocks.historyPanel(props);
    return <button type="button">履歴表示切替</button>;
  },
}));

beforeEach(() => {
  mocks.id = undefined;
  mocks.canCreate = false;
  mocks.canEdit = true;
  mocks.canDelete = false;
  mocks.searchParams = "";
  mocks.setSearchParams.mockReset();
  mocks.historyPanel.mockReset();
  mocks.useGetExaminations.mockReset();
  mocks.useGetExaminations.mockReturnValue({ data: [] });
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

describe("ExaminationForm — history pivot wiring", () => {
  it("petIdをserver-side filterへ渡し、historyView=pivotを初期表示へ反映する", () => {
    mocks.searchParams = "petId=pet-1&medicalRecordId=record-1&historyView=pivot";

    render(<ExaminationForm />);

    expect(mocks.useGetExaminations).toHaveBeenLastCalledWith({
      petId: "pet-1",
      startDate: undefined,
      endDate: undefined,
    });
    expect(mocks.historyPanel).toHaveBeenLastCalledWith(
      expect.objectContaining({ historyView: "pivot" }),
    );
  });

  it("表示切替時に既存query parameterを保持する", () => {
    mocks.searchParams = "petId=pet-1&medicalRecordId=record-1";
    render(<ExaminationForm />);

    const props = mocks.historyPanel.mock.lastCall?.[0] as {
      onHistoryViewChange: (view: "cards" | "pivot") => void;
    };
    act(() => props.onHistoryViewChange("pivot"));

    const nextParams = mocks.setSearchParams.mock.lastCall?.[0] as URLSearchParams;
    expect(nextParams.toString()).toBe(
      "petId=pet-1&medicalRecordId=record-1&historyView=pivot",
    );
  });

  it("view-only時も読み取り専用の履歴切替をdisabled fieldset外に置く", () => {
    mocks.canEdit = false;
    mocks.id = "exam-1";

    render(<ExaminationForm />);

    const historyToggle = screen.getByRole("button", { name: "履歴表示切替" });
    expect(historyToggle.closest("fieldset")).toBeNull();
    expect(historyToggle.closest("form")).toBeNull();
    expect(historyToggle).not.toBeDisabled();
  });
});
