import { useCallback, useMemo } from "react";
import { X, Briefcase } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
}

const PLACEHOLDER = "__none__";

export function ReservationTypeOccupationsSection({ clinicId, reservationTypeId }: Props) {
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

  const handleUnlink = useCallback(
    (id: number) => {
      unlinkMutate(id);
    },
    [unlinkMutate],
  );

  const linkedBadges = useMemo(
    () =>
      linked.map((item) => (
        <div
          key={item.id}
          className={`inline-flex items-center gap-1 text-sm px-2 py-0.5 rounded-full border ${C.borderMedium} ${C.text} ${C.bgWhite}`}
        >
          <span>{item.occupation?.name ?? "—"}</span>
          <button
            type="button"
            onClick={() => handleUnlink(item.id)}
            className={`${C.text40} ${C.hoverTextDanger} transition-colors`}
            aria-label={`${item.occupation?.name ?? "職種"} の紐付けを解除`}
          >
            <X className={ICON.smXs} />
          </button>
        </div>
      )),
    [linked, handleUnlink],
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
    <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
      <div className="flex items-center gap-1.5 mb-3">
        <Briefcase className={ICON.smXs} style={{ color: C.text50 }} />
        <p className={`text-xs font-medium ${C.text50}`}>紐付け職種（出勤なし → 予約不可）</p>
      </div>

      {/* 紐付き職種バッジ */}
      {linked.length > 0 ? <div className="flex flex-wrap gap-1.5 mb-3">{linkedBadges}</div> : null}

      {/* 追加セレクト */}
      {availableOccupations.length > 0 ? (
        <Select value={PLACEHOLDER} onValueChange={handleLink}>
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="職種を追加..." />
          </SelectTrigger>
          <SelectContent>{occupationSelectItems}</SelectContent>
        </Select>
      ) : (
        <p className={`text-sm ${C.text40}`}>
          {allOccupations.length === 0 ? "職種マスタが未設定です" : "すべての職種を紐付け済みです"}
        </p>
      )}
    </div>
  );
}
