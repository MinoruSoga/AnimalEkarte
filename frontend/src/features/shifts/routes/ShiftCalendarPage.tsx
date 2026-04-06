import { useState, useCallback } from "react";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ResourceShifts } from "@/types/generated/models";
import { useGetShifts } from "../api/get-shifts";
import { useStaffsForShift } from "../api/get-staffs";
import { ShiftCalendar as ShiftCalendarGrid } from "../components/ShiftCalendar/ShiftCalendar";
import { usePermission } from "@/features/auth";

function getInitialYearMonth(): string {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  return `${year}-${month}`;
}

export function ShiftCalendarPage() {
  const [yearMonth, setYearMonth] = useState<string>(getInitialYearMonth);
  const { canCreate, canEdit } = usePermission("shifts");
  const [selectedStaffId, setSelectedStaffId] = useState<string>("all");

  const shiftsQuery = useGetShifts({
    date: yearMonth,
    staff_id: selectedStaffId !== "all" ? selectedStaffId : undefined,
  });

  const staffsQuery = useStaffsForShift();

  const handlePrevMonth = useCallback(() => {
    setYearMonth((prev) => {
      const [year, month] = prev.split("-").map(Number);
      const date = new Date(year, month - 2, 1);
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`;
    });
  }, []);

  const handleNextMonth = useCallback(() => {
    setYearMonth((prev) => {
      const [year, month] = prev.split("-").map(Number);
      const date = new Date(year, month, 1);
      return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`;
    });
  }, []);

  const handleStaffChange = useCallback((staffId: string) => {
    setSelectedStaffId(staffId);
  }, []);

  const shifts = shiftsQuery.data ?? [];
  const staffs = staffsQuery.data ?? [];

  return (
    <PageLayout title="シフト管理" resource={ResourceShifts} maxWidth="max-w-full">
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
          canCreate={canCreate}
          canEdit={canEdit}
          onPrevMonth={handlePrevMonth}
          onNextMonth={handleNextMonth}
          onStaffChange={handleStaffChange}
        />
      )}
    </PageLayout>
  );
}
