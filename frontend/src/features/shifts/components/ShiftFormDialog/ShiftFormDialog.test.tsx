import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ComponentProps } from "react";
import { describe, expect, it, vi, beforeEach } from "vitest";

import type { Shift, ShiftTemplate } from "../../types";
import { ShiftFormDialog } from "./ShiftFormDialog";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const createShiftMock = vi.fn();
const updateShiftMock = vi.fn();
const deleteShiftMock = vi.fn();
const toastError = vi.fn();
const getShiftByIdMock = vi.fn();
const getShiftOperationResultMock = vi.fn();
const getShiftTemplatesMock = vi.fn((): { data: ShiftTemplate[] } => ({ data: [] }));

vi.mock("../../api/get-shift-templates", () => ({
  useGetShiftTemplates: () => getShiftTemplatesMock(),
}));

vi.mock("../../api/delete-shift", () => ({
  useDeleteShift: () => ({ mutateAsync: deleteShiftMock, isPending: false }),
}));

vi.mock("../../api/create-shift", () => ({
  createShift: (...args: unknown[]) => createShiftMock(...args),
}));

vi.mock("../../api/update-shift", () => ({
  updateShift: (...args: unknown[]) => updateShiftMock(...args),
}));

vi.mock("../../api/get-shifts", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../api/get-shifts")>();
  return {
    ...actual,
    getShiftById: (...args: unknown[]) => getShiftByIdMock(...args),
  };
});

vi.mock("../../api/shift-operation", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../../api/shift-operation")>();
  return {
    ...actual,
    newOperationKey: () => "op-key-fixed",
    getShiftOperationResult: (...args: unknown[]) => getShiftOperationResultMock(...args),
  };
});

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

vi.mock("sonner", () => ({
  toast: { error: (...args: unknown[]) => toastError(...args), success: vi.fn() },
}));

const EDIT_SHIFT: Shift = {
  id: "shift-1",
  clinic_id: "clinic-1",
  staff_id: "staff-1",
  staff_name: "テストスタッフ",
  date: "2026-08-01",
  shift_type: "full",
  starts_at: "2026-08-01T09:00:00+09:00",
  ends_at: "2026-08-01T18:00:00+09:00",
  version: 3,
  interval_state: "known",
  notes: "",
  breaks: [],
  created_at: "2026-07-01T00:00:00+09:00",
  updated_at: "2026-07-01T00:00:00+09:00",
};

const RECEIPT = {
  operation_key: "op-key-fixed",
  intent: "update" as const,
  shift_id: "shift-1",
  resulting_version: 4,
  shift: { ...EDIT_SHIFT, version: 4 },
};

const CREATE_RECEIPT = {
  ...RECEIPT,
  intent: "create" as const,
  shift_id: "new-shift",
};

const BASE_PROPS: ComponentProps<typeof ShiftFormDialog> = {
  open: true,
  onClose: vi.fn(),
  staffId: "staff-1",
  staffName: "テストスタッフ",
  date: "2026-08-01",
  canCreate: true,
  canEdit: true,
};

function renderDialog(props: Partial<ComponentProps<typeof ShiftFormDialog>> = {}) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const view = render(
    <QueryClientProvider client={queryClient}>
      <ShiftFormDialog {...BASE_PROPS} {...props} />
    </QueryClientProvider>,
  );
  return {
    ...view,
    rerenderDialog: (next: Partial<ComponentProps<typeof ShiftFormDialog>> = {}) =>
      view.rerender(
        <QueryClientProvider client={queryClient}>
          <ShiftFormDialog {...BASE_PROPS} {...props} {...next} />
        </QueryClientProvider>,
      ),
  };
}

describe("ShiftFormDialog responsive layout", () => {
  it("開始・終了時刻はmobileで1列、sm以上で2列になる", () => {
    renderDialog();

    const grid = screen.getByLabelText("開始時刻").parentElement?.parentElement;
    expect(grid).toHaveClass("grid-cols-1", "sm:grid-cols-2");
    expect(grid).not.toHaveClass("grid-cols-2");
  });
});

