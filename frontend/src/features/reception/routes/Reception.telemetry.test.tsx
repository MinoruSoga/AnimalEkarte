import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { createTestWrapper } from "@/testing/utils";
import type { ColumnData } from "@/types";
import { Reception } from "./Reception";

/**
 * Reception.tsx → useReceptionTelemetry → ReceptionTelemetryStrip の実配線を通した
 * 常時有効（Phase 2）の配線回帰テスト。
 *
 * ReceptionTelemetryStrip 単体テストだけでは「本物の配線経由で waitStats が渡る」ことは
 * 証明できないため、Reception ルート経由で件数・待ち時間が表示されることを検証する。
 * Strip 内部の Phase1/Phase2 prop 分岐は ReceptionTelemetryStrip.test.tsx で既に
 * カバー済みのため、ここでは再検証しない。
 */

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({ canView: true, canCreate: true, canEdit: true, canDelete: true })),
}));

const columnsFixture: ColumnData[] = [
  {
    title: "受付予約",
    appointments: [
      {
        id: "1",
        time: "10:00",
        visitDate: "2026-07-05",
        ownerName: "山田",
        petType: "犬",
        petName: "ポチ",
        visitType: "初診",
        reservationType: "一般診療",
        reservationTypeId: "1",
        reservationCategory: "general",
        isDesignated: false,
        doctorId: "",
        petId: "10",
        ownerId: "20",
        status: "pending",
        source: "manual",
        checkedInAt: undefined,
      },
    ],
  },
  {
    title: "受付済",
    appointments: [
      {
        id: "2",
        time: "09:30",
        visitDate: "2026-07-05",
        ownerName: "鈴木",
        petType: "猫",
        petName: "ミルク",
        visitType: "再診",
        reservationType: "一般診療",
        reservationTypeId: "1",
        reservationCategory: "general",
        isDesignated: false,
        doctorId: "",
        petId: "11",
        ownerId: "21",
        status: "checked_in",
        source: "manual",
        checkedInAt: new Date(Date.now() - 32 * 60_000).toISOString(),
      },
    ],
  },
];

vi.mock("../hooks/use-reception-kanban", () => ({
  useReceptionKanban: vi.fn(() => ({
    columns: columnsFixture,
    filteredColumns: columnsFixture,
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

describe("Reception — テレメトリ配線（Phase 2 常時有効）", () => {
  it("本日受付件数と平均待ち/最長待ちが実データで表示される", () => {
    renderReception();

    expect(screen.getByText("本日受付")).toBeInTheDocument();
    // 全カラム合算 = 2件（受付予約1 + 受付済1）
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("平均待ち")).toBeInTheDocument();
    expect(screen.getByText("最長待ち")).toBeInTheDocument();
    expect(screen.getByText("32分")).toBeInTheDocument();
    expect(screen.getByText("32分 — ミルク")).toBeInTheDocument();
  });

  it("受付カンバンは mobile で単一列、sm 以上で複数列にする", () => {
    renderReception();

    const kanbanGrid = screen.getByRole("region", { name: "受付予約 — 1件" }).parentElement;
    expect(kanbanGrid).toHaveClass("grid-cols-1", "sm:grid-cols-2");
  });
});
