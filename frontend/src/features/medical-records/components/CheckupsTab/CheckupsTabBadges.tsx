import { C, PALETTE } from "@/lib/design-tokens";

export type LstepStatus = "synced" | "not-linked" | "opt-out";

export function LstepStatusBadge({ status }: { status: LstepStatus }) {
  if (status === "synced") {
    return (
      <span
        className="inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full text-white"
        style={{ backgroundColor: PALETTE.lineGreen }}
      >
        LINE通知対象
      </span>
    );
  }
  if (status === "not-linked") {
    return (
      <span
        className={`inline-flex items-center text-xs font-medium px-2 py-0.5 rounded-full border ${C.textNotice} ${C.borderNotice} ${C.bgNotice40}`}
      >
        LINE未連携
      </span>
    );
  }
  return (
    <span
      className={`inline-flex items-center text-xs font-medium px-2 py-0.5 rounded-full border ${C.text40} ${C.borderMediumLight} ${C.bgPage30}`}
    >
      LINE受信拒否
    </span>
  );
}
