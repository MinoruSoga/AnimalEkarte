import { C } from "@/lib/design-tokens";
import type { Owner } from "@/types";

interface OwnerReportPanelProps {
  owner: Owner;
}

interface OwnerField {
  label: string;
  value: string;
  mono?: boolean;
}

/**
 * #158 R4: 飼主情報の常時固定表示。ペット切替で消えず、履歴スクロール中も上部バーに残る。
 * 業務ツール向けに 1〜2 行のコンパクトな identity ストリップとして横並び表示する（カードにしない）。
 */
export function OwnerReportPanel({ owner }: OwnerReportPanelProps) {
  const fields: OwnerField[] = [
    { label: "電話", value: owner.phone || "-", mono: true },
    { label: "会員区分", value: owner.membershipType || "-" },
  ];
  if (owner.email) {
    fields.push({ label: "メール", value: owner.email });
  }

  return (
    <div className="flex min-w-0 flex-wrap items-baseline gap-x-3 gap-y-0.5">
      <span className={`text-xs ${C.text50}`}>飼主No</span>
      <span className={`font-mono text-sm ${C.text}`}>{owner.id}</span>
      <h1 className={`max-w-full truncate text-base font-semibold ${C.text}`}>
        {owner.ownerName || "-"}
      </h1>
      {owner.ownerNameKana ? (
        <span className={`truncate text-xs ${C.text50}`}>{owner.ownerNameKana}</span>
      ) : null}
      <span aria-hidden className={`${C.text25}`}>
        |
      </span>
      {fields.map((field) => (
        <span key={field.label} className="flex items-baseline gap-1">
          <span className={`text-xs ${C.text50}`}>{field.label}</span>
          <span className={`text-sm ${field.mono ? "font-mono" : ""} ${C.text}`}>{field.value}</span>
        </span>
      ))}
    </div>
  );
}