describe("ShiftFormDialog BUG-036 required times", () => {
  beforeEach(() => {
    createShiftMock.mockReset().mockResolvedValue(CREATE_RECEIPT);
    updateShiftMock.mockReset().mockResolvedValue(RECEIPT);
    deleteShiftMock.mockReset();
    toastError.mockReset();
    getShiftByIdMock.mockReset();
    getShiftOperationResultMock.mockReset();
  });

  it("全日で開始・終了が空のまま保存するとエラーを出し API を呼ばない", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("開始時刻と終了時刻を入力してください")).toBeInTheDocument();
    expect(createShiftMock).not.toHaveBeenCalled();
  });

  // EMR-319: 必須時刻エラーで両時刻入力に aria-invalid + describedby(shift-time-error)。
  it("時刻エラー時に開始・終了時刻 Input が aria-invalid で共通エラー文を指す", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.click(screen.getByRole("button", { name: "保存" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveAttribute("id", "shift-time-error");
    for (const label of ["開始時刻", "終了時刻"]) {
      const input = screen.getByLabelText(label);
      expect(input).toHaveAttribute("aria-invalid", "true");
      expect(input).toHaveAttribute("aria-describedby", "shift-time-error");
    }
  });

  it("全日で時刻を入れて保存すると契約の明示日時で createShift が呼ばれる", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByLabelText("開始時刻"), "09:00");
    await user.type(screen.getByLabelText("終了時刻"), "18:00");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(createShiftMock).toHaveBeenCalledTimes(1);
    });
    expect(createShiftMock.mock.calls[0][0]).toStrictEqual({
      operation_key: "op-key-fixed",
      staff_id: "staff-1",
      date: "2026-08-01",
      shift_type: "full",
      starts_at: "2026-08-01T09:00:00+09:00",
      ends_at: "2026-08-01T18:00:00+09:00",
      breaks: [],
      notes: "",
    });
    expect(toastError).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });
});

describe("ShiftFormDialog mutation permission re-check", () => {
  beforeEach(() => {
    createShiftMock.mockReset().mockResolvedValue(CREATE_RECEIPT);
    updateShiftMock.mockReset().mockResolvedValue(RECEIPT);
    deleteShiftMock.mockReset().mockResolvedValue({
      operation_key: "op-key-fixed",
      intent: "delete",
      shift_id: "shift-1",
      resulting_version: null,
      shift: null,
    });
    toastError.mockReset();
    getShiftByIdMock.mockReset();
    getShiftOperationResultMock.mockReset();
  });

  it("canCreate 未指定（canEdit=true）の新規保存は fail-closed で createShift しない", async () => {
    const user = userEvent.setup();
    renderDialog({ canCreate: undefined, canEdit: true });

    await user.type(screen.getByLabelText("開始時刻"), "09:00");
    await user.type(screen.getByLabelText("終了時刻"), "18:00");

    const form = screen.getByLabelText("開始時刻").closest("form");
    expect(form).not.toBeNull();

    await act(async () => {
      form?.requestSubmit();
    });

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(createShiftMock).not.toHaveBeenCalled();
  });

  it("canCreate=false の新規保存は createShift せず toast.error する", async () => {
    const user = userEvent.setup();
    const { rerenderDialog } = renderDialog({ canCreate: true, canEdit: true });

    await user.type(screen.getByLabelText("開始時刻"), "09:00");
    await user.type(screen.getByLabelText("終了時刻"), "18:00");

    const form = screen.getByLabelText("開始時刻").closest("form");
    expect(form).not.toBeNull();

    rerenderDialog({ canCreate: false, canEdit: true });

    await act(async () => {
      form?.requestSubmit();
    });

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(createShiftMock).not.toHaveBeenCalled();
    expect(updateShiftMock).not.toHaveBeenCalled();
  });

  it("canEdit=false の更新保存は updateShift せず toast.error する", async () => {
    const { rerenderDialog } = renderDialog({
      editShift: EDIT_SHIFT,
      canCreate: true,
      canEdit: true,
    });

    const form = screen.getByLabelText("開始時刻").closest("form");
    expect(form).not.toBeNull();

    rerenderDialog({ editShift: EDIT_SHIFT, canCreate: true, canEdit: false });

    await act(async () => {
      form?.requestSubmit();
    });

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(updateShiftMock).not.toHaveBeenCalled();
    expect(createShiftMock).not.toHaveBeenCalled();
  });

  it("canDelete=true なら削除確認後に期待版と操作キー付きで deleteShift を呼ぶ", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    renderDialog({
      editShift: EDIT_SHIFT,
      canDelete: true,
      onClose,
    });

    await user.click(screen.getByRole("button", { name: "削除" }));
    const confirmDialog = await screen.findByRole("alertdialog");
    const confirmButtons = screen.getAllByRole("button", { name: "削除" });
    expect(confirmDialog).toBeInTheDocument();
    await user.click(confirmButtons[confirmButtons.length - 1]);

    expect(deleteShiftMock).toHaveBeenCalledWith({
      shiftId: "shift-1",
      expectedVersion: 3,
      operationKey: "op-key-fixed",
    });
    expect(toastError).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canDelete=false なら削除確定で mutate せず toast.error する", async () => {
    const user = userEvent.setup();
    const { rerenderDialog } = renderDialog({
      editShift: EDIT_SHIFT,
      canDelete: true,
    });

    await user.click(screen.getByRole("button", { name: "削除" }));
    await screen.findByRole("alertdialog");

    rerenderDialog({ editShift: EDIT_SHIFT, canDelete: false });

    const confirmButtons = screen.getAllByRole("button", { name: "削除" });
    await user.click(confirmButtons[confirmButtons.length - 1]);

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(deleteShiftMock).not.toHaveBeenCalled();
  });
});

