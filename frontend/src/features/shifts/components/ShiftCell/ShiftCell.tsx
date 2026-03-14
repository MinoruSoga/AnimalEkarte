import { memo } from "react";
import type { Shift } from "../../types";
import { SHIFT_TYPE_LABELS, SHIFT_TYPE_COLORS } from "../../types";

interface ShiftCellProps {
  shift: Shift | undefined;
  onAdd: () => void;
  onEdit: (shift: Shift) => void;
}

export const ShiftCell = memo(function ShiftCell({ shift, onAdd, onEdit }: ShiftCellProps) {
  if (!shift) {
    return (
      <button
        onClick={onAdd}
        className="w-full h-full min-h-[36px] flex items-center justify-center text-gray-300 hover:text-gray-500 hover:bg-gray-50 rounded transition-colors text-xs"
        type="button"
        aria-label="シフトを追加"
      >
        +
      </button>
    );
  }

  const colorClass = SHIFT_TYPE_COLORS[shift.shift_type];
  const label = SHIFT_TYPE_LABELS[shift.shift_type];

  return (
    <button
      onClick={() => onEdit(shift)}
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
  );
});
