import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { ShiftFormDialog } from "./ShiftFormDialog";

vi.mock("../../api/get-shift-templates", () => ({
  useGetShiftTemplates: () => ({ data: [] }),
}));

vi.mock("../../api/delete-shift", () => ({
  useDeleteShift: () => ({ mutate: vi.fn(), isPending: false }),
}));

describe("ShiftFormDialog responsive layout", () => {
  it("開始・終了時刻はmobileで1列、sm以上で2列になる", () => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <ShiftFormDialog
          open
          onClose={vi.fn()}
          staffId="staff-1"
          staffName="テストスタッフ"
          date="2026-07-21"
        />
      </QueryClientProvider>,
    );

    const grid = screen.getByLabelText("開始時刻").parentElement?.parentElement;
    expect(grid).toHaveClass("grid-cols-1", "sm:grid-cols-2");
    expect(grid).not.toHaveClass("grid-cols-2");
  });
});
