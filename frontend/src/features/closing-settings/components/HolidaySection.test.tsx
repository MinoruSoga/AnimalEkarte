import { describe, it, expect, vi, beforeEach } from "vitest";
import { act, render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HolidaySection } from "./HolidaySection";
import type { ClosingHoliday } from "../api/holidays";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const { mockCreateMutateAsync, mockDeleteMutateAsync, mockToastSuccess, mockToastError } =
  vi.hoisted(() => ({
    mockCreateMutateAsync: vi.fn(),
    mockDeleteMutateAsync: vi.fn(),
    mockToastSuccess: vi.fn(),
    mockToastError: vi.fn(),
  }));

vi.mock("sonner", () => ({ toast: { success: mockToastSuccess, error: mockToastError } }));

vi.mock("../api/holidays", () => ({
  useCreateHoliday: () => ({ mutateAsync: mockCreateMutateAsync }),
  useDeleteHoliday: () => ({ mutateAsync: mockDeleteMutateAsync }),
}));

function makeHoliday(overrides: Partial<ClosingHoliday> = {}): ClosingHoliday {
  return {
    id: 1,
    clinic_id: 1,
    date: "2026-08-15",
    reason: "院内研修",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("HolidaySection", () => {
  beforeEach(() => {
    mockCreateMutateAsync.mockReset().mockResolvedValue(undefined);
    mockDeleteMutateAsync.mockReset().mockResolvedValue(undefined);
    mockToastSuccess.mockClear();
    mockToastError.mockClear();
  });

  it("holidays が空のとき空状態メッセージを表示する", () => {
    render(<HolidaySection holidays={[]} canCreate={true} canDelete={true} />);
    expect(screen.getByText("個別休診日は登録されていません")).toBeInTheDocument();
  });

  it("holidays を一覧表示する（reason 未設定の場合は理由を表示しない）", () => {
    render(
      <HolidaySection
        holidays={[
          makeHoliday({ date: "2026-08-15", reason: "院内研修" }),
          makeHoliday({ id: 2, date: "2026-09-01", reason: "" }),
        ]}
        canCreate={true}
        canDelete={true}
      />,
    );
    expect(screen.getByText("2026-08-15")).toBeInTheDocument();
    expect(screen.getByText("院内研修")).toBeInTheDocument();
    expect(screen.getByText("2026-09-01")).toBeInTheDocument();
  });

  it("削除ボタンで確認ダイアログを開き、確定で deleteMutation.mutateAsync が date で呼ばれる", async () => {
    render(
      <HolidaySection
        holidays={[makeHoliday({ date: "2026-08-15" })]}
        canCreate={true}
        canDelete={true}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "2026-08-15の休診日を削除" }));

    expect(screen.getByText("休診日を削除しますか？")).toBeInTheDocument();
    expect(mockDeleteMutateAsync).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "削除" }));

    await waitFor(() => expect(mockDeleteMutateAsync).toHaveBeenCalledWith("2026-08-15"));
  });

  it("削除確認ダイアログのキャンセルでは mutate しない", async () => {
    render(
      <HolidaySection
        holidays={[makeHoliday({ date: "2026-08-15" })]}
        canCreate={true}
        canDelete={true}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "2026-08-15の休診日を削除" }));
    fireEvent.click(screen.getByRole("button", { name: "キャンセル" }));

    await waitFor(() =>
      expect(screen.queryByText("休診日を削除しますか？")).not.toBeInTheDocument(),
    );
    expect(mockDeleteMutateAsync).not.toHaveBeenCalled();
  });

  it("行追加コントロールはマスター共通の「新規登録」ラベルを表示する", () => {
    render(<HolidaySection holidays={[]} canCreate={true} canDelete={true} />);
    expect(screen.getByRole("button", { name: "新規登録" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "追加" })).not.toBeInTheDocument();
  });

  it("canCreate=false のとき新規登録ボタンを表示しない", () => {
    render(<HolidaySection holidays={[]} canCreate={false} canDelete={false} />);
    expect(screen.queryByRole("button", { name: "新規登録" })).not.toBeInTheDocument();
  });

  it("追加ボタンでフォームを表示し、キャンセルで閉じる", () => {
    render(<HolidaySection holidays={[]} canCreate={true} canDelete={true} />);
    fireEvent.click(screen.getByRole("button", { name: "新規登録" }));
    expect(screen.getByText("新しい休診日")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "キャンセル" }));
    expect(screen.queryByText("新しい休診日")).not.toBeInTheDocument();
  });

  it("日付を入れてキャンセル→再オープンすると日付入力と hidden 値がリセットされる", async () => {
    const user = userEvent.setup();
    render(<HolidaySection holidays={[]} canCreate={true} canDelete={true} />);

    await user.click(screen.getByRole("button", { name: "新規登録" }));
    const input = screen.getByRole<HTMLInputElement>("textbox", { name: "日付" });
    await user.click(input);
    await user.type(input, "2026/10/20");
    await user.tab();
    expect(input.value).not.toBe("");

    await user.click(screen.getByRole("button", { name: "キャンセル" }));
    await user.click(screen.getByRole("button", { name: "新規登録" }));

    const reopened = screen.getByRole<HTMLInputElement>("textbox", { name: "日付" });
    expect(reopened.value).toBe("");
    const hidden = document.querySelector<HTMLInputElement>('input[name="date"]');
    expect(hidden?.value ?? "").toBe("");
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockDeleteMutateAsync).not.toHaveBeenCalled();
  });

  it("追加・キャンセル・削除を44px以上とし、削除対象を名前で識別できる", () => {
    render(
      <HolidaySection
        holidays={[makeHoliday({ date: "2026-08-15" })]}
        canCreate={true}
        canDelete={true}
      />,
    );

    const addButton = screen.getByRole("button", { name: "新規登録" });
    expect(addButton).toHaveClass("min-h-11");
    const deleteButton = screen.getByRole("button", {
      name: "2026-08-15の休診日を削除",
    });
    expect(deleteButton).toHaveClass("size-11");

    fireEvent.click(addButton);
    expect(screen.getByRole("button", { name: "キャンセル" })).toHaveClass("min-h-11");
  });

  it("追加フォームはmobileで1列、sm以上で2列になる", () => {
    render(<HolidaySection holidays={[]} canCreate={true} canDelete={true} />);
    fireEvent.click(screen.getByRole("button", { name: "新規登録" }));

    const grid = screen.getByRole("textbox", { name: "日付" }).parentElement?.parentElement
      ?.parentElement;
    expect(grid).toHaveClass("grid-cols-1", "sm:grid-cols-2");
    expect(grid).not.toHaveClass("grid-cols-2");
  });

  it("日付と理由を入力して送信すると createMutation が呼ばれ、送信後フォームが閉じる", async () => {
    const user = userEvent.setup();
    render(<HolidaySection holidays={[]} canCreate={true} canDelete={true} />);
    await user.click(screen.getByRole("button", { name: "新規登録" }));

    const dateInput = screen.getByRole("textbox", { name: "日付" });
    await user.type(dateInput, "2026/10/10");
    await user.tab();
    await user.type(screen.getByLabelText("理由・メモ"), "設備点検");
    // 行追加は「新規登録」、フォーム内送信は「追加」のまま。
    await user.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() =>
      expect(mockCreateMutateAsync).toHaveBeenCalledWith({
        date: "2026-10-10",
        reason: "設備点検",
      }),
    );
    await waitFor(() => expect(screen.queryByText("新しい休診日")).not.toBeInTheDocument());
  });

  it("理由未入力の場合 reason は undefined として渡される", async () => {
    const user = userEvent.setup();
    render(<HolidaySection holidays={[]} canCreate={true} canDelete={true} />);
    await user.click(screen.getByRole("button", { name: "新規登録" }));

    const dateInput = screen.getByRole("textbox", { name: "日付" });
    await user.type(dateInput, "2026/10/10");
    await user.tab();
    await user.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() =>
      expect(mockCreateMutateAsync).toHaveBeenCalledWith({ date: "2026-10-10", reason: undefined }),
    );
  });

  it("canCreate=false のとき formAction は mutate せず toast.error する", async () => {
    const user = userEvent.setup();
    const { rerender } = render(<HolidaySection holidays={[]} canCreate={true} canDelete={true} />);
    await user.click(screen.getByRole("button", { name: "新規登録" }));
    const dateInput = screen.getByRole("textbox", { name: "日付" });
    await user.type(dateInput, "2026/10/10");
    await user.tab();

    const form = dateInput.closest("form");
    expect(form).not.toBeNull();

    rerender(<HolidaySection holidays={[]} canCreate={false} canDelete={true} />);

    await act(async () => {
      form?.requestSubmit();
    });

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(screen.getByText("新しい休診日")).toBeInTheDocument();
  });

  it("canDelete=false のとき delete は mutate せず toast.error する", async () => {
    render(
      <HolidaySection
        holidays={[makeHoliday({ date: "2026-08-15" })]}
        canCreate={false}
        canDelete={false}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "2026-08-15の休診日を削除" }));
    fireEvent.click(screen.getByRole("button", { name: "削除" }));

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(mockDeleteMutateAsync).not.toHaveBeenCalled();
    expect(mockToastSuccess).not.toHaveBeenCalled();
  });

  // EMR-319: 日付必須が空で invalid になると赤枠+エラー文+aria-invalid が出る。
  it("日付が空のまま invalid になると aria-invalid とエラー文が出て、入力で解除される", async () => {
    render(<HolidaySection holidays={[]} canCreate={true} canDelete={true} />);
    fireEvent.click(screen.getByRole("button", { name: "新規登録" }));

    const dateInput = screen.getByRole("textbox", { name: "日付" });
    fireEvent.invalid(dateInput);

    expect(dateInput).toHaveAttribute("aria-invalid", "true");
    expect(dateInput).toHaveAttribute("aria-describedby", "holiday-date-error");
    expect(document.getElementById("holiday-date-error")).toHaveTextContent(
      "日付を入力してください",
    );

    const user = userEvent.setup();
    await user.type(dateInput, "2026/10/10");
    await user.tab();

    expect(dateInput).not.toHaveAttribute("aria-invalid");
    expect(document.getElementById("holiday-date-error")).toBeNull();
  });
});
