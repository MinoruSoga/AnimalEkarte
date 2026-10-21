import { useMemo } from "react";
import { PrimaryButton } from "../components/PrimaryButton";
import type { ReservationFlow } from "../types/models";
import { LIFF_REVISIT_INTAKE_ACTIVE } from "../lib/revisit-intake-activation";
import { formatJapaneseDate, formatTimeHHMM } from "@/shared-liff/jst-date";
import {
  REVISIT_INTAKE_UNAVAILABLE_MESSAGE,
  isRevisitIntakeAppointmentIdUsable,
} from "../lib/revisit-intake-appointment-id";

interface CompletePageProps {
  reservationId: string | null;
  notes: string;
  flow: ReservationFlow;
  onMyReservations: () => void;
  onNewReservation: () => void;
  /** EMR-474: 任意の再診事前問診へ進む。予約 ID が無いときは渡さない（CTA 非表示）。 */
  onRevisitIntake?: () => void;
}

function extractConfirmationNumber(notes: string): string | null {
  const match = notes.match(/R-\d{8}-\d{4}/);
  return match ? match[0] : null;
}

export function CompletePage({
  reservationId,
  notes,
  flow,
  onMyReservations,
  onNewReservation,
  onRevisitIntake,
}: CompletePageProps) {
  const confirmationNumber = useMemo(() => extractConfirmationNumber(notes), [notes]);
  const displayNumber =
    confirmationNumber ?? (reservationId !== null ? `R-${reservationId.padStart(6, "0")}` : null);

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1 items-center justify-center px-4 py-8">
        <div className="text-center mb-6">
          <div className="text-6xl mb-4" aria-hidden="true">
            ✅
          </div>
          <h1 className="text-xl font-bold text-noah-teal-dark mb-2">ご予約を承りました</h1>
        </div>

        {/* 予約詳細カード */}
        <div className="w-full bg-white rounded-xl border border-noah-border p-5 mb-4">
          {displayNumber ? (
            <div className="mb-4 pb-4 border-b border-noah-border-light">
              <p className="text-xs text-noah-text-sub mb-1">予約番号</p>
              <p className="text-lg font-bold text-noah-text font-mono">{displayNumber}</p>
            </div>
          ) : null}

          <div className="space-y-3 text-sm">
            {flow.date ? (
              <div className="flex">
                <span className="text-noah-text-sub w-16 shrink-0">日時</span>
                <span className="text-noah-text font-medium">
                  {formatJapaneseDate(flow.date, true)}{" "}
                  {`${formatTimeHHMM(flow.startTime)}〜${formatTimeHHMM(flow.endTime)}`}
                </span>
              </div>
            ) : null}

            {flow.courseName ? (
              <div className="flex">
                <span className="text-noah-text-sub w-16 shrink-0">コース</span>
                <span className="text-noah-text font-medium">{flow.courseName}</span>
              </div>
            ) : null}

            {flow.staffName ? (
              <div className="flex">
                <span className="text-noah-text-sub w-16 shrink-0">担当</span>
                <span className="text-noah-text font-medium">
                  {flow.staffId === 0 ? "指名なし" : flow.staffName}
                </span>
              </div>
            ) : null}
          </div>
        </div>

        {/* EMR-474: 任意の再診事前問診。未回答でも予約は確定済みで、受付も止めない。
            回答先はこの予約の実際の対象（BE の current_target）で、ペットごとの予約作成は約束しない。 */}
        {LIFF_REVISIT_INTAKE_ACTIVE &&
        reservationId !== null &&
        !isRevisitIntakeAppointmentIdUsable(reservationId) ? (
          <div className="w-full bg-white rounded-xl border border-noah-border p-4 mb-4">
            <p role="status" className="text-sm text-noah-text-sub">
              {REVISIT_INTAKE_UNAVAILABLE_MESSAGE}
            </p>
          </div>
        ) : LIFF_REVISIT_INTAKE_ACTIVE && onRevisitIntake ? (
          <div className="w-full bg-white rounded-xl border border-noah-border p-4 mb-4">
            <p className="text-sm text-noah-text mb-1">再診の方は、来院前に問診へ回答できます。</p>
            <p className="text-xs text-noah-text-sub mb-3">
              回答は任意です。回答しなくても予約は確定しています。受付前であれば、予約一覧から訂正できます。
            </p>
            <button
              type="button"
              onClick={onRevisitIntake}
              className="w-full min-h-11 py-3 px-4 border border-noah-teal rounded-xl text-noah-teal-dark font-semibold hover:bg-noah-teal-light"
            >
              事前問診に回答する（任意）
            </button>
          </div>
        ) : null}

        {/* キャンセル案内 */}
        <div className="w-full bg-white/60 rounded-xl p-4 mb-6 text-center">
          <p className="text-xs text-noah-text-sub leading-relaxed">
            キャンセルは「予約確認・キャンセル」から行えます。
          </p>
        </div>

        <div className="w-full space-y-3">
          <PrimaryButton onClick={onMyReservations}>予約確認・キャンセル</PrimaryButton>
          <button
            type="button"
            onClick={onNewReservation}
            className="w-full py-3 px-4 border border-noah-border-input rounded-xl text-noah-text-sub font-semibold hover:bg-white"
          >
            新しい予約をする
          </button>
        </div>
      </div>
    </div>
  );
}
