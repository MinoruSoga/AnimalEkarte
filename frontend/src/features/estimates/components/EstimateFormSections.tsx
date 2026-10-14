import { memo } from "react";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { formatCurrency } from "@/lib/format/number";
import { C } from "@/lib/design-tokens";

import { CREATE_STATUS_OPTIONS, EDIT_STATUS_OPTIONS } from "../constants/estimate-status-options";
import type { EstimateStatus } from "../types";

// rendering-hoist-jsx: SelectItem リストは静的なのでモジュール定数に巻き上げ
const EDIT_STATUS_SELECT_ITEMS = EDIT_STATUS_OPTIONS.map((opt) => (
  <SelectItem key={opt.value} value={opt.value}>
    {opt.label}
  </SelectItem>
));

const CREATE_STATUS_SELECT_ITEMS = CREATE_STATUS_OPTIONS.map((opt) => (
  <SelectItem key={opt.value} value={opt.value}>
    {opt.label}
  </SelectItem>
));

// rerender-memo: 基本情報セクション（タイトル/ステータス/有効期限）を独立 memo 化
// 金額・テキストフィールドの変更では再レンダーしない
interface BasicInfoSectionProps {
  title: string;
  status: EstimateStatus;
  validUntil: string;
  isEdit: boolean;
  onChange: (key: string, value: unknown) => void;
  titleError?: string;
  statusError?: string;
}

export const BasicInfoSection = memo(function BasicInfoSection({
  title,
  status,
  validUntil,
  isEdit,
  onChange,
  titleError,
  statusError,
}: BasicInfoSectionProps) {
  return (
    <>
      {/* タイトル */}
      <div className="space-y-1.5">
        <Label htmlFor="title" className={`text-sm font-medium ${C.text}`}>
          タイトル <span className={C.textRequired}>*</span>
        </Label>
        <Input
          id="title"
          value={title}
          onChange={(e) => onChange("title", e.target.value)}
          placeholder="見積書タイトルを入力"
          className="h-11 text-sm"
          aria-invalid={titleError ? true : undefined}
          aria-describedby={titleError ? "title-error" : undefined}
        />
        <FormFieldError id="title-error" message={titleError} />
      </div>

      {/* ステータス */}
      <div className="space-y-1.5">
        <Label htmlFor="status" className={`text-sm font-medium ${C.text}`}>
          ステータス
        </Label>
        <Select value={status} onValueChange={(v) => onChange("status", v as EstimateStatus)}>
          <SelectTrigger
            id="status"
            className="h-11 text-sm w-full sm:w-[200px]"
            aria-invalid={statusError ? true : undefined}
            aria-describedby={statusError ? "status-error" : undefined}
          >
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {isEdit ? EDIT_STATUS_SELECT_ITEMS : CREATE_STATUS_SELECT_ITEMS}
          </SelectContent>
        </Select>
        <FormFieldError id="status-error" message={statusError} />
      </div>

      {/* 有効期限 */}
      <div className="space-y-1.5">
        <Label htmlFor="validUntil" className={`text-sm font-medium ${C.text}`}>
          有効期限
        </Label>
        <DatePicker
          id="validUntil"
          value={validUntil ? validUntil.slice(0, 10) : ""}
          onChange={(v) => onChange("validUntil", v)}
          placeholder="有効期限を選択…"
          className="w-full sm:w-[220px]"
        />
      </div>
    </>
  );
});

// rerender-memo: 金額サマリセクションを独立 memo 化
// タイトル/ステータス/テキストフィールドの変更では再レンダーしない
interface AmountSectionProps {
  subtotal: number;
  taxTotal: number;
  insuranceAmount: number;
  discountAmount: number;
  totalAmount: number;
}

function AmountSummaryRow({ label, value }: { label: string; value: number }) {
  return (
    <div className="space-y-1.5">
      <div className={`text-sm font-medium ${C.text}`}>{label}</div>
      <div className={`h-11 flex items-center text-sm font-mono ${C.text}`}>
        {formatCurrency(value)}
      </div>
    </div>
  );
}

export const AmountSection = memo(function AmountSection({
  subtotal,
  taxTotal,
  insuranceAmount,
  discountAmount,
  totalAmount,
}: AmountSectionProps) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <AmountSummaryRow label="小計（明細合計）" value={subtotal} />
      <AmountSummaryRow label="消費税（内税分を含む）" value={taxTotal} />
      <AmountSummaryRow label="保険適用額" value={insuranceAmount} />
      <AmountSummaryRow label="割引額" value={discountAmount} />
      <AmountSummaryRow label="合計金額" value={totalAmount} />
    </div>
  );
});

// rerender-memo: テキストセクション（コメント/備考）を独立 memo 化
// 金額・基本情報フィールドの変更では再レンダーしない
interface TextSectionProps {
  comment: string;
  notes: string;
  onChange: (key: string, value: unknown) => void;
}

export const TextSection = memo(function TextSection({
  comment,
  notes,
  onChange,
}: TextSectionProps) {
  return (
    <>
      {/* コメント */}
      <div className="space-y-1.5">
        <Label htmlFor="comment" className={`text-sm font-medium ${C.text}`}>
          コメント
        </Label>
        <Textarea
          id="comment"
          value={comment}
          onChange={(e) => onChange("comment", e.target.value)}
          placeholder="飼主向けコメントを入力"
          className="text-sm min-h-[80px] resize-none"
        />
      </div>

      {/* 備考 */}
      <div className="space-y-1.5">
        <Label htmlFor="notes" className={`text-sm font-medium ${C.text}`}>
          備考（社内メモ）
        </Label>
        <Textarea
          id="notes"
          value={notes}
          onChange={(e) => onChange("notes", e.target.value)}
          placeholder="社内メモを入力"
          className="text-sm min-h-[80px] resize-none"
        />
      </div>
    </>
  );
});
