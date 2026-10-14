import { useCallback, useMemo, type ReactNode } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";
import { KanbanColumn } from "../components/KanbanColumn";
import type { ReceptionAppointment } from "../api/types";
import type { ColumnData } from "@/types";
import { canAddColumnEntryOnDate, receptionColumnAddKind } from "../routes/reception-model";

interface UseReceptionColumnViewArgs {
  filteredColumns: ColumnData[];
  /** 通常予約（reservations:create）。受付予約列の「＋」はこれだけで出る。 */
  canCreateReservation: boolean;
  /**
   * PR665: 当日 walk-in 受付（reception=1 / checked_in）は予約権限に加えて
   * reception:create を要求する。通常予約の導線は狭めない。
   */
  canCreateReception: boolean;
  canEditReservation: boolean;
  advanceStatus: (appointment: ReceptionAppointment) => void;
  onCardClick: (appointment: ReceptionAppointment) => void;
  /** EMR-243: ボード表示中の日付（JST "YYYY-MM-DD"、検証済み）。 */
  selectedDate: string;
  /** selectedDate が当日なら true。非本日では checked_in 起点を出さない。 */
  isToday: boolean;
}

export function useReceptionColumnView({
  filteredColumns,
  canCreateReservation,
  canCreateReception = true,
  canEditReservation,
  advanceStatus,
  onCardClick,
  selectedDate,
  isToday,
}: UseReceptionColumnViewArgs): {
  columnElements: ReactNode;
  appointmentColumnTitleMap: Map<string, string>;
  goToNewReservation: (query: string) => void;
} {
  const navigate = useNavigate();

  const handleRecordOpen = useCallback(
    (appointment: ReceptionAppointment, columnTitle: string) => {
      // EMR-74: トリミング予約のレコード起動では汎用の in_consultation 遷移
      // （advanceStatus → PATCH /reservations）を発火しない。trimming の「施術中」は
      // trimming 記録側が intent 経路で管理する。
      if (
        columnTitle === "受付済" &&
        canEditReservation === true &&
        appointment.reservationCategory !== "trimming"
      ) {
        advanceStatus(appointment);
      }
    },
    [advanceStatus, canEditReservation],
  );

  // 当日受付ページから新規予約作成モーダルを自動オープンする遷移ヘルパー。
  // EMR-243: 非本日表示では予約作成クエリへ選択日を付与し、戻り先にも ?date= を保持する。
  const goToNewReservation = useCallback(
    (query: string) => {
      // PR665: walk-in (reception=1) は両権限、通常予約 (newReservation=1) は予約権限のみ。
      const isWalkIn = query.startsWith("reception=");
      const allowed = isWalkIn
        ? canCreateReservation === true && canCreateReception === true
        : canCreateReservation === true;
      if (!allowed) return;
      const datedQuery = isToday ? query : `${query}&date=${selectedDate}`;
      navigate(`${paths.reservations.getHref()}?${datedQuery}`, {
        state: {
          from: isToday ? paths.home.getHref() : `${paths.home.getHref()}?date=${selectedDate}`,
        },
      });
    },
    [navigate, isToday, selectedDate, canCreateReservation, canCreateReception],
  );

  // 受付予約ボード → 通常の新規予約（confirmed → 受付予約カラム）。
  // 受付済ボード → 受付 walk-in（checked_in → 受付済カラム、route=reception）。
  // 非本日では reception 系は canAddColumnEntryOnDate でボタン自体を出さない。
  const handleAddClick = useCallback(
    (columnTitle: string) => {
      goToNewReservation(`${receptionColumnAddKind(columnTitle)}=1`);
    },
    [goToNewReservation],
  );

  const addClickHandlers = useMemo(() => {
    const handlers = new Map<string, (() => void) | undefined>();
    for (const column of filteredColumns) {
      if (!canAddColumnEntryOnDate(column.title, isToday)) {
        handlers.set(column.title, undefined);
        continue;
      }
      // PR665: 列種別で必要権限が変わる。受付予約=通常予約、その他=当日 walk-in。
      const isWalkInColumn = receptionColumnAddKind(column.title) === "reception";
      const allowed = isWalkInColumn
        ? canCreateReservation === true && canCreateReception === true
        : canCreateReservation === true;
      handlers.set(column.title, allowed ? () => handleAddClick(column.title) : undefined);
    }
    return handlers;
  }, [filteredColumns, handleAddClick, canCreateReservation, canCreateReception, isToday]);

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
