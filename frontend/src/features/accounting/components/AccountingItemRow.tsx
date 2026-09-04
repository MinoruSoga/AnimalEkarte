import { useState } from "react";
import { ChevronDown } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { TableCell, TableRow } from "@/components/ui/table";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { TaxRateSelector } from "@/components/shared/TaxRateSelector/TaxRateSelector";
import { TaxTypeSelector } from "@/components/shared/TaxTypeSelector/TaxTypeSelector";
import { C } from "@/lib/design-tokens";
import type { TaxType } from "@/types/generated/models";
import { formatCurrency, formatCurrencyOrDash } from "@/lib/format/number";

import { useGetBillingItemDiscountSuggestions } from "../api/get-discount-suggestions";
import type { AccountingItem, ItemCategory } from "../types";

// FE5-26: 正本は src/constants/item-category.ts へ移設。
import { CATEGORY_LABELS } from "@/constants/item-category";

interface DiscountCellProps {
  item: AccountingItem;
  canEdit: boolean;
  accountingId: string | undefined;
  onUpdateItemDiscount: ((itemId: string, discountAmount: number) => void) | undefined;
}

function DiscountCell({ item, canEdit, accountingId, onUpdateItemDiscount }: DiscountCellProps) {
  const [open, setOpen] = useState(false);
  const { data: suggestions = [], isFetching } = useGetBillingItemDiscountSuggestions(
    accountingId !== undefined ? item.id : undefined,
    open,
  );

  if (accountingId === undefined || onUpdateItemDiscount === undefined || !canEdit) {
    return (
      <span className={`text-sm ${C.text50}`}>{formatCurrencyOrDash(item.discountAmount)}</span>
    );
  }

  return (
    <div className="flex items-center gap-1 justify-center">
      <Input
        id={`discount-${item.id}`}
        aria-label={`割引額: ${item.name} (ID ${item.id})`}
        type="number"
        min={0}
        defaultValue={item.discountAmount}
        onBlur={(e) =>
          onUpdateItemDiscount(item.id, Math.max(0, parseInt(e.target.value, 10) || 0))
        }
        className="w-20 min-h-11 text-right"
      />
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <button
            type="button"
            aria-label={`割引候補: ${item.name} (ID ${item.id})`}
            className={`min-h-11 min-w-11 flex items-center justify-center rounded-xs ${C.text50} ${C.hoverBgLight} transition-colors shrink-0`}
          >
            <ChevronDown className="h-3.5 w-3.5" />
          </button>
        </PopoverTrigger>
        <PopoverContent align="end" className="w-64 p-2">
          {isFetching ? (
            <p className={`text-xs ${C.text50} py-2 text-center`}>読み込み中...</p>
          ) : suggestions.length === 0 ? (
            <p className={`text-xs ${C.text50} py-2 text-center`}>適用可能な割引はありません</p>
          ) : (
            <ul className="space-y-0.5">
              {suggestions.map((s, i) => (
                <li key={s.campaign_id ?? `owner-${i}`}>
                  <button
                    type="button"
                    aria-label={`割引を適用: ${s.name} -${s.amount}円 (品目ID ${item.id})`}
                    onClick={() => {
                      onUpdateItemDiscount(item.id, s.amount);
                      setOpen(false);
                    }}
                    className={`w-full min-h-11 min-w-11 text-left px-2 py-1.5 rounded-xs text-xs ${C.hoverBgLight} transition-colors`}
                  >
                    <span className="font-medium">{s.name}</span>
                    <span className={`ml-1 ${C.text50}`}>
                      {s.discount_type === "rate"
                        ? `${s.discount_value}%`
                        : formatCurrency(s.discount_value)}
                    </span>
                    <span className={`float-right font-mono ${C.textRedIcon}`}>
                      -{formatCurrency(s.amount)}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </PopoverContent>
      </Popover>
    </div>
  );
}

interface AccountingItemRowProps {
  item: AccountingItem;
  accountingId: string | undefined;
  canEdit: boolean;
  canDelete: boolean;
  onDeleteItem: (id: string) => void;
  onUpdateItemTax: ((itemId: string, taxType: TaxType, taxRate: number) => void) | undefined;
  onUpdateItemDiscount: ((itemId: string, discountAmount: number) => void) | undefined;
}

export function AccountingItemRow({
  item,
  accountingId,
  canEdit,
  canDelete,
  onDeleteItem,
  onUpdateItemTax,
  onUpdateItemDiscount,
}: AccountingItemRowProps) {
  return (
    <TableRow className="h-12">
      <TableCell>
        <Badge variant="outline" className="font-normal text-xs whitespace-nowrap">
          {CATEGORY_LABELS[item.category as ItemCategory] ?? "その他"}
        </Badge>
      </TableCell>
      <TableCell className="font-medium whitespace-nowrap">
        {item.name}
        {item.source === "medical_record" ? (
          <span className={`ml-2 text-2xs ${C.textBrand} ${C.bgBrand5} px-1.5 py-0.5 rounded`}>
            カルテ連携
          </span>
        ) : null}
        {item.source === "trimming" ? (
          <span className={`ml-2 text-2xs ${C.textBrand} ${C.bgBrand5} px-1.5 py-0.5 rounded`}>
            トリミング
          </span>
        ) : null}
      </TableCell>
      <TableCell className="text-right">{formatCurrency(item.unitPrice)}</TableCell>
      <TableCell className="text-center">
        <div className="flex items-center justify-center gap-2">{item.quantity}</div>
      </TableCell>
      <TableCell className="text-center">
        <DiscountCell
          item={item}
          canEdit={canEdit}
          accountingId={accountingId}
          onUpdateItemDiscount={onUpdateItemDiscount}
        />
      </TableCell>
      <TableCell className="text-center">
        {accountingId !== undefined && onUpdateItemTax !== undefined && canEdit ? (
          <TaxTypeSelector
            value={item.taxType}
            onChange={(v) => onUpdateItemTax(item.id, v, item.taxRate)}
            ariaLabel={`課税区分: ${item.name} (ID ${item.id})`}
            className="min-h-11 min-w-11"
          />
        ) : (
          <span className={`text-sm ${C.text50}`}>
            {item.taxType === "excluded" ? "外税" : item.taxType === "included" ? "内税" : "非課税"}
          </span>
        )}
      </TableCell>
      <TableCell className="text-center">
        {accountingId !== undefined && onUpdateItemTax !== undefined && canEdit ? (
          <TaxRateSelector
            value={item.taxRate}
            onChange={(v) => onUpdateItemTax(item.id, item.taxType, v)}
            ariaLabel={`税率: ${item.name} (ID ${item.id})`}
            className="min-h-11 min-w-11"
          />
        ) : (
          <span className={`text-sm ${C.text50}`}>{Math.round(item.taxRate * 100)}%</span>
        )}
      </TableCell>
      <TableCell className="text-right font-mono text-sm whitespace-nowrap">
        {formatCurrency(item.taxAmount)}
      </TableCell>
      <TableCell className="text-center">
        {item.isInsuranceApplicable ? (
          <span className={`${C.textStatusGreen} font-bold text-xs`}>●</span>
        ) : (
          <span className={`${C.text20} text-xs`}>-</span>
        )}
      </TableCell>
      <TableCell className="text-right font-medium">
        {formatCurrency(item.subtotal + item.taxAmount)}
      </TableCell>
      <TableCell>
        {accountingId === undefined || (item.source === "manual" && canDelete) ? (
          <DeleteIconButton onClick={() => onDeleteItem(item.id)} />
        ) : null}
      </TableCell>
    </TableRow>
  );
}
