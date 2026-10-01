// React/Framework
import { useState, useMemo, useCallback } from "react";

// External
import { pointerWithin, KeyboardSensor, PointerSensor, useSensor, useSensors } from "@dnd-kit/core";
import { sortableKeyboardCoordinates } from "@dnd-kit/sortable";
import { useSearchParams } from "react-router";
import { formatDateWithWeekday } from "@/lib/format/date";

// Internal
import { buildJSTWallDateTime, todayJSTISO } from "@/lib/jst-date";
import {
  ResourceReservations,
  ResourceMedicalRecords,
  ResourceAccounting,
  ResourceHospitalization,
} from "@/types/generated/models";
import { usePermission } from "@/hooks/use-permission";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";

import { useReceptionKanban } from "../hooks/use-reception-kanban";
import { useReceptionTelemetry } from "../hooks/use-reception-telemetry";
import { useReceptionDragHandlers } from "../hooks/use-reception-drag-handlers";
import { useReceptionModalHandlers } from "../hooks/use-reception-modal-handlers";
import { ReceptionPageBody } from "./ReceptionPagePanels";
import { useReceptionColumnView } from "../hooks/use-reception-column-view";
import { resolveReceptionDateParam } from "./reception-model";

export function Reception() {
  // EMR-243: `?date=YYYY-MM-DD` で表示日を切替える。未指定/不正は当日 JST。
  // todayJSTISO() は毎レンダー評価する（30秒ポーリング/60秒テレメトリtickで
  // JST 日跨ぎにも追従）。明示指定された date は日跨ぎしてもピン留めされる。
  const [searchParams, setSearchParams] = useSearchParams();
  const dateParam = searchParams.get("date");
  const today = todayJSTISO();
  const selectedDate = resolveReceptionDateParam(dateParam, today);
  const isToday = selectedDate === today;
  const handleDateChange = useCallback(
    (value: string) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        // 空文字クリアは param 削除で当日へフォールバックさせる
        if (value) {
          next.set("date", value);
        } else {
          next.delete("date");
        }
        return next;
      });
    },
    [setSearchParams],
  );
  const {
    canCreate: canCreateReservation,
    canEdit: canEditReservation,
    canDelete: canDeleteReservation,
  } = usePermission(ResourceReservations);
  const { canCreate: canCreateMedicalRecord } = usePermission(ResourceMedicalRecords);
  const { canCreate: canCreateAccounting } = usePermission(ResourceAccounting);
  const { canCreate: canCreateHospitalization } = usePermission(ResourceHospitalization);
  const {
    columns,
    filteredColumns,
    isUpdatingStatus,
    isLoading,
    isError,
    staffs,
    moveCard,
    advanceStatus,
    cancelAppointment,
    updateAppointment,
    filters,
  } = useReceptionKanban({
    canEditReservation,
    canDeleteReservation,
    date: selectedDate,
  });

  // スタッフAPIから医師フィルター選択肢を動的生成
  const doctors = useMemo(
    () => [
      { id: "all", name: "全て" },
      ...staffs.flatMap((s) => (s.isActive ? [{ id: s.name, name: s.name }] : [])),
      { id: "医師指名なし", name: "医師指名なし" },
    ],
    [staffs],
  );

  const {
    cancelConfirmOpen,
    cancelTarget,
    closeCancelConfirm,
    closeEditModal,
    editingAppointment,
    executeCancel,
    handleAdvanceStatus,
    handleCancelAppointment,
    handleCardClick,
    handleEditAppointment,
    handleEditSave,
    isEditModalOpen,
    modalOpen,
    selectedAppointment,
    setModalOpen,
  } = useReceptionModalHandlers({
    advanceStatus,
    cancelAppointment,
    updateAppointment,
    canEditReservation,
    canDeleteReservation,
  });

  const [isFilterOpen, setIsFilterOpen] = useState(false);
  // accessibility-rules.md §4: PointerSensor だけを指定すると dnd-kit の既定センサー
  // （KeyboardSensor 含む）が失われるため、キーボード操作用に明示的に追加する。
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );
  const { handleDragEnd } = useReceptionDragHandlers(columns, moveCard);

  // 受付ヘッダー テレメトリ（change-ui.md）: 集計は必ず columns（フィルタ非適用）から算出する。
  // filteredColumns を渡すと「本日受付」件数がフィルタ操作で変動してしまう。
  const telemetry = useReceptionTelemetry(columns);
  // 表示日ラベル: 当日だけでなく `?date=` で選んだ暦日も同じ書式で出す。
  // "YYYY-MM-DD" は buildJSTWallDateTime 経由でローカル壁時計に置く（UTC 深夜罠を避ける）。
  const dateLabel = formatDateWithWeekday(buildJSTWallDateTime(selectedDate, "00:00"));

  const { columnElements, appointmentColumnTitleMap, goToNewReservation } = useReceptionColumnView({
    filteredColumns,
    canCreateReservation,
    canEditReservation,
    advanceStatus,
    onCardClick: handleCardClick,
    selectedDate,
    isToday,
  });

  if (isLoading) {
    return <LoadingFallback />;
  }

  if (isError) {
    return <ErrorFallback message="受付データの取得に失敗しました" />;
  }

  return (
    <ReceptionPageBody
      dateLabel={dateLabel}
      selectedDate={selectedDate}
      isToday={isToday}
      onDateChange={handleDateChange}
      isFilterOpen={isFilterOpen}
      onToggleFilter={() => setIsFilterOpen((prev) => !prev)}
      canCreateReservation={canCreateReservation}
      onNewReception={() =>
        // 本日: 従来通り当日受付（checked_in）。非本日: 選択日の通常予約作成。
        goToNewReservation(isToday ? "reception=1" : "newReservation=1")
      }
      telemetry={telemetry}
      selectedVisitTypes={filters.selectedVisitTypes}
      selectedDoctor={filters.selectedDoctor}
      doctors={doctors}
      isTrimmingOnly={filters.isTrimmingOnly}
      onToggleVisitType={filters.toggleVisitType}
      onSelectedDoctorChange={filters.setSelectedDoctor}
      onTrimmingOnlyChange={filters.setIsTrimmingOnly}
      sensors={sensors}
      collisionDetection={pointerWithin}
      onDragEnd={handleDragEnd}
      isUpdatingStatus={isUpdatingStatus}
      columnElements={columnElements}
      modalOpen={modalOpen}
      onCloseModal={() => setModalOpen(false)}
      onConfirm={canEditReservation === true ? handleAdvanceStatus : undefined}
      onEdit={canEditReservation === true ? handleEditAppointment : undefined}
      onCancel={canDeleteReservation === true ? handleCancelAppointment : undefined}
      selectedAppointment={selectedAppointment}
      currentStatus={
        selectedAppointment ? appointmentColumnTitleMap.get(selectedAppointment.id) : undefined
      }
      canCreateMedicalRecord={canCreateMedicalRecord}
      canCreateAccounting={canCreateAccounting}
      canCreateHospitalization={canCreateHospitalization}
      isEditModalOpen={isEditModalOpen}
      onCloseEditModal={closeEditModal}
      onEditSave={handleEditSave}
      editingAppointment={editingAppointment}
      canEditReservation={canEditReservation}
      cancelConfirmOpen={cancelConfirmOpen}
      onCloseCancelConfirm={closeCancelConfirm}
      onConfirmCancel={executeCancel}
      cancelDescription={
        cancelTarget
          ? `${cancelTarget.petName}（${cancelTarget.ownerName}）の予約を取り消します。`
          : ""
      }
    />
  );
}
