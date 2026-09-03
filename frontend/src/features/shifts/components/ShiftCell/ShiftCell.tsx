import { memo, useCallback } from "react";
import { C } from "@/lib/design-tokens";
import type { Shift, ShiftType } from "../../types";
import { SHIFT_TYPE_LABELS } from "../../types";

// rendering-hoist-jsx: 静的カラーマップはモジュール定数に巻き上げ（ShiftCell 専用）
const SHIFT_TYPE_COLORS: Record<ShiftType, string> = {
  full: `${C.bgBrandLight} ${C.textBrandDark} ${C.borderBrandLight}`,
  morning: `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreenAlt}`,
  afternoon: `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreenAlt}`,
  off: `${C.bgStatusGray} ${C.textStatusGray} ${C.borderMuted}`,
  paid_leave: `${C.bgStatusPurple} ${C.textStatusPurple} ${C.borderPurpleLight}`,
};

function formatTimeToHHmm(time: string): string {
  if (!time) return "";
  const parts = time.split(":");
  if (parts.length >= 2) {
    return `${parts[0].padStart(2, "0")}:${parts[1].padStart(2, "0")}`;
  }
  return time;
}

function getShiftTimeRangeLabel(shift: Shift): string {
  if (!shift.start_time && !shift.end_time) return "時刻なし";
  const start = shift.start_time ? formatTimeToHHmm(shift.start_time) : "開始時刻未設定";
  const end = shift.end_time ? formatTimeToHHmm(shift.end_time) : "終了時刻未設定";
  return `${start}〜${end}`;
}

interface ShiftCellProps {
  shift: Shift | undefined;
  // rerender-dependencies: primitive props + stable handlers の分離パターン
  staffId: string;
  staffName: string;
  dateStr: string;
  canCreate: boolean;
  canEdit: boolean;
  onAddShift: (staffId: string, staffName: string, dateStr: string) => void;
  onEditShift: (staffId: string, staffName: string, shift: Shift) => void;
}

export const ShiftCell = memo(function ShiftCell({
  shift,
  staffId,
  staffName,
  dateStr,
  canCreate,
  canEdit,
  onAddShift,
  onEditShift,
}: ShiftCellProps) {
  const handleAdd = useCallback(() => {
    onAddShift(staffId, staffName, dateStr);
  }, [onAddShift, staffId, staffName, dateStr]);

  const handleEdit = useCallback(() => {
    if (shift) onEditShift(staffId, staffName, shift);
  }, [onEditShift, staffId, staffName, shift]);

  if (!shift) {
    return canCreate ? (
      <button
        onClick={handleAdd}
        className={`w-full h-full min-h-11 flex items-center justify-center ${C.text20} ${C.hoverText60} ${C.hoverBgPage} rounded transition-colors text-xs`}
        type="button"
        aria-label={`${staffName} ${dateStr} シフトを追加`}
      >
        +
      </button>
    ) : (
      <div className="w-full h-full min-h-[36px]" />
    );
  }

  const colorClass = SHIFT_TYPE_COLORS[shift.shift_type];
  const label = SHIFT_TYPE_LABELS[shift.shift_type];
  const timeRangeLabel = getShiftTimeRangeLabel(shift);
  const startDisplay = shift.start_time ? formatTimeToHHmm(shift.start_time) : "";
  const endDisplay = shift.end_time ? formatTimeToHHmm(shift.end_time) : "";

  return canEdit ? (
    <button
      onClick={handleEdit}
      className={`w-full min-h-11 px-1 py-1 rounded border text-xs font-medium transition-opacity hover:opacity-80 overflow-hidden ${colorClass}`}
      type="button"
      aria-label={`${staffName} ${dateStr} ${label}シフト（${timeRangeLabel}）を編集`}
    >
      <span className="block leading-tight">{label}</span>
      {startDisplay ? (
        <span className="block leading-tight text-2xs opacity-70">
          {startDisplay}
          {endDisplay ? `〜${endDisplay}` : ""}
        </span>
      ) : null}
    </button>
  ) : (
    <div
      className={`w-full min-h-[36px] px-1 py-1 rounded border text-xs font-medium overflow-hidden ${colorClass}`}
    >
      <span className="block leading-tight">{label}</span>
      {startDisplay ? (
        <span className="block leading-tight text-2xs opacity-70">
          {startDisplay}
          {endDisplay ? `〜${endDisplay}` : ""}
        </span>
      ) : null}
    </div>
  );
});
