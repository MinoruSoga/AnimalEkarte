import { C } from "@/lib/design-tokens";
import type { Owner } from "@/types";

interface OwnerReportPanelProps {
  owner: Owner;
}

/**
 * #158 R4: 飼主情報の常時固定表示パネル。ペット切替で消えない。
 */
export function OwnerReportPanel({ owner }: OwnerReportPanelProps) {
  return (
    <div className={`rounded-lg border ${C.borderLight} ${C.bgWhite} p-4`}>
      <div className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
        <span className={`text-xs ${C.text50}`}>飼主No</span>
        <span className={`text-sm font-mono ${C.text}`}>{owner.id}</span>
        <h1 className={`text-lg font-semibold ${C.text}`}>{owner.ownerName || "-"}</h1>
        {owner.ownerNameKana ? (
          <span className={`text-sm ${C.text50}`}>{owner.ownerNameKana}</span>
        ) : null}
      </div>
      <dl className="mt-2 flex flex-wrap gap-x-6 gap-y-1">
        <div className="flex items-baseline gap-1.5">
          <dt className={`text-xs ${C.text50}`}>電話</dt>
          <dd className={`text-sm font-mono ${C.text}`}>{owner.phone || "-"}</dd>
        </div>
        <div className="flex items-baseline gap-1.5">
          <dt className={`text-xs ${C.text50}`}>会員区分</dt>
          <dd className={`text-sm ${C.text}`}>{owner.membershipType || "-"}</dd>
        </div>
        {owner.email ? (
          <div className="flex items-baseline gap-1.5">
            <dt className={`text-xs ${C.text50}`}>メール</dt>
            <dd className={`text-sm ${C.text}`}>{owner.email}</dd>
          </div>
        ) : null}
      </dl>
    </div>
  );
}
