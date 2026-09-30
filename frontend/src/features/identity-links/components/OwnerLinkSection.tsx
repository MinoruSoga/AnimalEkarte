import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
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
  // 解除は不可逆性が高いため、確認ダイアログ経由でのみ onUnlinkOwner を呼ぶ。
  const [unlinkTarget, setUnlinkTarget] = useState<OwnerSearchItem | null>(null);
  return (
    <section
      className={`rounded border p-4 space-y-3 ${C.borderLight} ${C.bgWhite}`}
      aria-label="飼主リンク"
    >
      <h2 className={`font-semibold ${C.textInk}`}>飼主リンク</h2>
      <label className="block text-sm">
        <span className={C.textInkMuted}>検索</span>
        <Input
          className="mt-1"
          value={ownerQuery}
          onChange={(e) => setOwnerQuery(e.target.value)}
          placeholder="氏名・カナ・電話"
        />
      </label>
      {ownerQuery.trim() !== "" && ownerHits.length === 0 ? (
        <p className={`text-sm ${C.text60}`}>条件に一致する飼主が見つかりません。</p>
      ) : null}
      <ul className="space-y-1 max-h-40 overflow-auto text-sm">
        {ownerHits.map((o) => {
          const isSelected = selectedOwners.some(
            (s) => s.clinic_id === o.clinic_id && s.owner_id === o.owner_id,
          );
          return (
            <li key={`${o.clinic_id}-${o.owner_id}`}>
              <button
                type="button"
                className={`w-full text-left px-2 py-1 min-h-11 flex items-center rounded ${C.hoverBgLight}`}
                aria-pressed={isSelected}
                onClick={() => toggleOwner(o)}
              >
                [医院 {o.clinic_id}] {o.name} ({o.phone})
              </button>
            </li>
          );
        })}
      </ul>
      <div className={`text-sm ${C.textInkSecondary}`}>
        選択: {selectedOwners.map((o) => `${o.clinic_id}/${o.owner_id}`).join(", ") || "なし"}
        {ownerGroupId != null ? ` / 連携グループ #${ownerGroupId}` : null}
      </div>
      {canEdit ? (
        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            disabled={pending || selectedOwners.length < 2}
            onClick={onLinkOwners}
          >
            飼主をリンク
          </Button>
          {selectedOwners.map((o) => (
            <Button
              key={`unlink-o-${o.clinic_id}-${o.owner_id}`}
              type="button"
              variant="outline"
              className="text-sm"
              disabled={pending || resolveOwnerGroupId(o) == null}
              onClick={() => setUnlinkTarget(o)}
            >
              連携解除 {o.clinic_id}/{o.owner_id}
            </Button>
          ))}
        </div>
      ) : null}
      <ConfirmDialog
        open={unlinkTarget !== null}
        onClose={() => setUnlinkTarget(null)}
        onConfirm={() => {
          if (!unlinkTarget) return;
          onUnlinkOwner(unlinkTarget);
          setUnlinkTarget(null);
        }}
        title="飼主の連携を解除しますか？"
        description={
          unlinkTarget
            ? `医院 ${unlinkTarget.clinic_id} の飼主「${unlinkTarget.name}」を同一飼主グループから解除します。解除後、この飼主はグループ内の他医院データと同一人物として扱われなくなり、連携表示の対象外になります。`
            : undefined
        }
        confirmLabel="解除する"
        cancelLabel="キャンセル"
        variant="destructive"
        isPending={pending}
      />
    </section>
  );
}
