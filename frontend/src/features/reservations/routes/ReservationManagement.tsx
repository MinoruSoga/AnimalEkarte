import { useState } from "react";
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
import { CalendarIcon, Plus, ChevronLeft, ChevronRight, Stethoscope } from "lucide-react";
import { FormHeader, PrimaryButton } from "@/components/shared/Form";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { getCalendarViewLabel, getReservationTypeColor } from "@/utils/status-helpers";
import { typedSetter } from "@/lib/type-utils";
import type { CalendarView } from "../types";
import { CALENDAR_VIEW_VALUES, RESERVATION_TYPE_VALUES } from "../types";
import { MonthView } from "../components/MonthView";
import { WeekView } from "../components/WeekView";
import { ReservationFormModal } from "@/components/shared/ReservationFormModal";
import { ReservationDetailModal } from "../components/ReservationDetailModal";
import { useReservationManagement } from "../hooks/useReservationManagement";

/** Navigation step per calendar view */
const VIEW_NAV_PREV: Record<CalendarView, (d: Date) => Date> = {
  month: (d) => subMonths(d, 1),
  week: (d) => subWeeks(d, 1),
};
const VIEW_NAV_NEXT: Record<CalendarView, (d: Date) => Date> = {
  month: (d) => addMonths(d, 1),
  week: (d) => addWeeks(d, 1),
};

export const ReservationManagement = () => {
  const [currentDate, setCurrentDate] = useState(new Date());
  const [view, setView] = useState<CalendarView>("week");
  const [doctorFilter, setDoctorFilter] = useState("all");

  const {
    appointments,
    isLoading,
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

  /** Unique doctor names derived from current appointments */
  const doctorNames = Array.from(
    new Set(appointments.map((a) => a.doctor).filter(Boolean))
  ).sort();

  /** Filtered appointments */
  const filteredAppointments =
    doctorFilter === "all"
      ? appointments
      : appointments.filter((a) => a.doctor === doctorFilter);

  const navigateToday = () => setCurrentDate(new Date());
  const navigatePrevious = () => setCurrentDate(VIEW_NAV_PREV[view](currentDate));
  const navigateNext = () => setCurrentDate(VIEW_NAV_NEXT[view](currentDate));

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center bg-[#F7F6F3]">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-[#37352F]" />
          <p className="mt-2 text-[#37352F]/60">読み込み中...</p>
        </div>
      </div>
    );
  }

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
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-4">
            <div className="flex items-center bg-white rounded-md border border-[rgba(55,53,47,0.16)] p-1 shadow-sm">
              <Button variant="ghost" size="icon" className="h-10 w-10" onClick={navigatePrevious}>
                <ChevronLeft className="size-5" />
              </Button>
              <Button variant="ghost" size="sm" className="h-10 px-4 text-sm font-medium" onClick={navigateToday}>
                今日
              </Button>
              <Button variant="ghost" size="icon" className="h-10 w-10" onClick={navigateNext}>
                <ChevronRight className="size-5" />
              </Button>
            </div>
            <h2 className="text-xl font-bold text-[#37352F] flex items-center gap-2">
              {format(currentDate, "yyyy年 M月", { locale: ja })}
            </h2>

            {/* Reservation Type Legend */}
            <div className="flex items-center gap-3 ml-4">
              {RESERVATION_TYPE_VALUES.map((type) => (
                <div key={type} className="flex items-center gap-1.5">
                  <span className={`w-2.5 h-2.5 rounded-sm ${getReservationTypeColor(type).split(" ")[0]}`} />
                  <span className="text-xs text-[#37352F]/60">{type}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="flex items-center gap-2">
            {/* Doctor Filter */}
            <Select value={doctorFilter} onValueChange={setDoctorFilter}>
              <SelectTrigger className="w-[160px] bg-white border-[rgba(55,53,47,0.16)] h-10 text-sm">
                <Stethoscope className="size-3.5 text-[#37352F]/40 flex-shrink-0" />
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

            <Select value={view} onValueChange={typedSetter(setView, CALENDAR_VIEW_VALUES)}>
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

        {/* Calendar View */}
        <div className="flex-1 overflow-hidden flex flex-col">
          {view === "month" ? (
            <MonthView
              currentDate={currentDate}
              appointments={filteredAppointments}
              onAppointmentClick={handleOpenDetail}
            />
          ) : (
            <WeekView
              currentDate={currentDate}
              appointments={filteredAppointments}
              onAppointmentClick={handleOpenDetail}
              onTimeSlotClick={handleTimeSlotClick}
              onAppointmentUpdate={handleAppointmentUpdate}
            />
          )}
        </div>
      </div>

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
};
