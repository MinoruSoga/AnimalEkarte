import { C } from "@/lib/design-tokens";
import type { OwnerSearchItem } from "@/types/generated/identitylink-responses";

interface OwnerLinkSectionProps {
  canEdit: boolean;
  pending: boolean;
  ownerQuery: string;
  setOwnerQuery: (value: string) => void;
  ownerHits: OwnerSearchItem[];
  selectedOwners: OwnerSearchItem[];
  ownerGroupId: number | null;
  toggleOwner: (item: OwnerSearchItem) => void;
  onLinkOwners: () => void;
  onUnlinkOwner: (item: OwnerSearchItem) => void;
  resolveOwnerGroupId: (item: OwnerSearchItem) => number | null;
}

export function OwnerLinkSection({
  canEdit,
  pending,
  ownerQuery,
  setOwnerQuery,
  ownerHits,
  selectedOwners,
  ownerGroupId,
  toggleOwner,
  onLinkOwners,
  onUnlinkOwner,
  resolveOwnerGroupId,
}: OwnerLinkSectionProps) {
  return (
    <section className={`rounded border p-4 space-y-3 ${C.borderLight} ${C.bgWhite}`} aria-label="飼主リンク">
      <h2 className={`font-semibold ${C.textInk}`}>飼主リンク</h2>
      <label className="block text-sm">
        <span className={C.textInkMuted}>検索</span>
        <input
          className={`mt-1 w-full rounded border px-3 py-2 ${C.borderLight} ${C.bgWhite} ${C.textInk}`}
          value={ownerQuery}
          onChange={(e) => setOwnerQuery(e.target.value)}
          placeholder="氏名・カナ・電話"
        />
      </label>
      <ul className="space-y-1 max-h-40 overflow-auto text-sm">
        {ownerHits.map((o) => (
          <li key={`${o.clinic_id}-${o.owner_id}`}>
            <button
              type="button"
              className={`w-full text-left px-2 py-1 rounded ${C.bgHover}`}
              onClick={() => toggleOwner(o)}
            >
              [医院 {o.clinic_id}] {o.name} ({o.phone})
            </button>
          </li>
        ))}
      </ul>
      <div className={`text-sm ${C.textInkSecondary}`}>
        選択: {selectedOwners.map((o) => `${o.clinic_id}/${o.owner_id}`).join(", ") || "なし"}
        {ownerGroupId != null ? ` / 連携グループ #${ownerGroupId}` : null}
      </div>
      {canEdit ? (
        <div className="flex flex-wrap gap-2">
          <button
            type="button"
            className={`px-3 py-1.5 rounded text-sm ${C.bgBrand} ${C.textOnBrand}`}
            disabled={pending || selectedOwners.length < 2}
            onClick={onLinkOwners}
          >
            飼主をリンク
          </button>
          {selectedOwners.map((o) => (
            <button
              key={`unlink-o-${o.clinic_id}-${o.owner_id}`}
              type="button"
              className={`px-3 py-1.5 rounded text-sm border ${C.borderLight}`}
              disabled={pending || resolveOwnerGroupId(o) == null}
              onClick={() => onUnlinkOwner(o)}
            >
              連携解除 {o.clinic_id}/{o.owner_id}
            </button>
          ))}
        </div>
      ) : null}
    </section>
  );
}
