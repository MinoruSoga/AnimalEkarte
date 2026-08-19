import { useEffect, type ReactNode } from "react";
import { act, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ExaminationForm } from "./ExaminationForm";

const mocks = vi.hoisted(() => ({
  id: undefined as string | undefined,
  canCreate: false,
  canEdit: true,
  canDelete: false,
  canUnconfirm: false,
  isPersistedConfirmed: false,
  isPatientChangeLocked: true,
  useExaminationForm: vi.fn(),
  unconfirmDialog: vi.fn(),
  patientChangeDialog: vi.fn(),
  useGetExaminations: vi.fn(),
  useGetStaffs: vi.fn(),
  historyPanel: vi.fn(),
  formFieldsMounted: vi.fn(),
  formFieldsUnmounted: vi.fn(),
  formFieldsProps: vi.fn(),
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
  usePermission: (resource: string) => ({
    canCreate: resource === "examination-unconfirm" ? false : mocks.canCreate,
    canEdit:
      resource === "examination-unconfirm" ? mocks.canUnconfirm : mocks.canEdit,
    canDelete: resource === "examination-unconfirm" ? false : mocks.canDelete,
    canView: true,
  }),
}));

vi.mock("@/hooks/use-master-items", () => ({
  useMasterItems: () => ({ data: [], isLoading: false }),
}));

vi.mock("@/features/master", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/features/master")>();
  return {
    ...actual,
    useGetStaffs: mocks.useGetStaffs,
  };
});

