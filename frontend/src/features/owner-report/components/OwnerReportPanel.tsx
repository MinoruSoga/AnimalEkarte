import { C } from "@/lib/design-tokens";
import type { Owner } from "@/types";

interface OwnerReportPanelProps {
  owner: Owner;
}

/**
 * #158 R4: 飼主情報の常時固定表示。ペット切替で消えず、履歴スクロール中も上部バーに残る。
 * 業務ツール向けに 1〜2 行のコンパクトな identity ストリップとして横並び表示する（カードにしない）。
 */
export function OwnerReportPanel({ owner }: OwnerReportPanelProps) {
  return (
    <div className="flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-0.5">
      <span
        className={`rounded-md border px-2 py-1 text-2xs font-semibold ${C.borderLight} ${C.bgPage} ${C.text70}`}
      >
        飼主ペットレポート
      </span>
      <span className={`text-xs ${C.text50}`}>飼主No</span>
      <span className={`text-sm tabular-nums ${C.text}`}>{owner.id}</span>
      <h1 className={`max-w-full truncate text-xl font-semibold ${C.text}`}>
        {owner.ownerName || "-"}
      </h1>
      {owner.ownerNameKana ? (
        <span className={`truncate text-xs ${C.text50}`}>{owner.ownerNameKana}</span>
      ) : null}
      <span aria-hidden className={C.text25}>
        |
      </span>
      <span className={`text-sm tabular-nums ${C.text}`}>{owner.phone || "-"}</span>
    </div>
  );
}
