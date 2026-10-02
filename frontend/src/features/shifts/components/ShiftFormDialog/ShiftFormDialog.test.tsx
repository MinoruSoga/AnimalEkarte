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
const getShiftTemplatesMock = vi.fn((): { data: ShiftTemplate[] } => ({ data: [] }));

vi.mock("../../api/get-shift-templates", () => ({
  useGetShiftTemplates: () => getShiftTemplatesMock(),
}));

vi.mock("../../api/delete-shift", () => ({
  useDeleteShift: () => ({ mutate: deleteShiftMock, isPending: false }),
}));

vi.mock("../../api/create-shift", () => ({
  createShift: (...args: unknown[]) => createShiftMock(...args),
}));

vi.mock("../../api/update-shift", () => ({
  updateShift: (...args: unknown[]) => updateShiftMock(...args),
}));

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
  start_time: "09:00",
  end_time: "18:00",
  notes: "",
  breaks: [],
  created_at: "",
  updated_at: "",
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
    createShiftMock.mockReset();
    updateShiftMock.mockReset();
    deleteShiftMock.mockReset();
    toastError.mockReset();
  });

  it("全日で開始・終了が空のまま保存するとエラーを出し API を呼ばない", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("開始時刻と終了時刻を入力してください")).toBeInTheDocument();
    expect(createShiftMock).not.toHaveBeenCalled();
  });

  it("全日で時刻を入れて保存すると createShift が呼ばれる", async () => {
    const user = userEvent.setup();
    createShiftMock.mockResolvedValue({});
    renderDialog();

    await user.type(screen.getByLabelText("開始時刻"), "09:00");
    await user.type(screen.getByLabelText("終了時刻"), "18:00");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(createShiftMock).toHaveBeenCalledTimes(1);
    });
    expect(createShiftMock.mock.calls[0][0]).toMatchObject({
      staff_id: "staff-1",
      date: "2026-08-01",
      shift_type: "full",
      start_time: "09:00",
      end_time: "18:00",
    });
    expect(toastError).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });
});

describe("ShiftFormDialog mutation permission re-check", () => {
  beforeEach(() => {
    createShiftMock.mockReset().mockResolvedValue({});
    updateShiftMock.mockReset().mockResolvedValue({});
    deleteShiftMock.mockReset();
    toastError.mockReset();
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

  it("canDelete=true なら削除確認後に deleteShift を呼ぶ", async () => {
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

    expect(deleteShiftMock).toHaveBeenCalledWith("shift-1", expect.any(Object));
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

// EMR-241: カテゴリ名による時刻フィールド非表示を廃止。
// 全シフト種別で時刻入力を表示・有効にし、off/paid_leave は両方空で保存可。
// 片方だけの時刻入力は全カテゴリで拒否。空時刻テンプレートを勤務種別に適用しても
// 黙って保存せずバリデーションエラーを表示する。
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

describe("ShiftFormDialog EMR-241 all-category time fields", () => {
  beforeEach(() => {
    createShiftMock.mockReset().mockResolvedValue({});
    updateShiftMock.mockReset().mockResolvedValue({});
    deleteShiftMock.mockReset();
    toastError.mockReset();
    getShiftTemplatesMock.mockReset().mockReturnValue({ data: [] });
  });

  it("休日(off)選択時も開始/終了時刻が有効で休憩セクションが表示される", async () => {
    const user = userEvent.setup();
    renderDialog();

    await selectOption(user, "シフト種別", "休日");

    expect(screen.getByLabelText("開始時刻")).toBeEnabled();
    expect(screen.getByLabelText("終了時刻")).toBeEnabled();
    expect(screen.getByText("休憩時間")).toBeInTheDocument();
  });

  it("off で両方空の時刻のまま保存すると createShift が呼ばれる", async () => {
    const user = userEvent.setup();
    renderDialog();

    await selectOption(user, "シフト種別", "休日");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(createShiftMock).toHaveBeenCalledTimes(1);
    });
    const input = createShiftMock.mock.calls[0][0] as {
      shift_type: string;
      start_time?: string;
      end_time?: string;
    };
    expect(input.shift_type).toBe("off");
    expect(input.start_time).toBeUndefined();
    expect(input.end_time).toBeUndefined();
  });

  it("paid_leave で両方空の時刻のまま保存すると createShift が呼ばれる", async () => {
    const user = userEvent.setup();
    renderDialog();

    await selectOption(user, "シフト種別", "有休");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(createShiftMock).toHaveBeenCalledTimes(1);
    });
    expect(createShiftMock.mock.calls[0][0]).toMatchObject({ shift_type: "paid_leave" });
  });

  it("off に入力した時刻を捨てずに送信する", async () => {
    const user = userEvent.setup();
    renderDialog();

    await selectOption(user, "シフト種別", "休日");
    await user.type(screen.getByLabelText("開始時刻"), "10:00");
    await user.type(screen.getByLabelText("終了時刻"), "15:00");
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(createShiftMock).toHaveBeenCalledTimes(1);
    });
    expect(createShiftMock.mock.calls[0][0]).toMatchObject({
      shift_type: "off",
      start_time: "10:00",
      end_time: "15:00",
    });
  });

  it("full で開始時刻だけ入力して保存するとエラーで API を呼ばない", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.type(screen.getByLabelText("開始時刻"), "09:00");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("開始時刻と終了時刻を入力してください")).toBeInTheDocument();
    expect(createShiftMock).not.toHaveBeenCalled();
  });

  it("off で終了時刻だけ入力して保存するとエラーで API を呼ばない", async () => {
    const user = userEvent.setup();
    renderDialog();

    await selectOption(user, "シフト種別", "休日");
    await user.type(screen.getByLabelText("終了時刻"), "18:00");
    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByText("開始時刻と終了時刻を入力してください")).toBeInTheDocument();
    expect(createShiftMock).not.toHaveBeenCalled();
  });

  it('編集で勤務種別を off に変え時刻を消すと "" でクリアを送る', async () => {
    const user = userEvent.setup();
    renderDialog({ editShift: EDIT_SHIFT });

    await selectOption(user, "シフト種別", "休日");
    await user.clear(screen.getByLabelText("開始時刻"));
    await user.clear(screen.getByLabelText("終了時刻"));
    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(updateShiftMock).toHaveBeenCalledTimes(1);
    });
    expect(updateShiftMock.mock.calls[0][1]).toMatchObject({
      shift_type: "off",
      start_time: "",
      end_time: "",
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
