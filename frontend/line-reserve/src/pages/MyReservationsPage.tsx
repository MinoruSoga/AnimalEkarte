import { useState, useEffect, useCallback } from 'react';
import type { Reservation } from '../types/models';
import { liffApi } from '../api/liff-api';
import { BackButton } from '../components/BackButton';

interface MyReservationsPageProps {
  clinicId: string;
  idToken: string;
  onBack: () => void;
}

function formatTime(hhmm: string): string {
  if (!hhmm || hhmm.length < 4) return hhmm;
  return `${hhmm.slice(0, 2)}:${hhmm.slice(2, 4)}`;
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const [year, month, day] = dateStr.split('-');
  return `${year}年${Number(month)}月${Number(day)}日`;
}

const STATUS_LABELS: Record<string, string> = {
  confirmed: '確定',
  cancelled: 'キャンセル済',
  completed: '完了',
};

const STATUS_COLORS: Record<string, string> = {
  confirmed: 'bg-green-100 text-green-800',
  cancelled: 'bg-gray-100 text-gray-600',
  completed: 'bg-blue-100 text-blue-800',
};

export function MyReservationsPage({
  clinicId,
  idToken,
  onBack,
}: MyReservationsPageProps) {
  const [reservations, setReservations] = useState<Reservation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [cancellingId, setCancellingId] = useState<number | null>(null);

  const loadReservations = useCallback(() => {
    setLoading(true);
    liffApi.getMyReservations(clinicId, idToken)
      .then(data => {
        setReservations(data);
        setError(null);
      })
      .catch(() => {
        setError('予約一覧の取得に失敗しました');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [clinicId, idToken]);

  useEffect(() => {
    loadReservations();
  }, [loadReservations]);

  const handleCancel = useCallback(async (id: number) => {
    const confirmed = window.confirm('この予約をキャンセルしますか？');
    if (!confirmed) return;

    setCancellingId(id);
    try {
      await liffApi.cancelReservation(clinicId, id, idToken);
      setReservations(prev =>
        prev.map(r => r.id === id ? { ...r, status: 'cancelled' as const } : r)
      );
    } catch {
      alert('キャンセルに失敗しました。もう一度お試しください。');
    } finally {
      setCancellingId(null);
    }
  }, [clinicId, idToken]);

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
            <div className="py-8 text-center text-red-500">{error}</div>
          ) : reservations.length === 0 ? (
            <div className="py-12 text-center text-noah-text-sub">
              <p className="text-4xl mb-3" aria-hidden="true">📅</p>
              <p>予約はありません</p>
            </div>
          ) : (
            <div className="space-y-3">
              {reservations.map(reservation => (
                <div
                  key={reservation.id}
                  className="bg-white rounded-xl border border-gray-200 shadow-sm p-4"
                >
                  <div className="flex items-start justify-between mb-2">
                    <div>
                      <p className="font-semibold text-noah-text">{reservation.course_name}</p>
                      <p className="text-sm text-noah-text-sub">{reservation.staff_name}</p>
                    </div>
                    <span
                      className={`text-xs px-2 py-1 rounded-full font-medium ${
                        STATUS_COLORS[reservation.status] ?? 'bg-gray-100 text-gray-600'
                      }`}
                    >
                      {STATUS_LABELS[reservation.status] ?? reservation.status}
                    </span>
                  </div>

                  <div className="text-sm text-noah-text space-y-1">
                    <p>📅 {formatDate(reservation.date)}</p>
                    <p>
                      🕐 {formatTime(reservation.start_time)} 〜 {formatTime(reservation.end_time)}
                    </p>
                  </div>

                  {reservation.status === 'confirmed' ? (
                    <div className="mt-3 pt-3 border-t border-gray-100">
                      <button
                        type="button"
                        onClick={() => handleCancel(reservation.id)}
                        disabled={cancellingId === reservation.id}
                        className="text-sm text-red-500 hover:text-red-700 disabled:opacity-50"
                      >
                        {cancellingId === reservation.id ? 'キャンセル中...' : 'キャンセルする'}
                      </button>
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
