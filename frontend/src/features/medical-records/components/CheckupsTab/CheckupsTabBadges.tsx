import { BADGE } from "@/lib/design-tokens";

export type LstepStatus = "synced" | "not-linked" | "opt-out";

const LSTEP_STATUS_STYLES: Record<LstepStatus, { label: string; badgeClass: string }> = {
  synced: { label: "LINE通知対象", badgeClass: BADGE.green },
  "not-linked": { label: "LINE未連携", badgeClass: BADGE.yellow },
  "opt-out": { label: "LINE受信拒否", badgeClass: BADGE.gray },
};

export function LstepStatusBadge({ status }: { status: LstepStatus }) {
  const { label, badgeClass } = LSTEP_STATUS_STYLES[status];
  return (
    <span
      className={`inline-flex items-center text-xs font-medium px-2 py-0.5 rounded-full border ${badgeClass}`}
    >
      {label}
    </span>
  );
}
