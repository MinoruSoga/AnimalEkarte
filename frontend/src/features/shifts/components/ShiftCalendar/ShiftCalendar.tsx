import { ICON } from "@/lib/design-tokens";
import { lazy, memo, Suspense, useCallback, useMemo, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { Shift } from "@/features/shifts/types";
import { ShiftCell } from "@/features/shifts/components/ShiftCell/ShiftCell";

const ShiftFormDialog = lazy(() =>
  import("@/features/shifts/components/ShiftFormDialog/ShiftFormDialog").then((m) => ({ default: m.ShiftFormDialog }))
);

export interface StaffItem {
  id: string;
  name: string;
}

// ─── ヘッダー列（静的 JSX）: rendering-hoist-jsx ───────────────────────
const STAFF_HEADER = (
  <div className="sticky left-0 z-20 bg-white px-3 py-2 text-xs font-semibold text-gray-600 border-r border-b border-gray-200 min-w-[100px]">
    スタッフ
  </div>
);

interface DialogState {
  open: boolean;
  staffId: string;
  staffName: string;
  date: string;
  editShift: Shift | undefined;
}

const CLOSED_DIALOG: DialogState = {
  open: false,
  staffId: "",
  staffName: "",
  date: "",
  editShift: undefined,
};

interface ShiftCalendarProps {
  yearMonth: string; // YYYY-MM
  shifts: Shift[];
  staffs: StaffItem[];
  selectedStaffId: string;
  canCreate: boolean;
  canEdit: boolean;
  onPrevMonth: () => void;
  onNextMonth: () => void;
  onStaffChange: (staffId: string) => void;
}

export const ShiftCalendar = memo(function ShiftCalendar({
  yearMonth,
  shifts,
  staffs,
  selectedStaffId,
  canCreate,
  canEdit,
  onPrevMonth,
  onNextMonth,
  onStaffChange,
}: ShiftCalendarProps) {
  const [dialog, setDialog] = useState<DialogState>(CLOSED_DIALOG);

  // 当月の日付リストを生成
  const days = useMemo(() => {
    const [year, month] = yearMonth.split("-").map(Number);
    const daysInMonth = new Date(year, month, 0).getDate();
    return Array.from({ length: daysInMonth }, (_, i) => {
      const d = i + 1;
      const dateStr = `${yearMonth}-${String(d).padStart(2, "0")}`;
      const dayOfWeek = new Date(year, month - 1, d).getDay();
      return { day: d, dateStr, dayOfWeek };
    });
  }, [yearMonth]);

  // 表示対象スタッフ
  const visibleStaffs = useMemo(
    () =>
      selectedStaffId === "all"
        ? staffs
        : staffs.filter((s) => s.id === selectedStaffId),
    [staffs, selectedStaffId],
  );

  // シフトを (staffId, date) でインデックス化
  const shiftIndex = useMemo(() => {
    const idx = new Map<string, Shift>();
    for (const shift of shifts) {
      idx.set(`${shift.staff_id}__${shift.date}`, shift);
    }
    return idx;
  }, [shifts]);

  // API 由来 JSX リスト: js-cache-function-results
  const staffSelectItems = useMemo(
    () =>
      staffs.map((s) => (
        <SelectItem key={s.id} value={s.id}>
          {s.name}
        </SelectItem>
      )),
    [staffs],
  );

  const handleAddShift = useCallback(
    (staffId: string, staffName: string, date: string) => {
      setDialog({ open: true, staffId, staffName, date, editShift: undefined });
    },
    [],
  );

  const handleEditShift = useCallback(
    (staffId: string, staffName: string, shift: Shift) => {
      setDialog({
        open: true,
        staffId,
        staffName,
        date: shift.date,
        editShift: shift,
      });
    },
    [],
  );

  const handleCloseDialog = useCallback(() => {
    setDialog(CLOSED_DIALOG);
  }, []);

  const [displayYear, displayMonth] = yearMonth.split("-");
  const displayLabel = `${displayYear}年${Number(displayMonth)}月`;

  return (
    <div className="flex flex-col h-full">
      {/* ツールバー */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200 bg-white shrink-0">
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="icon"
            onClick={onPrevMonth}
            aria-label="前月"
          >
            <ChevronLeft className={ICON.action} />
          </Button>
          <span className="text-base font-semibold min-w-[120px] text-center">
            {displayLabel}
          </span>
          <Button
            variant="outline"
            size="icon"
            onClick={onNextMonth}
            aria-label="翌月"
          >
            <ChevronRight className={ICON.action} />
          </Button>
        </div>

        <div className="flex items-center gap-2">
          <Select value={selectedStaffId} onValueChange={onStaffChange}>
            <SelectTrigger className="w-[160px]">
              <SelectValue placeholder="スタッフ選択" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全スタッフ</SelectItem>
              {staffSelectItems}
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* カレンダーグリッド */}
      <div className="flex-1 overflow-auto">
        <div className="inline-block min-w-full">
          {/* ヘッダー行（日付） */}
          <div className="flex sticky top-0 z-10 bg-white border-b border-gray-200">
            {STAFF_HEADER}
            {days.map(({ day, dateStr, dayOfWeek }) => {
              const isSun = dayOfWeek === 0;
              const isSat = dayOfWeek === 6;
              const colorClass = isSun
                ? "text-red-500"
                : isSat
                  ? "text-blue-500"
                  : "text-gray-700";
              return (
                <div
                  key={dateStr}
                  className={`min-w-[52px] w-[52px] px-1 py-2 text-center text-xs font-medium border-r border-gray-200 ${colorClass}`}
                >
                  <div>{day}</div>
                  <div className="text-[10px] opacity-70">
                    {["日", "月", "火", "水", "木", "金", "土"][dayOfWeek]}
                  </div>
                </div>
              );
            })}
          </div>

          {/* スタッフ行 */}
          {visibleStaffs.map((staff) => (
            <div key={staff.id} className="flex border-b border-gray-100 hover:bg-gray-50/50">
              {/* スタッフ名列 */}
              <div className="sticky left-0 z-10 bg-white min-w-[100px] px-3 py-2 border-r border-gray-200 flex items-center">
                <span className="text-xs font-medium text-gray-800 truncate max-w-[88px]">
                  {staff.name}
                </span>
              </div>

              {/* 日付セル */}
              {days.map(({ dateStr }) => {
                const shift = shiftIndex.get(`${staff.id}__${dateStr}`);
                return (
                  <div
                    key={dateStr}
                    className="min-w-[52px] w-[52px] border-r border-gray-100 p-0.5"
                  >
                    <ShiftCell
                      shift={shift}
                      staffId={staff.id}
                      staffName={staff.name}
                      dateStr={dateStr}
                      canCreate={canCreate}
                      canEdit={canEdit}
                      onAddShift={handleAddShift}
                      onEditShift={handleEditShift}
                    />
                  </div>
                );
              })}
            </div>
          ))}

          {visibleStaffs.length === 0 ? (
            <div className="flex items-center justify-center py-16 text-sm text-gray-400">
              スタッフが見つかりません
            </div>
          ) : null}
        </div>
      </div>

      {/* シフト追加・編集ダイアログ */}
      <Suspense fallback={null}>
        <ShiftFormDialog
          open={dialog.open}
          onClose={handleCloseDialog}
          staffId={dialog.staffId}
          staffName={dialog.staffName}
          date={dialog.date}
          editShift={dialog.editShift}
        />
      </Suspense>
    </div>
  );
});
