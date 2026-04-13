import { memo, useMemo, useCallback, useState } from "react";
import { isBefore, startOfDay, format } from "date-fns";
import { ja } from "date-fns/locale";
import { C, ICON } from "@/lib/design-tokens";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar as CalendarIcon, Clock, ArrowRight, ChevronsUpDown, Check } from "lucide-react";
import { cn } from "@/lib/utils";
import { useMasterItems } from "@/features/master";
import { useGetReservationTypesGrouped, useGetOnDutyStaffs } from "@/features/reservations";
import { MasterLink } from "@/components/shared/MasterLink";
import { isOneOf } from "@/lib/type-utils";
import type { Appointment } from "@/types";

const TRIGGER_CLASS =
  `h-9 text-sm bg-white ${C.borderMediumLight} ${C.text} ${C.hoverBgSubtle} transition-colors`;

const VISIT_TYPE_VALUES = ["first", "revisit"] as const;

function generateTimeOptions(): string[] {
  const times: string[] = [];
  for (let h = 0; h < 24; h++) {
    times.push(`${h}:00`);
    times.push(`${h}:15`);
    times.push(`${h}:30`);
    times.push(`${h}:45`);
  }
  return times;
}

const TIME_OPTIONS = generateTimeOptions();

interface FieldLabelProps {
  children: React.ReactNode;
  required?: boolean;
  trailing?: React.ReactNode;
}

function FieldLabel({ children, required, trailing }: FieldLabelProps) {
  return (
    <div className="flex items-center justify-between gap-2">
      <Label className={`text-[12px] ${C.text40} tracking-wide font-medium`}>
        {children}
        {required ? (
          <span style={{ color: C.danger }} className="ml-1" aria-hidden="true">*</span>
        ) : null}
      </Label>
      {trailing ? <div>{trailing}</div> : null}
    </div>
  );
}

interface ReservationFormFieldsProps {
  formData: Partial<Appointment>;
  onChange: (data: Partial<Appointment>) => void;
  validationErrors?: Record<string, string>;
  onClearError?: (field: string) => void;
  /** 定休日の日付文字列セット (YYYY-MM-DD 形式) — BUG-343 */
  holidayDates?: Set<string>;
  /** カレンダーの月が変わったときに呼ばれるコールバック (YYYY-MM 形式) — BUG-343 */
  onMonthChange?: (yearMonth: string) => void;
}

