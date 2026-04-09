import { useMemo } from 'react';
import { PrimaryButton } from '../components/PrimaryButton';
import type { ReservationFlow } from '../types/models';

interface CompletePageProps {
  reservationId: number;
  notes: string;
  flow: ReservationFlow;
  onMyReservations: () => void;
  onNewReservation: () => void;
}

function extractConfirmationNumber(notes: string): string | null {
  const match = notes.match(/R-\d{8}-\d{4}/);
  return match ? match[0] : null;
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const d = new Date(dateStr + 'T00:00:00');
  const weekdays = ['日', '月', '火', '水', '木', '金', '土'];
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  const w = weekdays[d.getDay()];
  return `${y}年${m}月${day}日(${w})`;
}

function formatTime(startTime: string, endTime: string): string {
  const fmt = (t: string) => `${t.slice(0, 2)}:${t.slice(2)}`;
  return `${fmt(startTime)}〜${fmt(endTime)}`;
}

export function CompletePage({
  reservationId,
  notes,
  flow,
  onMyReservations,
  onNewReservation,
}: CompletePageProps) {
  const confirmationNumber = useMemo(() => extractConfirmationNumber(notes), [notes]);
  const displayNumber = confirmationNumber ?? (reservationId > 0 ? `R-${String(reservationId).padStart(6, '0')}` : null);

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1 items-center justify-center px-4 py-8">
        <div className="text-center mb-6">
          <div className="text-6xl mb-4" aria-hidden="true">✅</div>
          <h1 className="text-xl font-bold text-noah-teal-dark mb-2">ご予約を承りました</h1>
        </div>

        {/* 予約詳細カード */}
        <div className="w-full bg-white rounded-xl border border-gray-200 shadow-sm p-5 mb-4">
          {displayNumber ? (
            <div className="mb-4 pb-4 border-b border-gray-100">
              <p className="text-xs text-noah-text-sub mb-1">予約番号</p>
              <p className="text-lg font-bold text-noah-text font-mono">{displayNumber}</p>
            </div>
          ) : null}

          <div className="space-y-3 text-sm">
            {flow.date ? (
              <div className="flex">
                <span className="text-noah-text-sub w-16 shrink-0">日時</span>
                <span className="text-noah-text font-medium">
                  {formatDate(flow.date)} {formatTime(flow.startTime, flow.endTime)}
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
                  {flow.staffId === 0 ? '指名なし' : flow.staffName}
                </span>
              </div>
            ) : null}
          </div>
        </div>

        {/* キャンセル案内 */}
        <div className="w-full bg-white/60 rounded-xl p-4 mb-6 text-center">
          <p className="text-xs text-noah-text-sub leading-relaxed">
            キャンセルは「予約確認・キャンセル」から行えます。
          </p>
        </div>

        <div className="w-full space-y-3">
          <PrimaryButton onClick={onMyReservations}>
            予約確認・キャンセル
          </PrimaryButton>
          <button
            type="button"
            onClick={onNewReservation}
            className="w-full py-3 px-4 border border-gray-300 rounded-xl text-noah-text-sub font-semibold hover:bg-white"
          >
            新しい予約をする
          </button>
        </div>
      </div>
    </div>
  );
}
