import { fireEvent, render, renderHook, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";
import type { ReactNode } from "react";

import type { ColumnData } from "@/types";

import type { ReceptionAppointment } from "../api/types";
import { useReceptionColumnView } from "./use-reception-column-view";

/**
 * EMR-74: 受付済カラムのレコード起動時、一般予約は advanceStatus（in_consultation PATCH）
 * を発火するが、トリミング予約は発火してはならないことを固定する。
 * handleRecordOpen は公開 API でないため、KanbanColumn をモックして onRecordOpen を起動する。
 */

interface MockKanbanColumnProps {
  data: ColumnData;
  onAddClick?: () => void;
  onCardClick: (appointment: ReceptionAppointment) => void;
  onRecordOpen?: (appointment: ReceptionAppointment, columnTitle: string) => void;
}

vi.mock("../components/KanbanColumn", () => ({
  KanbanColumn: ({ data, onRecordOpen }: MockKanbanColumnProps) => (
    <button
      type="button"
      onClick={() => {
        const appointment = data.appointments[0];
        if (appointment) onRecordOpen?.(appointment, data.title);
      }}
    >
      {`open-${data.title}`}
    </button>
  ),
}));

function makeAppointment(
  overrides: Partial<ReceptionAppointment> & { id: string },
): ReceptionAppointment {
  return {
    time: "09:00",
    visitDate: "2026-06-01",
    end: new Date(2026, 5, 1, 9, 30, 0),
    ownerName: "山田",
    petType: "犬",
    petName: "ポチ",
    visitType: "再診",
    reservationType: "一般診察",
    reservationTypeId: "1",
    reservationCategory: "general",
    isDesignated: false,
    doctor: undefined,
    doctorId: "",
    petId: "10",
    ownerId: "20",
    status: "checked_in",
    notes: undefined,
    source: "manual",
    ...overrides,
  };
}

function renderColumnView({
  columns,
  canEditReservation = true,
}: {
  columns: ColumnData[];
  canEditReservation?: boolean;
}) {
  const advanceStatus = vi.fn();
  const wrapper = ({ children }: { children: ReactNode }) => (
    <MemoryRouter>{children}</MemoryRouter>
  );
  const { result } = renderHook(
    () =>
      useReceptionColumnView({
        filteredColumns: columns,
        canCreateReservation: false,
        canEditReservation,
        advanceStatus,
        onCardClick: vi.fn(),
      }),
    { wrapper },
  );
  render(<>{result.current.columnElements}</>);
  return { advanceStatus };
}

describe("useReceptionColumnView — レコード起動時のステータス遷移", () => {
  it("トリミング予約の受付済レコード起動では advanceStatus を呼ばない", () => {
    const trimming = makeAppointment({
      id: "t1",
      reservationCategory: "trimming",
      reservationType: "シャンプーコース",
    });
    const { advanceStatus } = renderColumnView({
      columns: [{ title: "受付済", appointments: [trimming] }],
    });

    fireEvent.click(screen.getByRole("button", { name: "open-受付済" }));

    expect(advanceStatus).not.toHaveBeenCalled();
  });

  it("施術中（in_consultation）のトリミング予約でも advanceStatus を呼ばない", () => {
    const trimmingInProgress = makeAppointment({
      id: "t2",
      reservationCategory: "trimming",
      reservationType: "シャンプーコース",
      status: "in_consultation",
    });
    const { advanceStatus } = renderColumnView({
      columns: [{ title: "受付済", appointments: [trimmingInProgress] }],
    });

    fireEvent.click(screen.getByRole("button", { name: "open-受付済" }));

    expect(advanceStatus).not.toHaveBeenCalled();
  });

  it("一般予約の受付済レコード起動では advanceStatus を呼ぶ", () => {
    const general = makeAppointment({ id: "g1" });
    const { advanceStatus } = renderColumnView({
      columns: [{ title: "受付済", appointments: [general] }],
    });

    fireEvent.click(screen.getByRole("button", { name: "open-受付済" }));

    expect(advanceStatus).toHaveBeenCalledWith(general);
  });

  it("edit 権限がない場合は一般予約でも advanceStatus を呼ばない", () => {
    const general = makeAppointment({ id: "g2" });
    const { advanceStatus } = renderColumnView({
      columns: [{ title: "受付済", appointments: [general] }],
      canEditReservation: false,
    });

    fireEvent.click(screen.getByRole("button", { name: "open-受付済" }));

    expect(advanceStatus).not.toHaveBeenCalled();
  });
});