// EMR-514: off/paid_leave は終日非勤務。spec 24-shift-calendar の入力制御に従い
// 時刻・休憩入力を無効化し、保存時は入力値を捨てず state 保持したまま契約どおり
// 明示NULL・breaks[] を送る。片方だけの時刻入力は全カテゴリで拒否。
const EMPTY_FULL_TEMPLATE: ShiftTemplate = {
  id: "tpl-empty",
  clinic_id: "clinic-1",
  name: "空時刻テンプレート",
  shift_type: "full",
  start_time: "",
  end_time: "",
  notes: "",
  sort_order: 1,
  is_active: true,
  breaks: [],
  created_at: "",
  updated_at: "",
};

async function selectOption(
  user: ReturnType<typeof userEvent.setup>,
  comboboxName: string,
  optionName: string,
) {
  await user.click(screen.getByRole("combobox", { name: comboboxName }));
  const listbox = await screen.findByRole("listbox");
  await user.click(within(listbox).getByRole("option", { name: optionName }));
}

describe("ShiftFormDialog EMR-514 off/paid_leave contract", () => {
  beforeEach(() => {
    createShiftMock.mockReset().mockResolvedValue(CREATE_RECEIPT);
    updateShiftMock.mockReset().mockResolvedValue(RECEIPT);
    deleteShiftMock.mockReset();
    toastError.mockReset();
    getShiftTemplatesMock.mockReset().mockReturnValue({ data: [] });
  });

  it("休日(off)選択時は時刻・休憩入力が無効化される（spec 24-shift-calendar 入力制御）", async () => {
    const user = userEvent.setup();
    renderDialog();

    await selectOption(user, "シフト種別", "休日");

    expect(screen.getByLabelText("開始時刻")).toBeDisabled();
    expect(screen.getByLabelText("終了時刻")).toBeDisabled();
    expect(screen.getByText("休憩時間")).toBeInTheDocument();
  });

  it("off で保存すると明示NULL・breaks[] の createShift になる", async () => {
    const user = userEvent.setup();
    renderDialog();

    await selectOption(user, "シフト種別", "休日");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(createShiftMock).toHaveBeenCalledTimes(1);
    });
    const input = createShiftMock.mock.calls[0][0] as {
      shift_type: string;
      starts_at: string | null;
      ends_at: string | null;
      breaks: unknown[];
    };
    expect(input.shift_type).toBe("off");
    expect(input.starts_at).toBeNull();
    expect(input.ends_at).toBeNull();
    expect(input.breaks).toStrictEqual([]);
  });

  it("paid_leave で保存しても明示NULL・breaks[] を送る", async () => {
    const user = userEvent.setup();
    renderDialog();

    await selectOption(user, "シフト種別", "有休");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(createShiftMock).toHaveBeenCalledTimes(1);
    });
    const input = createShiftMock.mock.calls[0][0] as {
      shift_type: string;
      starts_at: string | null;
      breaks: unknown[];
    };
    expect(input.shift_type).toBe("paid_leave");
    expect(input.starts_at).toBeNull();
    expect(input.breaks).toStrictEqual([]);
  });

  it("off で入力した時刻は捨てず state保持し、送信は明示NULLになる", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByLabelText("開始時刻"), "10:00");
    await user.type(screen.getByLabelText("終了時刻"), "15:00");
    await selectOption(user, "シフト種別", "休日");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(createShiftMock).toHaveBeenCalledTimes(1);
    });
    const sent = createShiftMock.mock.calls[0][0] as {
      shift_type: string;
      starts_at: string | null;
      ends_at: string | null;
    };
    expect(sent.shift_type).toBe("off");
    expect(sent.starts_at).toBeNull();
    expect(sent.ends_at).toBeNull();

    // 入力値は捨てない: 勤務種別へ戻すと保持した時刻が入力欄に残る
    await selectOption(user, "シフト種別", "全日");
    expect(screen.getByLabelText("開始時刻")).toHaveValue("10:00");
    expect(screen.getByLabelText("終了時刻")).toHaveValue("15:00");
  });

  it("full で開始時刻だけ入力して保存するとエラーで API を呼ばない", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByLabelText("開始時刻"), "09:00");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("開始時刻と終了時刻を入力してください")).toBeInTheDocument();
    expect(createShiftMock).not.toHaveBeenCalled();
  });

  it("勤務種別で終了時刻だけ入力して保存するとエラーで API を呼ばない", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByLabelText("終了時刻"), "18:00");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("開始時刻と終了時刻を入力してください")).toBeInTheDocument();
    expect(createShiftMock).not.toHaveBeenCalled();
  });

  it("編集で勤務→休日へ変えると両日時NULL・breaks[] の明示遷移を送る", async () => {
    const user = userEvent.setup();
    renderDialog({ editShift: EDIT_SHIFT });

    await selectOption(user, "シフト種別", "休日");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(updateShiftMock).toHaveBeenCalledTimes(1);
    });
    expect(updateShiftMock.mock.calls[0][1]).toStrictEqual({
      operation_key: "op-key-fixed",
      expected_version: 3,
      shift_type: "off",
      starts_at: null,
      ends_at: null,
      breaks: [],
    });
  });

  it("空時刻の勤務種別テンプレートを適用して保存するとエラーで黙って保存しない", async () => {
    const user = userEvent.setup();
    getShiftTemplatesMock.mockReturnValue({ data: [EMPTY_FULL_TEMPLATE] });
    renderDialog();

    await selectOption(user, "テンプレートから入力", "空時刻テンプレート");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("開始時刻と終了時刻を入力してください")).toBeInTheDocument();
    expect(createShiftMock).not.toHaveBeenCalled();
  });
});

