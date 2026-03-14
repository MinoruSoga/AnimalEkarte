import { useState, useCallback } from "react";
import { PageLayout } from "@/components/shared/PageLayout";
import { useShifts } from "../api/get-shifts";
import { useStaffsForShift } from "../api/get-staffs";
import { ShiftCalendar as ShiftCalendarGrid } from "../components/ShiftCalendar/ShiftCalendar";

function getInitialYearMonth(): string {
  const now = new Date();
  const y = now.getFullYear();
  const m = String(now.getMonth() + 1).padStart(2, "0");
  return `${y}-${m}`;
}

export function ShiftCalendarPage() {
  const [yearMonth, setYearMonth] = useState<string>(getInitialYearMonth);
  const [selectedStaffId, setSelectedStaffId] = useState<string>("all");

  const shiftsQuery = useShifts({
    date: yearMonth,
    staff_id: selectedStaffId !== "all" ? selectedStaffId : undefined,
  });

  const staffsQuery = useStaffsForShift();

  const handlePrevMonth = useCallback(() => {
    setYearMonth((prev) => {
      const [y, m] = prev.split("-").map(Number);
      const d = new Date(y, m - 2, 1);
      return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
    });
  }, []);

  const handleNextMonth = useCallback(() => {
    setYearMonth((prev) => {
      const [y, m] = prev.split("-").map(Number);
      const d = new Date(y, m, 1);
      return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
    });
  }, []);

  const handleStaffChange = useCallback((staffId: string) => {
    setSelectedStaffId(staffId);
  }, []);

  const shifts = shiftsQuery.data ?? [];
  const staffs = staffsQuery.data ?? [];

  return (
    <PageLayout title="シフト管理" maxWidth="max-w-full">
      {shiftsQuery.isError ? (
        <div className="flex items-center justify-center h-64 text-sm text-red-500">
          シフトデータの取得に失敗しました
        </div>
      ) : (
        <ShiftCalendarGrid
          yearMonth={yearMonth}
          shifts={shifts}
          staffs={staffs}
          selectedStaffId={selectedStaffId}
          onPrevMonth={handlePrevMonth}
          onNextMonth={handleNextMonth}
          onStaffChange={handleStaffChange}
        />
      )}
    </PageLayout>
  );
}
