import { useCallback } from 'react';
import { liffApi } from '../api/liff-api';
import { ProgressDots } from '../components/ProgressDots';
import { BackButton } from '../components/BackButton';
import { AutoAdvanceHint } from '../components/AutoAdvanceHint';
import { formatTimeHHMM } from '@/shared-liff/jst-date';
import { useFetchState } from '@/shared-liff/use-fetch-state';
import { getStepProgress } from '../lib/step-progress';

interface TimeSelectPageProps {
  clinicId: string;
  idToken: string;
  courseId: number;
  staffId: number;
  date: string;
  isTrimming: boolean;
  onSelect: (startTime: string, endTime: string) => void;
  onBack: () => void;
}

export function TimeSelectPage({
  clinicId,
  idToken,
  courseId,
  staffId,
  date,
  isTrimming,
  onSelect,
  onBack,
}: TimeSelectPageProps) {
  const fetcher = useCallback(
    () => liffApi.getAvailableTimes(clinicId, courseId, staffId, date, idToken),
    [clinicId, courseId, staffId, date, idToken],
  );
  // R-F22/R-F23: ステータス別メッセージ解決と再試行導線を共通フックに統合。
  const { data: times, loading, error, retry } = useFetchState(fetcher, '空き時間の取得');
  // SD-16: トリミング分岐で挿入される2ステップ分、以降のフロー全体の total を一貫させる
  const { current, total } = getStepProgress('timeSelect', isTrimming);

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={current} total={total} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-2">時間を選択</h2>
          <AutoAdvanceHint step="timeSelect" />
        </div>

        <div className="flex-1">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-noah-text-sub">読み込み中...</div>
            </div>
          ) : error ? (
            <div className="px-4 py-8 text-center text-red-500">
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
          ) : !times || times.length === 0 ? (
            <div className="px-4 py-8 text-center text-noah-text-sub">
              この日の空き時間はありません
            </div>
          ) : (
            <div className="bg-white border-t border-gray-200">
              {times.map(time => (
                <button
                  key={time.start_time}
                  type="button"
                  onClick={() => onSelect(time.start_time, time.end_time)}
                  className="w-full flex items-center justify-between px-4 py-4 border-b border-gray-200 hover:bg-noah-teal-light active:bg-noah-teal-light text-left"
                >
                  <span className="text-noah-text font-medium">
                    {time.display_time || formatTimeHHMM(time.start_time)}
                    {' 〜 '}
                    {formatTimeHHMM(time.end_time)}
                  </span>
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 text-gray-400"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                    aria-hidden="true"
                  >
                    <path
                      fillRule="evenodd"
                      d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
                      clipRule="evenodd"
                    />
                  </svg>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
