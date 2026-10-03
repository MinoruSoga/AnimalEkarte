import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

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

import { usePermission } from "@/hooks/use-permission";

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
      {
        id: "10",
        name: "田中 医師",
        isActive: true,
        occupationName: "獣医師",
        staffType: "doctor",
      },
      {
        id: "11",
        name: "鈴木 医師",
        isActive: true,
        occupationName: "獣医師",
        staffType: "doctor",
      },
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

/** 追加フォームを開き、健診種別を選択した状態にする（ui/select = Radix: trigger click → portal option click） */
async function openAddFormWithType(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByText("記録を追加"));
  await user.click(screen.getByLabelText("健診種別"));
  await user.click(screen.getByRole("option", { name: "定期健診" }));
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
  vi.mocked(usePermission).mockReturnValue({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  });
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

  it("追加フォームに担当医セレクトが表示される", async () => {
    const user = userEvent.setup();
    renderComponent();
    await user.click(screen.getByText("記録を追加"));

    // Radix option は trigger を開いたときだけ portal に描画される
    await user.click(screen.getByLabelText("担当医"));

    // 未選択肢と staff options が存在する
    expect(screen.getByRole("option", { name: "担当医" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "田中 医師" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "鈴木 医師" })).toBeInTheDocument();
  });

  it("担当医を選択して追加すると doctor_id が payload に含まれる", async () => {
    const user = userEvent.setup();
    renderComponent();
    await openAddFormWithType(user);

    // 担当医セレクト（追加フォームの Label と trigger が htmlFor/id で対応）
    await user.click(screen.getByLabelText("担当医"));
    await user.click(screen.getByRole("option", { name: "田中 医師" }));

    await user.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(mutateAsyncMock).toHaveBeenCalledWith(expect.objectContaining({ doctor_id: 10 }));
    });
  });

  it("担当医未選択の場合 doctor_id は null で送信される", async () => {
    const user = userEvent.setup();
    renderComponent();
    await openAddFormWithType(user);

    // 担当医を選ばずにそのまま追加（trigger は未選択ラベルのまま）
    expect(screen.getByLabelText("担当医")).toHaveTextContent("担当医");
    await user.click(screen.getByRole("button", { name: "追加" }));

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
    const user = userEvent.setup();
    render(<CheckupsTab medicalRecordId="mr-1" />);

    // 編集ボタンをクリック
    await user.click(screen.getByTitle("編集"));

    // 担当医セレクト（編集行）は既存値「田中 医師」を表示し、"-" への変更でクリアされる
    const doctorTrigger = screen.getByRole("combobox", { name: "担当医 (2026-05-01)" });
    expect(doctorTrigger).toHaveTextContent("田中 医師");
    await user.click(doctorTrigger);
    await user.click(screen.getByRole("option", { name: "-" }));
    // ui/select は value 変化で key remount するため、選択後は trigger を取り直す
    expect(screen.getByRole("combobox", { name: "担当医 (2026-05-01)" })).toHaveTextContent("-");

    // 保存ボタン (Check アイコン) をクリック
    await user.click(screen.getByTitle("保存"));

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

  it("健診種別選択後に動的フィールド（所見）を表示する", async () => {
    const user = userEvent.setup();
    renderComponent();
    await openAddFormWithType(user);

    expect(screen.getByTestId("dynamic-checkup-fields")).toBeInTheDocument();
    expect(screen.getByText("所見")).toBeInTheDocument();
  });

  it("入力した所見を create 後に field-results へ PUT する", async () => {
    const mutateAsyncMock = vi.fn().mockResolvedValue({ id: "c-new" });
    mockCreateMutateAsync(mutateAsyncMock);

    const user = userEvent.setup();
    renderComponent();
    await openAddFormWithType(user);
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

    const user = userEvent.setup();
    renderComponent();
    await openAddFormWithType(user);
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

    const user = userEvent.setup();
    renderComponent();
    await openAddFormWithType(user);
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

    const user = userEvent.setup();
    renderComponent();
    await openAddFormWithType(user);
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

// ─────────────────────────────────────────────────────────────
// EMR-226: raw <select> → ui/select 移行後の保持動作
// ─────────────────────────────────────────────────────────────

describe("CheckupsTab — shared ui/select preserved behaviors (EMR-226)", () => {
  beforeEach(() => {
    vi.mocked(useGetCheckups).mockReturnValue({
      data: [CHECKUP_WITH_DOCTOR],
      isLoading: false,
    } as ReturnType<typeof useGetCheckups>);
  });

  it("編集行は健診種別・担当医の既存値を trigger に表示する", async () => {
    const user = userEvent.setup();
    render(<CheckupsTab medicalRecordId="mr-1" />);

    await user.click(screen.getByTitle("編集"));

    expect(screen.getByRole("combobox", { name: "健診種別 (2026-05-01)" })).toHaveTextContent(
      "定期健診",
    );
    expect(screen.getByRole("combobox", { name: "担当医 (2026-05-01)" })).toHaveTextContent(
      "田中 医師",
    );
  });

  it("健診種別はキーボードだけで選択できる（Enter で開き矢印と Enter で確定）", async () => {
    const user = userEvent.setup();
    renderComponent();
    await user.click(screen.getByText("記録を追加"));

    const typeTrigger = screen.getByLabelText("健診種別");
    typeTrigger.focus();
    await user.keyboard("{Enter}");
    // 未選択肢（先頭）→ 定期健診 へ移動して確定
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{Enter}");

    // ui/select は value 変化で key remount するため、選択後は trigger を取り直す
    expect(screen.getByLabelText("健診種別")).toHaveTextContent("定期健診");
  });

  it("canEdit=false では編集ボタンが出ず行内セレクトへ到達できない（locked-state denial）", () => {
    vi.mocked(usePermission).mockReturnValue({
      canView: true,
      canCreate: true,
      canEdit: false,
      canDelete: false,
    });
    render(<CheckupsTab medicalRecordId="mr-1" />);

    expect(screen.queryByTitle("編集")).not.toBeInTheDocument();
    expect(screen.queryByRole("combobox")).not.toBeInTheDocument();
  });
});
