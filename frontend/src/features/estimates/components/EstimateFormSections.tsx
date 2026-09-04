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
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
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
        />
        <FormFieldError message={titleError} />
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
  canEditDiscount: boolean;
  onChange: (key: string, value: unknown) => void;
}

export const AmountSection = memo(function AmountSection({
  subtotal,
  taxTotal,
  insuranceAmount,
  discountAmount,
  totalAmount,
  canEditDiscount,
  onChange,
}: AmountSectionProps) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <div className="space-y-1.5">
        <Label htmlFor="subtotal" className={`text-sm font-medium ${C.text}`}>
          小計（税抜）
        </Label>
        <NumberInput
          id="subtotal"
          min={0}
          value={subtotal}
          onChange={(v) => onChange("subtotal", Number(v))}
          suffix="円"
          className="h-11 text-sm"
        />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="taxTotal" className={`text-sm font-medium ${C.text}`}>
          消費税
        </Label>
        <NumberInput
          id="taxTotal"
          min={0}
          value={taxTotal}
          onChange={(v) => onChange("taxTotal", Number(v))}
          suffix="円"
          className="h-11 text-sm"
        />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="insuranceAmount" className={`text-sm font-medium ${C.text}`}>
          保険適用額
        </Label>
        <NumberInput
          id="insuranceAmount"
          min={0}
          value={insuranceAmount}
          onChange={(v) => onChange("insuranceAmount", Number(v))}
          suffix="円"
          className="h-11 text-sm"
        />
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="discountAmount" className={`text-sm font-medium ${C.text}`}>
          割引額
        </Label>
        <NumberInput
          id="discountAmount"
          min={0}
          value={discountAmount}
          disabled={!canEditDiscount}
          onChange={(v) => onChange("discountAmount", Number(v))}
          suffix="円"
          className="h-11 text-sm"
        />
        {!canEditDiscount ? (
          <p className={`text-xs ${C.text50}`}>割引額の変更には権限が必要です</p>
        ) : null}
      </div>
      <div className="space-y-1.5">
        <Label htmlFor="totalAmount" className={`text-sm font-medium ${C.text}`}>
          合計金額
        </Label>
        <NumberInput
          id="totalAmount"
          min={0}
          value={totalAmount}
          onChange={(v) => onChange("totalAmount", Number(v))}
          suffix="円"
          className="h-11 text-sm"
        />
      </div>
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
