import { useCallback, useMemo, type ReactNode } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";
import { KanbanColumn } from "../components/KanbanColumn";
import type { ReceptionAppointment } from "../api/types";
import type { ColumnData } from "@/types";
import { NO_ADD_BUTTON_COLUMNS } from "../routes/reception-model";

interface UseReceptionColumnViewArgs {
  filteredColumns: ColumnData[];
  canCreateReservation: boolean;
  canEditReservation: boolean;
  advanceStatus: (appointment: ReceptionAppointment) => void;
  onCardClick: (appointment: ReceptionAppointment) => void;
}

export function useReceptionColumnView({
  filteredColumns,
  canCreateReservation,
  canEditReservation,
  advanceStatus,
  onCardClick,
}: UseReceptionColumnViewArgs): {
  columnElements: ReactNode;
  appointmentColumnTitleMap: Map<string, string>;
  goToNewReservation: (query: string) => void;
} {
  const navigate = useNavigate();

  const handleRecordOpen = useCallback(
    (appointment: ReceptionAppointment, columnTitle: string) => {
      if (columnTitle === "受付済" && canEditReservation === true) {
        advanceStatus(appointment);
      }
    },
    [advanceStatus, canEditReservation],
  );

  // 当日受付ページから新規予約作成モーダルを自動オープンする遷移ヘルパー。
  const goToNewReservation = useCallback(
    (query: string) => {
      navigate(`${paths.reservations.getHref()}?${query}`, {
        state: { from: paths.home.getHref() },
      });
    },
    [navigate],
  );

  // 受付予約ボード → 通常の新規予約（confirmed → 受付予約カラム）。
  // 受付済ボード → 受付 walk-in（checked_in → 受付済カラム、route=reception）。
  const handleAddClick = useCallback(
    (columnTitle: string) => {
      goToNewReservation(columnTitle === "受付予約" ? "newReservation=1" : "reception=1");
    },
    [goToNewReservation],
  );

  const addClickHandlers = useMemo(() => {
    const handlers = new Map<string, (() => void) | undefined>();
    // BUG-132: create 権限がない場合は「新規追加」ボタンを非表示
    if (canCreateReservation !== true) return handlers;
    for (const column of filteredColumns) {
      handlers.set(
        column.title,
        NO_ADD_BUTTON_COLUMNS.has(column.title) ? undefined : () => handleAddClick(column.title),
      );
    }
    return handlers;
  }, [filteredColumns, handleAddClick, canCreateReservation]);

  const columnElements = useMemo(
    () =>
      filteredColumns.map((column) => (
        <KanbanColumn
          key={column.title}
          data={column}
          onAddClick={addClickHandlers.get(column.title)}
          onCardClick={onCardClick}
          onRecordOpen={handleRecordOpen}
        />
      )),
    [filteredColumns, addClickHandlers, onCardClick, handleRecordOpen],
  );

  // js-set-map-lookups: レンダーパスの O(n²) find+some を O(1) Map ルックアップへ変換
  const appointmentColumnTitleMap = useMemo(() => {
    const map = new Map<string, string>();
    for (const col of filteredColumns) {
      for (const apt of col.appointments) {
        map.set(apt.id, col.title);
      }
    }
    return map;
  }, [filteredColumns]);

  return { columnElements, appointmentColumnTitleMap, goToNewReservation };
}
