import type { ReactNode } from "react";

import Plus from "lucide-react/dist/esm/icons/plus";
import X from "lucide-react/dist/esm/icons/x";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { FieldHelp } from "@/components/shared/FieldHelp";
import { C, ICON } from "@/lib/design-tokens";

import { SHIFT_TYPE_LABELS, type ShiftType } from "../types";
import type { TemplateFormData } from "../lib/shift-template-form-model";

const SHIFT_TYPE_OPTIONS = (Object.entries(SHIFT_TYPE_LABELS) as [ShiftType, string][]).map(
  ([value, label]) => (
    <SelectItem key={value} value={value}>
      {label}
    </SelectItem>
  ),
);

interface ShiftTemplatePropertiesProps {
  formData: TemplateFormData;
  isTimeHidden: boolean;
  readOnly: boolean;
  onField: <K extends keyof TemplateFormData>(key: K, value: TemplateFormData[K]) => void;
  onBreakChange: (index: number, field: "break_start" | "break_end", value: string) => void;
  onAddBreak: () => void;
  onRemoveBreak: (index: number) => void;
}

export function ShiftTemplateProperties({
  formData,
  isTimeHidden,
  readOnly,
  onField,
  onBreakChange,
  onAddBreak,
  onRemoveBreak,
}: ShiftTemplatePropertiesProps) {
  return (
    <>
      <div className="py-1">
        <PropertyRow
          label="ステータス"
          description="有効／無効を切り替えます。無効なテンプレートはシフト割り当ての選択肢に表示されなくなります。"
        >
          <button
            type="button"
            onClick={() => onField("is_active", !formData.is_active)}
            disabled={readOnly}
            className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 min-h-11 cursor-pointer disabled:cursor-default`}
          >
            <StatusPill isActive={formData.is_active} />
          </button>
        </PropertyRow>

        <PropertyRow
          label="シフト種別"
          description="勤務の種別です（出勤・休みなど）。種別によって時刻入力の要否が変わります。"
        >
          <Select
            value={formData.shift_type}
            onValueChange={(v) => onField("shift_type", v as ShiftType)}
            disabled={readOnly}
          >
            <SelectTrigger className="h-11 text-sm border-0 shadow-none bg-transparent px-1.5">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{SHIFT_TYPE_OPTIONS}</SelectContent>
          </Select>
        </PropertyRow>

        {!isTimeHidden ? (
          <>
            <PropertyRow label="開始時刻" description="このテンプレートの勤務開始時刻です。">
              <PropInput
                type="time"
                ariaLabel="開始時刻"
                value={formData.start_time}
                onChange={(v) => onField("start_time", v)}
                readOnly={readOnly}
              />
            </PropertyRow>
            <PropertyRow label="終了時刻" description="このテンプレートの勤務終了時刻です。">
              <PropInput
                type="time"
                ariaLabel="終了時刻"
                value={formData.end_time}
                onChange={(v) => onField("end_time", v)}
                readOnly={readOnly}
              />
            </PropertyRow>
          </>
        ) : null}

        <PropertyRow
          label="メモ"
          description="このシフトテンプレートに関する補足メモです。内部向けの記録として使われます。"
        >
          <PropInput
            ariaLabel="メモ"
            value={formData.notes}
            onChange={(v) => onField("notes", v)}
            placeholder="補足情報など"
            readOnly={readOnly}
          />
        </PropertyRow>
      </div>

      {!isTimeHidden ? (
        <BreakEditor
          breaks={formData.breaks}
          readOnly={readOnly}
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
  readOnly: boolean;
  onBreakChange: (index: number, field: "break_start" | "break_end", value: string) => void;
  onAddBreak: () => void;
  onRemoveBreak: (index: number) => void;
}

function BreakEditor({
  breaks,
  readOnly,
  onBreakChange,
  onAddBreak,
  onRemoveBreak,
}: BreakEditorProps) {
  return (
    <div className="mt-4">
      <div className="flex items-center justify-between mb-2">
        <span className={`flex items-center gap-1 text-sm font-medium ${C.text}`}>
          休憩時間
          <FieldHelp
            label="休憩時間"
            content="勤務中の休憩時間帯です。「追加」で複数の休憩を登録できます。"
          />
        </span>
        {!readOnly ? (
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
        ) : null}
      </div>
      {breaks.map((b, i) => (
        <div key={i} className="flex items-center gap-2 mb-2">
          <Input
            type="time"
            aria-label={`休憩${i + 1} 開始時刻`}
            value={b.break_start}
            onChange={(e) => onBreakChange(i, "break_start", e.target.value)}
            className="flex-1 h-11 text-sm"
            readOnly={readOnly}
          />
          <span className={`text-xs ${C.text50}`}>〜</span>
          <Input
            type="time"
            aria-label={`休憩${i + 1} 終了時刻`}
            value={b.break_end}
            onChange={(e) => onBreakChange(i, "break_end", e.target.value)}
            className="flex-1 h-11 text-sm"
            readOnly={readOnly}
          />
          {!readOnly ? (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0"
              onClick={() => onRemoveBreak(i)}
              aria-label={`休憩${i + 1}を削除`}
            >
              <X className={ICON.smXs} aria-hidden="true" />
            </Button>
          ) : null}
        </div>
      ))}
      {breaks.length === 0 ? <p className={`text-xs ${C.text40}`}>休憩なし</p> : null}
    </div>
  );
}

function PropertyRow({
  label,
  description,
  children,
}: {
  label: string;
  /** 項目の説明文。指定すると行右端に ⓘ ツールチップを表示する */
  description?: string;
  children: ReactNode;
}) {
  return (
    <div
      className={`flex gap-2 py-2 px-2 -mx-2 rounded-xxs ${C.hoverBgLight} transition-colors min-h-[40px]`}
    >
      <div
        className={`w-[120px] shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}
      >
        {label}
      </div>
      <div className="flex-1 flex items-center">{children}</div>
      {description ? <FieldHelp label={label} content={description} /> : null}
    </div>
  );
}

function PropInput({
  value,
  onChange,
  placeholder,
  type = "text",
  ariaLabel,
  readOnly = false,
}: {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  type?: string;
  ariaLabel?: string;
  readOnly?: boolean;
}) {
  return (
    <input
      type={type}
      aria-label={ariaLabel}
      className={`w-full bg-transparent text-sm ${C.text} outline-none border-none px-1.5 py-0.5 min-h-11 rounded-xxs ${C.hoverBgLight} transition-colors ${C.textPlaceholder} focus-visible:ring-2 ${C.focusRingAccent40}`}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder ?? "空"}
      readOnly={readOnly}
    />
  );
}
