import { GripVertical } from "lucide-react";

import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C, ICON, STYLE } from "@/lib/design-tokens";

import type { Cage } from "../api/cages";
import {
  CAGE_SIZE_LABELS,
  CAGE_TYPE_LABELS,
  formatCagePrice,
} from "./cage-side-panel-model";

interface CageRowOverlayProps {
  cage: Cage;
}

export function CageRowOverlay({ cage }: CageRowOverlayProps) {
  return (
    <div
      className={`flex items-center h-12 ${C.bgWhite} border ${C.borderLight} rounded-xs ${STYLE.dragOverlayShadow} cursor-grabbing`}
      style={{ width: "100%" }}
    >
      <div className={`w-11 shrink-0 flex items-center justify-center ${C.text50}`}>
        <GripVertical className={ICON.action} />
      </div>
      <div className={`flex-1 min-w-0 text-base font-medium ${C.text} px-3`}>{cage.name}</div>
      <div className={`w-[100px] shrink-0 text-base ${C.text65}`}>
        {CAGE_TYPE_LABELS[cage.cageType]}
      </div>
      <div className={`w-[90px] shrink-0 text-base ${C.text65}`}>
        {CAGE_SIZE_LABELS[cage.cageSize]}
      </div>
      <div className={`w-[120px] shrink-0 text-right pr-4 font-mono text-base ${C.text}`}>
        {formatCagePrice(cage.price)}
      </div>
      <div className="w-[90px] shrink-0 flex justify-center">
        <StatusPill isActive={cage.isActive} />
      </div>
      <div className="w-[80px] shrink-0" />
    </div>
  );
}
