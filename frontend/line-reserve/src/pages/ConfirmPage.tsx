import { useState, useCallback } from 'react';
import axios from 'axios';
import liff from '@line/liff';
import type { ReservationFlow } from '../types/models';
import { liffApi } from '../api/liff-api';
import { LIFF_MOCK } from '../lib/liff-config';
import { ProgressDots } from '../components/ProgressDots';
import { PrimaryButton } from '../components/PrimaryButton';
import { BackButton } from '../components/BackButton';

interface SlotTakenResponse {
  error: string;
  code: string;
  redirect_step: number;
}

interface ConfirmPageProps {
  clinicId: string;
  idToken: string;
  flow: ReservationFlow;
  reservationNotice?: string;
  cancelNotice?: string;
  privacyPolicy?: string;
  onConfirm: (reservationId: number, notes: string) => void;
  onSlotTaken: (message: string, redirectStep: number) => void;
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

function formatDatePadded(dateStr: string): string {
  if (!dateStr) return '';
  const [year, month, day] = dateStr.split('-');
  const d = new Date(Number(year), Number(month) - 1, Number(day));
  const weekDays = ['日', '月', '火', '水', '木', '金', '土'];
  return `${year}年${month}月${day}日(${weekDays[d.getDay()]})`;
}

/** 予約完了後に LINE トーク画面へメッセージを送信する */
async function sendLiffMessage(flow: ReservationFlow, notes: string): Promise<void> {
  if (LIFF_MOCK) return; // ローカル開発時はスキップ
  if (!liff.isInClient()) return; // LINE アプリ外では送信不可

  const confirmNum = notes.match(/R-\d{8}-\d{4}/)?.[0] ?? '';
  const time = `${formatTime(flow.startTime)}〜${formatTime(flow.endTime)}`;

  const petList = flow.customerInfo.pets
    .map(p => p.type ? `${p.name}(${p.type})` : p.name)
    .join('、');

  const text = [
    'ご予約を承りました。',
    '',
    `■ 予約番号: ${confirmNum}`,
    `■ 日時: ${formatDatePadded(flow.date)} ${time}`,
    `■ コース: ${flow.courseName}`,
    flow.staffId > 0 ? `■ 担当: ${flow.staffName}` : '',
    petList ? `■ ペット: ${petList}` : '',
    '',
    'キャンセルはLINEメニューの',
    '「予約確認・キャンセル」から行えます。',
  ].filter(Boolean).join('\n');

  try {
    await liff.sendMessages([{ type: 'text', text }]);
  } catch {
    // メッセージ送信失敗は予約自体には影響しないので無視
  }
}

export function ConfirmPage({
  clinicId,
  idToken,
  flow,
  reservationNotice,
  cancelNotice,
  privacyPolicy,
  onConfirm,
  onSlotTaken,
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
          course_id: flow.courseId,
          staff_id: flow.staffId,
          date: flow.date,
          start_time: flow.startTime,
          end_time: flow.endTime,
          customer_fields: {
            customer_name: flow.customerInfo.name,
            phone: flow.customerInfo.phone,
            owner_name: flow.customerInfo.ownerName,
            pets: flow.customerInfo.pets.map(p => ({
              name: p.name,
              type: p.type,
              is_new: p.isNew,
            })),
          },
          request_text: flow.requestText,
        },
        idToken,
      );
      await sendLiffMessage(flow, reservation.notes);
      onConfirm(reservation.id, reservation.notes);
    } catch (err: unknown) {
      if (axios.isAxiosError(err) && err.response?.status === 409) {
        const data = err.response.data as SlotTakenResponse;
        const message = data.error || '選択された時間枠は既に予約が入っています。';
        const redirectStep = data.redirect_step || 4;
        onSlotTaken(message, redirectStep);
        return;
      }
      setError('予約の確定に失敗しました。もう一度お試しください。');
    } finally {
      setSubmitting(false);
    }
  }, [clinicId, idToken, flow, onConfirm, onSlotTaken]);

  const petDisplay = flow.customerInfo.pets.length > 0
    ? flow.customerInfo.pets.map(p => p.type ? `${p.name}（${p.type}）` : p.name).join('、')
    : '—';

  const rows: Array<{ label: string; value: string }> = [
    { label: 'お名前', value: flow.customerInfo.name },
    { label: '電話番号', value: flow.customerInfo.phone },
    { label: '飼い主名', value: flow.customerInfo.ownerName || '—' },
    { label: 'ペット', value: petDisplay },
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
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={7} total={8} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-4">予約内容の確認</h2>
        </div>

        <div className="flex-1 px-4 space-y-4">
          {/* 警告メッセージ */}
          <div className="bg-yellow-50 border border-yellow-200 rounded-xl px-4 py-3">
            <p className="text-sm text-yellow-800 font-medium">
              まだ予約は完了していません
            </p>
            <p className="text-xs text-yellow-700 mt-1">
              「予約を確定する」ボタンを押してください
            </p>
          </div>

          {/* 予約内容テーブル */}
          <div className="bg-white rounded-xl border border-gray-200 overflow-hidden shadow-sm">
            {rows.map(row => (
              <div
                key={row.label}
                className="flex border-b border-gray-100 last:border-b-0"
              >
                <div className="w-28 flex-shrink-0 px-4 py-3 bg-noah-teal-light text-sm text-noah-text-sub font-medium">
                  {row.label}
                </div>
                <div className="flex-1 px-4 py-3 text-sm text-noah-text break-words">
                  {row.value}
                </div>
              </div>
            ))}
          </div>

          {reservationNotice ? (
            <div className="bg-blue-50 border border-blue-200 rounded-xl px-4 py-3">
              <p className="text-xs font-medium text-blue-700 mb-1">ご予約にあたって</p>
              <p className="text-sm text-blue-800 whitespace-pre-wrap">{reservationNotice}</p>
            </div>
          ) : null}

          {cancelNotice ? (
            <div className="bg-gray-50 border border-gray-200 rounded-xl px-4 py-3">
              <p className="text-xs font-medium text-gray-500 mb-1">キャンセルポリシー</p>
              <p className="text-sm text-gray-700 whitespace-pre-wrap">{cancelNotice}</p>
            </div>
          ) : null}

          {error ? (
            <div className="bg-red-50 border border-red-200 rounded-lg px-4 py-3">
              <p className="text-sm text-red-700" role="alert">{error}</p>
            </div>
          ) : null}
        </div>

        <div className="px-4 py-6 space-y-3">
          {privacyPolicy ? (
            <p className="text-xs text-center text-gray-500 whitespace-pre-wrap">{privacyPolicy}</p>
          ) : null}
          <PrimaryButton onClick={handleConfirm} disabled={submitting}>
            {submitting ? '送信中...' : '予約を確定する'}
          </PrimaryButton>
        </div>
      </div>
    </div>
  );
}
