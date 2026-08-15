import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { HolidaySection } from "./HolidaySection";
import type { ClosingHoliday } from "../api/holidays";

const { mockCreateMutateAsync, mockDeleteMutateAsync, mockToastSuccess } = vi.hoisted(() => ({
  mockCreateMutateAsync: vi.fn(),
  mockDeleteMutateAsync: vi.fn(),
  mockToastSuccess: vi.fn(),
}));

vi.mock("sonner", () => ({ toast: { success: mockToastSuccess, error: vi.fn() } }));

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
  });

  it("holidays が空のとき空状態メッセージを表示する", () => {
    render(<HolidaySection holidays={[]} />);
    expect(screen.getByText("個別休診日は登録されていません")).toBeInTheDocument();
  });

  it("holidays を一覧表示する（reason 未設定の場合は理由を表示しない）", () => {
    render(
      <HolidaySection
        holidays={[makeHoliday({ date: "2026-08-15", reason: "院内研修" }), makeHoliday({ id: 2, date: "2026-09-01", reason: "" })]}
      />,
    );
    expect(screen.getByText("2026-08-15")).toBeInTheDocument();
    expect(screen.getByText("院内研修")).toBeInTheDocument();
    expect(screen.getByText("2026-09-01")).toBeInTheDocument();
  });

  it("削除ボタンで deleteMutation.mutateAsync が date で呼ばれる", async () => {
    render(<HolidaySection holidays={[makeHoliday({ date: "2026-08-15" })]} />);
    fireEvent.click(
      screen.getByRole("button", { name: "2026-08-15の休診日を削除" }),
    );

    await waitFor(() => expect(mockDeleteMutateAsync).toHaveBeenCalledWith("2026-08-15"));
  });

  it("追加ボタンでフォームを表示し、キャンセルで閉じる", () => {
    render(<HolidaySection holidays={[]} />);
    fireEvent.click(screen.getByRole("button", { name: "追加" }));
    expect(screen.getByText("新しい休診日")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "キャンセル" }));
    expect(screen.queryByText("新しい休診日")).not.toBeInTheDocument();
  });

  it("追加・キャンセル・削除を44px以上とし、削除対象を名前で識別できる", () => {
    render(<HolidaySection holidays={[makeHoliday({ date: "2026-08-15" })]} />);

    const addButton = screen.getByRole("button", { name: "追加" });
    expect(addButton).toHaveClass("min-h-11");
    const deleteButton = screen.getByRole("button", {
      name: "2026-08-15の休診日を削除",
    });
    expect(deleteButton).toHaveClass("size-11");

    fireEvent.click(addButton);
    expect(screen.getByRole("button", { name: "キャンセル" })).toHaveClass("min-h-11");
  });

  it("追加フォームはmobileで1列、sm以上で2列になる", () => {
    render(<HolidaySection holidays={[]} />);
    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    const grid = screen.getByLabelText("日付").parentElement?.parentElement;
    expect(grid).toHaveClass("grid-cols-1", "sm:grid-cols-2");
    expect(grid).not.toHaveClass("grid-cols-2");
  });

  it("日付と理由を入力して送信すると createMutation が呼ばれ、送信後フォームが閉じる", async () => {
    render(<HolidaySection holidays={[]} />);
    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    fireEvent.change(screen.getByLabelText("日付"), { target: { value: "2026-10-10" } });
    fireEvent.change(screen.getByLabelText("理由・メモ"), { target: { value: "設備点検" } });
    // フォーム表示中は見出しの「追加」ボタンと送信用「追加」ボタンが両方存在するため、
    // フォーム内の送信ボタン（DOM 順で最後）を明示的に選ぶ。
    fireEvent.click(screen.getAllByRole("button", { name: "追加" }).at(-1)!);

    await waitFor(() =>
      expect(mockCreateMutateAsync).toHaveBeenCalledWith({ date: "2026-10-10", reason: "設備点検" }),
    );
    await waitFor(() => expect(screen.queryByText("新しい休診日")).not.toBeInTheDocument());
  });

  it("理由未入力の場合 reason は undefined として渡される", async () => {
    render(<HolidaySection holidays={[]} />);
    fireEvent.click(screen.getByRole("button", { name: "追加" }));

    fireEvent.change(screen.getByLabelText("日付"), { target: { value: "2026-10-10" } });
    fireEvent.click(screen.getAllByRole("button", { name: "追加" }).at(-1)!);

    await waitFor(() =>
      expect(mockCreateMutateAsync).toHaveBeenCalledWith({ date: "2026-10-10", reason: undefined }),
    );
  });
});
