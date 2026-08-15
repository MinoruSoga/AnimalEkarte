import { useCallback } from "react";
import type React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { SelectItem } from "@/components/ui/select";
import { C } from "@/lib/design-tokens";
import { todayJSTISO } from "@/lib/jst-date";
// asJsonb / toDisplayTime / toStorageTime は LineReservationSettingsFormModel.ts に分離済み
// (react-refresh/only-export-components: コンポーネントファイルからの値 export 禁止)。
import { toDisplayTime, toStorageTime } from "./line-reservation-settings-form-model";

export type BusinessHours = { start: string; end: string };
export type BreakHour = { start: string; end: string };
export type BusinessHoursByWeekday = Record<string, BusinessHours>;

export const TIME_SLOT_MODE_ITEMS = (
  <>
    <SelectItem value="minimize_gaps">空き時間を最小化</SelectItem>
    <SelectItem value="allow_gaps">空き時間を許容</SelectItem>
  </>
);

export const NO_STAFF_MODE_ITEMS = (
  <>
    <SelectItem value="first_available">最初の空き</SelectItem>
    <SelectItem value="top_priority">優先度最上位</SelectItem>
  </>
);

export const WEEKDAYS = [
  { value: "1", label: "月" },
  { value: "2", label: "火" },
  { value: "3", label: "水" },
  { value: "4", label: "木" },
  { value: "5", label: "金" },
  { value: "6", label: "土" },
  { value: "0", label: "日" },
] as const;

export function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="space-y-3">
      <h2 className={`text-sm font-semibold ${C.text} border-b ${C.borderLight} pb-2`}>
        {title}
      </h2>
      <div className="space-y-4">{children}</div>
    </div>
  );
}

export function FieldRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-1 gap-4 items-start pt-1 sm:grid-cols-[200px_1fr]">
      <Label className={`text-sm ${C.text65} pt-1`}>{label}</Label>
      <div>{children}</div>
    </div>
  );
}

interface BreakHoursEditorProps {
  value: BreakHour[];
  onChange: (value: BreakHour[]) => void;
}

export function BreakHoursEditor({ value, onChange }: BreakHoursEditorProps) {
  const handleAdd = useCallback(() => {
    onChange([...value, { start: "1200", end: "1300" }]);
  }, [value, onChange]);

  const handleRemove = useCallback(
    (idx: number) => {
      onChange(value.filter((_, i) => i !== idx));
    },
    [value, onChange],
  );

  const handleChange = useCallback(
    (idx: number, field: "start" | "end", t: string) => {
      onChange(value.map((item, i) => (i === idx ? { ...item, [field]: toStorageTime(t) } : item)));
    },
    [value, onChange],
  );

  return (
    <div className="space-y-2">
      {value.map((item, idx) => (
        <div key={idx} className="flex flex-col items-stretch gap-2 sm:flex-row sm:items-center">
          <Input
            type="time"
            aria-label={`休憩時間 ${idx + 1} 開始`}
            value={toDisplayTime(item.start)}
            onChange={(event) => handleChange(idx, "start", event.target.value)}
            className="w-full sm:max-w-[120px]"
          />
          <span className={`text-sm ${C.textMuted}`}>〜</span>
          <Input
            type="time"
            aria-label={`休憩時間 ${idx + 1} 終了`}
            value={toDisplayTime(item.end)}
            onChange={(event) => handleChange(idx, "end", event.target.value)}
            className="w-full sm:max-w-[120px]"
          />
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => handleRemove(idx)}
            className={`w-full text-xs sm:w-auto ${C.textMuted}`}
          >
            削除
          </Button>
        </div>
      ))}
      <Button type="button" variant="outline" size="sm" onClick={handleAdd}>
        + 追加
      </Button>
    </div>
  );
}

interface ClosedDatesEditorProps {
  value: string[];
  onChange: (value: string[]) => void;
}

export function ClosedDatesEditor({ value, onChange }: ClosedDatesEditorProps) {
  const handleAdd = useCallback(() => {
    const today = todayJSTISO();
    onChange([...value, today]);
  }, [value, onChange]);

  const handleRemove = useCallback(
    (idx: number) => {
      onChange(value.filter((_, i) => i !== idx));
    },
    [value, onChange],
  );

  const handleChange = useCallback(
    (idx: number, date: string) => {
      onChange(value.map((item, i) => (i === idx ? date : item)));
    },
    [value, onChange],
  );

  return (
    <div className="space-y-2">
      {value.length === 0 ? (
        <p className={`text-sm ${C.textMuted}`}>特定定休日は設定されていません</p>
      ) : (
        value.map((date, idx) => (
          <div key={idx} className="flex flex-col items-stretch gap-2 sm:flex-row sm:items-center">
            <Input
              type="date"
              aria-label={`特定定休日 ${idx + 1}`}
              value={date}
              onChange={(event) => handleChange(idx, event.target.value)}
              className="w-full sm:max-w-[160px]"
            />
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => handleRemove(idx)}
              className={`w-full text-xs sm:w-auto ${C.textMuted}`}
            >
              削除
            </Button>
          </div>
        ))
      )}
      <Button type="button" variant="outline" size="sm" onClick={handleAdd}>
        + 追加
      </Button>
    </div>
  );
}

interface WeekdayHoursEditorProps {
  defaultHours: BusinessHours;
  value: BusinessHoursByWeekday;
  onChange: (value: BusinessHoursByWeekday) => void;
}

export function WeekdayHoursEditor({
  defaultHours,
  value,
  onChange,
}: WeekdayHoursEditorProps) {
  const handleChange = useCallback(
    (day: string, field: "start" | "end", time: string) => {
      onChange({
        ...value,
        [day]: { ...(value[day] ?? defaultHours), [field]: toStorageTime(time) },
      });
    },
    [value, defaultHours, onChange],
  );

  return (
    <div className="space-y-2 pl-1">
      {WEEKDAYS.map(({ value: day, label }) => {
        const hours = value[day] ?? defaultHours;
        return (
          <div key={day} className="flex flex-col items-stretch gap-3 sm:flex-row sm:items-center">
            <span className={`text-sm w-6 text-center font-medium ${C.text65}`}>{label}</span>
            <Input
              type="time"
              aria-label={`${label}曜日 営業開始`}
              value={toDisplayTime(hours.start)}
              onChange={(event) => handleChange(day, "start", event.target.value)}
              className="w-full sm:max-w-[120px]"
            />
            <span className={`text-sm ${C.textMuted}`}>〜</span>
            <Input
              type="time"
              aria-label={`${label}曜日 営業終了`}
              value={toDisplayTime(hours.end)}
              onChange={(event) => handleChange(day, "end", event.target.value)}
              className="w-full sm:max-w-[120px]"
            />
          </div>
        );
      })}
    </div>
  );
}
