import { ICON, C } from "@/lib/design-tokens";
import { useState, useMemo, useCallback, Suspense, lazy } from "react";
import { useSearchParams } from "react-router";
import { useClinicScope } from "@/hooks/use-clinic-scope";
import { addMonths, format, subMonths, addWeeks, subWeeks } from "date-fns";

import { CalendarIcon, Plus } from "lucide-react";
import { FormHeader } from "@/components/shared/Form/FormHeader";
import { PermissionBadges } from "@/components/shared/PermissionBadges/PermissionBadges";
import { ResourceReservations } from "@/types/generated/models";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { toJSTWallDate } from "@/lib/jst-date";
import { getReservationStatusLabel } from "@/lib/status-helpers";
import type { ReservationCreateMutations } from "@/types/reservation-create-mutations";
import type { CalendarView, Reservation } from "../types";

/** #116: キャンセル済み予約をカレンダーから非表示にする */
export function filterCalendarAppointments(appointments: Reservation[]): Reservation[] {
  return appointments.filter((a) => a.status !== "cancelled");
}

const ReservationFormModal = lazy(() =>
  import("@/components/shared/ReservationFormModal/ReservationFormModal").then((m) => ({
    default: m.ReservationFormModal,
  })),
);
import { useGetClinicHolidays } from "@/hooks/use-clinic-holidays";
import { useReservationManagement } from "../hooks/use-reservation-management";
import { useReservationTypeColorMap } from "../hooks/use-reservation-type-color-map";
import { usePermission } from "@/hooks/use-permission";
import { ReservationManagementCalendar } from "../components/ReservationManagementCalendar";
const ReservationDetailModal = lazy(() =>
  import("../components/ReservationDetailModal").then((m) => ({
    default: m.ReservationDetailModal,
  })),
);

/** Navigation step per calendar view */
const VIEW_NAV_PREV: Record<CalendarView, (d: Date) => Date> = {
  month: (d) => subMonths(d, 1),
  week: (d) => subWeeks(d, 1),
};
const VIEW_NAV_NEXT: Record<CalendarView, (d: Date) => Date> = {
  month: (d) => addMonths(d, 1),
  week: (d) => addWeeks(d, 1),
};

interface ReservationManagementProps {
  createMutations: ReservationCreateMutations;
}

export function ReservationManagement({ createMutations }: ReservationManagementProps) {
  const [currentDate, setCurrentDate] = useState(() => toJSTWallDate(new Date()));
  const { canCreate, canEdit, canDelete } = usePermission("reservations");
  const [view, setView] = useState<CalendarView>("week");
  const [doctorFilter, setDoctorFilter] = useState("all");
  const [sourceFilter, setSourceFilter] = useState("all");

  const [searchParams, setSearchParams] = useSearchParams();
  const days: 5 | 7 = searchParams.get("days") === "7" ? 7 : 5;
  const handleDaysChange = useCallback(
    (next: 5 | 7) => {
      setSearchParams(
        (prev) => {
          const params = new URLSearchParams(prev);
          if (next === 7) {
            params.set("days", "7");
          } else {
            params.delete("days");
          }
          return params;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  const { selectedClinicIds, isMultiClinic } = useClinicScope();

  const { activeEntries, colorMap: dynamicColorMap } = useReservationTypeColorMap();
  const yearMonth = format(currentDate, "yyyy-MM");
  const { data: clinicHolidays = [] } = useGetClinicHolidays(yearMonth);
  const holidayDates = useMemo(
    () => new Set(clinicHolidays.map((holiday) => holiday.date)),
    [clinicHolidays],
  );

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
    handleReservationUpdate,
    deleteConfirmOpen,
    deleteTarget,
    executeDelete,
    handleDeleteConfirmClose,
    statusConfirmOpen,
    statusConfirmTarget,
    executeStatusChange,
    handleStatusConfirmClose,
    petSelectConfirmOpen,
    setPetSelectConfirmOpen,
    handlePetSelectConfirm,
  } = useReservationManagement({
    currentDate,
    view,
    days,
    clinicIds: isMultiClinic ? selectedClinicIds : undefined,
    createMutations,
    permissions: { canCreate, canEdit, canDelete },
  });

  // BUG-069: Reservation → ReservationFormData 変換を行うラッパー
  // handleOpenForm は ReservationFormData を期待するが、詳細モーダルからは Reservation が来る
  const handleOpenFormFromReservation = useCallback(
    (reservation: Reservation) => {
      handleOpenForm({
        id: reservation.id,
        start: reservation.start,
        end: reservation.end,
        ownerName: reservation.ownerName,
        petName: reservation.petName,
        visitType: reservation.visitType,
        type: reservation.reservationTypeId ?? "",
        doctor: reservation.doctorId ?? "",
        isDesignated: reservation.isDesignated,
        status: reservation.status,
        notes: reservation.notes,
        petId: reservation.petId,
      });
    },
    [handleOpenForm],
  );

  const doctorNames = useMemo(
    () => Array.from(new Set(appointments.map((a) => a.doctor).filter(Boolean))).sort(),
    [appointments],
  );

  const filteredAppointments = useMemo(() => {
    let result = filterCalendarAppointments(appointments);
    if (doctorFilter !== "all") {
      result = result.filter((a) => a.doctor === doctorFilter);
    }
    if (sourceFilter !== "all") {
      result = result.filter((a) => a.source === sourceFilter);
    }
    return result;
  }, [appointments, doctorFilter, sourceFilter]);

  const navigateToday = useCallback(() => setCurrentDate(toJSTWallDate(new Date())), []);
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

      <ReservationManagementCalendar
        currentDate={currentDate}
        view={view}
        days={days}
        doctorFilter={doctorFilter}
        sourceFilter={sourceFilter}
        doctorNames={doctorNames}
        appointments={filteredAppointments}
        activeEntries={activeEntries}
        dynamicColorMap={dynamicColorMap}
        canCreate={canCreate}
        onViewChange={setView}
        onDoctorFilterChange={setDoctorFilter}
        onSourceFilterChange={setSourceFilter}
        onDaysChange={handleDaysChange}
        onNavigatePrevious={navigatePrevious}
        onNavigateToday={navigateToday}
        onNavigateNext={navigateNext}
        onAppointmentClick={handleOpenDetail}
        onMonthDateClick={handleMonthDateClick}
        onTimeSlotClick={handleTimeSlotClick}
        onAppointmentUpdate={handleReservationUpdate}
        holidayDates={holidayDates}
      />

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
          reservation={detailAppointment}
          onEdit={canEdit ? handleOpenFormFromReservation : undefined}
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

      {/* BUG-020: destructive status confirm (cancel / no_show) */}
      <ConfirmDialog
        open={statusConfirmOpen}
        onClose={handleStatusConfirmClose}
        onConfirm={executeStatusChange}
        title={
          statusConfirmTarget
            ? `ステータスを「${getReservationStatusLabel(statusConfirmTarget.status)}」に変更しますか？`
            : "ステータスを変更しますか？"
        }
        description={
          statusConfirmTarget
            ? `${statusConfirmTarget.reservation.petName || ""}${
                statusConfirmTarget.reservation.ownerName
                  ? `（${statusConfirmTarget.reservation.ownerName}）`
                  : ""
              } の予約ステータスを変更します。`
            : ""
        }
        confirmLabel="変更する"
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
