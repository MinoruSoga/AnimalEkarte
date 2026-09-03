import { ChevronLeft, ChevronRight, X } from "lucide-react";
import type { Calendar } from "@/components/ui/calendar";
import { C, ICON } from "@/lib/design-tokens";
import { cn } from "@/lib/utils";

export interface SingleDatePickerProps {
  mode?: "single";
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  id?: string;
  name?: string;
  disabledDays?: React.ComponentProps<typeof Calendar>["disabled"];
}

export interface RangeDatePickerProps {
  mode: "range";
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
  id?: string;
}

export type DatePickerProps = SingleDatePickerProps | RangeDatePickerProps;

// 日付パース/フォーマット・カレンダークラス定数は DatePickerModel.ts に分離済み
// (react-refresh/only-export-components: コンポーネントファイルからの値 export 禁止)。

const MONTH_LABELS = [
  "1月",
  "2月",
  "3月",
  "4月",
  "5月",
  "6月",
  "7月",
  "8月",
  "9月",
  "10月",
  "11月",
  "12月",
];

const NAV_BTN = `inline-flex min-h-11 min-w-11 items-center justify-center rounded p-1 ${C.text50} ${C.hoverBgPrimary10} ${C.hoverText} transition-colors`;

export function CalendarNav({
  displayMonth,
  onPrev,
  onNext,
  onTitleClick,
}: {
  displayMonth: Date;
  onPrev: () => void;
  onNext: () => void;
  onTitleClick: () => void;
}) {
  const year = displayMonth.getFullYear();
  const month = displayMonth.getMonth() + 1;

  return (
    <div className="flex items-center justify-between px-3 pb-1">
      <button type="button" onClick={onPrev} className={NAV_BTN} aria-label="前の月">
        <ChevronLeft className={ICON.action} />
      </button>
      <button
        type="button"
        onClick={onTitleClick}
        className={`min-h-11 min-w-11 rounded px-2 py-1 text-sm font-medium ${C.text} ${C.hoverBgPage} transition-colors`}
      >
        {year}年 {month}月
      </button>
      <button type="button" onClick={onNext} className={NAV_BTN} aria-label="次の月">
        <ChevronRight className={ICON.action} />
      </button>
    </div>
  );
}

export function YearNav({
  year,
  onPrevYear,
  onNextYear,
}: {
  year: number;
  onPrevYear: () => void;
  onNextYear: () => void;
}) {
  return (
    <div className="flex items-center justify-between px-3 pb-2 pt-1">
      <button type="button" onClick={onPrevYear} className={NAV_BTN} aria-label="前の年">
        <ChevronLeft className={ICON.action} />
      </button>
      <span className={`text-sm font-medium ${C.text}`}>{year}年</span>
      <button type="button" onClick={onNextYear} className={NAV_BTN} aria-label="次の年">
        <ChevronRight className={ICON.action} />
      </button>
    </div>
  );
}

export function MonthGrid({
  currentMonth,
  onSelect,
}: {
  currentMonth: number;
  onSelect: (month: number) => void;
}) {
  return (
    <div className="grid grid-cols-4 gap-1 px-3 pb-3">
      {MONTH_LABELS.map((label, i) => (
        <button
          key={label}
          type="button"
          onClick={() => onSelect(i)}
          className={cn(
            "min-h-11 min-w-11 rounded px-2 py-2 text-sm transition-colors",
            i === currentMonth
              ? `${C.bgBrand} ${C.textOnBrand} font-medium`
              : `${C.text} ${C.hoverBgPage}`,
          )}
        >
          {label}
        </button>
      ))}
    </div>
  );
}

export function ClearButton({
  onClick,
}: {
  onClick: (e: React.MouseEvent<HTMLButtonElement>) => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`ml-1 -my-px min-h-11 min-w-11 shrink-0 rounded p-0.5 ${C.text40} ${C.hoverBgPrimary10} ${C.hoverText}/70 cursor-pointer`}
      aria-label="日付をクリア"
    >
      <X className={ICON.action} />
    </button>
  );
}
