import { useActionState, useEffect, type ReactNode } from "react";
import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ExaminationForm } from "./ExaminationForm";

const mocks = vi.hoisted(() => ({
  id: undefined as string | undefined,
  canCreate: false,
  canEdit: true,
  canDelete: false,
  canUnconfirm: false,
  isPersistedConfirmed: false,
  isPersistedCompletedLocked: false,
  status: "依頼中" as string,
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
  patientInfoCard: vi.fn(),
  isPetDeceased: false,
  searchParams: "",
  setSearchParams: vi.fn(),
}));

vi.mock("react-router", () => ({
  useNavigate: () => vi.fn(),
  useLocation: () => ({ state: undefined }),
  useParams: () => ({ id: mocks.id }),
  useSearchParams: () => [new URLSearchParams(mocks.searchParams), mocks.setSearchParams],
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: (resource: string) => ({
    canCreate: resource === "examination-unconfirm" ? false : mocks.canCreate,
    canEdit: resource === "examination-unconfirm" ? mocks.canUnconfirm : mocks.canEdit,
    canDelete: resource === "examination-unconfirm" ? false : mocks.canDelete,
    canView: true,
  }),
}));

vi.mock("@/hooks/use-master-items", () => ({
  useGetMasterItems: () => ({ data: [], isLoading: false }),
}));

vi.mock("@/hooks/use-staffs", () => ({
  useGetStaffs: mocks.useGetStaffs,
}));

// ExaminationForm は未紐付け受信バナーを描画する。バナーは useQuery を使うため
// QueryClientProvider の無いフォーム単体テストでは null に差し替える。
vi.mock("@/components/shared/LabDeviceUnlinkedBanner/LabDeviceUnlinkedBanner", () => ({
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
  PageLayout: ({
    children,
    title,
    headerAction,
  }: {
    children: ReactNode;
    title: string;
    headerAction?: ReactNode;
  }) => (
    <div>
      <h1>{title}</h1>
      {headerAction}
      {children}
    </div>
  ),
}));

vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

vi.mock("@/components/shared/ConfirmDialog", () => ({
  ConfirmDialog: (props: {
    open: boolean;
    onConfirm: () => void;
    onClose: () => void;
    confirmLabel?: string;
    cancelLabel?: string;
    title?: string;
  }) =>
    props.open ? (
      <div role="dialog" aria-label={props.title}>
        <button type="button" onClick={props.onConfirm}>
          {props.confirmLabel ?? "確認"}
        </button>
        <button type="button" onClick={props.onClose}>
          {props.cancelLabel ?? "キャンセル"}
        </button>
      </div>
    ) : null,
}));

vi.mock("@/components/shared/PatientInfoCard", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/components/shared/PatientInfoCard")>()),
  PatientInfoCard: (props: unknown) => {
    mocks.patientInfoCard(props);
    return null;
  },
}));

vi.mock("../components/ExamItemsTable", () => ({
  ExamItemsTable: () => null,
}));

vi.mock("../components/ExaminationFormFields", () => ({
  ExaminationFormFields: (props: { staffList: { id: string; name: string }[] }) => {
    mocks.formFieldsProps(props);
    useEffect(() => {
      mocks.formFieldsMounted();
      return () => mocks.formFieldsUnmounted();
    }, []);
    return <div />;
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
  mocks.isPersistedCompletedLocked = false;
  mocks.status = "依頼中";
  mocks.isPatientChangeLocked = true;
  mocks.isPetDeceased = false;
  mocks.searchParams = "";
  mocks.setSearchParams.mockReset();
  mocks.historyPanel.mockReset();
  mocks.unconfirmDialog.mockReset();
  mocks.patientChangeDialog.mockReset();
  mocks.formFieldsMounted.mockReset();
  mocks.formFieldsUnmounted.mockReset();
  mocks.formFieldsProps.mockReset();
  mocks.patientInfoCard.mockReset();
  mocks.useGetExaminations.mockReset();
  mocks.useGetExaminations.mockReturnValue({ data: [] });
  mocks.useGetStaffs.mockReset();
  mocks.useGetStaffs.mockReturnValue({ data: [], isLoading: false });
  mocks.useExaminationForm.mockReset();
  mocks.useExaminationForm.mockImplementation(hookReturn);
});

function hookReturn() {
  return {
    formData: { status: mocks.status, petId: "pet-1" },
    setFormData: vi.fn(),
    isPetDeceased: mocks.isPetDeceased,
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
    isPersistedCompletedLocked: mocks.isPersistedCompletedLocked,
    isPersistedResultsLocked: mocks.isPersistedConfirmed,
    isPatientChangeLocked: mocks.isPatientChangeLocked,
  };
}

function deferredGate() {
  let resolve!: () => void;
  let reject!: (e: unknown) => void;
  const promise = new Promise<void>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

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
    expect(lastProps.staffList.map((s) => s.name)).not.toContain("お手入れ・オゾン療法");
    expect(lastProps.staffList.map((s) => s.id)).not.toEqual(
      expect.arrayContaining(["2", "3", "4", "5"]),
    );
  });
});

