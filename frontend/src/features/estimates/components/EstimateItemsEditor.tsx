import { memo, useMemo } from "react";
import { PlusCircle } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { CATEGORY_LABELS } from "@/constants/item-category";
import { formatTaxRatePercent } from "@/hooks/use-clinic-tax-rates";
import { formatCurrency } from "@/lib/format/number";
import { calcEstimateLineAmount } from "@/lib/estimate-totals";
import { C } from "@/lib/design-tokens";

import type { EstimateItemDraft } from "../hooks/use-estimate-form";

type LineTaxType = EstimateItemDraft["taxType"];
type LineCategory = EstimateItemDraft["category"];

const ITEMS_ERROR_ID = "estimate-items-error";

const TAX_TYPE_OPTIONS: { value: LineTaxType; label: string }[] = [
  { value: "included", label: "内税" },
  { value: "excluded", label: "外税" },
  { value: "exempt", label: "非課税" },
];

const CATEGORY_OPTIONS = Object.entries(CATEGORY_LABELS).map(([value, label]) => (
  <SelectItem key={value} value={value}>
    {label}
  </SelectItem>
));

export interface EstimateItemsEditorProps {
  items: EstimateItemDraft[];
  createItem: () => EstimateItemDraft;
  canEditDiscount: boolean;
  disabled?: boolean;
  error?: string;
  standardTaxRate: number;
  reducedTaxRate: number;
  onChange: (items: EstimateItemDraft[]) => void;
}

export const EstimateItemsEditor = memo(function EstimateItemsEditor({
  items,
  createItem,
  canEditDiscount,
  disabled,
  error,
  standardTaxRate,
  reducedTaxRate,
  onChange,
}: EstimateItemsEditorProps) {
  const rows = useMemo(
    () =>
      items.map((item) => {
        const update = (patch: Partial<EstimateItemDraft>) =>
          onChange(items.map((it) => (it.localId === item.localId ? { ...it, ...patch } : it)));
        const rateOptions = Array.from(
          new Set([item.taxRate, standardTaxRate, reducedTaxRate]),
        ).sort((a, b) => b - a);
        const { total: lineTotal } = calcEstimateLineAmount(item);
        return (
          <div
            key={item.localId}
            className={`grid grid-cols-1 gap-3 rounded-md border ${C.borderLight} p-3 sm:grid-cols-2 lg:grid-cols-4 items-end`}
          >
            <div className="space-y-1 min-w-0 sm:col-span-2">
              <Label htmlFor={`item-name-${item.localId}`} className="text-xs whitespace-nowrap">
                品目名
              </Label>
              <Input
                id={`item-name-${item.localId}`}
                value={item.name}
                onChange={(e) => update({ name: e.target.value })}
                placeholder="品目名を入力"
                className="h-11 text-sm"
                disabled={disabled}
              />
            </div>
            <div className="space-y-1 min-w-0">
              <Label
                htmlFor={`item-category-${item.localId}`}
                className="text-xs whitespace-nowrap"
              >
                カテゴリ
              </Label>
              <Select
                value={item.category}
                onValueChange={(v) => update({ category: v as LineCategory })}
                disabled={disabled}
              >
                <SelectTrigger id={`item-category-${item.localId}`} className="h-11 text-sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>{CATEGORY_OPTIONS}</SelectContent>
              </Select>
            </div>
            <div className="space-y-1 min-w-0">
              <Label htmlFor={`item-price-${item.localId}`} className="text-xs whitespace-nowrap">
                {item.taxType === "included" ? "単価（税込）" : "単価（税抜）"}
              </Label>
              <NumberInput
                id={`item-price-${item.localId}`}
                min={0}
                value={item.unitPrice}
                onChange={(v) => update({ unitPrice: Number(v) })}
                suffix="円"
                className="h-11 text-sm"
                disabled={disabled}
              />
            </div>
            <div className="space-y-1 min-w-0">
              <Label htmlFor={`item-qty-${item.localId}`} className="text-xs whitespace-nowrap">
                数量
              </Label>
              <NumberInput
                id={`item-qty-${item.localId}`}
                min={0}
                step={0.1}
                value={item.quantity}
                onChange={(v) => update({ quantity: Number(v) })}
                className="h-11 text-sm"
                disabled={disabled}
              />
            </div>
            <div className="space-y-1 min-w-0">
              <Label
                htmlFor={`item-tax-type-${item.localId}`}
                className="text-xs whitespace-nowrap"
              >
                課税区分
              </Label>
              <Select
                value={item.taxType}
                onValueChange={(v) => update({ taxType: v as LineTaxType })}
                disabled={disabled}
              >
                <SelectTrigger id={`item-tax-type-${item.localId}`} className="h-11 text-sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TAX_TYPE_OPTIONS.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-1 min-w-0">
              <Label
                htmlFor={`item-tax-rate-${item.localId}`}
                className="text-xs whitespace-nowrap"
              >
                税率
              </Label>
              <Select
                value={String(item.taxRate)}
                onValueChange={(v) => update({ taxRate: Number(v) })}
                disabled={disabled || item.taxType === "exempt"}
              >
                <SelectTrigger id={`item-tax-rate-${item.localId}`} className="h-11 text-sm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {rateOptions.map((rate) => (
                    <SelectItem key={rate} value={String(rate)}>
                      {formatTaxRatePercent(rate)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            {canEditDiscount ? (
              <div className="space-y-1 min-w-0">
                <Label
                  htmlFor={`item-discount-${item.localId}`}
                  className="text-xs whitespace-nowrap"
                >
                  値引（円）
                </Label>
                <NumberInput
                  id={`item-discount-${item.localId}`}
                  min={0}
                  value={item.discountAmount}
                  onChange={(v) => update({ discountAmount: Number(v) })}
                  suffix="円"
                  className="h-11 text-sm"
                  disabled={disabled}
                />
              </div>
            ) : null}
            <div className="flex items-end justify-between gap-1">
              <div className="space-y-1 pb-1">
                <span className={`text-xs ${C.text60}`}>小計</span>
                <div className={`text-sm font-mono font-medium ${C.text}`}>
                  {formatCurrency(lineTotal)}
                </div>
              </div>
              {disabled ? null : (
                <DeleteIconButton
                  title="行を削除"
                  onClick={() => onChange(items.filter((it) => it.localId !== item.localId))}
                />
              )}
            </div>
          </div>
        );
      }),
    [items, onChange, canEditDiscount, disabled, standardTaxRate, reducedTaxRate],
  );

  return (
    <div className="space-y-3" aria-describedby={error ? ITEMS_ERROR_ID : undefined}>
      <div className={`text-sm font-medium ${C.text}`}>見積明細</div>
      {items.length === 0 ? <p className={`text-sm ${C.text60}`}>明細行がありません</p> : rows}
      {disabled ? null : (
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className={`h-10 text-sm gap-2 ${C.text60} ${C.hoverText}`}
          onClick={() => onChange([...items, createItem()])}
        >
          <PlusCircle className="h-4 w-4" />
          明細行を追加
        </Button>
      )}
      <FormFieldError id={ITEMS_ERROR_ID} message={error} />
    </div>
  );
});
