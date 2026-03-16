import { useState, useMemo, useCallback, Suspense, lazy } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { addMonths, subMonths, addWeeks, subWeeks } from "date-fns";

import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { CalendarIcon, Plus, ChevronLeft, ChevronRight } from "lucide-react";
import { FormHeader } from "@/components/shared/Form/FormHeader";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { getCalendarViewLabel } from "@/utils/status-helpers";
import { typedSetter } from "@/lib/type-utils";
import type { CalendarView } from "../types";
import { CALENDAR_VIEW_VALUES } from "../types";
const ReservationFormModal = lazy(() =>
  import("@/components/shared/ReservationFormModal/ReservationFormModal").then((m) => ({
    default: m.ReservationFormModal,
  })),
);
import { useReservationManagement } from "../hooks/useReservationManagement";
import { useServiceTypeColorMap } from "@/hooks/use-service-type-color-map";

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
  const [currentDate, setCurrentDate] = useState(new Date());
  const [view, setView] = useState<CalendarView>("week");
  const [doctorFilter, setDoctorFilter] = useState("all");

  const { activeEntries, colorMap: dynamicColorMap } = useServiceTypeColorMap();

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

  const doctorNames = useMemo(
    () =>
      Array.from(
        new Set(appointments.map((a) => a.doctor).filter(Boolean)),
      ).sort(),
    [appointments],
  );

  const filteredAppointments = useMemo(
    () =>
      doctorFilter === "all"
        ? appointments
        : appointments.filter((a) => a.doctor === doctorFilter),
    [appointments, doctorFilter],
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

  return (
    <div className="flex-1 flex flex-col h-full bg-[#F7F6F3] min-w-0 w-full">
      <FormHeader
        title="予約管理"
        icon={<CalendarIcon className="size-5 text-[#37352F]" />}
        action={
          <div className="flex items-center gap-2">
            <PrimaryButton className="gap-2" onClick={() => handleOpenForm()}>
              <Plus className="size-4" />
              <span className="hidden sm:inline">新規予約</span>
              <span className="sm:hidden">予約</span>
            </PrimaryButton>
          </div>
        }
      />

      <div className="flex-1 flex flex-col p-4 overflow-hidden w-full min-w-0">
        {/* Toolbar */}
        <div className="flex flex-wrap items-center justify-between gap-2 mb-4">
          <div className="flex items-center gap-4">
            <div className="flex items-center bg-white rounded-md border border-[rgba(55,53,47,0.16)] p-1 shadow-sm">
              <Button
                variant="ghost"
                size="icon"
                className="h-10 w-10"
                onClick={navigatePrevious}
              >
                <ChevronLeft className="size-5" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="h-10 px-4 text-sm font-medium"
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
                <ChevronRight className="size-5" />
              </Button>
            </div>
            <h2 className="text-xl font-bold text-[#37352F] flex items-center gap-2">
              {format(currentDate, "yyyy年 M月", { locale: ja })}
            </h2>
          </div>

          <div className="flex items-center gap-2">
            {/* Doctor Filter */}
            <Select value={doctorFilter} onValueChange={setDoctorFilter}>
              <SelectTrigger className="w-[160px] bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm">
                <SelectValue placeholder="担当医で絞込" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">すべての医師</SelectItem>
                {doctorNames.map((name) => (
                  <SelectItem key={name} value={name}>
                    {name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            <Select
              value={view}
              onValueChange={typedSetter(setView, CALENDAR_VIEW_VALUES)}
            >
              <SelectTrigger className="w-[140px] bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm">
                <SelectValue placeholder="表示切替" />
              </SelectTrigger>
              <SelectContent>
                {CALENDAR_VIEW_VALUES.map((v) => (
                  <SelectItem key={v} value={v}>
                    {getCalendarViewLabel(v)}
                  </SelectItem>
                ))}
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
              <span className="text-xs text-[#37352F]/60">{entry.name}</span>
            </div>
          ))}
        </div>

        {/* Calendar View */}
        <div className="flex-1 overflow-hidden flex flex-col">
          <Suspense
            fallback={
              <div className="flex-1 flex items-center justify-center">
                <div className="text-center">
                  <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[#37352F]" />
                  <p className="mt-2 text-[#37352F]/60 text-sm">
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
                dynamicColorMap={dynamicColorMap}
              />
            ) : (
              <WeekView
                currentDate={currentDate}
                appointments={filteredAppointments}
                onAppointmentClick={handleOpenDetail}
                onTimeSlotClick={handleTimeSlotClick}
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
        />

        {/* Detail Modal */}
        <ReservationDetailModal
          isOpen={isDetailOpen}
          onClose={handleCloseDetail}
          appointment={detailAppointment}
          onEdit={handleOpenForm}
          onDelete={handleDelete}
          onCreateRecord={handleCreateRecord}
          onStatusChange={handleStatusChange}
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
