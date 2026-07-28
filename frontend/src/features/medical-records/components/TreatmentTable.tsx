// React/Framework
import { C, ICON, STYLE } from "@/lib/design-tokens";
import React, { memo } from "react";

// Shared
import { calcLineItemAmount } from "@/lib/line-item-helpers";

// External
import { Circle, X, PlusCircle } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

// rendering-hoist-jsx: 静的SelectItem定数をモジュールスコープに巻き上げ
const TREATMENT_STATUS_ITEMS = (
  <>
    <SelectItem value="pending">未完了</SelectItem>
    <SelectItem value="completed">完了</SelectItem>
    <SelectItem value="not_applicable">-</SelectItem>
  </>
);

export interface TreatmentItem {
  id: number;
  status?: string;
  content: string;
  memo: string;
  is_insurance: boolean;
  unitPrice: number;
  quantity: number;
  discountRate: number;
  discountAmount: number;
  is_selected?: boolean;
  inventoryId?: string;
}

interface TreatmentTableProps {
  items: TreatmentItem[];
  onUpdate: (id: number, field: keyof TreatmentItem, value: string | number | boolean) => void;
  onRemove: (id: number) => void;
  onOpenSearch?: () => void;
  onAddRow?: () => void;
  showStatus?: boolean;
  /** BUG-051: 確認済み等で全入力を無効化する */
  disabled?: boolean;
}

export const TreatmentTable = memo(function TreatmentTable({
  items,
  onUpdate,
  onRemove,
  onOpenSearch,
  onAddRow,
  showStatus = false,
  disabled = false,
}: TreatmentTableProps) {
  const gridColsClass = showStatus
    ? "grid-cols-[80px_3fr_2fr_0.8fr_1fr_0.8fr_1fr_1fr_1fr_0.8fr]"
    : "grid-cols-[3fr_2fr_0.8fr_1fr_0.8fr_1fr_1fr_1fr_0.8fr]";

  return (
    <div className={`flex-1 flex flex-col min-h-0 border ${C.borderMedium} rounded-lg ${C.bgWhite} overflow-hidden`}>
      {/* Header — DESIGN.md ex-data-table-cell: canvas-soft 背景 + eyebrow 相当タイポグラフィ（STYLE.sectionLabel） */}
      <div
        className={cn(
          `grid gap-0 border-b ${C.borderMedium} ${C.bgPage} ${STYLE.sectionLabel} min-h-[48px] items-center`,
          gridColsClass
        )}
      >
        {showStatus ? <HeaderCell align="center">ステータス</HeaderCell> : null}
        <HeaderCell>治療内容</HeaderCell>
        <HeaderCell>メモ</HeaderCell>
        <HeaderCell align="center">保険</HeaderCell>
        <HeaderCell align="right">単価</HeaderCell>
        <HeaderCell align="right">数量</HeaderCell>
        <HeaderCell align="right">割引(%)</HeaderCell>
        <HeaderCell align="right">値引(￥)</HeaderCell>
        <HeaderCell align="right">小計</HeaderCell>
        <HeaderCell align="center" last>
          操作
        </HeaderCell>
      </div>

      <ScrollArea className="flex-1">
        <div>
          {items.map((item) => (
            <div
              key={item.id}
              className={cn(
                `grid gap-0 border-b ${C.borderLight} ${C.bgWhite} text-sm ${C.text} items-center ${C.hoverBgPageHalf} transition-colors h-12 group`,
                gridColsClass
              )}
            >
              {showStatus ? (
                <Cell align="center">
                  <Select
                    value={item.status || "pending"}
                    onValueChange={(val) => onUpdate(item.id, "status", val)}
                    disabled={disabled}
                  >
                    <SelectTrigger className="h-full w-full border-none bg-transparent p-0 text-sm justify-center text-center font-medium focus:ring-0">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>{TREATMENT_STATUS_ITEMS}</SelectContent>
                  </Select>
                </Cell>
              ) : null}
              <Cell>
                <TableInput
                  value={item.content}
                  onChange={(val) => onUpdate(item.id, "content", val)}
                  placeholder="治療内容を入力"
                  disabled={disabled}
                />
              </Cell>
              <Cell>
                <TableInput
                  value={item.memo}
                  onChange={(val) => onUpdate(item.id, "memo", val)}
                  placeholder="メモ"
                  disabled={disabled}
                />
              </Cell>
              <Cell align="center">
                <button
                  type="button"
                  aria-pressed={item.is_insurance}
                  aria-label="保険適用"
                  disabled={disabled}
                  onClick={() => onUpdate(item.id, "is_insurance", !item.is_insurance)}
                  className={cn(
                    "h-full w-full flex items-center justify-center",
                    disabled ? "cursor-not-allowed opacity-60" : `cursor-pointer ${C.hoverBgPage}`
                  )}
                >
                  {item.is_insurance ? (
                    <Circle className={`${ICON.action} ${C.textRedIcon}`} />
                  ) : (
                    <X className={`${ICON.action} ${C.text25}`} />
                  )}
                </button>
              </Cell>
              <Cell>
                {/* BUG-072: 単価は0以上の整数・上限9億円 */}
                <TableInput
                  type="number"
                  step="1"
                  min="0"
                  max="999999999"
                  value={item.unitPrice}
                  onChange={(val) => onUpdate(item.id, "unitPrice", Number(val))}
                  align="right"
                  className="font-mono"
                  disabled={disabled}
                />
              </Cell>
              <Cell>
                {/* BUG-TRT-001: min="0.1" → min="1" に変更（float32精度 0.10000000149011612 問題の回避） */}
                {/* BUG-068: 0以下の入力は1にクランプ */}
                <TableInput
                  type="number"
                  step="1"
                  min="1"
                  value={item.quantity}
                  onChange={(val) => {
                    const n = Number(val);
                    onUpdate(item.id, "quantity", n <= 0 ? 1 : n);
                  }}
                  align="right"
                  className="font-mono"
                  disabled={disabled}
                />
              </Cell>
              <Cell>
                {/* BUG-072: 割引率は0〜100 */}
                <TableInput
                  type="number"
                  min="0"
                  max="100"
                  value={item.discountRate}
                  onChange={(val) => onUpdate(item.id, "discountRate", Number(val))}
                  align="right"
                  className="font-mono"
                  disabled={disabled}
                />
              </Cell>
              <Cell>
                {/* BUG-072: 値引額は0以上 */}
                <TableInput
                  type="number"
                  min="0"
                  value={item.discountAmount}
                  onChange={(val) => onUpdate(item.id, "discountAmount", Number(val))}
                  align="right"
                  className="font-mono"
                  disabled={disabled}
                />
              </Cell>
              <Cell align="right" className="font-mono font-medium px-2">
                {calcLineItemAmount(item).total.toLocaleString()}
              </Cell>
              <Cell align="center" last>
                {disabled ? null : (
                  <DeleteIconButton
                    onClick={() => onRemove(item.id)}
                    className="opacity-0 group-hover:opacity-100 transition-opacity"
                  />
                )}
              </Cell>
            </div>
          ))}
        </div>
      </ScrollArea>

      {/* BUG-051: disabled 時は行追加ボタンを非表示 */}
      {disabled ? null : (
        <div className={`p-2 ${C.bgPage} border-t ${C.borderMedium} flex justify-center`}>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className={`h-10 text-sm gap-2 ${C.text50} ${C.hoverTextBrand}`}
            onClick={onOpenSearch || onAddRow}
          >
            <PlusCircle className={ICON.action} />
            {onOpenSearch ? "行を追加（検索）" : "行を追加"}
          </Button>
        </div>
      )}
    </div>
  );
});

