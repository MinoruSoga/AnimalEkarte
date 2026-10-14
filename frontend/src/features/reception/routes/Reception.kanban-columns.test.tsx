import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { createTestWrapper } from "@/testing/TestUtils";
import type { ColumnData } from "@/types";
import { Reception } from "./Reception";
import { usePermission } from "@/hooks/use-permission";
const ResourceReception = "reception";
const ResourceReservations = "reservations";

/**
 * EMR-377 / VQ-001: desktop 5列カンバンで右端「会計待ち」列が見切れる。
 * 修復契約: 列は adaptive（flex-1 + 220px floor）に留め、xl 帯で 280px 固定を
 * 強制しない。溢れた分はカンバン内の横スクロールに閉じ込める。
 */

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({ canView: true, canCreate: true, canEdit: true, canDelete: true })),
}));

const fiveColumns: ColumnData[] = ["受付予約", "受付済", "診察中", "会計待ち", "会計済"].map(
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

describe("Reception creation permissions", () => {
  it.each([
    [false, false],
    [false, true],
    [true, false],
    [true, true],
  ])(
    "future-date header uses reservations permission: reception=%s reservations=%s",
    (reception, reservations) => {
      vi.mocked(usePermission).mockImplementation((resource) => ({
        canView: true,
        canCreate:
          resource === ResourceReception
            ? reception
            : resource === ResourceReservations
              ? reservations
              : true,
        canEdit: true,
        canDelete: true,
      }));
      render(<Reception />, {
        wrapper: createTestWrapper({ initialEntries: ["/reception?date=2099-01-01"] }),
      });
      expect(screen.queryByRole("button", { name: "新規予約登録" }) !== null).toBe(reservations);
    },
  );

  it.each([
    [false, false],
    [false, true],
    [true, false],
    [true, true],
  ])(
    "reception=%s reservations=%s requires both for new and add controls",
    (reception, reservations) => {
      vi.mocked(usePermission).mockImplementation((resource) => ({
        canView: true,
        canCreate:
          resource === ResourceReception
            ? reception
            : resource === ResourceReservations
              ? reservations
              : true,
        canEdit: true,
        canDelete: true,
      }));
      renderReception();
      // PR665: 当日のヘッダーボタンは walk-in 受付のため両権限で表示する。
      expect(screen.queryByRole("button", { name: "新規予約登録" }) !== null).toBe(
        reception && reservations,
      );
      // PR665: 受付予約列（通常予約）の「新規追加」は reservations:create のみで表示。
      // 両権限がなくても有効な予約作成導線を取り上げない。
      expect(screen.queryAllByRole("button", { name: "新規追加" }).length > 0).toBe(reservations);
      vi.mocked(usePermission).mockImplementation(() => ({
        canView: true,
        canCreate: true,
        canEdit: true,
        canDelete: true,
      }));
    },
  );
});

describe("Reception kanban desktop layout (EMR-377)", () => {
  it("desktop は lg:grid-cols-5 の grid で5列を表示し、強制横スクロールを持たない", () => {
    renderReception();

    const board = screen.getByRole("region", { name: /受付予約/ }).parentElement;
    expect(board).toHaveClass("grid", "lg:grid-cols-5", "overflow-y-auto");
    expect(board.className).not.toContain("lg:flex");
    expect(board.className).not.toContain("lg:overflow-x-auto");
  });

  it("5列全てが min-w-0 で grid に収まり、220px floor / flex-1 を強制しない", () => {
    renderReception();

    const columns = screen.getAllByRole("region");
    expect(columns).toHaveLength(5);
    for (const column of columns) {
      expect(column).toHaveClass("min-w-0");
      expect(column.className).not.toContain("lg:min-w-[220px]");
      expect(column.className).not.toContain("lg:flex-1");
      expect(column.className).not.toContain("xl:min-w-[280px]");
    }
  });

  it("右端の 会計済 列を含む5つの status region が全て描画される", () => {
    renderReception();

    for (const title of ["受付予約", "受付済", "診察中", "会計待ち", "会計済"]) {
      expect(screen.getByRole("region", { name: new RegExp(title) })).toBeInTheDocument();
    }
  });
});
