import { useLayoutEffect, useState, useRef } from "react";
import { C } from "@/lib/design-tokens";
import { calcAgeAt } from "@/lib/calc-age";
import { toJSTWallDate } from "@/lib/jst-date";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { useRevokePetDeath } from "@/hooks/use-revoke-pet-death";

interface PetDeceasedBannerProps {
  deceasedAt: string;
  deceasedReason?: string | null;
  birthDate?: string;
  petId: string;
  canEdit?: boolean;
  /** BUG-407: 死亡記録解除成功時に外側フォームのローカル状態へ同期するためのコールバック */
  onRevoked?: () => void;
}

// FE4-9 fix: `new Date(deceasedAt)` の後にブラウザローカル TZ の getter で整形していたため、
// 非 JST ブラウザでは前日が表示されるバグがあった。toJSTWallDate で JST 壁時計へ変換してから
// 整形する（曜日表示なしの現行フォーマットはそのまま維持）。
function formatDeceasedDate(deceasedAt: string): string {
  const d = toJSTWallDate(deceasedAt);
  const yyyy = d.getFullYear();
  const mm = d.getMonth() + 1;
  const dd = d.getDate();
  return `${yyyy}年${mm}月${dd}日`;
}

export function PetDeceasedBanner({
  deceasedAt,
  deceasedReason,
  birthDate,
  petId,
  canEdit = false,
  onRevoked,
}: PetDeceasedBannerProps) {
  const [confirmOpen, setConfirmOpen] = useState(false);
  const revokeButtonRef = useRef<HTMLButtonElement>(null);
  const mutation = useRevokePetDeath();
  const canEditRef = useRef(canEdit);
  useLayoutEffect(() => {
    canEditRef.current = canEdit;
  }, [canEdit]);

  const age = birthDate ? calcAgeAt(deceasedAt, birthDate) : null;
  const formattedDate = formatDeceasedDate(deceasedAt);
  const trimmedReason = deceasedReason?.trim() || "";

  const handleRevokeConfirm = () => {
    if (canEditRef.current !== true) return;
    mutation.mutate(petId, {
      onSuccess: () => onRevoked?.(),
      onSettled: () => setConfirmOpen(false),
    });
  };

  return (
    <div className={`rounded-lg border ${C.borderGray300} ${C.bgGray100} p-3 text-sm`}>
      <div className="flex items-center justify-between gap-2">
        <p className={`${C.text70}`}>
          {age !== null ? (
            <span>
              享年: <span className={`font-medium ${C.text}`}>{age}歳</span>
              <span className={`ml-1.5 ${C.text50}`}>
                （{formattedDate} 永眠）
              </span>
            </span>
          ) : (
            <span>
              <span className={C.text50}>{formattedDate} 永眠</span>
            </span>
          )}
        </p>

        {canEdit ? (
          <button
            ref={revokeButtonRef}
            type="button"
            onClick={() => setConfirmOpen(true)}
            className={`text-xs ${C.text50} underline hover:no-underline ${C.hoverText} transition-colors shrink-0`}
            disabled={mutation.isPending}
          >
            死亡記録を解除
          </button>
        ) : null}
      </div>

      {trimmedReason ? (
        <p className={`mt-1.5 ${C.text50}`}>
          死亡理由: <span className={C.text70}>{trimmedReason}</span>
        </p>
      ) : null}

      {canEdit ? (
        <ConfirmDialog
          open={confirmOpen}
          onClose={() => setConfirmOpen(false)}
          onConfirm={handleRevokeConfirm}
          title="死亡記録を解除しますか？"
          description="解除するとこのペットは「生存」ステータスに戻ります。Lステップタグは自動では復元されません。必要に応じて手動で再同期してください。"
          confirmLabel="解除する"
          cancelLabel="キャンセル"
          variant="destructive"
          isPending={mutation.isPending}
          triggerRef={revokeButtonRef as React.RefObject<HTMLElement>}
        />
      ) : null}
    </div>
  );
}