export const ReservationFormFields = memo(function ReservationFormFields({
  formData,
  onChange,
  validationErrors,
  onClearError: _onClearError,
  holidayDates,
  onMonthChange,
}: ReservationFormFieldsProps) {
  // BUG-341: グループ情報付きで取得（useMasterItems は group 情報を捨てるため専用フック使用）
  const { data: groupedReservationTypes = [] } = useGetReservationTypesGrouped();
  const [typeComboOpen, setTypeComboOpen] = useState(false);
  // 現在選択中の予約区分名（Combobox トリガー表示用）
  const selectedTypeLabel = useMemo(() => {
    for (const group of groupedReservationTypes) {
      const found = group.types.find((t) => String(t.id) === formData.type);
      if (found) return found.name;
    }
    return null;
  }, [groupedReservationTypes, formData.type]);

  const handleMonthChange = useCallback((month: Date) => {
    onMonthChange?.(format(month, "yyyy-MM"));
  }, [onMonthChange]);

  const isCalendarDateDisabled = useCallback((date: Date): boolean => {
    if (isBefore(date, startOfDay(new Date()))) return true;
    if (holidayDates) return holidayDates.has(format(date, "yyyy-MM-dd"));
    return false;
  }, [holidayDates]);
  const { data: staffItems } = useMasterItems("staff");
  // useMemo で参照を安定化（staffOptions の deps が毎レンダー新参照を受け取るのを防ぐ）
  const activeStaff = useMemo(() => staffItems.filter((s) => s.status === "active"), [staffItems]);

  // BUG-344: 選択日に出勤しているスタッフのみに絞り込む
  const selectedDateStr = formData.start ? format(formData.start, "yyyy-MM-dd") : null;
  const { data: onDutyStaffs } = useGetOnDutyStaffs(selectedDateStr);
  const staffOptions = useMemo(() => {
    if (selectedDateStr === null || onDutyStaffs === undefined) return activeStaff;
    // 出勤スタッフの ID セットで絞り込む
    const onDutyIdSet = new Set(onDutyStaffs.map((s) => String(s.id)));
    return activeStaff.filter((s) => onDutyIdSet.has(String(s.id)));
  }, [selectedDateStr, onDutyStaffs, activeStaff]);

  return (
    <div className="space-y-4">
      {/* Date + Time Group */}
      <div className={`rounded-lg border ${C.bgSubtle} p-3 space-y-3 ${C.borderMediumLight}`}>
        <div className="space-y-1.5">
          <FieldLabel required>日付</FieldLabel>
          <Popover>
            <PopoverTrigger asChild>
              <button
                type="button"
                className={cn(
                  `flex h-9 w-full items-center justify-between rounded border px-3 py-1 text-sm transition-colors ${C.borderMediumLight} ${C.text} bg-white ${C.hoverBgSubtle}`,
                  !formData.start && C.text40
                )}
              >
                <span className="flex items-center">
                  <CalendarIcon className={`mr-2 ${ICON.action}`} />
                  {formData.start ? (
                    format(formData.start, "yyyy/MM/dd (E)", { locale: ja })
                  ) : (
                    <span>日付を選択</span>
                  )}
                </span>
              </button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0">
              <Calendar
                mode="single"
                selected={formData.start}
                onSelect={(date) => {
                  if (!date) return;
                  const newStart = new Date(date);
                  const newEnd = new Date(date);

                  if (formData.start) {
                    newStart.setHours(formData.start.getHours(), formData.start.getMinutes());
                  }
                  if (formData.end) {
                    newEnd.setHours(formData.end.getHours(), formData.end.getMinutes());
                  }

                  onChange({ ...formData, start: newStart, end: newEnd });
                }}
                disabled={isCalendarDateDisabled}
                onMonthChange={handleMonthChange}
                initialFocus
              />
            </PopoverContent>
          </Popover>
          {validationErrors?.date ? (
            <FormFieldError id="res-date-error" message={validationErrors.date} />
          ) : null}
        </div>

        <div className="space-y-1.5">
          <div className={`flex items-center gap-2 text-[12px] ${C.text40} tracking-wide font-medium`}>
            <Clock className={ICON.action} />
            時間
            <span style={{ color: C.danger }} className="ml-0.5" aria-hidden="true">*</span>
          </div>
          <div className="flex items-center gap-2">
            <Select
              value={formData.start ? format(formData.start, "H:mm") : "10:00"}
              onValueChange={(v) => {
                if (!formData.start) return;
                const [h, m] = v.split(":").map(Number);
                const newStart = new Date(formData.start);
                newStart.setHours(h, m);
                onChange({ ...formData, start: newStart });
              }}
            >
              <SelectTrigger className={TRIGGER_CLASS}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="max-h-[200px]">
                {TIME_OPTIONS.map((time) => (
                  <SelectItem key={time} value={time}>
                    {time}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <ArrowRight className={`${ICON.action} ${C.text40} flex-shrink-0`} />
            <Select
              value={formData.end ? format(formData.end, "H:mm") : "11:00"}
              onValueChange={(v) => {
                if (!formData.end) return;
                const [h, m] = v.split(":").map(Number);
                const newEnd = new Date(formData.end);
                newEnd.setHours(h, m);
                onChange({ ...formData, end: newEnd });
              }}
            >
              <SelectTrigger className={TRIGGER_CLASS}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="max-h-[200px]">
                {TIME_OPTIONS.map((time) => (
                  <SelectItem key={time} value={time}>
                    {time}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          {validationErrors?.time ? (
            <FormFieldError id="res-time-error" message={validationErrors.time} />
          ) : null}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-1.5">
          <FieldLabel
            required
            trailing={
              <MasterLink
                category="reservationType"
                label="編集"
                className="text-[11px]"
              />
            }
          >
            予約区分
          </FieldLabel>
          {/* BUG-341: Combobox（検索＋グループ階層表示） */}
          <Popover open={typeComboOpen} onOpenChange={setTypeComboOpen}>
            <PopoverTrigger asChild>
              <button
                type="button"
                role="combobox"
                aria-expanded={typeComboOpen}
                className={cn(
                  `flex h-9 w-full items-center justify-between rounded border px-3 py-1 text-sm transition-colors ${C.borderMediumLight} ${C.text} bg-white ${C.hoverBgSubtle}`,
                  !selectedTypeLabel && C.text40
                )}
              >
                <span>{selectedTypeLabel ?? "選択してください"}</span>
                <ChevronsUpDown className={`${ICON.action} ${C.text40} flex-shrink-0`} />
              </button>
            </PopoverTrigger>
            <PopoverContent className="w-[--radix-popover-trigger-width] p-0 flex flex-col overflow-hidden max-h-[280px]" align="start">
              <Command className="flex flex-col overflow-hidden h-full">
                <CommandInput placeholder="予約区分を検索..." className="h-9 shrink-0" />
                <CommandList className="flex-1 overflow-y-auto min-h-0">
                  <CommandEmpty className={`py-4 text-sm ${C.text40}`}>
                    該当する予約区分がありません
                  </CommandEmpty>
                  {groupedReservationTypes.map((group) => (
                    <CommandGroup key={group.label} heading={group.label}>
                      {group.types.map((t) => (
                        <CommandItem
                          key={t.id}
                          value={`${group.label} ${t.name}`}
                          onSelect={() => {
                            onChange({ ...formData, type: String(t.id) });
                            setTypeComboOpen(false);
                          }}
                        >
                          <Check
                            className={cn(
                              `${ICON.action} flex-shrink-0`,
                              String(t.id) === formData.type ? "opacity-100" : "opacity-0"
                            )}
                          />
                          {t.name}
                        </CommandItem>
                      ))}
                    </CommandGroup>
                  ))}
                </CommandList>
              </Command>
            </PopoverContent>
          </Popover>
          {validationErrors?.type ? (
            <FormFieldError id="res-type-error" message={validationErrors.type} />
          ) : null}
        </div>
        <div className="space-y-1.5">
          <FieldLabel>来院区分</FieldLabel>
          <RadioGroup
            value={formData.visitType || ""}
            onValueChange={(v: string) => {
              if (isOneOf(v, VISIT_TYPE_VALUES)) {
                onChange({ ...formData, visitType: v });
              }
            }}
            className="flex gap-2 pt-1"
          >
            <div className="flex-1">
              <RadioGroupItem value="first" id="first" className="sr-only" />
              <Label
                htmlFor="first"
                className={cn(
                  `block h-9 rounded-full border-2 px-3 py-1.5 text-center text-sm font-medium cursor-pointer transition-colors ${C.text}`,
                  formData.visitType === "first"
                    ? `${C.borderDanger} ${C.bgDanger8}`
                    : `${C.borderMediumLight} bg-white ${C.hoverBgSubtle}`
                )}
              >
                初診
              </Label>
            </div>
            <div className="flex-1">
              <RadioGroupItem value="revisit" id="revisit" className="sr-only" />
              <Label
                htmlFor="revisit"
                className={cn(
                  `block h-9 rounded-full border-2 px-3 py-1.5 text-center text-sm font-medium cursor-pointer transition-colors ${C.text}`,
                  formData.visitType === "revisit"
                    ? `${C.borderAccent} ${C.bgAccent8}`
                    : `${C.borderMediumLight} bg-white ${C.hoverBgSubtle}`
                )}
              >
                再診
              </Label>
            </div>
          </RadioGroup>
        </div>
      </div>

      <div className="space-y-1.5">
        <FieldLabel
          trailing={
            <MasterLink
              category="staff"
              label="編集"
              className="text-[11px]"
            />
          }
        >
          担当者
        </FieldLabel>
        <Select
          value={formData.doctor || ""}
          onValueChange={(v) => onChange({ ...formData, doctor: v })}
        >
          <SelectTrigger className={TRIGGER_CLASS}>
            <SelectValue placeholder="選択してください" />
          </SelectTrigger>
          <SelectContent>
            {staffOptions.length > 0 ? (
              staffOptions.map((s) => (
                <SelectItem key={s.id} value={String(s.id)}>
                  {s.name}
                </SelectItem>
              ))
            ) : (
              <div className={`px-3 py-2 text-sm ${C.text40}`}>
                {selectedDateStr !== null
                  ? "この日に出勤しているスタッフがいません"
                  : "スタッフが登録されていません"}
              </div>
            )}
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-1.5">
        <FieldLabel>メモ</FieldLabel>
        <Textarea
          value={formData.notes || ""}
          onChange={(e) => onChange({ ...formData, notes: e.target.value })}
          placeholder="詳細や備考を入力..."
          className={`min-h-[80px] text-sm resize-none bg-white ${C.borderMedium} ${C.text} ${C.textPlaceholder} ${C.focusRingMedium}`}
        />
      </div>

      <div className="space-y-1.5">
        <FieldLabel>予約ソース</FieldLabel>
        <Select
          value={formData.source || "manual"}
          onValueChange={(v: string) => onChange({ ...formData, source: v as "manual" | "line" })}
        >
          <SelectTrigger className={TRIGGER_CLASS}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="manual">手動予約</SelectItem>
            <SelectItem value="line">LINE予約</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  );
});
