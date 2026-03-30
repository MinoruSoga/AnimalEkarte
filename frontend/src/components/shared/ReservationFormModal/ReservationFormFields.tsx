import { ICON } from "@/lib/design-tokens";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { FormFieldError } from "@/components/shared/FormFieldError";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Calendar as CalendarIcon, Clock, ArrowRight } from "lucide-react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { cn } from "@/lib/utils";
import { useMasterItems } from "@/hooks/use-master-items";
import { MasterLink } from "@/components/shared/MasterLink";
import { isOneOf } from "@/lib/type-utils";
import type { ReservationAppointment } from "@/types";

const TRIGGER_CLASS =
  "h-9 text-sm bg-white border-[rgba(55,53,47,0.12)] text-[#37352F] hover:bg-[#FAFAF8] transition-colors";

const VISIT_TYPE_VALUES = ["first", "revisit"] as const;

function generateTimeOptions(): string[] {
  const times: string[] = [];
  for (let h = 0; h < 24; h++) {
    times.push(`${h}:00`);
    times.push(`${h}:30`);
  }
  return times;
}

const TIME_OPTIONS = generateTimeOptions();

interface FieldLabelProps {
  children: React.ReactNode;
  trailing?: React.ReactNode;
}

function FieldLabel({ children, trailing }: FieldLabelProps) {
  return (
    <div className="flex items-center justify-between gap-2">
      <Label className="text-[12px] text-[#37352F]/40 tracking-wide font-medium">
        {children}
      </Label>
      {trailing ? <div>{trailing}</div> : null}
    </div>
  );
}

interface ReservationFormFieldsProps {
  formData: Partial<ReservationAppointment>;
  onChange: (data: Partial<ReservationAppointment>) => void;
  validationErrors?: Record<string, string>;
  onClearError?: (field: string) => void;
}

export function ReservationFormFields({
  formData,
  onChange,
  validationErrors,
  onClearError: _onClearError,
}: ReservationFormFieldsProps) {
  const { data: serviceTypes } = useMasterItems("serviceType");
  const { data: staffItems } = useMasterItems("staff");
  const activeStaff = staffItems.filter((s) => s.status === "active");

  return (
    <div className="space-y-4">
      {/* Date + Time Group */}
      <div className="rounded-lg border bg-[#FAFAF8] p-3 space-y-3 border-[rgba(55,53,47,0.12)]">
        <div className="space-y-1.5">
          <FieldLabel>日付</FieldLabel>
          <Popover>
            <PopoverTrigger asChild>
              <button
                type="button"
                className={cn(
                  "flex h-9 w-full items-center justify-between rounded border px-3 py-1 text-sm transition-colors border-[rgba(55,53,47,0.12)] text-[#37352F] bg-white hover:bg-[#FAFAF8]",
                  !formData.start && "text-muted-foreground"
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
                initialFocus
              />
            </PopoverContent>
          </Popover>
          {validationErrors?.date ? (
            <FormFieldError id="res-date-error" message={validationErrors.date} />
          ) : null}
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center gap-2 text-[12px] text-[#37352F]/40 tracking-wide font-medium">
            <Clock className={ICON.action} />
            時間
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
            <ArrowRight className={`${ICON.action} text-[#37352F]/40 flex-shrink-0`} />
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
            trailing={
              <MasterLink
                category="serviceType"
                label="編集"
                className="text-[11px]"
              />
            }
          >
            診療サービス
          </FieldLabel>
          <Select
            value={formData.type || ""}
            onValueChange={(v: string) => onChange({ ...formData, type: v })}
          >
            <SelectTrigger className={TRIGGER_CLASS}>
              <SelectValue placeholder="選択してください" />
            </SelectTrigger>
            <SelectContent>
              {serviceTypes.map((item) => (
                <SelectItem key={item.id} value={String(item.id)}>
                  {item.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {validationErrors?.type ? (
            <FormFieldError id="res-type-error" message={validationErrors.type} />
          ) : null}
        </div>
        <div className="space-y-1.5">
          <FieldLabel>予約区分</FieldLabel>
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
                  "block h-9 rounded-full border-2 px-3 py-1.5 text-center text-sm font-medium cursor-pointer transition-colors text-[#37352F]",
                  formData.visitType === "first"
                    ? "border-red-600 bg-red-50"
                    : "border-[rgba(55,53,47,0.12)] bg-white hover:bg-[#FAFAF8]"
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
                  "block h-9 rounded-full border-2 px-3 py-1.5 text-center text-sm font-medium cursor-pointer transition-colors text-[#37352F]",
                  formData.visitType === "revisit"
                    ? "border-blue-600 bg-blue-50"
                    : "border-[rgba(55,53,47,0.12)] bg-white hover:bg-[#FAFAF8]"
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
            {activeStaff.length > 0 ? (
              activeStaff.map((s) => (
                <SelectItem key={s.id} value={String(s.id)}>
                  {s.name}
                </SelectItem>
              ))
            ) : (
              <>
                <SelectItem value="医師A">医師A</SelectItem>
                <SelectItem value="医師B">医師B</SelectItem>
                <SelectItem value="スタッフA">スタッフA</SelectItem>
              </>
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
          className="min-h-[80px] text-sm resize-none bg-white border-[rgba(55,53,47,0.16)] text-[#37352F] placeholder:text-[#37352F]/25 focus-visible:ring-[rgba(55,53,47,0.16)]"
        />
      </div>
    </div>
  );
}
