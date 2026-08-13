import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ComponentProps } from "react";
import { describe, expect, it, vi, beforeEach } from "vitest";

import { ShiftFormDialog } from "./ShiftFormDialog";

const createShiftMock = vi.fn();
const updateShiftMock = vi.fn();

vi.mock("../../api/get-shift-templates", () => ({
  useGetShiftTemplates: () => ({ data: [] }),
}));

vi.mock("../../api/delete-shift", () => ({
  useDeleteShift: () => ({ mutate: vi.fn(), isPending: false }),
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

function renderDialog(
  props: Partial<ComponentProps<typeof ShiftFormDialog>> = {},
) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <ShiftFormDialog
        open
        onClose={vi.fn()}
        staffId="staff-1"
        staffName="テストスタッフ"
        date="2026-08-01"
        canEdit
        {...props}
      />
    </QueryClientProvider>,
  );
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
  });

  it("全日で開始・終了が空のまま保存するとエラーを出し API を呼ばない", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.click(screen.getByRole("button", { name: "保存" }));

    expect(
      await screen.findByText("開始時刻と終了時刻を入力してください"),
    ).toBeInTheDocument();
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
  });
});
