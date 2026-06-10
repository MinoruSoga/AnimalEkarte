import { useActionState, useCallback, useMemo, useState } from "react";
import {
  addDays,
  addMonths,
  endOfMonth,
  endOfWeek,
  format,
  isSameDay,
  isSameMonth,
  startOfMonth,
  startOfWeek,
  subMonths,
} from "date-fns";
import { ja } from "date-fns/locale";
import { ChevronLeft, ChevronRight, Plus, Repeat, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectTrigger, SelectValue } from "@/components/ui/select";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
import { AvailableSlotTypeSpecific, AvailableSlotTypeWeekly } from "@/types/generated/models";
import {
  useGetAvailableSlots,
  useCreateAvailableSlot,
  useDeleteAvailableSlot,
} from "../api/reservation-type-available-slots";
import { TIME_SELECT_ITEMS } from "./available-slot-options";
import type { ReservationTypeAvailableSlot } from "../api/reservation-type-available-slots";

const DAYS_OF_WEEK = ["日", "月", "火", "水", "木", "金", "土"] as const;

// 予約管理ページ MonthView と同一の曜日ヘッダー行
const HEADER_ROW = (
  <div className={`grid grid-cols-7 border-b ${C.borderMedium} ${C.bgPage}`}>
    {DAYS_OF_WEEK.map((d, i) => (
      <div
        key={d}
        className={`py-3 text-sm font-bold text-center ${
          i === 0 ? C.danger : i === 6 ? C.accent : C.text60
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
  /** テスト用に初期表示月を固定できる（省略時は当月） */
  initialMonth?: Date;
}

export function ReservationTypeAvailableSlotsCalendar({
  clinicId,
  reservationTypeId,
  initialMonth,
}: Props) {
  const { data: items = [] } = useGetAvailableSlots(clinicId, reservationTypeId);
  const createMutation = useCreateAvailableSlot(clinicId, reservationTypeId);
  const deleteMutation = useDeleteAvailableSlot(clinicId, reservationTypeId);

  const [currentMonth, setCurrentMonth] = useState<Date>(() => initialMonth ?? toJSTWallDate(new Date()));
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

  const handlePrevMonth = useCallback(() => setCurrentMonth((m) => subMonths(m, 1)), []);
  const handleNextMonth = useCallback(() => setCurrentMonth((m) => addMonths(m, 1)), []);
  const handleToday = useCallback(() => setCurrentMonth(toJSTWallDate(new Date())), []);

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

  const handleDelete = useCallback(
    (id: number) => {
      deleteMutation.mutate(id);
    },
    [deleteMutation],
  );

  const weeks = useMemo(() => {
    const monthStart = startOfMonth(currentMonth);
    const gridStart = startOfWeek(monthStart, { locale: ja });
    const gridEnd = endOfWeek(endOfMonth(monthStart), { locale: ja });

    const result: Date[][] = [];
    let day = gridStart;
    while (day <= gridEnd) {
      const week: Date[] = [];
      for (let i = 0; i < 7; i++) {
        week.push(day);
        day = addDays(day, 1);
      }
      result.push(week);
    }
    return result;
  }, [currentMonth]);

  const today = toJSTWallDate(new Date());

  const selectedWeeklySlots = selectedDate
    ? weeklyByDow.get(new Date(`${selectedDate}T00:00:00`).getDay()) ?? []
    : [];
  const selectedSpecificSlots = selectedDate ? specificByDate.get(selectedDate) ?? [] : [];

  return (
    <div className="flex-1 min-h-0 flex flex-col gap-3">
      {/* 予約管理ページと同一構成のナビゲーションバー */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-4">
          <div className={`flex items-center ${C.bgWhite} rounded-md border ${C.borderMedium} p-1 shadow-sm`}>
            <Button variant="ghost" size="icon" className="h-10 w-10" onClick={handlePrevMonth} aria-label="前の月">
              <ChevronLeft className={ICON.page} />
            </Button>
            <Button variant="ghost" size="sm" className="h-10 px-4 text-base font-medium" onClick={handleToday}>
              今日
            </Button>
            <Button variant="ghost" size="icon" className="h-10 w-10" onClick={handleNextMonth} aria-label="次の月">
              <ChevronRight className={ICON.page} />
            </Button>
          </div>
          <h2 className={`text-xl font-bold ${C.text}`}>
            {format(currentMonth, "yyyy年 M月", { locale: ja })}
          </h2>
        </div>

        {/* 凡例（予約管理ページの予約区分凡例と同形式） */}
        <div className="flex items-center gap-3 flex-wrap">
          <div className="flex items-center gap-1.5">
            <Repeat className={`${ICON.smXs} ${C.text50}`} />
            <span className={`text-base ${C.text60}`}>毎週の枠</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className={`${ICON.dotMd} rounded-full ${C.bgAccent}`} />
            <span className={`text-base ${C.text60}`}>特定日の枠</span>
          </div>
        </div>
      </div>

      {/* 予約管理 MonthView と同一スタイルの月グリッド */}
      <div className={`flex-1 min-h-0 overflow-y-auto flex flex-col border-l border-t ${C.borderMedium} rounded-lg ${C.bgWhite} shadow-sm`}>
        {HEADER_ROW}
        <div className="flex-1 flex flex-col">
          {weeks.map((week) => (
            <div className="grid grid-cols-7 flex-1" key={week[0].toISOString()}>
              {week.map((day) => {
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
                    aria-label={format(day, "yyyy年M月d日", { locale: ja })}
                    className={`h-full min-h-[120px] text-left border-b border-r ${C.borderLight} p-2 transition-colors cursor-pointer flex flex-col
                      ${isSelected ? C.bgAccent8 : `${C.bgWhite} ${C.hoverBgPage}`}
                      ${!isSameMonth(day, currentMonth) ? `${C.bgPage30} ${C.text30}` : C.text}
                    `}
                  >
                    <div className="flex justify-between items-start mb-2">
                      <span
                        className={`text-base font-bold size-7 flex items-center justify-center rounded-full ${
                          isSameDay(day, today) ? `${C.bgAccent} ${C.textWhite} shadow-sm` : ""
                        }`}
                      >
                        {format(day, "d")}
                      </span>
                    </div>
                    <div className="space-y-1.5 flex-1 overflow-hidden">
                      {chips.slice(0, 4).map(({ slot, weekly }) => (
                        <div
                          key={slot.id}
                          className={`text-sm px-2 py-1.5 rounded border leading-tight flex items-center gap-1 tabular-nums ${
                            weekly
                              ? `${C.bgPage} ${C.text50} ${C.borderLight}`
                              : `${C.bgAccentLight} ${C.textAccentDark} ${C.borderLight}`
                          }`}
                        >
                          {weekly ? <Repeat className="size-3 shrink-0" /> : null}
                          {slot.startTime}
                        </div>
                      ))}
                      {chips.length > 4 ? (
                        <div className={`text-sm ${C.text60} pl-1`}>他 {chips.length - 4} 件</div>
                      ) : null}
                    </div>
                  </button>
                );
              })}
            </div>
          ))}
        </div>
      </div>

      {/* 日別編集パネル */}
      <div className={`shrink-0 ${C.bgWhite} rounded-md border ${C.borderMedium} p-3 shadow-sm`}>
        {selectedDate === null ? (
          <p className={`text-sm ${C.text40}`}>日付をクリックすると、その日の枠を編集できます</p>
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
              </div>
            ) : null}

            {selectedSpecificSlots.length > 0 ? (
              <div className="flex items-center gap-1.5 flex-wrap">
                <span className={`text-xs ${C.text50}`}>この日の枠:</span>
                {selectedSpecificSlots.map((slot) => (
                  <span
                    key={slot.id}
                    className={`flex items-center gap-1 text-sm px-2 py-1 rounded border ${C.borderLight} ${C.bgAccentLight} ${C.textAccentDark} tabular-nums`}
                  >
                    {slot.startTime}
                    <button
                      type="button"
                      onClick={() => handleDelete(slot.id)}
                      aria-label={`${slot.startTime}の枠を削除`}
                      className={`${C.text40} ${C.hoverTextDanger} transition-colors`}
                    >
                      <Trash2 className="size-3" />
                    </button>
                  </span>
                ))}
              </div>
            ) : (
              <p className={`text-xs ${C.text40}`}>この日の特定日枠はありません</p>
            )}

            <form action={formAction} className="flex items-center gap-2 ml-auto">
              <Select value={startTime} onValueChange={setStartTime}>
                <SelectTrigger className={STYLE.selectCompact}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>{TIME_SELECT_ITEMS}</SelectContent>
              </Select>
              <SubmitButton loadingText="追加中..." className="h-8 text-sm px-3">
                <Plus className={ICON.smXs} />
                追加
              </SubmitButton>
            </form>
          </div>
        )}
      </div>
    </div>
  );
}
