import { ICON, C } from "@/lib/design-tokens";
import { useState, useMemo, useCallback, Suspense, lazy } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { addMonths, subMonths, addWeeks, subWeeks } from "date-fns";

import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { CalendarIcon, Plus, ChevronLeft, ChevronRight } from "lucide-react";
import { FormHeader } from "@/components/shared/Form/FormHeader";
import { PermissionBadges } from "@/components/shared/PermissionBadges/PermissionBadges";
import { ResourceReservations } from "@/types/generated/models";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { getCalendarViewLabel } from "@/utils/status-helpers";
import { typedSetter } from "@/lib/type-utils";
import type { CalendarView, ReservationAppointment } from "../types";
import { CALENDAR_VIEW_VALUES } from "../types";
const ReservationFormModal = lazy(() =>
  import("@/components/shared/ReservationFormModal/ReservationFormModal").then((m) => ({
    default: m.ReservationFormModal,
  })),
);
import { useReservationManagement } from "../hooks/use-reservation-management";
import { useReservationCategoryColorMap } from "@/features/master";
import { usePermission } from "@/features/auth";

const MonthView = lazy(() =>
  import("../components/MonthView").then((m) => ({ default: m.MonthView })),
);
const WeekView = lazy(() =>
  import("../components/WeekView").then((m) => ({ default: m.WeekView })),
);
const ReservationDetailModal = lazy(() =>
  import("../components/ReservationDetailModal").then((m) => ({
    default: m.ReservationDetailModal,
  })),
);

// rendering-hoist-jsx: 静的 SelectItem JSX をモジュール定数に巻き上げ
const SOURCE_FILTER_SELECT_ITEMS = (
  <>
    <SelectItem value="all">すべて</SelectItem>
    <SelectItem value="manual">手動予約</SelectItem>
    <SelectItem value="line">LINE予約</SelectItem>
  </>
);

const CALENDAR_VIEW_SELECT_ITEMS = CALENDAR_VIEW_VALUES.map((v) => (
  <SelectItem key={v} value={v}>
    {getCalendarViewLabel(v)}
  </SelectItem>
));

/** Navigation step per calendar view */
const VIEW_NAV_PREV: Record<CalendarView, (d: Date) => Date> = {
  month: (d) => subMonths(d, 1),
  week: (d) => subWeeks(d, 1),
};
const VIEW_NAV_NEXT: Record<CalendarView, (d: Date) => Date> = {
  month: (d) => addMonths(d, 1),
  week: (d) => addWeeks(d, 1),
};

