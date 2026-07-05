import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import type { ColumnData } from "@/types";
import { Reception } from "./Reception";

/**
 * Reception.tsx の実配線を通した回帰テスト（ReceptionTelemetryStrip 単体テストだけでは
 * 「Phase 1 = 件数のみ」が本物の配線経由でも成立することを証明できないためのギャップ埋め）。
 *
 * 「受付済 0 件」と「Phase 2 未稼働」を区別するため、fixture には checkedInAt 付きの
 * 受付済患者を意図的に含める。それでも平均待ち/最長待ちが表示されないことを示せて初めて、
 * ゲート（RECEPTION_TELEMETRY_PHASE2_ENABLED）が実データの有無ではなくフラグで
 * 判定していることの証明になる。
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
        // 実データが存在していても(=待ち時間を計算できる状態でも)、ゲートが false の間は
        // 表示してはならないことを示すため、意図的に checkedInAt を設定する。
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
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <Reception />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("Reception — テレメトリ配線（Phase 2 ゲート OFF、実際の現行状態）", () => {
  it("checked_in_at を持つ受付済患者が実在しても、ゲートが false の間は本日受付件数のみ表示する", () => {
    renderReception();

    expect(screen.getByText("本日受付")).toBeInTheDocument();
    // 全カラム合算 = 2件（受付予約1 + 受付済1）
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.queryByText("平均待ち")).not.toBeInTheDocument();
    expect(screen.queryByText("最長待ち")).not.toBeInTheDocument();
    expect(screen.queryByText(/32分/)).not.toBeInTheDocument();
  });
});
