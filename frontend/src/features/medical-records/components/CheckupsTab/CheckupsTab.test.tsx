import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";

import { CheckupsTab } from "./CheckupsTab";

// ── mock ──────────────────────────────────────────────────────────────────

// CheckupsTab は deep link 用に useSearchParams を読む。Router 無しで render できるよう差し替える。
vi.mock("react-router", async (importOriginal) => ({
  ...(await importOriginal<typeof import("react-router")>()),
  useSearchParams: () => [new URLSearchParams(), vi.fn()],
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({ canCreate: true, canEdit: true, canDelete: true })),
}));

const { replaceCheckupFieldResultsMock, handleApiErrorMock, toastSuccessMock } = vi.hoisted(() => ({
  replaceCheckupFieldResultsMock: vi.fn(),
  handleApiErrorMock: vi.fn(),
  toastSuccessMock: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { success: toastSuccessMock },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: handleApiErrorMock,
}));

vi.mock("../../api/checkups", () => ({
  useGetCheckups: vi.fn(() => ({ data: [], isLoading: false })),
  useCreateCheckup: vi.fn(() => ({
    mutate: vi.fn(),
    mutateAsync: vi.fn(),
    isPending: false,
  })),
  useUpdateCheckup: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
  useDeleteCheckup: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
}));

vi.mock("@/hooks/use-checkup-fields", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/hooks/use-checkup-fields")>();
  return {
    ...actual,
    useGetCheckupTypeFields: vi.fn(),
    replaceCheckupFieldResults: replaceCheckupFieldResultsMock,
  };
});

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllCheckupTypes: vi.fn(() => ({
    data: [{ id: "1", name: "定期健診" }],
  })),
}));

vi.mock("@/hooks/use-staffs", () => ({
  useGetStaffs: vi.fn(() => ({
    data: [
      { id: "10", name: "田中 医師", isActive: true, occupationName: "獣医師" },
      { id: "11", name: "鈴木 医師", isActive: true, occupationName: "獣医師" },
    ],
  })),
}));

import { useCreateCheckup, useUpdateCheckup, useGetCheckups } from "../../api/checkups";
import { useGetCheckupTypeFields, type CheckupTypeFieldRow } from "@/hooks/use-checkup-fields";

const SAMPLE_TEXT_FIELDS: CheckupTypeFieldRow[] = [
  {
    id: 1,
    checkupTypeId: 1,
    name: "所見",
    fieldType: "text",
    unit: "",
    options: [],
    isProvisional: false,
    sortOrder: 1,
  },
];

const CHECKUP_WITH_DOCTOR = {
  id: "c1",
  medical_record_id: "mr-1",
  checkup_type_id: "1",
  date: "2026-05-01",
  next_date: null,
  doctor_id: "10",
  result: "Normal",
  created_at: "2026-05-01T00:00:00Z",
  updated_at: "2026-05-01T00:00:00Z",
  checkup_type: { id: "1", name: "定期健診" },
  doctor: { id: "10", name: "田中 医師" },
};

// ── helpers ──────────────────────────────────────────────────────────────

function renderComponent() {
  return render(<CheckupsTab medicalRecordId="mr-1" />);
}

/** 追加フォームを開き、健診種別を選択した状態にする */
function openAddFormWithType() {
  fireEvent.click(screen.getByText("記録を追加"));
  // 健診種別セレクト：「選択」という option を持つ最初の combobox
  const selects = screen.getAllByRole("combobox");
  const typeSelect = selects.find(
    (s) => (s as HTMLSelectElement).querySelector("option[value='']")?.textContent === "選択",
  ) as HTMLSelectElement;
  fireEvent.change(typeSelect, { target: { value: "1" } });
  return selects;
}

function mockCreateMutateAsync(mutateAsync: ReturnType<typeof vi.fn>) {
  vi.mocked(useCreateCheckup).mockReturnValue({
    mutate: vi.fn(),
    mutateAsync,
    isPending: false,
  } as ReturnType<typeof useCreateCheckup>);
}

