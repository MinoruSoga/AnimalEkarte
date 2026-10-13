import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { createTestWrapper } from "@/testing/TestUtils";
import type { ColumnData } from "@/types";
import { Reception } from "./Reception";

/**
 * EMR-377 / VQ-001: desktop 5列カンバンで右端「会計待ち」列が見切れる。
 * 修復契約: 列は adaptive（flex-1 + 220px floor）に留め、xl 帯で 280px 固定を
 * 強制しない。溢れた分はカンバン内の横スクロールに閉じ込める。
 */

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({ canView: true, canCreate: true, canEdit: true, canDelete: true })),
}));

const fiveColumns: ColumnData[] = ["受付予約", "受付済", "診察中", "会計待ち", "完了"].map(
  (title) => ({ title, appointments: [] }),
);

vi.mock("../hooks/use-reception-kanban", () => ({
  useReceptionKanban: vi.fn(() => ({
    columns: fiveColumns,
    filteredColumns: fiveColumns,
    isLoading: false,
    isError: false,
    isUpdatingStatus: false,
    staffs: [],
    moveCard: vi.fn(),
    advanceStatus: vi.fn(),
    cancelAppointment: vi.fn(),
    updateAppointment: vi.fn(),
    filters: {
      selectedVisitTypes: ["初診", "再診"],
      selectedDoctor: "all",
      isTrimmingOnly: false,
      setSelectedDoctor: vi.fn(),
      setIsTrimmingOnly: vi.fn(),
      toggleVisitType: vi.fn(),
    },
  })),
}));

function renderReception() {
  return render(<Reception />, { wrapper: createTestWrapper({ router: true }) });
}

describe("Reception kanban desktop layout (EMR-377)", () => {
  it("desktop flex はボード内 overflow-x-auto に横スクロールを閉じ込める", () => {
    renderReception();

    const board = screen.getByRole("region", { name: /受付予約/ }).parentElement;
    expect(board).toHaveClass("lg:flex", "lg:overflow-x-auto");
  });

  it("5列全てが adaptive（flex-1 + 220px floor）で、280px の xl 固定最小幅を強制しない", () => {
    renderReception();

    const columns = screen.getAllByRole("region");
    expect(columns).toHaveLength(5);
    for (const column of columns) {
      expect(column).toHaveClass("lg:flex-1", "lg:min-w-[220px]");
      expect(column.className).not.toContain("xl:min-w-[280px]");
    }
  });
});
