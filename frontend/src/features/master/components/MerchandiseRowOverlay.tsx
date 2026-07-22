import { GripVertical } from "lucide-react";

import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";

import type { FrontendMerchandiseItem } from "../api/merchandise-items";
import {
  formatMerchandiseTaxRate,
  MERCHANDISE_CATEGORY_LABELS,
} from "./merchandise-side-panel-model";

interface MerchandiseRowOverlayProps {
  item: FrontendMerchandiseItem;
}

export function MerchandiseRowOverlay({ item }: MerchandiseRowOverlayProps) {
  return (
    <div
      className={`flex items-center h-12 ${C.bgWhite} border ${C.borderLight} rounded-xs ${STYLE.dragOverlayShadow} cursor-grabbing`}
      style={{ width: "100%" }}
    >
      <div className={`w-11 shrink-0 flex items-center justify-center ${C.text50}`}>
        <GripVertical className={ICON.action} />
      </div>
      <div className={`flex-1 min-w-0 text-base font-medium ${C.text} px-3`}>{item.name}</div>
      <div className={`w-[90px] shrink-0 text-base ${C.text70} text-center`}>
        {MERCHANDISE_CATEGORY_LABELS[item.category] ?? item.category}
      </div>
      <div className={`w-[120px] shrink-0 text-right pr-4 font-mono text-base ${C.text}`}>
        {formatCurrency(item.unitPrice)}
      </div>
      <div className={`w-[80px] shrink-0 text-base ${C.text70} text-center`}>
        {formatMerchandiseTaxRate(item.taxRate)}
      </div>
      <div className="w-[90px] shrink-0 flex justify-center">
        <StatusPill isActive={item.isActive} />
      </div>
      <div className="w-[80px] shrink-0" />
    </div>
  );
}
