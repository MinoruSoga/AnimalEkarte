import { useCallback, useState, useTransition } from "react";
import { liffApi } from "../api/liff-api";
import type { Reservation } from "../types/models";
import { BackButton } from "../components/BackButton";
import { formatJSTApplicationDate } from "@/shared-liff/jst-date";
import { useFetchState } from "@/shared-liff/use-fetch-state";
import type { RevisitIntakeDraftStore } from "../hooks/use-revisit-intake-drafts";
import { LIFF_REVISIT_INTAKE_ACTIVE } from "../lib/revisit-intake-activation";
import {
  REVISIT_INTAKE_UNAVAILABLE_MESSAGE,
  isRevisitIntakeAppointmentIdUsable,
} from "../lib/revisit-intake-appointment-id";

interface MyReservationsPageProps {
  clinicId: string;
  idToken: string;
  onBack: () => void;
  /**
   * EMR-474: 予約カード行内の再診問診 CTA（LIFF に予約詳細画面が無いため、カード行内を予約詳細相当とする）。
   * 未指定なら CTA を出さない。
   */
  onOpenRevisitIntake?: (reservation: Reservation) => void;
  /**
   * 未確定の送信（memory-only の operation key）が残る予約では、取消後も結果照会の入口を残すために使う。
   * 未指定なら取消済み予約には入口を出さない。
   */
  draftStore?: RevisitIntakeDraftStore;
}

function formatDate(dateStr: string): string {
  if (!dateStr) return "";
  const [year, month, day] = dateStr.split("-");
  return `${year}年${Number(month)}月${Number(day)}日`;
}

function formatCreatedAt(isoStr: string): string {
  return formatJSTApplicationDate(isoStr);
}

const STATUS_LABELS: Record<string, string> = {
  confirmed: "確定",
  pending: "確認中",
  cancelled: "キャンセル済",
  checked_in: "受付済",
  in_consultation: "診察中",
  accounting: "会計中",
  completed: "完了",
  no_show: "未来院",
};

// FE-RC-051: 状態バッジは brand-tokens.css の semantic token（noah-success/warning/info/danger/neutral）を使う。
const STATUS_COLORS: Record<string, string> = {
  confirmed: "bg-noah-success-bg text-noah-success-text",
  pending: "bg-noah-warning-bg text-noah-warning-text",
  cancelled: "bg-noah-neutral-bg text-noah-neutral-text",
  checked_in: "bg-noah-info-bg text-noah-info-text",
  in_consultation: "bg-noah-info-bg text-noah-info-text",
  accounting: "bg-noah-info-bg text-noah-info-text",
  completed: "bg-noah-info-bg text-noah-info-text",
  no_show: "bg-noah-danger-bg-strong text-noah-danger-text",
};

// EMR-474 / D6: 取消済み以外に導線を出し、編集可否は問診 GET の current_target.editable だけが決める。
// 受付後（以降）は書き込みを BE が editable=false で止めるため、閲覧の入口として「回答を見る」を出す。
// no_show は FE で read-only 固定にせず、サーバー判定に従う。
const VIEW_ONLY_ENTRY_STATUSES: ReadonlySet<Reservation["status"]> = new Set([
  "checked_in",
  "in_consultation",
  "accounting",
  "completed",
]);
const REVISIT_ENTRY_STATUSES: ReadonlySet<Reservation["status"]> = new Set([
  "confirmed",
  "pending",
  "no_show",
  ...VIEW_ONLY_ENTRY_STATUSES,
]);

const revisitCtaClassName =
  "min-h-11 w-full rounded-lg border border-noah-teal px-3 text-sm font-semibold text-noah-teal-dark hover:bg-noah-teal-light";

