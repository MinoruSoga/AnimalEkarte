import { C, STYLE } from "@/lib/design-tokens";
import { DAY_OF_WEEK_LABELS } from "@/constants/day-of-week";
import { EmptyState } from "@/components/shared/DataStates";
import { lazy, memo, Suspense, useCallback, useMemo, useState } from "react";
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";
import { CalendarNavToolbar } from "@/components/shared/CalendarNavToolbar";
import type { Shift } from "../../types";
import { ShiftCell } from "../../components/ShiftCell/ShiftCell";
import type { ClinicHoliday } from "../../api/clinic-holidays";

const ShiftFormDialog = lazy(() =>
  import("../../components/ShiftFormDialog/ShiftFormDialog").then((m) => ({ default: m.ShiftFormDialog }))
);

export interface StaffItem {
  id: string;
  name: string;
}

// ─── ヘッダー列（静的 JSX）: rendering-hoist-jsx ───────────────────────
const STAFF_HEADER = (
  <div className={`sticky left-0 z-20 ${C.bgWhite} px-3 py-2 text-xs font-semibold ${C.text60} border-r border-b ${C.borderLight} w-[120px] min-w-[120px] shrink-0`}>
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
  holidays: ClinicHoliday[];
  selectedStaffId: string;
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
  onPrevMonth: () => void;
  onNextMonth: () => void;
  onStaffChange: (staffId: string) => void;
  onDateHeaderClick?: (dateStr: string) => void;
}

export const ShiftCalendar = memo(function ShiftCalendar({
  yearMonth,
  shifts,
  staffs,
  holidays,
  selectedStaffId,
  canCreate,
  canEdit,
  canDelete,
  onPrevMonth,
  onNextMonth,
  onStaffChange,
  onDateHeaderClick,
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

  // 定休日を日付文字列でインデックス化
  const holidaySet = useMemo(() => new Set(holidays.map((h) => h.date)), [holidays]);

  // シフトを (staffId, date) でインデックス化
  const shiftIndex = useMemo(() => {
    const idx = new Map<string, Shift>();
    for (const shift of shifts) {
      idx.set(`${shift.staff_id}__${shift.date}`, shift);
    }
    return idx;
  }, [shifts]);

  // スタッフ選択肢({value,label}): js-cache-function-results
  const staffFilterOptions = useMemo<SearchableSelectOption[]>(
    () => [
      { value: "all", label: "全スタッフ" },
      ...staffs.map((s) => ({ value: s.id, label: s.name })),
    ],
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
      <div className={`flex items-center justify-between px-4 py-3 border-b ${C.borderLight} ${C.bgWhite} shrink-0`}>
        <CalendarNavToolbar
          layout="inline"
          variant="outline"
          size="auto"
          onPrev={onPrevMonth}
          onNext={onNextMonth}
          prevAriaLabel="前月"
          nextAriaLabel="翌月"
          label={
            <span className="text-base font-semibold min-w-[120px] text-center">
              {displayLabel}
            </span>
          }
        />

        <div className="flex items-center gap-2">
          <SearchableSelect
            value={selectedStaffId}
            onValueChange={onStaffChange}
            options={staffFilterOptions}
            ariaLabel="スタッフ絞り込み"
            placeholder="スタッフ選択"
            searchPlaceholder="スタッフを検索..."
            className="w-[160px]"
          />
        </div>
      </div>

      {/* カレンダーグリッド */}
      <div className="flex-1 overflow-auto">
        <div className="inline-block min-w-full">
          {/* ヘッダー行（日付） */}
          <div className={`flex sticky top-0 z-10 ${C.bgWhite} border-b ${C.borderLight}`}>
            {STAFF_HEADER}
            {days.map(({ day, dateStr, dayOfWeek }) => {
              const isSun = dayOfWeek === 0;
              const isSat = dayOfWeek === 6;
              const isHoliday = holidaySet.has(dateStr);
              const colorClass = isHoliday
                ? C.danger
                : isSun
                  ? C.danger
                  : isSat
                    ? C.textBrand
                    : C.text70;
              const isClickable = !!onDateHeaderClick && canCreate;
              return (
                <div
                  key={dateStr}
                  role={isClickable ? "button" : undefined}
                  tabIndex={isClickable ? 0 : undefined}
                  aria-label={isClickable ? `${dateStr} の定休日設定` : undefined}
                  onClick={isClickable ? () => onDateHeaderClick(dateStr) : undefined}
                  onKeyDown={isClickable ? (e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); onDateHeaderClick(dateStr); } } : undefined}
                  className={`min-w-[52px] w-[52px] px-1 py-2 text-center text-xs font-medium border-r ${C.borderLight} ${colorClass} relative${isClickable ? ` cursor-pointer ${STYLE.tableRowHover} transition-colors` : ""}`}
                >
                  <div>{day}</div>
                  <div className="text-2xs opacity-70">
                    {DAY_OF_WEEK_LABELS[dayOfWeek]}
                  </div>
                  {isHoliday ? (
                    <div className={`absolute bottom-0.5 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full ${C.bgStatusRedDot}`} aria-label="定休日" />
                  ) : null}
                </div>
              );
            })}
          </div>

          {/* スタッフ行 */}
          {visibleStaffs.map((staff) => (
            <div key={staff.id} className={`flex border-b ${C.borderDivider} ${C.hoverBgPageHalf}`}>
              {/* スタッフ名列 */}
              <div className={`sticky left-0 z-10 ${C.bgWhite} w-[120px] min-w-[120px] shrink-0 px-3 py-2 border-r ${C.borderLight} flex items-center overflow-x-auto`}>
                <span className={`text-xs font-medium ${C.text} whitespace-nowrap`}>
                  {staff.name}
                </span>
              </div>

              {/* 日付セル */}
              {days.map(({ dateStr }) => {
                const shift = shiftIndex.get(`${staff.id}__${dateStr}`);
                return (
                  <div
                    key={dateStr}
                    className={`min-w-[52px] w-[52px] border-r ${C.borderDivider} p-0.5`}
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
            <div className="flex items-center justify-center py-16">
              <EmptyState message="スタッフが見つかりません" />
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
          canEdit={canEdit}
          canDelete={canDelete}
        />
      </Suspense>
    </div>
  );
});
