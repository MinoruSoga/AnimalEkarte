// React/Framework
import { ICON, C, BADGE } from "@/lib/design-tokens";

// External
import { Pill, Stethoscope, Utensils, ClipboardList, MoreHorizontal } from "lucide-react";

// Internal
import { SelectItem } from "@/components/ui/select";

// Types
import type { CarePlanItem, CarePlanItemType, CarePlanTiming } from "../../api/care-plan-items";

// ---- Static constants ----

const TYPE_OPTIONS: { value: CarePlanItemType; label: string }[] = [
  { value: "food", label: "食事" },
  { value: "medicine", label: "投薬" },
  { value: "treatment", label: "処置・検査" },
  { value: "instruction", label: "指示・その他" },
  { value: "item", label: "持ち物" },
];

export const TIMING_OPTIONS: { value: CarePlanTiming; label: string }[] = [
  { value: "morning", label: "朝" },
  { value: "noon", label: "昼" },
  { value: "night", label: "夜" },
];

export const TYPE_SELECT_ITEMS = (
  <>
    {TYPE_OPTIONS.map((opt) => (
      <SelectItem key={opt.value} value={opt.value}>
        {opt.label}
      </SelectItem>
    ))}
  </>
);

// ---- Helper functions ----

export function TypeIcon({ type }: { type: CarePlanItemType }) {
  if (type === "food") return <Utensils className={`${ICON.action} ${C.textDiscount} shrink-0`} />;
  if (type === "medicine") return <Pill className={`${ICON.action} ${C.textBrand} shrink-0`} />;
  if (type === "treatment")
    return <Stethoscope className={`${ICON.action} ${C.textStatusPurple} shrink-0`} />;
  if (type === "instruction")
    return <ClipboardList className={`${ICON.action} ${C.textStatusGreen} shrink-0`} />;
  return <MoreHorizontal className={`${ICON.action} ${C.text40} shrink-0`} />;
}

export function StatusBadge({ status }: { status: CarePlanItem["status"] }) {
  if (status === "active") {
    return (
      <span
        className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${BADGE.blueNoBorder}`}
      >
        実施中
      </span>
    );
  }
  if (status === "completed") {
    return (
      <span
        className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${BADGE.greenNoBorder}`}
      >
        完了
      </span>
    );
  }
  return (
    <span
      className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${BADGE.grayNoBorder}`}
    >
      中止
    </span>
  );
}

export function TimingBadges({ timing }: { timing: CarePlanTiming[] }) {
  return (
    <div className="flex gap-1">
      {TIMING_OPTIONS.map((opt) => (
        <span
          key={opt.value}
          className={`inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium ${
            timing.includes(opt.value) ? BADGE.blueNoBorder : `${C.bgPage} ${C.text30}`
          }`}
        >
          {opt.label}
        </span>
      ))}
    </div>
  );
}
