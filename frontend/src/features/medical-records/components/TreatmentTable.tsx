import React from "react";
import { Button } from "../../../components/ui/button";
import { Input } from "../../../components/ui/input";
import { ScrollArea } from "../../../components/ui/scroll-area";
import { Circle, X, Trash2, PlusCircle } from "lucide-react";
import { cn } from "../../../components/ui/utils";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../../../components/ui/select";

export interface TreatmentItem {
  id: number;
  status?: string;
  content: string;
  memo: string;
  insurance: boolean;
  unitPrice: number;
  quantity: number;
  discountRate: number;
  discountAmount: number;
  selected?: boolean;
  inventoryId?: string;
}

interface TreatmentTableProps {
  items: TreatmentItem[];
  onUpdate: (id: number, field: keyof TreatmentItem, value: any) => void;
  onRemove: (id: number) => void;
  onOpenSearch?: () => void;
  onAddRow?: () => void;
  showStatus?: boolean;
}

export function TreatmentTable({
  items,
  onUpdate,
  onRemove,
  onOpenSearch,
  onAddRow,
  showStatus = false,
}: TreatmentTableProps) {
  const gridColsClass = showStatus
    ? "grid-cols-[80px_3fr_2fr_0.8fr_1fr_0.8fr_1fr_1fr_1fr_0.8fr]"
    : "grid-cols-[3fr_2fr_0.8fr_1fr_0.8fr_1fr_1fr_1fr_0.8fr]";

  return (
    <div className="flex-1 flex flex-col min-h-0 border border-[rgba(55,53,47,0.16)] rounded-lg bg-white overflow-hidden shadow-sm">
      {/* Header */}
      <div
        className={cn(
          "grid gap-0 border-b border-[rgba(55,53,47,0.16)] bg-[#F7F6F3] text-sm font-bold text-[#37352F]/80 min-h-[48px] items-center",
          gridColsClass
        )}
      >
        {showStatus && <HeaderCell align="center">ステータス</HeaderCell>}
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
                "grid gap-0 border-b border-[rgba(55,53,47,0.09)] bg-white text-sm text-[#37352F] items-center hover:bg-[#F7F6F3]/50 transition-colors h-12 group",
                gridColsClass
              )}
            >
              {showStatus && (
                <Cell align="center">
                  <Select
                    value={item.status || "未完了"}
                    onValueChange={(val) => onUpdate(item.id, "status", val)}
                  >
                    <SelectTrigger className="h-full w-full border-none bg-transparent p-0 text-sm justify-center text-center font-medium focus:ring-0">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="未完了">未完了</SelectItem>
                      <SelectItem value="完了">完了</SelectItem>
                      <SelectItem value="-">-</SelectItem>
                    </SelectContent>
                  </Select>
                </Cell>
              )}
              <Cell>
                <TableInput
                  value={item.content}
                  onChange={(val) => onUpdate(item.id, "content", val)}
                  placeholder="治療内容を入力"
                />
              </Cell>
              <Cell>
                <TableInput
                  value={item.memo}
                  onChange={(val) => onUpdate(item.id, "memo", val)}
                  placeholder="メモ"
                />
              </Cell>
              <Cell
                align="center"
                onClick={() => onUpdate(item.id, "insurance", !item.insurance)}
                className="cursor-pointer hover:bg-gray-50"
              >
                {item.insurance ? (
                  <Circle className="size-4 text-[#EA3323]" />
                ) : (
                  <X className="size-4 text-gray-300" />
                )}
              </Cell>
              <Cell>
                <TableInput
                  type="number"
                  value={item.unitPrice}
                  onChange={(val) => onUpdate(item.id, "unitPrice", Number(val))}
                  align="right"
                  className="font-mono"
                />
              </Cell>
              <Cell>
                <TableInput
                  type="number"
                  value={item.quantity}
                  onChange={(val) => onUpdate(item.id, "quantity", Number(val))}
                  align="right"
                  className="font-mono"
                />
              </Cell>
              <Cell>
                <TableInput
                  type="number"
                  value={item.discountRate}
                  onChange={(val) => onUpdate(item.id, "discountRate", Number(val))}
                  align="right"
                  className="font-mono"
                />
              </Cell>
              <Cell>
                <TableInput
                  type="number"
                  value={item.discountAmount}
                  onChange={(val) => onUpdate(item.id, "discountAmount", Number(val))}
                  align="right"
                  className="font-mono"
                />
              </Cell>
              <Cell align="right" className="font-mono font-medium px-2">
                {(
                  item.unitPrice * item.quantity -
                  item.discountAmount
                ).toLocaleString()}
              </Cell>
              <Cell align="center" last>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-10 w-10 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded opacity-0 group-hover:opacity-100 transition-opacity"
                  onClick={() => onRemove(item.id)}
                >
                  <Trash2 className="size-4" />
                </Button>
              </Cell>
            </div>
          ))}
        </div>
      </ScrollArea>

      <div className="p-2 bg-[#F7F6F3] border-t border-[rgba(55,53,47,0.16)] flex justify-center">
        <Button
          variant="ghost"
          size="sm"
          className="h-10 text-sm gap-2 text-muted-foreground hover:text-primary"
          onClick={onOpenSearch || onAddRow}
        >
          <PlusCircle className="size-4" />
          {onOpenSearch ? "行を追加（検索）" : "行を追加"}
        </Button>
      </div>
    </div>
  );
}

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
        "p-2 border-[rgba(55,53,47,0.16)] h-full flex items-center",
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
        "h-full flex items-center border-[rgba(55,53,47,0.09)] p-0",
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
  align = "left",
  className,
}: {
  value: string | number;
  onChange: (val: string) => void;
  placeholder?: string;
  type?: string;
  align?: "left" | "right";
  className?: string;
}) {
  return (
    <Input
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className={cn(
        "h-full w-full border-none bg-transparent rounded-none focus-visible:ring-0 px-3 text-sm shadow-none",
        align === "right" && "text-right",
        className
      )}
      placeholder={placeholder}
    />
  );
}