describe("ShiftFormDialog EMR-514 PATCH minimal diff", () => {
  beforeEach(() => {
    updateShiftMock.mockReset().mockResolvedValue(RECEIPT);
    toastError.mockReset();
  });

  it("メモだけ変更した更新は未指定項目を送らず維持を示す", async () => {
    const user = userEvent.setup();
    renderDialog({ editShift: EDIT_SHIFT });

    const notes = screen.getByLabelText("メモ");
    await user.type(notes, " afternoon メモ");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(updateShiftMock).toHaveBeenCalledTimes(1);
    });
    expect(updateShiftMock).toHaveBeenCalledWith("shift-1", {
      operation_key: "op-key-fixed",
      expected_version: 3,
      notes: " afternoon メモ",
    });
  });

  it("変更がない更新は API を呼ばず案内を出す", async () => {
    const user = userEvent.setup();
    renderDialog({ editShift: EDIT_SHIFT });

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("変更がありません")).toBeInTheDocument();
    expect(updateShiftMock).not.toHaveBeenCalled();
  });
});

describe("ShiftFormDialog EMR-514 unknown breaks", () => {
  beforeEach(() => {
    updateShiftMock.mockReset().mockResolvedValue(RECEIPT);
    toastError.mockReset();
  });

  const EDIT_WITH_UNKNOWN: Shift = {
    ...EDIT_SHIFT,
    interval_state: "unknown",
    starts_at: null,
    ends_at: null,
    breaks: [
      { interval_state: "unknown", starts_at: null, ends_at: null },
      {
        interval_state: "known",
        starts_at: "2026-08-01T12:00:00+09:00",
        ends_at: "2026-08-01T13:00:00+09:00",
      },
    ],
  };

  it("unknown勤務は時刻の訂正確認まで保存自体が停止する（M2）", async () => {
    const user = userEvent.setup();
    renderDialog({ editShift: EDIT_WITH_UNKNOWN });

    // unknown は未解決として表示（値の補完をしない）
    expect(screen.getByText(/時刻未解決/)).toBeInTheDocument();
    expect(screen.getByTestId("unresolved-breaks-note")).toBeInTheDocument();

    // メモだけ変更して保存 → ゲートで停止（updateShiftは呼ばれない）
    await user.type(screen.getByLabelText("メモ"), "a");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText(/訂正確認するまで保存できません/)).toBeInTheDocument();
    expect(updateShiftMock).not.toHaveBeenCalled();
  });

  it("known休日から勤務種別へ変えるときは時刻必須（M1: 時刻なしPATCHを送らない）", async () => {
    const user = userEvent.setup();
    const KNOWN_OFF: Shift = {
      ...EDIT_SHIFT,
      shift_type: "off",
      starts_at: null,
      ends_at: null,
    };
    renderDialog({ editShift: KNOWN_OFF });

    await selectOption(user, "シフト種別", "全日");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("開始時刻と終了時刻を入力してください")).toBeInTheDocument();
    expect(updateShiftMock).not.toHaveBeenCalled();
  });

  it("日跨ぎ既存勤務の時刻変更なし保存は時刻を送らず検証もしない（B3）", async () => {
    const user = userEvent.setup();
    const OVERNIGHT: Shift = {
      ...EDIT_SHIFT,
      starts_at: "2026-08-01T22:00:00+09:00",
      ends_at: "2026-08-02T02:00:00+09:00",
    };
    renderDialog({ editShift: OVERNIGHT });

    // 時刻入力欄は 22:00 / 02:00（end <= start だが既存値）
    expect(screen.getByLabelText("開始時刻")).toHaveValue("22:00");
    expect(screen.getByLabelText("終了時刻")).toHaveValue("02:00");
    await user.type(screen.getByLabelText("メモ"), "a");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(updateShiftMock).toHaveBeenCalledTimes(1);
    });
    const req = updateShiftMock.mock.calls[0][1] as Record<string, unknown>;
    // 時刻未変更のため日時は送らない（未指定=維持）
    expect(req).not.toHaveProperty("starts_at");
    expect(req).not.toHaveProperty("ends_at");
    expect(req.notes).toBe("a");
  });

  it("休憩付き日跨ぎ勤務も、休憩未変更ならメモだけ保存できる（B3残）", async () => {
    const user = userEvent.setup();
    const OVERNIGHT_WITH_BREAK: Shift = {
      ...EDIT_SHIFT,
      starts_at: "2026-08-01T22:00:00+09:00",
      ends_at: "2026-08-02T02:00:00+09:00",
      breaks: [
        {
          interval_state: "known",
          starts_at: "2026-08-01T23:00:00+09:00",
          ends_at: "2026-08-01T23:30:00+09:00",
        },
      ],
    };
    renderDialog({ editShift: OVERNIGHT_WITH_BREAK });

    // 休憩 23:00〜23:30 は勤務時間内検証の文字列比較ではアウトだが未変更のため検証しない
    await user.type(screen.getByLabelText("メモ"), "a");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(updateShiftMock).toHaveBeenCalledTimes(1);
    });
    const req = updateShiftMock.mock.calls[0][1] as Record<string, unknown>;
    expect(req).not.toHaveProperty("starts_at");
    expect(req).not.toHaveProperty("ends_at");
    expect(req).not.toHaveProperty("breaks");
    expect(req.notes).toBe("a");
  });

  it("未解決休憩があるunknown勤務の休日への変更は訂正確認まで停止する（B5/M2）", async () => {
    const user = userEvent.setup();
    renderDialog({ editShift: EDIT_WITH_UNKNOWN });

    await selectOption(user, "シフト種別", "休日");
    await user.click(screen.getByRole("button", { name: "保存" }));

    // unknown記録は時刻の訂正確認まで全保存が停止する（work→offの全置換消去も含め不可）
    expect(await screen.findByText(/訂正確認するまで保存できません/)).toBeInTheDocument();
    expect(updateShiftMock).not.toHaveBeenCalled();
  });

  it("訂正する操作で unknown休憩を編集行へ移し、明示時刻入力後に配列全置換を送れる", async () => {
    const user = userEvent.setup();
    renderDialog({ editShift: EDIT_WITH_UNKNOWN });

    await user.click(screen.getByRole("button", { name: "休憩（未解決）1を訂正する" }));

    // 補完はしない: 追加された行は空欄（既定値を入れない）
    expect(screen.getByLabelText("休憩2 開始時刻")).toHaveValue("");
    // 時刻の訂正確認（ゲート通過）と休憩の明示入力
    await user.type(screen.getByLabelText("開始時刻"), "09:00");
    await user.type(screen.getByLabelText("終了時刻"), "18:00");
    await user.type(screen.getByLabelText("休憩2 開始時刻"), "12:00");
    await user.type(screen.getByLabelText("休憩2 終了時刻"), "13:00");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(updateShiftMock).toHaveBeenCalledTimes(1);
    });
    const req = updateShiftMock.mock.calls[0][1] as {
      starts_at?: string;
      ends_at?: string;
      breaks?: { starts_at: string; ends_at: string }[];
    };
    // 旧NULL → 明示日時の遷移を必ず送る
    expect(req.starts_at).toBe("2026-08-01T09:00:00+09:00");
    expect(req.ends_at).toBe("2026-08-01T18:00:00+09:00");
    // 配列全置換（既知行＋訂正行）。unknown行は要求へ流用しない
    expect(req.breaks).toStrictEqual([
      { starts_at: "2026-08-01T12:00:00+09:00", ends_at: "2026-08-01T13:00:00+09:00" },
      { starts_at: "2026-08-01T12:00:00+09:00", ends_at: "2026-08-01T13:00:00+09:00" },
    ]);
  });
});

