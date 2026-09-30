import { useActionState, useCallback, useMemo, useState } from "react";
import { useNavigate } from "react-router";
import { addDays, addWeeks, format, isSameDay, startOfWeek, subWeeks } from "date-fns";
import { ja } from "date-fns/locale";
import { Plus, Repeat, Trash2 } from "lucide-react";
import { Select, SelectContent, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { CalendarNavToolbar } from "@/components/shared/CalendarNavToolbar";
import { FieldHelp } from "@/components/shared/FieldHelp";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { SUPPORT_WIDGET_LAYOUT } from "@/constants/support-widget-layout";
import { toJSTWallDate } from "@/lib/jst-date";
import { DISPLAY_DATE_FORMAT } from "@/lib/format/date";
import { paths } from "@/config/paths";
import { AvailableSlotTypeSpecific, AvailableSlotTypeWeekly } from "@/types/generated/models";
import {
  useGetAvailableSlots,
  useCreateAvailableSlot,
  useDeleteAvailableSlot,
} from "../api/reservation-type-available-slots";
import { TIME_SELECT_ITEMS } from "./AvailableSlotOptions";
import type { ReservationTypeAvailableSlot } from "../api/reservation-type-available-slots";

const DAYS_OF_WEEK = ["月", "火", "水", "木", "金", "土", "日"] as const;

const HEADER_ROW = (
  <div className={`grid grid-cols-7 border-b ${C.borderMedium} ${C.bgPage}`}>
    {DAYS_OF_WEEK.map((d, i) => (
      <div
        key={d}
        className={`py-3 text-sm font-bold text-center ${
          i === 5 ? C.textBrand : i === 6 ? C.textNotionRed : C.text60
        }`}
      >
        {d}
      </div>
    ))}
  </div>
);

const DEFAULT_START_TIME = "09:00";

function byStartTime(a: ReservationTypeAvailableSlot, b: ReservationTypeAvailableSlot) {
  return a.startTime.localeCompare(b.startTime);
}

interface Props {
  clinicId: string;
  reservationTypeId: string;
  /** テスト用に初期表示週を固定できる（省略時は当週） */
  initialMonth?: Date;
}

export function ReservationTypeAvailableSlotsCalendar({
  clinicId,
  reservationTypeId,
  initialMonth,
}: Props) {
  const navigate = useNavigate();
  const { data: items = [] } = useGetAvailableSlots(clinicId, reservationTypeId);
  const createMutation = useCreateAvailableSlot(clinicId, reservationTypeId);
  const deleteMutation = useDeleteAvailableSlot(clinicId, reservationTypeId);
  const { mutate } = deleteMutation;

  const [currentWeek, setCurrentWeek] = useState<Date>(
    () => initialMonth ?? toJSTWallDate(new Date()),
  );
  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const [startTime, setStartTime] = useState(DEFAULT_START_TIME);

  const { specificByDate, weeklyByDow } = useMemo(() => {
    const byDate = new Map<string, ReservationTypeAvailableSlot[]>();
    const byDow = new Map<number, ReservationTypeAvailableSlot[]>();
    for (const item of items) {
      if (!item.isActive) continue;
      if (item.availableType === AvailableSlotTypeSpecific && item.specificDate) {
        const key = item.specificDate.slice(0, 10);
        byDate.set(key, [...(byDate.get(key) ?? []), item]);
      } else if (item.availableType === AvailableSlotTypeWeekly && item.dayOfWeek !== undefined) {
        byDow.set(item.dayOfWeek, [...(byDow.get(item.dayOfWeek) ?? []), item]);
      }
    }
    for (const list of byDate.values()) {
      list.sort(byStartTime);
    }
    for (const list of byDow.values()) {
      list.sort(byStartTime);
    }
    return { specificByDate: byDate, weeklyByDow: byDow };
  }, [items]);

  const handlePrevWeek = useCallback(() => {
    setCurrentWeek((w) => subWeeks(w, 1));
    setSelectedDate(null);
  }, []);
  const handleNextWeek = useCallback(() => {
    setCurrentWeek((w) => addWeeks(w, 1));
    setSelectedDate(null);
  }, []);
  const handleToday = useCallback(() => {
    setCurrentWeek(toJSTWallDate(new Date()));
    setSelectedDate(null);
  }, []);

  const [, formAction] = useActionState(async () => {
    if (!selectedDate) return null;
    try {
      await createMutation.mutateAsync({
        available_type: AvailableSlotTypeSpecific,
        start_time: startTime,
        is_active: true,
        specific_date: selectedDate,
      });
    } catch {
      // エラー通知は useCreateAvailableSlot の onError で表示済み
    }
    return null;
  }, null);

  // 破壊的削除は ConfirmDialog 経由（直行削除禁止）
  const [pendingDelete, setPendingDelete] = useState<ReservationTypeAvailableSlot | null>(null);

  const handleDeleteRequest = useCallback((slot: ReservationTypeAvailableSlot) => {
    setPendingDelete(slot);
  }, []);

  const handleDeleteCancel = useCallback(() => {
    setPendingDelete(null);
  }, []);

  const handleDeleteConfirm = useCallback(() => {
    if (pendingDelete === null) return;
    mutate(pendingDelete.id);
    setPendingDelete(null);
  }, [mutate, pendingDelete]);

  const weekDays = useMemo(() => {
    const weekStart = startOfWeek(currentWeek, { weekStartsOn: 1 });
    return Array.from({ length: 7 }, (_, index) => addDays(weekStart, index));
  }, [currentWeek]);

  const weekLabel = useMemo(() => {
    const weekStart = weekDays[0];
    const weekEnd = weekDays[6];
    if (!weekStart || !weekEnd) return "";
    return `${format(weekStart, DISPLAY_DATE_FORMAT, { locale: ja })} - ${format(weekEnd, DISPLAY_DATE_FORMAT, { locale: ja })}`;
  }, [weekDays]);

  const today = toJSTWallDate(new Date());

  const selectedWeeklySlots = selectedDate
    ? (weeklyByDow.get(new Date(`${selectedDate}T00:00:00`).getDay()) ?? [])
    : [];
  const selectedSpecificSlots = selectedDate ? (specificByDate.get(selectedDate) ?? []) : [];

  const isDuplicate = useMemo(() => {
    if (!selectedDate) return false;
    const slots = specificByDate.get(selectedDate) ?? [];
    return slots.some((s) => s.startTime === startTime);
  }, [selectedDate, specificByDate, startTime]);

  return (
    <div className="flex-1 min-h-0 flex flex-col gap-3">
      {/* 予約管理ページと同一構成のナビゲーションバー */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <CalendarNavToolbar
          onPrev={handlePrevWeek}
          onToday={handleToday}
          onNext={handleNextWeek}
          prevAriaLabel="前の週"
          nextAriaLabel="次の週"
          label={<h2 className={`text-xl font-bold ${C.text}`}>{weekLabel}</h2>}
        />

        {/* 凡例（予約管理ページの予約区分凡例と同形式） */}
        <div className="flex items-center gap-3 flex-wrap">
          <div className="flex items-center gap-1.5">
            <Repeat className={`${ICON.smXs} ${C.text50}`} />
            <span className={`text-base ${C.text60}`}>毎週の枠</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className={`${ICON.dotMd} rounded-full ${C.bgBrandDot}`} />
            <span className={`text-base ${C.text60}`}>特定日の枠</span>
          </div>
        </div>
      </div>

      {/* 予約管理画面と同じ密度の週グリッド */}
      <div
        data-testid="reservation-slots-calendar-scroll"
        className={`flex-1 min-h-0 min-w-0 max-w-full overflow-x-auto overflow-y-auto relative border-l border-t ${C.borderMedium} rounded-lg ${C.bgWhite} ${SUPPORT_WIDGET_LAYOUT.scrollBottomClearance}`}
      >
        {/* 7日 × 44px。狭幅ではこのcalendar内だけ横scrollし、document overflowを発生させない。 */}
        <div className="flex min-h-full w-full min-w-[308px] flex-col">
          {HEADER_ROW}
          <div className="grid grid-cols-7 flex-1 min-h-[420px]">
            {weekDays.map((day) => {
              const dateKey = format(day, "yyyy-MM-dd");
              const weeklySlots = weeklyByDow.get(day.getDay()) ?? [];
              const specificSlots = specificByDate.get(dateKey) ?? [];
              const chips = [
                ...weeklySlots.map((slot) => ({ slot, weekly: true })),
                ...specificSlots.map((slot) => ({ slot, weekly: false })),
              ].sort((a, b) => byStartTime(a.slot, b.slot));
              const isSelected = selectedDate === dateKey;

              return (
                <button
                  type="button"
                  key={dateKey}
                  onClick={() => setSelectedDate(dateKey)}
                  aria-label={format(day, DISPLAY_DATE_FORMAT, { locale: ja })}
                  aria-pressed={isSelected}
                  className={`h-full min-h-[420px] min-w-11 text-left border-b border-r ${C.borderLight} p-3 transition-colors cursor-pointer flex flex-col
                    ${isSelected ? C.bgBrand8 : `${C.bgWhite} ${C.hoverBgPage}`}
                  `}
                >
                  <div className="flex justify-between items-start mb-3">
                    <div className="min-w-0">
                      <span
                        className={`text-base font-bold size-8 flex items-center justify-center rounded-full ${
                          isSameDay(day, today)
                            ? `${C.bgActionPrimarySolid} ${C.textOnActionPrimary}`
                            : C.text
                        }`}
                      >
                        {format(day, "d")}
                      </span>
                      <div className={`text-xs ${C.text50} mt-1`}>
                        {format(day, "M月", { locale: ja })}
                      </div>
                    </div>
                    <span className={`text-xs ${C.text60}`}>{chips.length}件</span>
                  </div>
                  <div className="space-y-1.5 flex-1 overflow-hidden">
                    {chips.map(({ slot, weekly }) => (
                      <div
                        key={slot.id}
                        className={`text-sm px-2 py-1.5 rounded border leading-tight flex items-center gap-1 tabular-nums ${
                          weekly
                            ? `${C.bgPage} ${C.text50} ${C.borderLight}`
                            : `${C.bgBrandLight} ${C.textBrandDark} ${C.borderLight}`
                        }`}
                      >
                        {weekly ? <Repeat className="size-3 shrink-0" /> : null}
                        {slot.startTime}
                      </div>
                    ))}
                  </div>
                </button>
              );
            })}
          </div>
        </div>
      </div>

      {/* 日別編集パネル */}
      <div className={`shrink-0 ${C.bgWhite} rounded-md border ${C.borderMedium} p-3`}>
        {selectedDate === null ? (
          <p className={`text-sm ${C.text60}`}>日付をクリックすると、その日の枠を編集できます</p>
        ) : (
          <div className="flex flex-wrap items-center gap-x-6 gap-y-2">
            <p className={`text-base font-bold ${C.text}`}>
              {format(new Date(`${selectedDate}T00:00:00`), "M月d日（E）", { locale: ja })}
            </p>

            {selectedWeeklySlots.length > 0 ? (
              <div className="flex items-center gap-1.5 flex-wrap">
                <span className={`text-xs ${C.text50}`}>毎週の枠（リストで管理）:</span>
                {selectedWeeklySlots.map((slot) => (
                  <span
                    key={slot.id}
                    className={`flex items-center gap-1 text-sm px-2 py-1 rounded border ${C.borderLight} ${C.bgPage} ${C.text50} tabular-nums`}
                  >
                    <Repeat className="size-3 shrink-0" />
                    {slot.startTime}
                  </span>
                ))}
                <button
                  type="button"
                  onClick={() =>
                    navigate(
                      `${paths.settings.reservationType.getHref()}?typeId=${reservationTypeId}`,
                    )
                  }
                  className={`inline-flex items-center min-h-11 text-xs ${C.text60} ${C.hoverTextBrand} transition-colors`}
                >
                  毎週枠は予約区分マスタで編集 →
                </button>
              </div>
            ) : null}

            {selectedSpecificSlots.length > 0 ? (
              <div className="flex items-center gap-1.5 flex-wrap">
                <span className={`text-xs ${C.text50}`}>この日の枠:</span>
                {selectedSpecificSlots.map((slot) => (
                  <span
                    key={slot.id}
                    className={`flex items-center gap-1 text-sm px-2 py-1 rounded border ${C.borderLight} ${C.bgBrandLight} ${C.textBrandDark} tabular-nums`}
                  >
                    {slot.startTime}
                    <button
                      type="button"
                      onClick={() => handleDeleteRequest(slot)}
                      aria-label={`${slot.startTime}の枠を削除`}
                      className={`-m-3.5 flex min-h-11 min-w-11 items-center justify-center rounded-xxs outline-none ${C.text50} ${C.hoverTextDanger} focus-visible:ring-2 ${C.focusRingAccent40} transition-colors`}
                    >
                      <Trash2 className="size-3" />
                    </button>
                  </span>
                ))}
              </div>
            ) : (
              <p className={`text-xs ${C.text60}`}>この日の特定日枠はありません</p>
            )}

            <form action={formAction} className="flex items-center gap-2 ml-auto flex-wrap">
              <FieldHelp label="開始時刻" content="この日に追加する予約枠の開始時刻です。" />
              <Select value={startTime} onValueChange={setStartTime}>
                <SelectTrigger className={STYLE.selectCompact} aria-label="開始時刻">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>{TIME_SELECT_ITEMS}</SelectContent>
              </Select>
              {isDuplicate ? (
                <p className={`text-xs ${C.text60}`}>この時刻は既に登録済みです</p>
              ) : null}
              <SubmitButton
                colorVariant="primary"
                disabled={isDuplicate}
                loadingText="追加中..."
                className="h-8 text-sm px-3"
              >
                <Plus className={ICON.smXs} />
                追加
              </SubmitButton>
            </form>
          </div>
        )}
      </div>

      <ConfirmDialog
        open={pendingDelete !== null}
        onClose={handleDeleteCancel}
        onConfirm={handleDeleteConfirm}
        title="予約可能枠を削除しますか？"
        description={
          pendingDelete === null
            ? undefined
            : `「${pendingDelete.specificDate?.slice(0, 10) ?? ""} ${pendingDelete.startTime}」の予約可能枠を削除します。この操作は取り消せません。`
        }
        confirmLabel="削除"
        variant="destructive"
        isPending={deleteMutation.isPending}
      />
    </div>
  );
}
