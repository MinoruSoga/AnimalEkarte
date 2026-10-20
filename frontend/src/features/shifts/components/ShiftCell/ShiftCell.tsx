import { memo, useCallback } from "react";
import { C } from "@/lib/design-tokens";
import { formatJSTTime } from "@/lib/jst-date";
import type { Shift, ShiftType } from "../../types";
import { SHIFT_TYPE_LABELS } from "../../types";

// rendering-hoist-jsx: 静的カラーマップはモジュール定数に巻き上げ（ShiftCell 専用）
const SHIFT_TYPE_COLORS: Record<ShiftType, string> = {
  full: `${C.bgBrandLight} ${C.textBrandDark} ${C.borderBrandLight}`,
  morning: `${C.bgAccentLight} ${C.textAccentDark} ${C.borderAccentBadge}`,
  afternoon: `${C.bgDiscountLight} ${C.textDiscount} ${C.borderOrangeBadge}`,
  off: `${C.bgStatusGray} ${C.textStatusGray} ${C.borderMuted}`,
  paid_leave: `${C.bgStatusPurple} ${C.textStatusPurple} ${C.borderPurpleLight}`,
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
  // EMR-514: interval_state=unknown は旧日時が未解決。NULL を空欄・勤務可と解釈せず
  // 「時刻未確認」と明示し、推測補完しない。off/paid_leave のNULLは終日非勤務として「終日」。
  const isUnknown = shift.interval_state === "unknown";
  const isNonWork = shift.shift_type === "off" || shift.shift_type === "paid_leave";
  const timeRangeLabel = isUnknown
    ? "時刻未確認"
    : shift.starts_at
      ? `${formatJSTTime(shift.starts_at)}〜${shift.ends_at ? formatJSTTime(shift.ends_at) : "終了時刻未設定"}`
      : isNonWork
        ? "終日"
        : "時刻未設定";
  const timeDisplay = isUnknown
    ? "時刻未確認"
    : shift.starts_at
      ? `${formatJSTTime(shift.starts_at)}${shift.ends_at ? `〜${formatJSTTime(shift.ends_at)}` : ""}`
      : "";
  const interactive = canEdit;

  return interactive ? (
    <button
      onClick={handleEdit}
      className={`w-full min-h-11 px-1 py-1 rounded border text-xs font-medium transition-opacity overflow-hidden ${colorClass}`}
      type="button"
      aria-label={`${staffName} ${dateStr} ${label}シフト（${timeRangeLabel}）を編集`}
    >
      <span className="block leading-tight">
        {label}
        {isUnknown ? <span className="sr-only">（時刻未確認）</span> : null}
      </span>
      {timeDisplay ? (
        <span className="block leading-tight text-2xs opacity-70">{timeDisplay}</span>
      ) : null}
    </button>
  ) : (
    <div
      className={`w-full min-h-[36px] px-1 py-1 rounded border text-xs font-medium overflow-hidden ${colorClass}`}
    >
      <span className="block leading-tight">
        {label}
        {isUnknown ? <span className="sr-only">（時刻未確認）</span> : null}
      </span>
      {timeDisplay ? (
        <span className="block leading-tight text-2xs opacity-70">{timeDisplay}</span>
      ) : null}
    </div>
  );
});