beforeEach(() => {
  replaceCheckupFieldResultsMock.mockReset();
  replaceCheckupFieldResultsMock.mockResolvedValue(undefined);
  handleApiErrorMock.mockReset();
  toastSuccessMock.mockReset();
  vi.mocked(useGetCheckups).mockReturnValue({
    data: [],
    isLoading: false,
  } as ReturnType<typeof useGetCheckups>);
  vi.mocked(useGetCheckupTypeFields).mockReturnValue({
    data: [],
  } as ReturnType<typeof useGetCheckupTypeFields>);
});

// ── tests ─────────────────────────────────────────────────────────────────

describe("CheckupsTab — finalized lock", () => {
  it("isFinalized=true のとき「記録を追加」ボタンが表示されない", () => {
    render(<CheckupsTab medicalRecordId="mr-1" isFinalized={true} />);
    expect(screen.queryByText("記録を追加")).toBeNull();
  });

  it("isFinalized=true のとき閲覧専用メッセージが表示される", () => {
    render(<CheckupsTab medicalRecordId="mr-1" isFinalized={true} />);
    expect(screen.getByText("確定済みカルテのため健診情報は編集できません")).toBeInTheDocument();
  });

  it("isFinalized=false のとき「記録を追加」ボタンが表示される", () => {
    render(<CheckupsTab medicalRecordId="mr-1" isFinalized={false} />);
    expect(screen.getByText("記録を追加")).toBeInTheDocument();
  });
});

describe("CheckupsTab — doctor field", () => {
  let mutateAsyncMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mutateAsyncMock = vi.fn().mockResolvedValue({ id: "c-new" });
    vi.mocked(useCreateCheckup).mockReturnValue({
      mutate: vi.fn(),
      mutateAsync: mutateAsyncMock,
      isPending: false,
    } as ReturnType<typeof useCreateCheckup>);
  });

  it("追加フォームに担当医セレクトが表示される", () => {
    renderComponent();
    fireEvent.click(screen.getByText("記録を追加"));
    // 担当医 option と staff options が存在する
    expect(screen.getByRole("option", { name: "担当医" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "田中 医師" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "鈴木 医師" })).toBeInTheDocument();
  });

  it("担当医を選択して追加すると doctor_id が payload に含まれる", async () => {
    renderComponent();
    openAddFormWithType();

    // 担当医セレクト：「担当医」という option を持つ combobox
    const doctorSelect = screen
      .getAllByRole("combobox")
      .find(
        (s) => (s as HTMLSelectElement).querySelector("option[value='']")?.textContent === "担当医",
      ) as HTMLSelectElement;
    expect(doctorSelect).toBeDefined();
    fireEvent.change(doctorSelect, { target: { value: "10" } });

    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(mutateAsyncMock).toHaveBeenCalledWith(expect.objectContaining({ doctor_id: 10 }));
    });
  });

  it("担当医未選択の場合 doctor_id は null で送信される", async () => {
    renderComponent();
    openAddFormWithType();

    // 担当医を選ばずにそのまま追加
    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(mutateAsyncMock).toHaveBeenCalledWith(expect.objectContaining({ doctor_id: null }));
    });
  });
});

// ─────────────────────────────────────────────────────────────
// Issue #59: 担当医クリア (doctor_id_clear flag)
// ─────────────────────────────────────────────────────────────

describe("CheckupsTab — doctor clear (Issue #59)", () => {
  let updateMutateMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    updateMutateMock = vi.fn();
    vi.mocked(useUpdateCheckup).mockReturnValue({
      mutate: updateMutateMock,
      isPending: false,
    } as ReturnType<typeof useUpdateCheckup>);
    vi.mocked(useGetCheckups).mockReturnValue({
      data: [CHECKUP_WITH_DOCTOR],
      isLoading: false,
    } as ReturnType<typeof useGetCheckups>);
  });

  it("担当医を '-' に変更して保存すると doctor_id_clear=true が payload に含まれる", async () => {
    render(<CheckupsTab medicalRecordId="mr-1" />);

    // 編集ボタンをクリック
    fireEvent.click(screen.getByTitle("編集"));

    // 担当医セレクト："-" option を持つ combobox
    const doctorSelect = screen
      .getAllByRole("combobox")
      .find(
        (s) => (s as HTMLSelectElement).querySelector("option[value='']")?.textContent === "-",
      ) as HTMLSelectElement;
    expect(doctorSelect).toBeDefined();
    fireEvent.change(doctorSelect, { target: { value: "" } });

    // 保存ボタン (Check アイコン) をクリック
    fireEvent.click(screen.getByTitle("保存"));

    await waitFor(() => {
      expect(updateMutateMock).toHaveBeenCalledWith(
        expect.objectContaining({
          checkupId: "c1",
          input: expect.objectContaining({ doctor_id_clear: true }),
        }),
        expect.anything(),
      );
    });
  });
});