// ExaminationForm は未紐付け受信バナーを描画する。バナーは useQuery を使うため
// QueryClientProvider の無いフォーム単体テストでは null に差し替える。
vi.mock("@/features/lab-device", () => ({
  LabDeviceUnlinkedBanner: () => null,
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

vi.mock("../api/get-examination-print-snapshot", () => ({
  useGetExaminationPrintSnapshot: () => ({ data: undefined, isLoading: false }),
}));

vi.mock("../components/ExaminationPrintArea", () => ({
  ExaminationPrintArea: () => null,
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
  ExaminationFormFields: (props: {
    staffList: { id: string; name: string }[];
  }) => {
    mocks.formFieldsProps(props);
    useEffect(() => {
      mocks.formFieldsMounted();
      return () => mocks.formFieldsUnmounted();
    }, []);
    return <button type="submit">保存</button>;
  },
}));

vi.mock("../components/ExaminationUnconfirmDialog", () => ({
  ExaminationUnconfirmDialog: (props: unknown) => {
    mocks.unconfirmDialog(props);
    return <button type="button">確定解除</button>;
  },
}));

vi.mock("../components/ExaminationPatientChangeDialog", () => ({
  ExaminationPatientChangeDialog: (props: unknown) => {
    mocks.patientChangeDialog(props);
    return <button type="button">患者を変更</button>;
  },
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
  mocks.canUnconfirm = false;
  mocks.isPersistedConfirmed = false;
  mocks.isPatientChangeLocked = true;
  mocks.searchParams = "";
  mocks.setSearchParams.mockReset();
  mocks.historyPanel.mockReset();
  mocks.unconfirmDialog.mockReset();
  mocks.patientChangeDialog.mockReset();
  mocks.formFieldsMounted.mockReset();
  mocks.formFieldsUnmounted.mockReset();
  mocks.formFieldsProps.mockReset();
  mocks.useGetExaminations.mockReset();
  mocks.useGetExaminations.mockReturnValue({ data: [] });
  mocks.useGetStaffs.mockReset();
  mocks.useGetStaffs.mockReturnValue({ data: [], isLoading: false });
  mocks.useExaminationForm.mockReset();
  mocks.useExaminationForm.mockImplementation(() => ({
    formData: { status: "依頼中", petId: "pet-1" },
    setFormData: vi.fn(),
    petSelection: {
      selectedPets: [
        {
          id: "pet-1",
          ownerId: "owner-1",
          ownerName: "飼主",
          name: "ポチ",
          species: "犬",
        },
      ],
    },
    formAction: vi.fn(),
    formState: { success: false, timestamp: 0 },
    handleDelete: vi.fn(),
    isEdit: mocks.id !== undefined,
    isSaving: false,
    isDeleting: false,
    formItems: [],
    setInspectionValue: vi.fn(),
    addManualItem: vi.fn(),
    removeItem: vi.fn(),
    setItemName: vi.fn(),
    handleUnconfirm: vi.fn(),
    isUnconfirming: false,
    isPersistedConfirmed: mocks.isPersistedConfirmed,
    isPersistedCompletedLocked: false,
    isPersistedResultsLocked: mocks.isPersistedConfirmed,
    isPatientChangeLocked: mocks.isPatientChangeLocked,
  }));
});

describe("ExaminationForm — doctor candidate filter (BUG-005)", () => {
  it("担当医候補は同院 active doctor のみを渡し resource/nurse/inactive を除外する", () => {
    mocks.useGetStaffs.mockReturnValue({
      data: [
        {
          id: "1",
          name: "林文明",
          staffType: "doctor",
          isActive: true,
        },
        {
          id: "2",
          name: "お手入れ・オゾン療法",
          staffType: "resource",
          isActive: true,
        },
        {
          id: "3",
          name: "看護師A",
          staffType: "nurse",
          isActive: true,
        },
        {
          id: "4",
          name: "退職医",
          staffType: "doctor",
          isActive: false,
        },
        {
          id: "5",
          name: "トリマーB",
          staffType: "trimmer",
          isActive: true,
        },
      ],
      isLoading: false,
    });

    render(<ExaminationForm />);

    expect(mocks.formFieldsProps).toHaveBeenCalled();
    const lastProps = mocks.formFieldsProps.mock.calls.at(-1)?.[0] as {
      staffList: { id: string; name: string }[];
    };
    expect(lastProps.staffList).toEqual([{ id: "1", name: "林文明" }]);
    expect(lastProps.staffList.map((s) => s.name)).not.toContain(
      "お手入れ・オゾン療法",
    );
    expect(lastProps.staffList.map((s) => s.id)).not.toEqual(
      expect.arrayContaining(["2", "3", "4", "5"]),
    );
  });
});

describe("ExaminationForm — mutation permission wiring", () => {
  it("route id変更時はrecord-scopedフォームstateをremountする", () => {
    mocks.id = "examination-1";
    const view = render(<ExaminationForm />);

    expect(mocks.formFieldsMounted).toHaveBeenCalledOnce();
    expect(mocks.formFieldsUnmounted).not.toHaveBeenCalled();

    mocks.id = "examination-2";
    view.rerender(<ExaminationForm />);

    expect(mocks.formFieldsUnmounted).toHaveBeenCalledOnce();
    expect(mocks.formFieldsMounted).toHaveBeenCalledTimes(2);
  });

  it("create/edit/delete の現在値を hook へ渡す", () => {
    const view = render(<ExaminationForm />);

    expect(mocks.useExaminationForm).toHaveBeenLastCalledWith(
      undefined,
      undefined,
      {
        canCreate: false,
        canEdit: true,
        canDelete: false,
        canUnconfirm: false,
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
        canUnconfirm: false,
      },
    );
  });

  it("新規作成は作成権限があっても編集権限なしならフォームをdisabledにする", () => {
    mocks.canCreate = true;
    mocks.canEdit = false;

    render(<ExaminationForm />);

    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();
  });

  it("編集は作成権限なしでも編集権限があればフォームを有効にする", () => {
    mocks.id = "examination-1";
    mocks.canCreate = false;
    mocks.canEdit = true;

    render(<ExaminationForm />);

    expect(screen.getByRole("button", { name: "保存" })).toBeEnabled();
  });

  it("dedicated確定解除権限を通常の編集権限と独立してhookへ渡す", () => {
    mocks.id = "examination-1";
    mocks.canEdit = false;
    mocks.canUnconfirm = true;

    render(<ExaminationForm />);

    expect(mocks.useExaminationForm).toHaveBeenLastCalledWith(
      "examination-1",
      undefined,
      expect.objectContaining({ canEdit: false, canUnconfirm: true }),
    );
  });

  it("確定済みかつdedicated権限ありなら通常fieldset外に確定解除を表示する", () => {
    mocks.id = "examination-1";
    mocks.canEdit = false;
    mocks.canUnconfirm = true;
    mocks.isPersistedConfirmed = true;

    render(<ExaminationForm />);

    const button = screen.getByRole("button", { name: "確定解除" });
    expect(button.closest("fieldset")).toBeNull();
    expect(button.closest("form")).toBeNull();
  });

  it("通常編集権限だけでは確定解除を表示しない", () => {
    mocks.id = "examination-1";
    mocks.canEdit = true;
    mocks.canUnconfirm = false;
    mocks.isPersistedConfirmed = true;

    render(<ExaminationForm />);

    expect(
      screen.queryByRole("button", { name: "確定解除" }),
    ).not.toBeInTheDocument();
  });

  it("初回confirm前かつ編集権限ありのときだけ患者変更を表示する", () => {
    mocks.id = "examination-1";
    mocks.canEdit = true;
    mocks.isPatientChangeLocked = false;

    const view = render(<ExaminationForm />);
    expect(
      screen.getByRole("button", { name: "患者を変更" }),
    ).toBeInTheDocument();

    mocks.isPatientChangeLocked = true;
    view.rerender(<ExaminationForm />);
    expect(
      screen.queryByRole("button", { name: "患者を変更" }),
    ).not.toBeInTheDocument();
  });
});

describe("ExaminationForm — history pivot wiring", () => {
  it("petIdをserver-side filterへ渡し、historyView=pivotを初期表示へ反映する", () => {
    mocks.searchParams =
      "petId=pet-1&medicalRecordId=record-1&historyView=pivot";

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

    const nextParams = mocks.setSearchParams.mock
      .lastCall?.[0] as URLSearchParams;
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
