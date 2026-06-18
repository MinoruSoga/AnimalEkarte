import { C } from "@/lib/design-tokens";
import type { Owner } from "@/types";
import { formatDMPreference } from "../lib/dm-preference";

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
  // レガシー EMR（Figma 37:142）の飼主欄に揃える: 郵便番号 + 住所、勤務先、勤務先電話。
  // 既存 useGetOwner のデータに含まれる項目のみ表示し、欠落値は行ごと出さない。
  const address = [owner.postalCode ? `〒${owner.postalCode}` : "", owner.address1, owner.address2]
    .filter(Boolean)
    .join(" ");

  const fields: OwnerField[] = [
    { label: "電話", value: owner.phone || "-", mono: true },
    { label: "会員区分", value: owner.membershipType || "-" },
  ];
  if (address) {
    fields.push({ label: "住所", value: address });
  }
  if (owner.email) {
    fields.push({ label: "メール", value: owner.email });
  }
  if (owner.company) {
    fields.push({ label: "勤務先", value: owner.company });
  }
  if (owner.companyPhone) {
    fields.push({ label: "勤務先TEL", value: owner.companyPhone, mono: true });
  }
  // DM 区分は未設定（undefined/null）の場合は行を出さず、設定済み（必要/不要）のみ表示する。
  // 既存飼主は値を持たないため、レポートを「不要」表記で埋めない。
  if (owner.dmPreference != null) {
    fields.push({ label: "DM", value: formatDMPreference(owner.dmPreference) });
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