describe("ExaminationForm — mutation permission wiring", () => {
  it("削除・キャンセル・保存はフォーム外のヘッダーに置き保存だけを入力フォームに接続する", () => {
    mocks.id = "examination-1";
    mocks.canDelete = true;
    render(<ExaminationForm />);

    const save = screen.getByRole("button", { name: "保存" });
    expect(save).toHaveAttribute("type", "submit");
    expect(save).toHaveAttribute("form", document.querySelector("form")?.id);
    for (const name of ["削除", "キャンセル", "保存"]) {
      expect(screen.getByRole("button", { name }).closest("form")).toBeNull();
    }
    expect(screen.getByRole("button", { name: "削除" })).toHaveAttribute("type", "button");
    expect(screen.getByRole("button", { name: "キャンセル" })).toHaveAttribute("type", "button");
  });

  it("A-S02-01: 未確定ならローカルステータスが確定でもヘッダーから保存できる", () => {
    mocks.id = "examination-1";
    mocks.status = "確定";
    render(<ExaminationForm />);

    expect(screen.getByRole("button", { name: "保存" })).toBeEnabled();
  });

  it("PR669: 削除権限は編集権限と独立し、閲覧者でも削除ボタンを実行できる", async () => {
    const user = userEvent.setup();
    const handleDelete = vi.fn();
    mocks.useExaminationForm.mockImplementation(() => ({
      ...hookReturn(),
      handleDelete,
    }));
    mocks.id = "examination-1";
    mocks.canEdit = false;
    mocks.canDelete = true;

    render(<ExaminationForm />);

    expect(screen.getByRole("heading", { name: "検査詳細" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();

    const deleteButton = screen.getByRole("button", { name: "削除" });
    expect(deleteButton).toBeEnabled();
    await user.click(deleteButton);
    // 確認ダイアログが開き、confirm で handleDelete が呼ばれる
    const dialog = screen.getByRole("dialog", { name: "検査記録を削除しますか？" });
    await user.click(within(dialog).getByRole("button", { name: "削除" }));
    expect(handleDelete).toHaveBeenCalledTimes(1);
  });

  it("PR669: 削除実行中（isDeleting）は削除ボタンを disabled にする", () => {
    mocks.useExaminationForm.mockImplementation(() => ({
      ...hookReturn(),
      isDeleting: true,
    }));
    mocks.id = "examination-1";
    mocks.canEdit = false;
    mocks.canDelete = true;

    render(<ExaminationForm />);

    expect(screen.getByRole("button", { name: "削除中..." })).toBeDisabled();
  });

  it("PR669: 編集権限のみでは削除ボタンを表示しない", () => {
    mocks.id = "examination-1";
    mocks.canEdit = true;
    mocks.canDelete = false;

    render(<ExaminationForm />);

    expect(screen.getByRole("button", { name: "保存" })).toBeEnabled();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
  });

  it("完了封印中は詳細、ステータス変更後に保存可能なら編集として表示する", () => {
    mocks.id = "examination-1";
    mocks.canDelete = true;
    mocks.isPersistedCompletedLocked = true;
    mocks.status = "完了";
    const view = render(<ExaminationForm />);

    expect(screen.getByRole("heading", { name: "検査詳細" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
    mocks.status = "確定";
    view.rerender(<ExaminationForm />);
    expect(screen.getByRole("heading", { name: "検査編集" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "保存" })).toBeEnabled();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
  });

  it("確定済みレコードは詳細扱いで編集・削除を表示しない", () => {
    mocks.id = "examination-1";
    mocks.isPersistedConfirmed = true;
    mocks.canDelete = true;

    render(<ExaminationForm />);

    expect(screen.getByRole("heading", { name: "検査詳細" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "キャンセル" })).toBeEnabled();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
  });

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

    expect(mocks.useExaminationForm).toHaveBeenLastCalledWith(undefined, undefined, {
      canCreate: false,
      canEdit: true,
      canDelete: false,
      canUnconfirm: false,
    });

    mocks.id = "examination-1";
    mocks.canCreate = true;
    mocks.canEdit = false;
    mocks.canDelete = true;
    view.rerender(<ExaminationForm />);

    expect(mocks.useExaminationForm).toHaveBeenLastCalledWith("examination-1", undefined, {
      canCreate: true,
      canEdit: false,
      canDelete: true,
      canUnconfirm: false,
    });
  });

  it("新規作成は作成権限があっても編集権限なしならフォームをdisabledにする", () => {
    mocks.canCreate = true;
    mocks.canEdit = false;

    render(<ExaminationForm />);

    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
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

    expect(screen.queryByRole("button", { name: "確定解除" })).not.toBeInTheDocument();
  });

  it("初回confirm前かつ編集権限ありのときだけ患者変更を表示する", () => {
    mocks.id = "examination-1";
    mocks.canEdit = true;
    mocks.isPatientChangeLocked = false;

    const view = render(<ExaminationForm />);
    expect(screen.getByRole("button", { name: "患者を変更" })).toBeInTheDocument();

    mocks.isPatientChangeLocked = true;
    view.rerender(<ExaminationForm />);
    expect(screen.queryByRole("button", { name: "患者を変更" })).not.toBeInTheDocument();
  });
});

describe("ExaminationForm — 死亡ペット render側二重防壁 (FE-RC-002)", () => {
  it("死亡ペットでは権限があってもfieldsetをdisabledにし、理由バナーとPatientInfoCardのdeceased statusを出す", () => {
    mocks.id = "examination-1";
    mocks.canCreate = true;
    mocks.canEdit = true;
    mocks.canDelete = true;
    mocks.isPetDeceased = true;

    render(<ExaminationForm />);

    expect(document.querySelector("fieldset")).toBeDisabled();
    // 死亡ペットは handler 側再検査に加え render 側でも削除を無効化する（二重防壁）。
    expect(screen.getByRole("button", { name: "削除" })).toBeDisabled();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.getByText("死亡したペットの検査記録は保存できません")).toBeInTheDocument();
    expect(mocks.patientInfoCard).toHaveBeenLastCalledWith(
      expect.objectContaining({ status: "deceased" }),
    );
  });

  it("生存ペットではfieldsetを有効にし、理由バナーを出さない", () => {
    mocks.canCreate = true;
    mocks.canEdit = true;
    mocks.isPetDeceased = false;

    render(<ExaminationForm />);

    expect(document.querySelector("fieldset")).not.toBeDisabled();
    expect(screen.getByRole("button", { name: "保存" })).toBeEnabled();
    expect(screen.queryByText("死亡したペットの検査記録は保存できません")).not.toBeInTheDocument();
    expect(mocks.patientInfoCard).toHaveBeenLastCalledWith(
      expect.objectContaining({ status: "alive" }),
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
    expect(nextParams.toString()).toBe("petId=pet-1&medicalRecordId=record-1&historyView=pivot");
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

describe("ExaminationForm — header pending contract (PR666)", () => {
  function usePendingHookImpl(mutate: () => void, gate: ReturnType<typeof deferredGate>) {
    return function useExaminationFormPending() {
      const [formState, formAction, isSaving] = useActionState(
        async () => {
          mutate();
          await gate.promise;
          return { success: false, timestamp: 0 };
        },
        { success: false, timestamp: 0 },
      );
      return { ...hookReturn(), formState, formAction, isSaving };
    };
  }

  it("保存中は外部ヘッダーボタンを disabled+保存中... にし二重送信しない", async () => {
    const user = userEvent.setup();
    const gate = deferredGate();
    const mutate = vi.fn();
    mocks.useExaminationForm.mockImplementation(usePendingHookImpl(mutate, gate));
    mocks.id = "exam-1";
    mocks.canEdit = true;

    render(<ExaminationForm />);
    const save = screen.getByRole("button", { name: "保存" });

    try {
      await user.click(save);
      await waitFor(() => {
        expect(save).toBeDisabled();
        expect(save).toHaveTextContent("保存中...");
      });
      expect(mutate).toHaveBeenCalledTimes(1);

      await user.click(save);
      expect(mutate).toHaveBeenCalledTimes(1);
    } finally {
      await act(async () => {
        gate.resolve();
      });
    }
    await waitFor(() => expect(save).toBeEnabled());
  });

  it("アクション失敗後もボタンは再び有効になる", async () => {
    const user = userEvent.setup();
    const gate = deferredGate();
    const mutate = vi.fn();
    mocks.useExaminationForm.mockImplementation(function useExaminationFormRejected() {
      const [formState, formAction, isSaving] = useActionState(
        async () => {
          mutate();
          try {
            await gate.promise;
            return { success: false, timestamp: 0 };
          } catch {
            return { success: false, timestamp: 0, error: "failed" };
          }
        },
        { success: false, timestamp: 0 },
      );
      return { ...hookReturn(), formState, formAction, isSaving };
    });
    mocks.id = "exam-1";
    mocks.canEdit = true;

    render(<ExaminationForm />);
    const save = screen.getByRole("button", { name: "保存" });
    await user.click(save);
    await waitFor(() => expect(save).toBeDisabled());
    await act(async () => {
      gate.reject(new Error("network"));
    });
    await waitFor(() => {
      expect(save).toBeEnabled();
      expect(save).toHaveTextContent("保存");
    });
  });
});
