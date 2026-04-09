import { useState, useCallback } from 'react';
import type { ReservationFlow } from '../types/models';
import { liffApi } from '../api/liff-api';
import { ProgressDots } from '../components/ProgressDots';
import { PrimaryButton } from '../components/PrimaryButton';
import { BackButton } from '../components/BackButton';

interface ConfirmPageProps {
  clinicId: string;
  idToken: string;
  flow: ReservationFlow;
  onConfirm: (reservationId: number, notes: string) => void;
  onBack: () => void;
}

function formatTime(hhmm: string): string {
  if (!hhmm || hhmm.length < 4) return hhmm;
  return `${hhmm.slice(0, 2)}:${hhmm.slice(2, 4)}`;
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '';
  const [year, month, day] = dateStr.split('-');
  const d = new Date(Number(year), Number(month) - 1, Number(day));
  const weekDays = ['日', '月', '火', '水', '木', '金', '土'];
  return `${year}年${Number(month)}月${Number(day)}日（${weekDays[d.getDay()]}）`;
}

export function ConfirmPage({
  clinicId,
  idToken,
  flow,
  onConfirm,
  onBack,
}: ConfirmPageProps) {
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleConfirm = useCallback(async () => {
    if (!flow.courseId) return;
    setSubmitting(true);
    setError(null);

    try {
      const reservation = await liffApi.createReservation(
        clinicId,
        {
          customer_name: flow.customerInfo.name,
          phone: flow.customerInfo.phone,
          owner_name: flow.customerInfo.ownerName,
          pet_name: flow.customerInfo.petName,
          pet_type: flow.customerInfo.petType,
          course_id: flow.courseId,
          staff_id: flow.staffId,
          date: flow.date,
          start_time: flow.startTime,
          end_time: flow.endTime,
          request_text: flow.requestText,
        },
        idToken,
      );
      onConfirm(reservation.id, reservation.notes);
    } catch {
      setError('予約の確定に失敗しました。もう一度お試しください。');
    } finally {
      setSubmitting(false);
    }
  }, [clinicId, idToken, flow, onConfirm]);

  const rows: Array<{ label: string; value: string }> = [
    { label: 'お名前', value: flow.customerInfo.name },
    { label: '電話番号', value: flow.customerInfo.phone },
    { label: '飼い主名', value: flow.customerInfo.ownerName || '—' },
    { label: 'ペット名', value: flow.customerInfo.petName || '—' },
    { label: 'ペット種類', value: flow.customerInfo.petType || '—' },
    { label: 'コース', value: flow.courseName },
    { label: 'スタッフ', value: flow.staffName || '指名なし' },
    { label: '日付', value: formatDate(flow.date) },
    {
      label: '時間',
      value: `${formatTime(flow.startTime)} 〜 ${formatTime(flow.endTime)}`,
    },
    { label: 'ご要望', value: flow.requestText || '—' },
  ];

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={7} total={8} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-gray-800 mb-4">予約内容の確認</h2>
        </div>

        <div className="flex-1 px-4 space-y-4">
          {/* 警告メッセージ */}
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg px-4 py-3">
            <p className="text-sm text-yellow-800 font-medium">
              まだ予約は完了していません
            </p>
            <p className="text-xs text-yellow-700 mt-1">
              「予約を確定する」ボタンを押してください
            </p>
          </div>

          {/* 予約内容テーブル */}
          <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
            {rows.map(row => (
              <div
                key={row.label}
                className="flex border-b border-gray-100 last:border-b-0"
              >
                <div className="w-28 flex-shrink-0 px-4 py-3 bg-gray-50 text-sm text-gray-600 font-medium">
                  {row.label}
                </div>
                <div className="flex-1 px-4 py-3 text-sm text-gray-800 break-words">
                  {row.value}
                </div>
              </div>
            ))}
          </div>

          {error ? (
            <div className="bg-red-50 border border-red-200 rounded-lg px-4 py-3">
              <p className="text-sm text-red-700" role="alert">{error}</p>
            </div>
          ) : null}
        </div>

        <div className="px-4 py-6">
          <PrimaryButton onClick={handleConfirm} disabled={submitting}>
            {submitting ? '送信中...' : '予約を確定する'}
          </PrimaryButton>
        </div>
      </div>
    </div>
  );
}
