import { useCallback, useMemo, useState } from "react";
import { X, Briefcase } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import { FieldHelp } from "@/components/shared/FieldHelp";
import { C, STYLE, ICON } from "@/lib/design-tokens";
import { useGetAllOccupations } from "../api/occupations";
import {
  useGetReservationTypeOccupations,
  useLinkOccupation,
  useUnlinkOccupation,
} from "../api/reservation-type-occupations";

// ─────────────────────────────────────────────────
// Component
// ─────────────────────────────────────────────────

interface Props {
  clinicId: string;
  reservationTypeId: string;
  /** 参照権限のみのパネル表示。紐付け追加・解除の mutation UI を描画しない */
  readOnly?: boolean;
}

const PLACEHOLDER = "__none__";

export function ReservationTypeOccupationsSection({
  clinicId,
  reservationTypeId,
  readOnly = false,
}: Props) {
  const { data: linked = [] } = useGetReservationTypeOccupations(clinicId, reservationTypeId);
  const { data: allOccupations = [] } = useGetAllOccupations();
  const linkMutation = useLinkOccupation(clinicId, reservationTypeId);
  const { mutate } = linkMutation;
  const unlinkMutation = useUnlinkOccupation(clinicId, reservationTypeId);
  const { mutate: unlinkMutate } = unlinkMutation;

  const linkedOccupationIds = useMemo(() => new Set(linked.map((l) => l.occupationId)), [linked]);

  // 未紐付けの職種のみを選択肢に表示
  const availableOccupations = useMemo(
    () => allOccupations.filter((o) => !linkedOccupationIds.has(Number(o.id))),
    [allOccupations, linkedOccupationIds],
  );

  // エラー通知は use-reservation-type-occupations 側の onError に一本化（二重トースト防止）
  const handleLink = useCallback(
    (value: string) => {
      if (value === PLACEHOLDER) return;
      mutate(Number(value));
    },
    [mutate],
  );

  // 紐付け解除はレコード削除なので ConfirmDialog 経由（直行削除禁止）
  const [pendingUnlink, setPendingUnlink] = useState<{ id: number; name: string } | null>(null);

  const handleUnlinkRequest = useCallback((id: number, name: string) => {
    setPendingUnlink({ id, name });
  }, []);

  const handleUnlinkCancel = useCallback(() => {
    setPendingUnlink(null);
  }, []);

  const handleUnlinkConfirm = useCallback(() => {
    if (pendingUnlink === null) return;
    unlinkMutate(pendingUnlink.id);
    setPendingUnlink(null);
  }, [pendingUnlink, unlinkMutate]);

  const linkedBadges = useMemo(
    () =>
      linked.map((item) => {
        const occupationName = item.occupation?.name ?? "—";
        return (
          <div
            key={item.id}
            className={`inline-flex items-center gap-1 text-sm px-2 py-0.5 rounded-full border ${C.borderMedium} ${C.text} ${C.bgWhite}`}
          >
            <span>{occupationName}</span>
            {readOnly ? null : (
              <button
                type="button"
                onClick={() => handleUnlinkRequest(item.id, occupationName)}
                className={`-m-3.5 flex min-h-11 min-w-11 items-center justify-center rounded-full outline-none ${C.text50} ${C.hoverTextDanger} focus-visible:ring-2 ${C.focusRingAccent40} transition-colors`}
                aria-label={`${occupationName} の紐付けを解除`}
              >
                <X className={ICON.smXs} />
              </button>
            )}
          </div>
        );
      }),
    [linked, handleUnlinkRequest, readOnly],
  );

  const occupationSelectItems = useMemo(
    () =>
      availableOccupations.map((o) => (
        <SelectItem key={o.id} value={o.id}>
          {o.name}
        </SelectItem>
      )),
    [availableOccupations],
  );

  return (
    <div data-testid="linked-occupations-section" className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
      <div className="flex items-center gap-1.5 mb-3">
        <Briefcase className={`${ICON.smXs} ${C.text50}`} />
        <p className={`text-xs font-medium ${C.text50}`}>紐付け職種（出勤なし → 予約不可）</p>
        <FieldHelp
          label="紐付け職種"
          content="この予約区分を担当できる職種です。紐付けた職種のスタッフが出勤しない日は、この区分の予約ができなくなります。"
        />
      </div>

      {/* 紐付き職種バッジ */}
      {linked.length > 0 ? <div className="flex flex-wrap gap-1.5 mb-3">{linkedBadges}</div> : null}

      {/* 追加セレクト */}
      {readOnly ? null : availableOccupations.length > 0 ? (
        <Select value={PLACEHOLDER} onValueChange={handleLink}>
          <SelectTrigger className={STYLE.selectCompact} aria-label="職種を追加">
            <SelectValue placeholder="職種を追加..." />
          </SelectTrigger>
          <SelectContent>{occupationSelectItems}</SelectContent>
        </Select>
      ) : (
        <p className={`text-sm ${C.text60}`}>
          {allOccupations.length === 0 ? "職種マスタが未設定です" : "すべての職種を紐付け済みです"}
        </p>
      )}

      <ConfirmDialog
        open={pendingUnlink !== null}
        onClose={handleUnlinkCancel}
        onConfirm={handleUnlinkConfirm}
        title="職種の紐付けを解除しますか？"
        description={
          pendingUnlink === null
            ? undefined
            : `「${pendingUnlink.name}」の紐付けを解除します。解除後はこの職種の出勤状況に関係なく予約できるようになります。`
        }
        confirmLabel="解除"
        variant="destructive"
        isPending={unlinkMutation.isPending}
      />
    </div>
  );
}
