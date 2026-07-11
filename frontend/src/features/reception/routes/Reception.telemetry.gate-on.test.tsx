import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { createTestWrapper } from "@/testing/utils";
import type { ColumnData } from "@/types";
import { Reception } from "./Reception";

/**
 * RECEPTION_TELEMETRY_PHASE2_ENABLED を true に切り替えた将来のフォローアップコミット後の
 * 挙動をシミュレートする。実体（useReceptionTelemetry の計算ロジック）は本物のまま、
 * ゲート定数だけを差し替える — Reception.tsx の配線が「ゲートが true なら waitStats を渡す」
 * ことを保証する。ReceptionTelemetryStrip 自体の Phase1/Phase2 分岐ロジックは
 * ReceptionTelemetryStrip.test.tsx で既に検証済みのため、ここでは再検証しない。
 */

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn(() => ({ canView: true, canCreate: true, canEdit: true, canDelete: true })),
}));

vi.mock("../hooks/use-reception-telemetry", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../hooks/use-reception-telemetry")>();
  return { ...actual, RECEPTION_TELEMETRY_PHASE2_ENABLED: true };
});

const columnsFixture: ColumnData[] = [
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

describe("Reception — テレメトリ配線（Phase 2 ゲート ON、フォローアップ後の想定挙動）", () => {
  it("ゲートが true になると平均待ち/最長待ちが実データで表示される", () => {
    renderReception();

    expect(screen.getByText("本日受付")).toBeInTheDocument();
    expect(screen.getByText("平均待ち")).toBeInTheDocument();
    expect(screen.getByText("最長待ち")).toBeInTheDocument();
    expect(screen.getByText("32分")).toBeInTheDocument();
    expect(screen.getByText("32分 — ミルク")).toBeInTheDocument();
  });
});
