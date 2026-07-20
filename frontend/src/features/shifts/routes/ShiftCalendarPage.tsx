import { lazy, Suspense, useState, useCallback } from "react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ResourceShifts } from "@/types/generated/models";
import { useGetShifts } from "../api/get-shifts";
import { useGetStaffsForShift } from "../api/get-staffs";
import { ShiftCalendar as ShiftCalendarGrid } from "../components/ShiftCalendar/ShiftCalendar";
import { usePermission } from "@/hooks/use-permission";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { useGetClinicHolidays } from "../api/clinic-holidays";
import type { ClinicHoliday } from "../api/clinic-holidays";
import { todayJSTISO } from "@/lib/jst-date";
import { LAYOUT } from "@/lib/design-tokens";

const ClinicHolidayModal = lazy(() =>
  import("../components/ClinicHolidayModal/ClinicHolidayModal").then((m) => ({ default: m.ClinicHolidayModal }))
);

function getInitialYearMonth(): string {
  return todayJSTISO().slice(0, 7);
}

function shiftYearMonth(value: string, delta: number): string {
  const [year, month] = value.split("-").map(Number);
  const zeroBased = year * 12 + (month - 1) + delta;
  const nextYear = Math.floor(zeroBased / 12);
  const nextMonth = (zeroBased % 12) + 1;
  return `${nextYear}-${String(nextMonth).padStart(2, "0")}`;
}

interface HolidayModalState {
  open: boolean;
  date: string;
}

const CLOSED_HOLIDAY_MODAL: HolidayModalState = { open: false, date: "" };

export function ShiftCalendarPage() {
  const [yearMonth, setYearMonth] = useState<string>(getInitialYearMonth);
  const { canCreate, canEdit, canDelete } = usePermission("shifts");
  const [selectedStaffId, setSelectedStaffId] = useState<string>("all");
  const [holidayModal, setHolidayModal] = useState<HolidayModalState>(CLOSED_HOLIDAY_MODAL);

  const shiftsQuery = useGetShifts({
    date: yearMonth,
    staff_id: selectedStaffId !== "all" ? selectedStaffId : undefined,
  });

  const staffsQuery = useGetStaffsForShift();
  const holidaysQuery = useGetClinicHolidays(yearMonth);

  const handlePrevMonth = useCallback(() => {
    setYearMonth((prev) => shiftYearMonth(prev, -1));
  }, []);

  const handleNextMonth = useCallback(() => {
    setYearMonth((prev) => shiftYearMonth(prev, 1));
  }, []);

  const handleStaffChange = useCallback((staffId: string) => {
    setSelectedStaffId(staffId);
  }, []);

  const handleDateHeaderClick = useCallback((dateStr: string) => {
    setHolidayModal({ open: true, date: dateStr });
  }, []);

  const handleHolidayModalClose = useCallback(() => {
    setHolidayModal(CLOSED_HOLIDAY_MODAL);
  }, []);

  const shifts = shiftsQuery.data ?? [];
  const staffs = staffsQuery.data ?? [];
  const holidays = holidaysQuery.data ?? [];

  // 定休日を date → ClinicHoliday のマップに変換（モーダルで existing 取得用）
  const holidayMap = new Map<string, ClinicHoliday>(holidays.map((h) => [h.date, h]));
  const modalExisting = holidayModal.date ? holidayMap.get(holidayModal.date) : undefined;

  if (shiftsQuery.isLoading || staffsQuery.isLoading) {
    return (
      <PageLayout title="シフト管理" resource={ResourceShifts} maxWidth={LAYOUT.pageContentMaxWidth.full}>
        <LoadingFallback />
      </PageLayout>
    );
  }

  if (shiftsQuery.isError || staffsQuery.isError) {
    return (
      <PageLayout title="シフト管理" resource={ResourceShifts} maxWidth={LAYOUT.pageContentMaxWidth.full}>
        <ErrorFallback message="シフトデータの取得に失敗しました" />
      </PageLayout>
    );
  }

  return (
    <PageLayout title="シフト管理" resource={ResourceShifts} maxWidth={LAYOUT.pageContentMaxWidth.full}>
      <ShiftCalendarGrid
        yearMonth={yearMonth}
        shifts={shifts}
        staffs={staffs}
        holidays={holidays}
        selectedStaffId={selectedStaffId}
        canCreate={canCreate}
        canEdit={canEdit}
        canDelete={canDelete}
        onPrevMonth={handlePrevMonth}
        onNextMonth={handleNextMonth}
        onStaffChange={handleStaffChange}
        onDateHeaderClick={canCreate ? handleDateHeaderClick : undefined}
      />

      <Suspense fallback={null}>
        <ClinicHolidayModal
          open={holidayModal.open}
          onClose={handleHolidayModalClose}
          date={holidayModal.date}
          existing={modalExisting}
          canEdit={canEdit}
        />
      </Suspense>
    </PageLayout>
  );
}
