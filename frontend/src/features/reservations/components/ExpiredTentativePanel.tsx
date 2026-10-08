import { memo } from "react";
import { AlertTriangle, CalendarCheck, CalendarClock, CircleSlash } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useGetPet } from "@/hooks/use-pet";
import { C, ICON } from "@/lib/design-tokens";
import { isPetDeceasedForClinicalWrite } from "@/lib/transforms/pet";
import type { Reservation } from "../types";
import { EXPIRED_TENTATIVE_LABEL, tentativeIntendedDate } from "../lib/tentative";

interface ExpiredTentativePanelProps {
  /** 期限超過の tentative 行（呼び出し側で isExpiredTentativeReservation 済み） */
  appointments: Reservation[];
  /** 確定ボタンの表示可否（confirm-tentative = reservations:edit 相当） */
  canConfirm?: boolean;
  /** 再設定ボタンの表示可否（PUT next-visit-intent = reservations:create 相当） */
  canReset?: boolean;
  /** 終了ボタンの表示可否（DELETE next-visit-intent = reservations:edit 相当） */
  canEnd?: boolean;
  onConfirm: (reservation: Reservation) => void;
  onReset: (reservation: Reservation) => void;
  onEnd: (reservation: Reservation) => void;
}

interface ExpiredTentativeRowProps {
  reservation: Reservation;
  canConfirm: boolean;
  canReset: boolean;
  canEnd: boolean;
  onConfirm: (reservation: Reservation) => void;
  onReset: (reservation: Reservation) => void;
  onEnd: (reservation: Reservation) => void;
}

/**
 * 行単位コンポーネント — useGetPet を行ごとに呼ぶための境界。
 * 「確定」は tentative を実予約化する臨床 write のため死亡ペットは隠す
 * （EMR-177 positive-match 契約: pet 不明時は表示維持し BE 側ガードに委譲）。
 * 「再設定」「終了」は intent 操作のまま（実予約を作らない）のでゲートしない。
 */
function ExpiredTentativeRow({
  reservation,
  canConfirm,
  canReset,
  canEnd,
  onConfirm,
  onReset,
  onEnd,
}: ExpiredTentativeRowProps) {
  const { data: pet } = useGetPet(reservation.petId ?? "");
  const isDeceasedPet = pet ? isPetDeceasedForClinicalWrite(pet) : false;

  return (
    <li className={`flex flex-wrap items-center gap-x-3 gap-y-1 rounded-md ${C.bgWhite} px-3 py-2`}>
      <div className="min-w-0 flex-1">
        <span className={`text-sm font-medium ${C.text}`}>
          {reservation.petName}
          {reservation.ownerName ? `（${reservation.ownerName}）` : ""}
        </span>
        <span className={`ml-2 text-xs ${C.text60}`}>
          予定日: {tentativeIntendedDate(reservation)}
        </span>
      </div>
      <div className="flex gap-1.5">
        {canConfirm && !isDeceasedPet ? (
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => onConfirm(reservation)}
            className={`h-9 gap-1 text-xs ${C.borderMedium} ${C.bgWhite} ${C.text}`}
            aria-label={`${reservation.petName}の次回予定を確定`}
          >
            <CalendarCheck className={ICON.xs} />
            確定
          </Button>
        ) : null}
        {canReset ? (
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => onReset(reservation)}
            className={`h-9 gap-1 text-xs ${C.borderMedium} ${C.bgWhite} ${C.text}`}
            aria-label={`${reservation.petName}の次回予定を再設定`}
          >
            <CalendarClock className={ICON.xs} />
            再設定
          </Button>
        ) : null}
        {canEnd ? (
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => onEnd(reservation)}
            className={`h-9 gap-1 text-xs ${C.borderMedium} ${C.bgWhite} ${C.text60}`}
            aria-label={`${reservation.petName}の次回予定を終了`}
          >
            <CircleSlash className={ICON.xs} />
            終了
          </Button>
        ) : null}
      </div>
    </li>
  );
}

/**
 * EMR-308 A3: 意図日が過ぎた tentative（次回予定・未予約）の actionable list。
 * カレンダーグリッドには過去日の幻影を出さないため、ここで一覧化して
 * 「確定」（枠あり予約へ昇格）/「再設定」（新しい意図日を入れ直す）/「終了」（意思表示を閉じる）を提供する。
 * パネル自体の操作失敗は各 mutation の toast で通知し、カレンダー機能をブロックしない。
 */
export const ExpiredTentativePanel = memo(function ExpiredTentativePanel({
  appointments,
  canConfirm = false,
  canReset = false,
  canEnd = false,
  onConfirm,
  onReset,
  onEnd,
}: ExpiredTentativePanelProps) {
  if (appointments.length === 0) return null;

  return (
    <section
      aria-label={EXPIRED_TENTATIVE_LABEL}
      className={`mx-4 mt-3 shrink-0 rounded-lg border ${C.borderWarning20} ${C.bgWarning50} px-3 py-2`}
    >
      <div className="flex items-center gap-2">
        <AlertTriangle className={`${ICON.action} ${C.textWarningIcon} shrink-0`} />
        <h2 className={`text-sm font-bold ${C.textWarning}`}>
          {EXPIRED_TENTATIVE_LABEL}
          <span className="ml-1 font-normal">（{appointments.length}件）</span>
        </h2>
        <p className={`text-2xs ${C.textWarning} hidden sm:inline`}>
          次回予定の日付を過ぎています。確定・再設定・終了のいずれかを選択してください。
        </p>
      </div>
      <ul className="mt-2 max-h-36 space-y-1.5 overflow-y-auto">
        {appointments.map((reservation) => (
          <ExpiredTentativeRow
            key={reservation.id}
            reservation={reservation}
            canConfirm={canConfirm}
            canReset={canReset}
            canEnd={canEnd}
            onConfirm={onConfirm}
            onReset={onReset}
            onEnd={onEnd}
          />
        ))}
      </ul>
    </section>
  );
});
