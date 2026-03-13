// React/Framework
import type { ReactNode } from "react";

// External
import { AlertTriangle, Package } from "lucide-react";

// Internal
import { StatusBadge } from "@/components/shared/StatusBadge";
import { C } from "@/lib/design-tokens";
import {
  getInventoryStatusColor,
  getInventoryStatusLabel,
} from "@/utils/status-helpers";

// Types
import type { InventoryItem } from "@/types";

// ─────────────────────────────────────────────────
// StockAlert sub-component
// ─────────────────────────────────────────────────

interface StockAlertProps {
  status: InventoryItem["status"];
}

/**
 * StockAlert
 *
 * Displays a compact warning banner when stock is low or out.
 * Renders nothing when status is "sufficient".
 */
export function StockAlert({ status }: StockAlertProps): ReactNode {
  if (status === "sufficient") return null;

  const isOutOfStock = status === "out_of_stock";

  return (
    <div
      className={`flex items-center gap-1.5 px-2 py-1 rounded text-xs font-medium ${
        isOutOfStock
          ? "bg-red-50 text-red-600 border border-red-200"
          : "bg-amber-50 text-amber-600 border border-amber-200"
      }`}
    >
      <AlertTriangle className="size-3 shrink-0" />
      <span>{isOutOfStock ? "在庫切れ" : "残少"}</span>
    </div>
  );
}

// ─────────────────────────────────────────────────
// Category label map
// ─────────────────────────────────────────────────

const CATEGORY_LABELS: Record<InventoryItem["category"], string> = {
  medicine: "医薬品",
  consumable: "消耗品",
  food: "フード",
  other: "その他",
};

// ─────────────────────────────────────────────────
// InventoryCard
// ─────────────────────────────────────────────────

export interface InventoryCardProps {
  item: InventoryItem;
  onClick?: (id: string) => void;
}

/**
 * InventoryCard
 *
 * Card view for a single inventory item. Shows:
 * - Product name and category badge
 * - Current stock / minimum stock with unit
 * - StockAlert warning when stock is low or out
 * - Status badge
 */
export function InventoryCard({ item, onClick }: InventoryCardProps) {
  const isLow = item.status !== "sufficient";

  return (
    <button
      type="button"
      onClick={() => onClick?.(item.id)}
      className={`w-full text-left bg-white border rounded-lg p-4 transition-colors ${
        isLow
          ? "border-amber-200 hover:border-amber-300"
          : `${C.borderLight} hover:border-[rgba(55,53,47,0.18)]`
      } hover:bg-[#F7F6F3] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring`}
    >
      {/* Header row */}
      <div className="flex items-start justify-between gap-2 mb-3">
        <div className="flex items-center gap-2 min-w-0">
          <span
            className={`shrink-0 size-8 flex items-center justify-center rounded-[3px] ${C.bgPage} ${C.text45}`}
          >
            <Package className="size-4" />
          </span>
          <span className={`text-sm font-medium ${C.text} truncate`}>
            {item.name}
          </span>
        </div>
        <StockAlert status={item.status} />
      </div>

      {/* Category */}
      <div className="flex items-center gap-1.5 mb-3">
        <span
          className={`inline-block px-1.5 py-0.5 text-xs rounded ${C.bgPage} ${C.text55} border ${C.borderLight}`}
        >
          {CATEGORY_LABELS[item.category]}
        </span>
        {item.location && (
          <span className={`text-xs ${C.text45} truncate`}>{item.location}</span>
        )}
      </div>

      {/* Stock info */}
      <div className="flex items-end justify-between gap-2">
        <div>
          <p className={`text-xs ${C.text40} mb-0.5`}>在庫数</p>
          <p className={`text-lg font-semibold font-mono ${C.text} leading-none`}>
            {item.quantity}
            <span className={`text-xs font-normal ${C.text55} ml-1`}>
              {item.unit}
            </span>
          </p>
          <p className={`text-xs ${C.text35} mt-0.5`}>
            最低在庫: {item.minStockLevel} {item.unit}
          </p>
        </div>

        <StatusBadge colorClass={getInventoryStatusColor(item.status)}>
          {getInventoryStatusLabel(item.status)}
        </StatusBadge>
      </div>

      {/* Expiry date */}
      {item.expiryDate && (
        <div className={`mt-2 pt-2 border-t ${C.borderLight}`}>
          <p className={`text-xs ${C.text40}`}>
            有効期限:{" "}
            <span className={`${C.text55}`}>{item.expiryDate}</span>
          </p>
        </div>
      )}
    </button>
  );
}
