import { STYLE } from "@/lib/design-tokens";
import { PERIOD_LABELS, PERIOD_OPTIONS, type CashRegisterPeriod } from "../lib/constants";

export type HistoryPeriodFilter = "all" | CashRegisterPeriod;

interface CashRegisterHistoryFiltersProps {
  year: number;
  month: number;
  periodFilter: HistoryPeriodFilter;
  yearOptions: number[];
  onYearChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
  onMonthChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
  onPeriodFilterChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
}

export function CashRegisterHistoryFilters({
  year,
  month,
  periodFilter,
  yearOptions,
  onYearChange,
  onMonthChange,
  onPeriodFilterChange,
}: CashRegisterHistoryFiltersProps) {
  return (
    <div className="flex flex-wrap gap-3 items-center">
      <div>
        <label htmlFor="hist_year" className={`${STYLE.formLabel} mr-2`}>
          年
        </label>
        <select
          id="hist_year"
          value={year}
          onChange={onYearChange}
          className={`${STYLE.formInput} rounded-xs border px-3 inline-block w-auto`}
        >
          {yearOptions.map((y) => (
            <option key={y} value={y}>
              {y}年
            </option>
          ))}
        </select>
      </div>
      <div>
        <label htmlFor="hist_month" className={`${STYLE.formLabel} mr-2`}>
          月
        </label>
        <select
          id="hist_month"
          value={month}
          onChange={onMonthChange}
          className={`${STYLE.formInput} rounded-xs border px-3 inline-block w-auto`}
        >
          {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
            <option key={m} value={m}>
              {m}月
            </option>
          ))}
        </select>
      </div>
      <div>
        <label htmlFor="hist_period" className={`${STYLE.formLabel} mr-2`}>
          区分
        </label>
        <select
          id="hist_period"
          value={periodFilter}
          onChange={onPeriodFilterChange}
          className={`${STYLE.formInput} rounded-xs border px-3 inline-block w-auto`}
        >
          <option value="all">すべて</option>
          {PERIOD_OPTIONS.map((p) => (
            <option key={p} value={p}>
              {PERIOD_LABELS[p]}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
}
