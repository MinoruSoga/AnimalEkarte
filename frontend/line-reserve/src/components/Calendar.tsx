import { useState, useCallback, useMemo } from "react";
import type { AvailableDate } from "../types/models";
import { addDaysISO, getJSTToday, parseISODate } from "@/shared-liff/jst-date";

interface CalendarProps {
  availableDates: AvailableDate[];
  selectedDate: string;
  onSelect: (date: string) => void;
  bookingWindow: number; // 今日から何日後まで予約可能
}

const WEEK_DAYS = ["日", "月", "火", "水", "木", "金", "土"] as const;

function formatDate(year: number, month: number, day: number): string {
  const m = String(month + 1).padStart(2, "0");
  const d = String(day).padStart(2, "0");
  return `${year}-${m}-${d}`;
}

export function Calendar({ availableDates, selectedDate, onSelect, bookingWindow }: CalendarProps) {
  const today = useMemo(() => {
    return getJSTToday();
  }, []);

  const maxDate = useMemo(() => {
    return parseISODate(
      addDaysISO(
        formatDate(today.getUTCFullYear(), today.getUTCMonth(), today.getUTCDate()),
        bookingWindow,
      ),
    );
  }, [today, bookingWindow]);

  const [viewYear, setViewYear] = useState<number>(today.getUTCFullYear());
  const [viewMonth, setViewMonth] = useState<number>(today.getUTCMonth()); // 0-based

  const availableSet = useMemo(() => {
    return new Set(availableDates.filter((d) => d.available).map((d) => d.date));
  }, [availableDates]);

  const reasonMap = useMemo(() => {
    const map = new Map<string, string>();
    const REASON_LABELS: Record<string, string> = {
      closed: "休診日",
      holiday: "祝日休診",
      staff_off: "スタッフ不在",
      no_slots: "満席",
    };
    for (const d of availableDates) {
      // 同一日に available:true と staff_off が混在しても、選択可能な日には理由を載せない
      if (!d.available && d.reason && !availableSet.has(d.date)) {
        map.set(d.date, REASON_LABELS[d.reason] ?? "予約不可");
      }
    }
    return map;
  }, [availableDates, availableSet]);

  const daysInMonth = useMemo(() => {
    return new Date(Date.UTC(viewYear, viewMonth + 1, 0)).getUTCDate();
  }, [viewYear, viewMonth]);

  const firstDayOfWeek = useMemo(() => {
    return new Date(Date.UTC(viewYear, viewMonth, 1)).getUTCDay(); // 0=日
  }, [viewYear, viewMonth]);

  const canGoPrev = useMemo(() => {
    const prevMonth = new Date(Date.UTC(viewYear, viewMonth - 1, 1));
    const todayMonthStart = new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), 1));
    return prevMonth >= todayMonthStart;
  }, [viewYear, viewMonth, today]);

  const canGoNext = useMemo(() => {
    const nextMonthStart = new Date(Date.UTC(viewYear, viewMonth + 1, 1));
    return nextMonthStart <= maxDate;
  }, [viewYear, viewMonth, maxDate]);

  const handlePrev = useCallback(() => {
    if (!canGoPrev) return;
    if (viewMonth === 0) {
      setViewYear((y) => y - 1);
      setViewMonth(11);
    } else {
      setViewMonth((m) => m - 1);
    }
  }, [canGoPrev, viewMonth]);

  const handleNext = useCallback(() => {
    if (!canGoNext) return;
    if (viewMonth === 11) {
      setViewYear((y) => y + 1);
      setViewMonth(0);
    } else {
      setViewMonth((m) => m + 1);
    }
  }, [canGoNext, viewMonth]);

  const cells = useMemo(() => {
    const result: Array<{ day: number | null; dateStr: string | null }> = [];
    for (let i = 0; i < firstDayOfWeek; i++) {
      result.push({ day: null, dateStr: null });
    }
    for (let day = 1; day <= daysInMonth; day++) {
      result.push({ day, dateStr: formatDate(viewYear, viewMonth, day) });
    }
    return result;
  }, [firstDayOfWeek, daysInMonth, viewYear, viewMonth]);

  return (
    <div className="bg-white rounded-lg border border-noah-border">
      {/* ヘッダー */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-noah-border">
        <button
          type="button"
          onClick={handlePrev}
          disabled={!canGoPrev}
          className="p-1 rounded disabled:opacity-30 hover:bg-noah-teal-light"
          aria-label="前月"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            viewBox="0 0 20 20"
            fill="currentColor"
            aria-hidden="true"
          >
            <path
              fillRule="evenodd"
              d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z"
              clipRule="evenodd"
            />
          </svg>
        </button>
        <span className="font-semibold text-noah-teal-dark">
          {viewYear}年{viewMonth + 1}月
        </span>
        <button
          type="button"
          onClick={handleNext}
          disabled={!canGoNext}
          className="p-1 rounded disabled:opacity-30 hover:bg-noah-teal-light"
          aria-label="翌月"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            viewBox="0 0 20 20"
            fill="currentColor"
            aria-hidden="true"
          >
            <path
              fillRule="evenodd"
              d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
              clipRule="evenodd"
            />
          </svg>
        </button>
      </div>

      {/* 曜日ヘッダー */}
      <div className="grid grid-cols-7 border-b border-noah-border">
        {WEEK_DAYS.map((day, idx) => (
          <div
            key={day}
            className={`text-center text-xs py-2 font-medium ${
              idx === 0 ? "text-noah-danger" : idx === 6 ? "text-noah-teal" : "text-noah-text-sub"
            }`}
          >
            {day}
          </div>
        ))}
      </div>

      {/* 日付グリッド */}
      <div className="grid grid-cols-7">
        {cells.map((cell, idx) => {
          if (cell.day === null || cell.dateStr === null) {
            return <div key={`empty-${idx}`} className="aspect-square" />;
          }

          const dateStr = cell.dateStr;
          const dayOfWeek = (firstDayOfWeek + cell.day - 1) % 7;
          const isAvailable = availableSet.has(dateStr);
          const isSelected = dateStr === selectedDate;
          const cellDate = new Date(Date.UTC(viewYear, viewMonth, cell.day));
          const isPast = cellDate < today;
          const isBeyondWindow = cellDate > maxDate;
          const isDisabled = !isAvailable || isPast || isBeyondWindow;

          const reasonLabel = isDisabled ? reasonMap.get(dateStr) : undefined;
          const ariaLabel = `${viewYear}年${viewMonth + 1}月${cell.day}日${reasonLabel ? `（${reasonLabel}）` : isDisabled ? "（予約不可）" : ""}`;

          return (
            <div key={dateStr} className="relative group">
              <button
                type="button"
                onClick={() => !isDisabled && onSelect(dateStr)}
                disabled={isDisabled}
                title={reasonLabel}
                className={`w-full aspect-square flex items-center justify-center text-sm transition-colors ${
                  isSelected
                    ? "bg-noah-teal text-white rounded-full mx-1 my-1"
                    : isDisabled
                      ? "text-noah-disabled cursor-not-allowed"
                      : dayOfWeek === 0
                        ? "text-noah-danger hover:bg-noah-danger-bg"
                        : dayOfWeek === 6
                          ? "text-noah-teal hover:bg-noah-teal-light"
                          : "text-noah-text hover:bg-noah-teal-light"
                }`}
                aria-label={ariaLabel}
              >
                {cell.day}
              </button>
              {reasonLabel ? (
                <span className="pointer-events-none absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:block whitespace-nowrap rounded bg-noah-tooltip-bg px-2 py-0.5 text-xs text-white z-10">
                  {reasonLabel}
                </span>
              ) : null}
            </div>
          );
        })}
      </div>
    </div>
  );
}
