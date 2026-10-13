import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ExaminationForm } from "./ExaminationForm";

const mocks = vi.hoisted(() => ({
  id: "examination-1" as string | undefined,
  navigate: vi.fn(),
  locationState: undefined as { from?: string } | undefined,
  isReadLoading: false,
  isReadNotFound: false,
  isReadError: false,
  retryRead: undefined as (() => void) | undefined,
}));

vi.mock("react-router", () => ({
  useNavigate: () => mocks.navigate,
  useLocation: () => ({ state: mocks.locationState }),
  useParams: () => ({ id: mocks.id }),
  useSearchParams: () => [new URLSearchParams(), vi.fn()],
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canCreate: true,
    canEdit: true,
    canDelete: true,
    canView: true,
  }),
}));

vi.mock("@/hooks/use-master-items", () => ({
  useGetMasterItems: () => ({ data: [], isLoading: false }),
}));

vi.mock("@/hooks/use-staffs", () => ({
  useGetStaffs: () => ({ data: [], isLoading: false }),
}));

vi.mock("@/hooks/use-unsaved-changes", () => ({
  useUnsavedChanges: () => ({
    isDirty: false,
    markDirty: vi.fn(),
    markClean: vi.fn(),
  }),
}));

vi.mock("../hooks/use-examination-form", () => ({
  useExaminationForm: () => ({
    formData: { status: "依頼中", petId: "pet-1" },
    setFormData: vi.fn(),
    petSelection: { selectedPets: [], setSelectedPets: vi.fn() },
    isPetDeceased: false,
    formAction: vi.fn(),
    formState: { success: false, timestamp: 0, fieldErrors: {} },
    fieldErrors: {},
    handleDelete: vi.fn(),
    isEdit: true,
    isReadLoading: mocks.isReadLoading,
    isReadNotFound: mocks.isReadNotFound,
    isReadError: mocks.isReadError,
    retryRead: mocks.retryRead,
    isSaving: false,
    isDeleting: false,
    formItems: [],
    setInspectionValue: vi.fn(),
    addManualItem: vi.fn(),
    removeItem: vi.fn(),
    setItemName: vi.fn(),
    handleUnconfirm: vi.fn(),
    isPersistedConfirmed: false,
    isPersistedCompletedLocked: false,
    isPersistedResultsLocked: false,
    isPatientChangeLocked: true,
  }),
}));

vi.mock("../api/get-examinations", () => ({
  useGetExaminations: () => ({ data: [] }),
}));

vi.mock("../api/get-examination-print-snapshot", () => ({
  useGetExaminationPrintSnapshot: () => ({ data: undefined, isLoading: false }),
}));

beforeEach(() => {
  mocks.id = "examination-1";
  mocks.navigate.mockReset();
  mocks.locationState = undefined;
  mocks.isReadLoading = false;
  mocks.isReadNotFound = false;
  mocks.isReadError = false;
  mocks.retryRead = undefined;
});

describe("ExaminationForm read states (EMR-383)", () => {
  it("読み込み中は検査ページ内で読み込み状態を示し、保存・削除を出さない", () => {
    mocks.isReadLoading = true;
    render(<ExaminationForm />);

    expect(screen.getByRole("heading", { name: "検査" })).toBeInTheDocument();
    expect(screen.getByRole("status", { name: "読み込み中" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "戻る" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
  });

  it("存在しないIDは非開示の説明と一覧戻りを出し、編集UIを出さない", async () => {
    const user = userEvent.setup();
    mocks.isReadNotFound = true;
    render(<ExaminationForm />);

    expect(screen.getByRole("heading", { name: "検査" })).toBeInTheDocument();
    expect(screen.getByText("検査記録が見つかりません")).toBeInTheDocument();
    expect(screen.getByText(/表示する権限がありません/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "一覧に戻る" }));
    expect(mocks.navigate).toHaveBeenCalledWith("/examinations");
  });

  it("取得失敗は説明・再試行・一覧戻りを出し、編集UIを出さない", async () => {
    const user = userEvent.setup();
    const retry = vi.fn();
    mocks.isReadError = true;
    mocks.retryRead = retry;
    render(<ExaminationForm />);

    expect(screen.getByText("検査記録の取得に失敗しました")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "再試行" }));
    expect(retry).toHaveBeenCalledTimes(1);

    await user.click(screen.getByRole("button", { name: "一覧に戻る" }));
    expect(mocks.navigate).toHaveBeenCalledWith("/examinations");
  });

  it("一覧に戻るは遷移元が別画面でも必ず検査一覧へ戻る (E001)", async () => {
    const user = userEvent.setup();
    mocks.isReadNotFound = true;
    mocks.locationState = { from: "/owners/990002" };
    render(<ExaminationForm />);

    await user.click(screen.getByRole("button", { name: "一覧に戻る" }));
    expect(mocks.navigate).toHaveBeenCalledWith("/examinations");
  });

  it("ヘッダー戻るは従来どおり遷移元を優先する (E001)", async () => {
    const user = userEvent.setup();
    mocks.isReadNotFound = true;
    mocks.locationState = { from: "/owners/990002" };
    render(<ExaminationForm />);

    await user.click(screen.getByRole("button", { name: "戻る" }));
    expect(mocks.navigate).toHaveBeenCalledWith("/owners/990002");
  });
});