describe("CheckupsTab — dynamic field results (BUG-004)", () => {
  beforeEach(() => {
    vi.mocked(useGetCheckupTypeFields).mockImplementation(
      (checkupTypeId) =>
        ({
          data: checkupTypeId === "1" ? SAMPLE_TEXT_FIELDS : [],
        }) as ReturnType<typeof useGetCheckupTypeFields>,
    );
  });

  it("健診種別選択後に動的フィールド（所見）を表示する", () => {
    renderComponent();
    openAddFormWithType();

    expect(screen.getByTestId("dynamic-checkup-fields")).toBeInTheDocument();
    expect(screen.getByText("所見")).toBeInTheDocument();
  });

  it("入力した所見を create 後に field-results へ PUT する", async () => {
    const mutateAsyncMock = vi.fn().mockResolvedValue({ id: "c-new" });
    mockCreateMutateAsync(mutateAsyncMock);

    renderComponent();
    openAddFormWithType();
    fireEvent.change(screen.getByLabelText("所見"), { target: { value: "異常なし" } });
    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(mutateAsyncMock).toHaveBeenCalled();
    });
    await waitFor(() => {
      expect(replaceCheckupFieldResultsMock).toHaveBeenCalledTimes(1);
    });
    expect(replaceCheckupFieldResultsMock).toHaveBeenCalledWith(
      "mr-1",
      "c-new",
      expect.arrayContaining([
        expect.objectContaining({
          checkup_type_field_id: 1,
          value_text: "異常なし",
        }),
      ]),
    );
    expect(mutateAsyncMock.mock.invocationCallOrder[0]).toBeLessThan(
      replaceCheckupFieldResultsMock.mock.invocationCallOrder[0],
    );
    expect(toastSuccessMock).toHaveBeenCalledWith("健診記録を追加しました");
  });

  it("所見が未入力なら field-results を PUT しない", async () => {
    const mutateAsyncMock = vi.fn().mockResolvedValue({ id: "c-new" });
    mockCreateMutateAsync(mutateAsyncMock);

    renderComponent();
    openAddFormWithType();
    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(mutateAsyncMock).toHaveBeenCalled();
    });
    expect(replaceCheckupFieldResultsMock).not.toHaveBeenCalled();
    expect(replaceCheckupFieldResultsMock).not.toHaveBeenCalledWith(
      expect.anything(),
      expect.anything(),
      [],
    );
  });

  it("create が失敗したら field-results を PUT しない", async () => {
    const mutateAsyncMock = vi.fn().mockRejectedValue(new Error("create failed"));
    mockCreateMutateAsync(mutateAsyncMock);

    renderComponent();
    openAddFormWithType();
    fireEvent.change(screen.getByLabelText("所見"), { target: { value: "異常なし" } });
    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(mutateAsyncMock).toHaveBeenCalled();
    });
    expect(replaceCheckupFieldResultsMock).not.toHaveBeenCalled();
    expect(toastSuccessMock).not.toHaveBeenCalled();
  });

  it("field-results の PUT が失敗したら成功トーストを出さない", async () => {
    const mutateAsyncMock = vi.fn().mockResolvedValue({ id: "c-new" });
    mockCreateMutateAsync(mutateAsyncMock);
    replaceCheckupFieldResultsMock.mockRejectedValue(new Error("put failed"));

    renderComponent();
    openAddFormWithType();
    fireEvent.change(screen.getByLabelText("所見"), { target: { value: "異常なし" } });
    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(replaceCheckupFieldResultsMock).toHaveBeenCalled();
    });
    expect(handleApiErrorMock).toHaveBeenCalled();
    expect(toastSuccessMock).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "追加" })).toBeInTheDocument();
  });
});
