import { fireEvent, render, renderHook, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ReactNode } from "react";

import type { ColumnData } from "@/types";

import type { ReceptionAppointment } from "../api/types";
import { useReceptionColumnView } from "./use-reception-column-view";

/**
 * EMR-74: 受付済カラムのレコード起動時、一般予約は advanceStatus（in_consultation PATCH）
 * を発火するが、トリミング予約は発火してはならないことを固定する。
 * handleRecordOpen は公開 API でないため、KanbanColumn をモックして onRecordOpen を起動する。
 */

// EMR-243: goToNewReservation の navigate 先クエリと from 保持を固定するため
// useNavigate を spy に差し替える（MemoryRouter は他の router 依存用に残す）。
const { navigateMock } = vi.hoisted(() => ({ navigateMock: vi.fn() }));

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return { ...actual, useNavigate: () => navigateMock };
});

interface MockKanbanColumnProps {
  data: ColumnData;
  onAddClick?: () => void;
  onCardClick: (appointment: ReceptionAppointment) => void;
  onRecordOpen?: (appointment: ReceptionAppointment, columnTitle: string) => void;
}

vi.mock("../components/KanbanColumn", () => ({
  KanbanColumn: ({ data, onAddClick, onRecordOpen }: MockKanbanColumnProps) => (
    <>
      <button
        type="button"
        onClick={() => {
          const appointment = data.appointments[0];
          if (appointment) onRecordOpen?.(appointment, data.title);
        }}
      >
        {`open-${data.title}`}
      </button>
      {onAddClick ? (
        <button type="button" onClick={onAddClick}>
          {`add-${data.title}`}
        </button>
      ) : null}
    </>
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
  canCreateReservation = false,
  selectedDate = "2026-06-01",
  isToday = true,
}: {
  columns: ColumnData[];
  canEditReservation?: boolean;
  canCreateReservation?: boolean;
  selectedDate?: string;
  isToday?: boolean;
}) {
  const advanceStatus = vi.fn();
  const wrapper = ({ children }: { children: ReactNode }) => (
    <MemoryRouter>{children}</MemoryRouter>
  );
  const { result } = renderHook(
    () =>
      useReceptionColumnView({
        filteredColumns: columns,
        canCreateReservation,
        canEditReservation,
        advanceStatus,
        onCardClick: vi.fn(),
        selectedDate,
        isToday,
      }),
    { wrapper },
  );
  render(<>{result.current.columnElements}</>);
  return { advanceStatus, result };
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

/**
 * EMR-243: `?date=` 表示日に応じた予約作成遷移と「＋」ボタン抑制を固定する。
 * 本日: reception=1 / newReservation=1 を従来どおり発行し、from="/" で戻る。
 * 非本日: `&date=<選択日>` を付与し from="/?date=<選択日>" で戻る。
 *        checked_in（当日受付）を作る reception=1 系の「＋」は出さない。
 */
describe("useReceptionColumnView — 日付切替 (EMR-243)", () => {
  beforeEach(() => {
    navigateMock.mockClear();
  });

  const twoColumns: ColumnData[] = [
    { title: "受付予約", appointments: [] },
    { title: "受付済", appointments: [] },
    { title: "診療中", appointments: [] },
  ];

  it("本日表示ではクエリをそのまま navigate し from は /", () => {
    const { result } = renderColumnView({ columns: twoColumns, isToday: true });

    result.current.goToNewReservation("reception=1");

    expect(navigateMock).toHaveBeenCalledWith("/reservations?reception=1", {
      state: { from: "/" },
    });
  });

  it("非本日表示ではクエリに date を付与し from は /?date= を指す", () => {
    const { result } = renderColumnView({
      columns: twoColumns,
      selectedDate: "2026-06-05",
      isToday: false,
    });

    result.current.goToNewReservation("newReservation=1");

    expect(navigateMock).toHaveBeenCalledWith("/reservations?newReservation=1&date=2026-06-05", {
      state: { from: "/?date=2026-06-05" },
    });
  });

  it("本日表示では 受付予約/受付済 の両方に「＋」を出し、受付済は reception=1 を発行する", () => {
    renderColumnView({ columns: twoColumns, canCreateReservation: true, isToday: true });

    fireEvent.click(screen.getByRole("button", { name: "add-受付済" }));
    expect(navigateMock).toHaveBeenCalledWith("/reservations?reception=1", {
      state: { from: "/" },
    });

    fireEvent.click(screen.getByRole("button", { name: "add-受付予約" }));
    expect(navigateMock).toHaveBeenCalledWith("/reservations?newReservation=1", {
      state: { from: "/" },
    });
  });

  it("非本日では 受付済（checked_in 起点）の「＋」を出さず、受付予約は date 付きで残す", () => {
    renderColumnView({
      columns: twoColumns,
      canCreateReservation: true,
      selectedDate: "2026-06-05",
      isToday: false,
    });

    expect(screen.queryByRole("button", { name: "add-受付済" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "add-診療中" })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "add-受付予約" }));
    expect(navigateMock).toHaveBeenCalledWith("/reservations?newReservation=1&date=2026-06-05", {
      state: { from: "/?date=2026-06-05" },
    });
  });
});
