import { memo, useCallback } from "react";
import { C } from "@/lib/design-tokens";
import {
  ShiftTypeFull,
  ShiftTypeMorning,
  ShiftTypeAfternoon,
  ShiftTypeOff,
  ShiftTypePaidLeave,
} from "@/types/generated/models";
import type { Shift, ShiftType } from "../../types";
import { SHIFT_TYPE_LABELS } from "../../types";

// rendering-hoist-jsx: 静的カラーマップはモジュール定数に巻き上げ（ShiftCell 専用）
const SHIFT_TYPE_COLORS: Record<ShiftType, string> = {
  [ShiftTypeFull]:      `${C.bgAccentLight} ${C.textAccentDark} ${C.borderAccentBadge}`,
  [ShiftTypeMorning]:   `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreenAlt}`,
  [ShiftTypeAfternoon]: `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreenAlt}`,
  [ShiftTypeOff]:       `${C.bgStatusGray} ${C.textStatusGray} ${C.borderMuted}`,
  [ShiftTypePaidLeave]: `${C.bgStatusPurple} ${C.textStatusPurple} ${C.borderPurpleLight}`,
};

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
        className={`w-full h-full min-h-[36px] flex items-center justify-center ${C.text20} ${C.hoverText60} ${C.hoverBgPage} rounded transition-colors text-xs`}
        type="button"
        aria-label="シフトを追加"
      >
        +
      </button>
    ) : (
      <div className="w-full h-full min-h-[36px]" />
    );
  }

  const colorClass = SHIFT_TYPE_COLORS[shift.shift_type];
  const label = SHIFT_TYPE_LABELS[shift.shift_type];

  return canEdit ? (
    <button
      onClick={handleEdit}
      className={`w-full min-h-[36px] px-1 py-1 rounded border text-xs font-medium transition-opacity hover:opacity-80 ${colorClass}`}
      type="button"
    >
      <span className="block leading-tight">{label}</span>
      {shift.start_time ? (
        <span className="block leading-tight text-[10px] opacity-70">
          {shift.start_time}
          {shift.end_time ? `〜${shift.end_time}` : ""}
        </span>
      ) : null}
    </button>
  ) : (
    <div className={`w-full min-h-[36px] px-1 py-1 rounded border text-xs font-medium ${colorClass}`}>
      <span className="block leading-tight">{label}</span>
      {shift.start_time ? (
        <span className="block leading-tight text-[10px] opacity-70">
          {shift.start_time}
          {shift.end_time ? `〜${shift.end_time}` : ""}
        </span>
      ) : null}
    </div>
  );
});