export function MyReservationsPage({
  clinicId,
  idToken,
  onBack,
  onOpenRevisitIntake,
  draftStore,
}: MyReservationsPageProps) {
  const [cancellingId, setCancellingId] = useState<string | null>(null);
  const [confirmingId, setConfirmingId] = useState<string | null>(null);
  const [cancelError, setCancelError] = useState<{ id: string; message: string } | null>(null);
  // FE-RC-024: 手動 pending state ではなく React 19 useTransition で保留状態を管理する。
  const [isCancelling, startCancelTransition] = useTransition();

  const fetcher = useCallback(
    () => liffApi.getMyReservations(clinicId, idToken),
    [clinicId, idToken],
  );
  // R-F22/R-F23: ステータス別メッセージ解決と再試行導線を共通フックに統合。
  // setReservations は cancel 成功時のローカルな楽観的更新にも使う。
  const {
    data: reservations,
    loading,
    error,
    retry,
    setData: setReservations,
  } = useFetchState(fetcher, "予約一覧の取得");

  const handleCancelRequest = (id: string) => {
    setCancelError(null);
    setConfirmingId(id);
  };

  const handleCancelDismiss = () => {
    setConfirmingId(null);
  };

  const handleCancelConfirm = (id: string) => {
    setConfirmingId(null);
    setCancelError(null);
    setCancellingId(id);
    startCancelTransition(async () => {
      try {
        await liffApi.cancelReservation(clinicId, id, idToken);
        setReservations((prev) =>
          (prev ?? []).map((r) => (r.id === id ? { ...r, status: "cancelled" as const } : r)),
        );
      } catch {
        setCancelError({ id, message: "キャンセルに失敗しました。もう一度お試しください。" });
      } finally {
        setCancellingId(null);
      }
    });
  };

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <div className="px-4 pt-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-4">マイ予約</h2>
        </div>

        <div className="flex-1 px-4">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-noah-text-sub">読み込み中...</div>
            </div>
          ) : error ? (
            <div className="py-8 text-center text-noah-danger">
              <p role="alert">{error.message}</p>
              {error.canRetry ? (
                <button
                  type="button"
                  onClick={retry}
                  className="mt-3 text-sm font-medium text-noah-teal-dark underline"
                >
                  再試行
                </button>
              ) : null}
            </div>
          ) : !reservations || reservations.length === 0 ? (
            <div className="py-12 text-center text-noah-text-sub">
              <p className="text-4xl mb-3" aria-hidden="true">
                📅
              </p>
              <p>予約はありません</p>
            </div>
          ) : (
            <div className="space-y-3">
              {reservations.map((reservation) => (
                <div
                  key={reservation.id}
                  className="bg-white rounded-xl border border-noah-border p-4"
                >
                  <div className="flex items-start justify-between gap-2 mb-2">
                    <div className="min-w-0 flex-1 break-words">
                      {/* EMR-519: ペット名は pet_id が現在 target の pet と一致するときだけ
                          表示する契約。このページは current_target を持たないため表示しない
                          （HOLD・補完しない）。 */}
                      <p className="font-semibold text-noah-text">{reservation.course_name}</p>
                      {reservation.staff_name ? (
                        <p className="text-sm text-noah-text-sub">{reservation.staff_name}</p>
                      ) : null}
                    </div>
                    <span
                      className={`shrink-0 whitespace-nowrap text-xs px-2 py-1 rounded-full font-medium ${
                        STATUS_COLORS[reservation.status] ??
                        "bg-noah-neutral-bg text-noah-neutral-text"
                      }`}
                    >
                      {STATUS_LABELS[reservation.status] ?? reservation.status}
                    </span>
                  </div>

                  <div className="text-sm text-noah-text space-y-1">
                    <p>📅 {formatDate(reservation.date)}</p>
                    <p>
                      🕐 {reservation.start_time} 〜 {reservation.end_time}
                    </p>
                  </div>

                  {reservation.created_at ? (
                    <p className="text-xs text-noah-text-faint mt-2">
                      {formatCreatedAt(reservation.created_at)}
                    </p>
                  ) : null}

                  {LIFF_REVISIT_INTAKE_ACTIVE &&
                  reservation.status === "cancelled" &&
                  onOpenRevisitIntake &&
                  isRevisitIntakeAppointmentIdUsable(reservation.id) &&
                  draftStore?.findPending(clinicId, reservation.id) !== undefined ? (
                    // 取消済みでは通常の問診導線を出さない。未確定の送信がある間だけ、結果照会に戻る入口を残す。
                    <div className="mt-3 pt-3 border-t border-noah-border-light">
                      <button
                        type="button"
                        onClick={() => onOpenRevisitIntake(reservation)}
                        className={revisitCtaClassName}
                      >
                        送信結果を確認する
                      </button>
                    </div>
                  ) : null}

                  {LIFF_REVISIT_INTAKE_ACTIVE &&
                  REVISIT_ENTRY_STATUSES.has(reservation.status) &&
                  onOpenRevisitIntake ? (
                    <div className="mt-3 pt-3 border-t border-noah-border-light">
                      {isRevisitIntakeAppointmentIdUsable(reservation.id) ? (
                        <button
                          type="button"
                          onClick={() => onOpenRevisitIntake(reservation)}
                          className={revisitCtaClassName}
                        >
                          {VIEW_ONLY_ENTRY_STATUSES.has(reservation.status)
                            ? "回答を見る"
                            : "問診を回答する／訂正する"}
                        </button>
                      ) : (
                        <p role="status" className="text-sm text-noah-text-sub">
                          {REVISIT_INTAKE_UNAVAILABLE_MESSAGE}
                        </p>
                      )}
                    </div>
                  ) : null}

                  {reservation.status === "confirmed" ? (
                    <div className="mt-3 pt-3 border-t border-noah-border-light">
                      {confirmingId === reservation.id ? (
                        <div className="space-y-2">
                          <p className="text-sm text-noah-text">本当にキャンセルしますか？</p>
                          <div className="flex gap-2">
                            <button
                              type="button"
                              onClick={() => handleCancelConfirm(reservation.id)}
                              disabled={Boolean(isCancelling && cancellingId === reservation.id)}
                              className="text-sm text-white bg-noah-danger hover:bg-noah-danger-hover disabled:opacity-50 px-3 py-1.5 rounded-lg"
                            >
                              {isCancelling && cancellingId === reservation.id
                                ? "キャンセル中..."
                                : "はい、キャンセルする"}
                            </button>
                            <button
                              type="button"
                              onClick={handleCancelDismiss}
                              disabled={Boolean(isCancelling && cancellingId === reservation.id)}
                              className="text-sm text-noah-text-sub hover:text-noah-text disabled:opacity-50 px-3 py-1.5 rounded-lg border border-noah-border"
                            >
                              いいえ
                            </button>
                          </div>
                        </div>
                      ) : (
                        <button
                          type="button"
                          onClick={() => handleCancelRequest(reservation.id)}
                          disabled={Boolean(isCancelling && cancellingId === reservation.id)}
                          className="text-sm text-noah-danger hover:text-noah-danger-text disabled:opacity-50"
                        >
                          キャンセルする
                        </button>
                      )}

                      {cancelError?.id === reservation.id ? (
                        <div className="mt-2 bg-noah-danger-bg border border-noah-danger-border rounded-lg px-4 py-3">
                          <p className="text-sm text-noah-danger-text" role="alert">
                            {cancelError.message}
                          </p>
                        </div>
                      ) : null}
                    </div>
                  ) : null}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
