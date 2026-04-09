import { useState, useEffect } from 'react';
import type { AvailableDate } from '../types/models';
import { liffApi } from '../api/liff-api';
import { ProgressDots } from '../components/ProgressDots';
import { PrimaryButton } from '../components/PrimaryButton';
import { BackButton } from '../components/BackButton';
import { Calendar } from '../components/Calendar';

interface DateSelectPageProps {
  clinicId: string;
  idToken: string;
  courseId: number;
  staffId: number;
  selectedDate: string;
  bookingWindow: number;
  onSelect: (date: string) => void;
  onNext: () => void;
  onBack: () => void;
}

export function DateSelectPage({
  clinicId,
  idToken,
  courseId,
  staffId,
  selectedDate,
  bookingWindow,
  onSelect,
  onNext,
  onBack,
}: DateSelectPageProps) {
  const [availableDates, setAvailableDates] = useState<AvailableDate[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    liffApi.getAvailableDates(clinicId, courseId, staffId, idToken)
      .then(data => {
        setAvailableDates(data);
        setError(null);
      })
      .catch(() => {
        setError('空き日程の取得に失敗しました');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [clinicId, courseId, staffId, idToken]);

  const formatSelectedDate = (date: string): string => {
    if (!date) return '';
    const [year, month, day] = date.split('-');
    const d = new Date(Number(year), Number(month) - 1, Number(day));
    const weekDays = ['日', '月', '火', '水', '木', '金', '土'];
    return `${year}年${Number(month)}月${Number(day)}日（${weekDays[d.getDay()]}）`;
  };

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={4} total={8} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-gray-800 mb-4">日付を選択</h2>
        </div>

        <div className="flex-1 px-4">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-gray-500">読み込み中...</div>
            </div>
          ) : error ? (
            <div className="py-8 text-center text-red-500">{error}</div>
          ) : (
            <Calendar
              availableDates={availableDates}
              selectedDate={selectedDate}
              onSelect={onSelect}
              bookingWindow={bookingWindow}
            />
          )}

          {selectedDate ? (
            <div className="mt-4 p-3 bg-line-green bg-opacity-10 rounded-lg text-center text-sm text-gray-800">
              選択中: {formatSelectedDate(selectedDate)}
            </div>
          ) : null}
        </div>

        <div className="px-4 py-6">
          <PrimaryButton onClick={onNext} disabled={!selectedDate}>
            次へ
          </PrimaryButton>
        </div>
      </div>
    </div>
  );
}