// --- Helper Components ---

function HeaderCell({
  children,
  align = "left",
  last = false,
}: {
  children: React.ReactNode;
  align?: "left" | "center" | "right";
  last?: boolean;
}) {
  return (
    <div
      className={cn(
        `p-2 ${C.borderMedium} h-full flex items-center`,
        !last && "border-r",
        align === "center" && "justify-center",
        align === "right" && "justify-end"
      )}
    >
      {children}
    </div>
  );
}

function Cell({
  children,
  align = "left",
  last = false,
  onClick,
  className,
}: {
  children: React.ReactNode;
  align?: "left" | "center" | "right";
  last?: boolean;
  onClick?: () => void;
  className?: string;
}) {
  return (
    <div
      className={cn(
        `h-full flex items-center ${C.borderLight} p-0`,
        !last && "border-r",
        align === "center" && "justify-center",
        align === "right" && "justify-end",
        className
      )}
      onClick={onClick}
    >
      {children}
    </div>
  );
}

function TableInput({
  value,
  onChange,
  placeholder,
  type = "text",
  step,
  min,
  max,
  align = "left",
  className,
  disabled,
}: {
  value: string | number;
  onChange: (val: string) => void;
  placeholder?: string;
  type?: string;
  step?: string;
  min?: string;
  max?: string;
  align?: "left" | "right";
  className?: string;
  disabled?: boolean;
}) {
  return (
    <Input
      type={type}
      step={step}
      min={min}
      max={max}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      disabled={disabled}
      className={cn(
        "h-full w-full border-none bg-transparent rounded-none focus-visible:ring-0 px-3 text-sm shadow-none",
        align === "right" && "text-right",
        disabled && "opacity-60 cursor-not-allowed",
        className
      )}
      placeholder={placeholder}
    />
  );
}
