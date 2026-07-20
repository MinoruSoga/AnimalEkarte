import type { ReactNode } from "react";

import Plus from "lucide-react/dist/esm/icons/plus";
import X from "lucide-react/dist/esm/icons/x";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C, ICON } from "@/lib/design-tokens";

import { SHIFT_TYPE_LABELS, type ShiftType } from "../types";
import type { TemplateFormData } from "./shift-template-form-model";

const SHIFT_TYPE_OPTIONS = (
  Object.entries(SHIFT_TYPE_LABELS) as [ShiftType, string][]
).map(([value, label]) => (
  <SelectItem key={value} value={value}>
    {label}
  </SelectItem>
));

interface ShiftTemplatePropertiesProps {
  formData: TemplateFormData;
  isTimeHidden: boolean;
  onField: <K extends keyof TemplateFormData>(key: K, value: TemplateFormData[K]) => void;
  onBreakChange: (index: number, field: "break_start" | "break_end", value: string) => void;
  onAddBreak: () => void;
  onRemoveBreak: (index: number) => void;
}

export function ShiftTemplateProperties({
  formData,
  isTimeHidden,
  onField,
  onBreakChange,
  onAddBreak,
  onRemoveBreak,
}: ShiftTemplatePropertiesProps) {
  return (
    <>
      <div className="py-1">
        <PropertyRow label="ステータス">
          <button
            type="button"
            onClick={() => onField("is_active", !formData.is_active)}
            className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
          >
            <StatusPill isActive={formData.is_active} />
          </button>
        </PropertyRow>

        <PropertyRow label="シフト種別">
          <Select
            value={formData.shift_type}
            onValueChange={(v) => onField("shift_type", v as ShiftType)}
          >
            <SelectTrigger className="h-7 text-sm border-0 shadow-none bg-transparent px-1.5">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{SHIFT_TYPE_OPTIONS}</SelectContent>
          </Select>
        </PropertyRow>

        {!isTimeHidden ? (
          <>
            <PropertyRow label="開始時刻">
              <PropInput
                type="time"
                ariaLabel="開始時刻"
                value={formData.start_time}
                onChange={(v) => onField("start_time", v)}
              />
            </PropertyRow>
            <PropertyRow label="終了時刻">
              <PropInput
                type="time"
                ariaLabel="終了時刻"
                value={formData.end_time}
                onChange={(v) => onField("end_time", v)}
              />
            </PropertyRow>
          </>
        ) : null}

        <PropertyRow label="メモ">
          <PropInput
            ariaLabel="メモ"
            value={formData.notes}
            onChange={(v) => onField("notes", v)}
            placeholder="補足情報など"
          />
        </PropertyRow>
      </div>

      {!isTimeHidden ? (
        <BreakEditor
          breaks={formData.breaks}
          onBreakChange={onBreakChange}
          onAddBreak={onAddBreak}
          onRemoveBreak={onRemoveBreak}
        />
      ) : null}
    </>
  );
}

interface BreakEditorProps {
  breaks: TemplateFormData["breaks"];
  onBreakChange: (index: number, field: "break_start" | "break_end", value: string) => void;
  onAddBreak: () => void;
  onRemoveBreak: (index: number) => void;
}

function BreakEditor({
  breaks,
  onBreakChange,
  onAddBreak,
  onRemoveBreak,
}: BreakEditorProps) {
  return (
    <div className="mt-4">
      <div className="flex items-center justify-between mb-2">
        <span className={`text-sm font-medium ${C.text}`}>休憩時間</span>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="h-7 px-2 text-xs"
          onClick={onAddBreak}
        >
          <Plus className={`${ICON.xxs} mr-1`} />
          追加
        </Button>
      </div>
      {breaks.map((b, i) => (
        <div key={i} className="flex items-center gap-2 mb-2">
          <Input
            type="time"
            aria-label={`休憩${i + 1} 開始時刻`}
            value={b.break_start}
            onChange={(e) => onBreakChange(i, "break_start", e.target.value)}
            className="flex-1 h-8 text-sm"
          />
          <span className={`text-xs ${C.text50}`}>〜</span>
          <Input
            type="time"
            aria-label={`休憩${i + 1} 終了時刻`}
            value={b.break_end}
            onChange={(e) => onBreakChange(i, "break_end", e.target.value)}
            className="flex-1 h-8 text-sm"
          />
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className="h-8 w-8 p-0"
            onClick={() => onRemoveBreak(i)}
          >
            <X className={ICON.smXs} />
          </Button>
        </div>
      ))}
      {breaks.length === 0 ? (
        <p className={`text-xs ${C.text40}`}>休憩なし</p>
      ) : null}
    </div>
  );
}

function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-xxs ${C.hoverBgLight} transition-colors min-h-[40px]`}
    >
      <div className={`w-[120px] shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}>
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
    </div>
  );
}

function PropInput({
  value,
  onChange,
  placeholder,
  type = "text",
  ariaLabel,
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  type?: string;
  ariaLabel?: string;
}) {
  return (
    <input
      type={type}
      aria-label={ariaLabel}
      className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 rounded-xxs ${C.hoverBgLight} transition-colors ${C.textPlaceholder}`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder ?? "空"}
    />
  );
}