export function ReservationManagement() {
  const [currentDate, setCurrentDate] = useState(() => new Date());
  const { canCreate, canEdit, canDelete } = usePermission("reservations");
  const [view, setView] = useState<CalendarView>("week");
  const [doctorFilter, setDoctorFilter] = useState("all");
  const [sourceFilter, setSourceFilter] = useState("all");

  const { activeEntries, colorMap: dynamicColorMap } = useReservationCategoryColorMap();

  const {
    appointments,
    isFormOpen,
    editingAppointment,
    handleOpenForm,
    handleCloseForm,
    handleSave,
    isDetailOpen,
    detailAppointment,
    handleOpenDetail,
    handleCloseDetail,
    handleStatusChange,
    handleDelete,
    handleCreateRecord,
    handleTimeSlotClick,
    handleAppointmentUpdate,
    deleteConfirmOpen,
    deleteTarget,
    executeDelete,
    handleDeleteConfirmClose,
    petSelectConfirmOpen,
    setPetSelectConfirmOpen,
    handlePetSelectConfirm,
  } = useReservationManagement();

  // BUG-069: ReservationAppointment → ReservationFormData 変換を行うラッパー
  // handleOpenForm は ReservationFormData を期待するが、詳細モーダルからは ReservationAppointment が来る
  const handleOpenFormFromAppointment = useCallback(
    (appointment: ReservationAppointment) => {
      handleOpenForm({
        id: appointment.id,
        start: appointment.start,
        end: appointment.end,
        ownerName: appointment.ownerName,
        petName: appointment.petName,
        visitType: appointment.visitType,
        type: appointment.reservationCategoryId ?? "",
        doctor: appointment.doctorId ?? "",
        isDesignated: appointment.isDesignated,
        status: appointment.status,
        notes: appointment.notes,
        petId: appointment.petId,
      });
    },
    [handleOpenForm],
  );

  const doctorNames = useMemo(
    () =>
      Array.from(
        new Set(appointments.map((a) => a.doctor).filter(Boolean)),
      ).sort(),
    [appointments],
  );

  // js-cache-function-results: API データから生成する JSX リストを useMemo でキャッシュ
  const doctorNameSelectItems = useMemo(
    () => doctorNames.map((name) => (
      <SelectItem key={name} value={name}>{name}</SelectItem>
    )),
    [doctorNames]
  );

  const filteredAppointments = useMemo(
    () => {
      let result = appointments;
      if (doctorFilter !== "all") {
        result = result.filter((a) => a.doctor === doctorFilter);
      }
      if (sourceFilter !== "all") {
        result = result.filter((a) => a.source === sourceFilter);
      }
      return result;
    },
    [appointments, doctorFilter, sourceFilter],
  );

  const navigateToday = useCallback(() => setCurrentDate(new Date()), []);
  const navigatePrevious = useCallback(
    () => setCurrentDate((prev) => VIEW_NAV_PREV[view](prev)),
    [view],
  );
  const navigateNext = useCallback(
    () => setCurrentDate((prev) => VIEW_NAV_NEXT[view](prev)),
    [view],
  );

  // BUG-076: 月表示の日付セルクリックで週表示に遷移
  const handleMonthDateClick = useCallback((date: Date) => {
    setCurrentDate(date);
    setView("week");
  }, []);

  return (
    <div className={`flex-1 flex flex-col h-full ${C.bgPage} min-w-0 w-full`}>
      <FormHeader
        title="予約管理"
        icon={<CalendarIcon className={`${ICON.page} ${C.text}`} />}
        action={
          <div className="flex items-center gap-2">
            <PermissionBadges resource={ResourceReservations} />
            {canCreate ? (
              <PrimaryButton className="gap-2" onClick={() => handleOpenForm()}>
                <Plus className={ICON.action} />
                <span className="hidden sm:inline">新規予約登録</span>
                <span className="sm:hidden">予約</span>
              </PrimaryButton>
            ) : null}
          </div>
        }
      />

      <div className="flex-1 flex flex-col p-4 overflow-hidden w-full min-w-0">
        {/* Toolbar */}
        <div className="flex flex-wrap items-center justify-between gap-2 mb-4">
          <div className="flex items-center gap-4">
            <div className={`flex items-center bg-white rounded-md border ${C.borderMedium} p-1 shadow-sm`}>
              <Button
                variant="ghost"
                size="icon"
                className="h-10 w-10"
                onClick={navigatePrevious}
              >
                <ChevronLeft className={ICON.page} />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-10 px-4 text-base font-medium"
                onClick={navigateToday}
              >
                今日
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="h-10 w-10"
                onClick={navigateNext}
              >
                <ChevronRight className={ICON.page} />
              </Button>
            </div>
            <h2 className={`text-xl font-bold ${C.text} flex items-center gap-2`}>
              {format(currentDate, "yyyy年 M月", { locale: ja })}
            </h2>
          </div>

          <div className="flex items-center gap-2">
            {/* Source Filter */}
            <Select value={sourceFilter} onValueChange={setSourceFilter}>
              <SelectTrigger className={`w-[140px] bg-white ${C.borderMedium} h-10 text-base`}>
                <SelectValue placeholder="予約ソース" />
              </SelectTrigger>
              <SelectContent>{SOURCE_FILTER_SELECT_ITEMS}</SelectContent>
            </Select>

            {/* Doctor Filter */}
            <Select value={doctorFilter} onValueChange={setDoctorFilter}>
              <SelectTrigger className={`w-[160px] bg-white ${C.borderMedium} h-10 text-base`}>
                <SelectValue placeholder="担当医で絞込" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">すべての医師</SelectItem>
                {doctorNameSelectItems}
              </SelectContent>
            </Select>

            <Select
              value={view}
              onValueChange={typedSetter(setView, CALENDAR_VIEW_VALUES)}
            >
              <SelectTrigger className={`w-[140px] bg-white ${C.borderMedium} h-10 text-base`}>
                <SelectValue placeholder="表示切替" />
              </SelectTrigger>
              <SelectContent>
                {CALENDAR_VIEW_SELECT_ITEMS}
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Legend row */}
        <div className="flex items-center gap-3 mb-3 flex-wrap">
          {activeEntries.map((entry) => (
            <div key={entry.name} className="flex items-center gap-1.5">
              <span
                className="w-2.5 h-2.5 rounded-full"
                style={entry.color.dotStyle}
              />
              <span className={`text-base ${C.text60}`}>{entry.name}</span>
            </div>
          ))}
        </div>

        {/* Calendar View */}
        <div className="flex-1 overflow-hidden flex flex-col">
          <Suspense
            fallback={
              <div className="flex-1 flex items-center justify-center">
                <div className="text-center">
                  <div className={`inline-block animate-spin rounded-full h-8 w-8 border-b-2 ${C.borderPrimary}`} />
                  <p className={`mt-2 ${C.text60} text-base`}>
                    読み込み中...
                  </p>
                </div>
              </div>
            }
          >
            {view === "month" ? (
              <MonthView
                currentDate={currentDate}
                appointments={filteredAppointments}
                onAppointmentClick={handleOpenDetail}
                onDateClick={handleMonthDateClick}
                dynamicColorMap={dynamicColorMap}
              />
            ) : (
              <WeekView
                currentDate={currentDate}
                appointments={filteredAppointments}
                onAppointmentClick={handleOpenDetail}
                onTimeSlotClick={canCreate ? handleTimeSlotClick : undefined}
                onAppointmentUpdate={handleAppointmentUpdate}
                dynamicColorMap={dynamicColorMap}
              />
            )}
          </Suspense>
        </div>
      </div>

      <Suspense fallback={null}>
        {/* Create/Edit Form */}
        <ReservationFormModal
          isOpen={isFormOpen}
          onClose={handleCloseForm}
          onSave={handleSave}
          initialData={editingAppointment}
          canCreate={canCreate}
          canEdit={canEdit}
        />

        {/* Detail Modal */}
        <ReservationDetailModal
          isOpen={isDetailOpen}
          onClose={handleCloseDetail}
          appointment={detailAppointment}
          onEdit={canEdit ? handleOpenFormFromAppointment : undefined}
          onDelete={canDelete ? handleDelete : undefined}
          onCreateRecord={handleCreateRecord}
          onStatusChange={canEdit ? handleStatusChange : undefined}
        />
      </Suspense>

      {/* Delete Confirm Dialog */}
      <ConfirmDialog
        open={deleteConfirmOpen}
        onClose={handleDeleteConfirmClose}
        onConfirm={executeDelete}
        title="予約を削除しますか？"
        description={
          deleteTarget
            ? `${deleteTarget.petName || ""}${deleteTarget.ownerName ? `（${deleteTarget.ownerName}）` : ""} の予約を削除します。この操作は取り消せません。`
            : ""
        }
        confirmLabel="削除する"
        variant="destructive"
      />

      {/* Pet Select Confirm Dialog */}
      <ConfirmDialog
        open={petSelectConfirmOpen}
        onClose={() => setPetSelectConfirmOpen(false)}
        onConfirm={handlePetSelectConfirm}
        title="ペットIDが紐付いていません"
        description="この予約にはペットIDが紐付いていません。ペット選択画面へ移動しますか？"
        confirmLabel="移動する"
      />
    </div>
  );
}