describe("ShiftFormDialog EMR-514 conflict & pending", () => {
  beforeEach(() => {
    updateShiftMock.mockReset();
    deleteShiftMock.mockReset();
    toastError.mockReset();
    getShiftByIdMock.mockReset();
    getShiftOperationResultMock.mockReset();
  });

  function axiosConflict(data: unknown) {
    return Object.assign(new Error("Request failed"), {
      isAxiosError: true,
      response: { status: 409, data },
    });
  }

  it("version_conflict は入力を保持し、最新版の採用で期待版が切り替わる（B1/M4）", async () => {
    const user = userEvent.setup();
    updateShiftMock
      .mockRejectedValueOnce(axiosConflict({ error: "conflict", code: "version_conflict" }))
      .mockResolvedValueOnce({
        ...RECEIPT,
        resulting_version: 5,
        shift: { ...EDIT_SHIFT, version: 5 },
      });
    getShiftByIdMock.mockResolvedValue({
      ...EDIT_SHIFT,
      version: 4,
      shift_type: "morning",
      starts_at: "2026-08-01T08:00:00+09:00",
    });
    renderDialog({ editShift: EDIT_SHIFT });

    await user.type(screen.getByLabelText("メモ"), "x");
    await user.click(screen.getByRole("button", { name: "保存" }));

    const area = await screen.findByTestId("shift-conflict");
    expect(area).toHaveTextContent("他の操作で更新されました");
    // 入力保持
    expect(screen.getByLabelText("メモ")).toHaveValue("x");

    await user.click(screen.getByRole("button", { name: "最新の勤務を取得" }));
    expect(await screen.findByTestId("latest-shift-summary")).toHaveTextContent("版 4");

    // 最新版の採用を明示 → 競合表示が消え、次の保存は期待版4で送られる
    await user.click(screen.getByRole("button", { name: /この最新版（版 4）を基準にする/ }));
    expect(screen.queryByTestId("shift-conflict")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(updateShiftMock).toHaveBeenCalledTimes(2);
    });
    expect(updateShiftMock.mock.calls[1][1]).toMatchObject({
      operation_key: "op-key-fixed",
      expected_version: 4,
    });
  });

  it("時刻判明済みでも未解決休憩が残ればメモ修正も停止する（R1/M2残）", async () => {
    const user = userEvent.setup();
    const KNOWN_TIMES_UNKNOWN_BREAK: Shift = {
      ...EDIT_SHIFT,
      interval_state: "unknown",
      breaks: [{ interval_state: "unknown", starts_at: null, ends_at: null }],
    };
    renderDialog({ editShift: KNOWN_TIMES_UNKNOWN_BREAK });

    // 時刻は既知で入力済みだが、未解決休憩があるため保存は停止
    await user.type(screen.getByLabelText("メモ"), "a");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(
      await screen.findByText(/未解決の休憩があります。訂正確認するまで保存できません/),
    ).toBeInTheDocument();
    expect(updateShiftMock).not.toHaveBeenCalled();
  });

  it("訂正確認後に保存できる（R1の出口）", async () => {
    const user = userEvent.setup();
    updateShiftMock.mockReset().mockResolvedValue(RECEIPT);
    const KNOWN_TIMES_UNKNOWN_BREAK: Shift = {
      ...EDIT_SHIFT,
      interval_state: "unknown",
      breaks: [{ interval_state: "unknown", starts_at: null, ends_at: null }],
    };
    renderDialog({ editShift: KNOWN_TIMES_UNKNOWN_BREAK });

    // 最初の保存は未解決休憩で停止（フォームエラーと常設ノートの両方に文言が出る）
    await user.type(screen.getByLabelText("メモ"), "a");
    await user.click(screen.getByRole("button", { name: "保存" }));
    const gateMessages = await screen.findAllByText(/未解決の休憩があります/);
    expect(gateMessages.length).toBeGreaterThan(0);
    expect(updateShiftMock).not.toHaveBeenCalled();

    // 訂正確認（空欄行に明示時刻を入力）→ 保存できる
    await user.click(screen.getByRole("button", { name: "休憩（未解決）1を訂正する" }));
    await user.type(screen.getByLabelText("休憩1 開始時刻"), "12:00");
    await user.type(screen.getByLabelText("休憩1 終了時刻"), "13:00");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(updateShiftMock).toHaveBeenCalledTimes(1);
    });
    const req = updateShiftMock.mock.calls[0][1] as {
      starts_at?: string;
      ends_at?: string;
      breaks?: unknown[];
    };
    expect(req.breaks).toHaveLength(1);
  });

  it("protected_bookings は予約一覧を表示し自動解除の文言を出さない", async () => {
    const user = userEvent.setup();
    updateShiftMock.mockRejectedValue(
      axiosConflict({
        error: "conflict",
        code: "protected_bookings",
        protected_bookings: [
          {
            appointment_id: "10",
            starts_at: "2026-08-01T10:00:00+09:00",
            ends_at: "2026-08-01T11:00:00+09:00",
            patient_label: "ハチ",
            reservation_type_label: "再診",
          },
        ],
      }),
    );
    renderDialog({ editShift: EDIT_SHIFT });

    await user.type(screen.getByLabelText("メモ"), "x");
    await user.click(screen.getByRole("button", { name: "保存" }));

    const area = await screen.findByTestId("shift-conflict");
    expect(area).toHaveTextContent("予約は自動解除されません");
    const list = within(await screen.findByTestId("protected-bookings"));
    expect(list.getByText(/ハチ（再診）/)).toBeInTheDocument();
    expect(list.getByText(/10:00〜11:00/)).toBeInTheDocument();
  });

  it("cross_clinic_overlap は契約の医院名と時間だけを表示する", async () => {
    const user = userEvent.setup();
    updateShiftMock.mockRejectedValue(
      axiosConflict({
        error: "conflict",
        code: "cross_clinic_overlap",
        conflicting_clinic_name: "B動物病院",
        overlap_starts_at: "2026-08-01T09:00:00+09:00",
        overlap_ends_at: "2026-08-01T10:00:00+09:00",
      }),
    );
    renderDialog({ editShift: EDIT_SHIFT });

    await user.type(screen.getByLabelText("メモ"), "x");
    await user.click(screen.getByRole("button", { name: "保存" }));

    const area = await screen.findByTestId("cross-clinic-overlap");
    expect(area).toHaveTextContent("B動物病院");
    expect(area).toHaveTextContent("09:00〜10:00");
  });

  it("通信断では新key再送を禁止し、同key結果照会でsucceededなら確定する", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    updateShiftMock.mockRejectedValueOnce(
      Object.assign(new Error("Network Error"), { isAxiosError: true }),
    );
    renderDialog({ editShift: EDIT_SHIFT, onClose });

    await user.type(screen.getByLabelText("メモ"), "x");
    await user.click(screen.getByRole("button", { name: "保存" }));

    const pending = await screen.findByTestId("pending-operation");
    expect(pending).toHaveTextContent("結果が未確定");
    expect(updateShiftMock).toHaveBeenCalledTimes(1);

    // unknown中は保存・削除を無効化（新key再送禁止）
    expect(screen.getByRole("button", { name: "保存" })).toBeDisabled();

    getShiftOperationResultMock.mockResolvedValue({
      operation_key: "op-key-fixed",
      state: "succeeded",
      receipt: RECEIPT,
    });
    await user.click(screen.getByRole("button", { name: "結果を照会" }));

    expect(getShiftOperationResultMock).toHaveBeenCalledWith("op-key-fixed", "update");
    await waitFor(() => {
      expect(onClose).toHaveBeenCalled();
    });
  });

  it("照会がunknownの間は再送せず停止を続ける", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    updateShiftMock.mockRejectedValueOnce(
      Object.assign(new Error("Network Error"), { isAxiosError: true }),
    );
    renderDialog({ editShift: EDIT_SHIFT, onClose });

    await user.type(screen.getByLabelText("メモ"), "x");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await screen.findByTestId("pending-operation");

    getShiftOperationResultMock.mockResolvedValue({
      operation_key: "op-key-fixed",
      state: "unknown",
      receipt: null,
    });
    await user.click(screen.getByRole("button", { name: "結果を照会" }));

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith(
        "結果はまだ確定していません。時間を置いてもう一度照会してください",
      );
    });
    expect(updateShiftMock).toHaveBeenCalledTimes(1);
    expect(onClose).not.toHaveBeenCalled();
  });

  it("照会がunknownの間はキャンセルが無効で、破棄して閉じるのみ出口になる（B2/M5/B10）", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    updateShiftMock.mockRejectedValueOnce(
      Object.assign(new Error("Network Error"), { isAxiosError: true }),
    );
    renderDialog({ editShift: EDIT_SHIFT, onClose });

    await user.type(screen.getByLabelText("メモ"), "x");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await screen.findByTestId("pending-operation");
    expect(screen.getByRole("button", { name: "キャンセル" })).toBeDisabled();

    // 明示的に破棄した場合のみ閉じられる（再送はしない）
    await user.click(screen.getByRole("button", { name: "破棄して閉じる" }));
    expect(onClose).toHaveBeenCalledTimes(1);
    expect(updateShiftMock).toHaveBeenCalledTimes(1);
  });

  it("照会succeededのreceiptが契約不整合なら確定せず停止する（B8/m3）", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    updateShiftMock.mockRejectedValueOnce(
      Object.assign(new Error("Network Error"), { isAxiosError: true }),
    );
    getShiftOperationResultMock.mockResolvedValue({
      operation_key: "op-key-fixed",
      state: "succeeded",
      receipt: {
        operation_key: "op-key-fixed",
        intent: "update",
        shift_id: "shift-999", // shift.id と不一致（意味検証で不整合）
        resulting_version: 4,
        shift: { ...EDIT_SHIFT, version: 4 },
      },
    });
    renderDialog({ editShift: EDIT_SHIFT, onClose });

    await user.type(screen.getByLabelText("メモ"), "x");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await screen.findByTestId("pending-operation");
    await user.click(screen.getByRole("button", { name: "結果を照会" }));

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith("操作結果の形式が不正です");
    });
    expect(onClose).not.toHaveBeenCalled();
  });

  it("削除もversion_conflict後に最新版を採用して再削除できる（R2/M4残）", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    deleteShiftMock
      .mockRejectedValueOnce(axiosConflict({ error: "conflict", code: "version_conflict" }))
      .mockResolvedValueOnce({
        operation_key: "op-key-fixed",
        intent: "delete",
        shift_id: "shift-1",
        resulting_version: null,
        shift: null,
      });
    getShiftByIdMock.mockResolvedValue({ ...EDIT_SHIFT, version: 4 });
    renderDialog({ editShift: EDIT_SHIFT, canDelete: true, onClose });

    await user.click(screen.getByRole("button", { name: "削除" }));
    const confirmButtons = screen.getAllByRole("button", { name: "削除" });
    await user.click(confirmButtons[confirmButtons.length - 1]);

    // 1回目は旧版で409 → conflict表示
    await screen.findByTestId("shift-conflict");
    await user.click(screen.getByRole("button", { name: "最新の勤務を取得" }));
    await screen.findByTestId("latest-shift-summary");
    await user.click(screen.getByRole("button", { name: /この最新版（版 4）を基準にする/ }));

    // 再削除は採用した版4で送られる
    await user.click(screen.getByRole("button", { name: "削除" }));
    const confirmButtons2 = screen.getAllByRole("button", { name: "削除" });
    await user.click(confirmButtons2[confirmButtons2.length - 1]);

    await waitFor(() => {
      expect(deleteShiftMock).toHaveBeenCalledTimes(2);
    });
    expect(deleteShiftMock.mock.calls[1][0]).toMatchObject({
      shiftId: "shift-1",
      expectedVersion: 4,
    });
    await waitFor(() => {
      expect(onClose).toHaveBeenCalled();
    });
  });

  it("削除の通信断も同じkeyの結果照会へ切り替える", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    deleteShiftMock.mockRejectedValue(
      Object.assign(new Error("Network Error"), { isAxiosError: true }),
    );
    renderDialog({ editShift: EDIT_SHIFT, canDelete: true, onClose });

    await user.click(screen.getByRole("button", { name: "削除" }));
    const confirmButtons = screen.getAllByRole("button", { name: "削除" });
    await user.click(confirmButtons[confirmButtons.length - 1]);

    await screen.findByTestId("pending-operation");
    getShiftOperationResultMock.mockResolvedValue({
      operation_key: "op-key-fixed",
      state: "succeeded",
      receipt: {
        operation_key: "op-key-fixed",
        intent: "delete",
        shift_id: "shift-1",
        resulting_version: null,
        shift: null,
      },
    });
    await user.click(screen.getByRole("button", { name: "結果を照会" }));

    expect(getShiftOperationResultMock).toHaveBeenCalledWith("op-key-fixed", "delete");
    await waitFor(() => {
      expect(onClose).toHaveBeenCalled();
    });
  });
});
