import { useCallback } from "react";
import { liffApi } from "../api/liff-api";
import { ProgressDots } from "../components/ProgressDots";
import { PrimaryButton } from "../components/PrimaryButton";
import { BackButton } from "../components/BackButton";
import { Calendar } from "../components/Calendar";
import { formatJapaneseDate } from "@/shared-liff/jst-date";
import { useFetchState } from "@/shared-liff/use-fetch-state";
import { getStepProgress } from "../lib/step-progress";
import { EXPLICIT_PRIMARY_CTA_LABEL } from "../lib/advance-policy";

interface DateSelectPageProps {
  clinicId: string;
  idToken: string;
  courseId: number;
  staffId: number;
  selectedDate: string;
  bookingWindow: number;
  isTrimming: boolean;
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
  isTrimming,
  onSelect,
  onNext,
  onBack,
}: DateSelectPageProps) {
  const fetcher = useCallback(
    () => liffApi.getAvailableDates(clinicId, courseId, staffId, idToken),
    [clinicId, courseId, staffId, idToken],
  );
  // R-F22/R-F23: ステータス別メッセージ解決と再試行導線を共通フックに統合。
  const { data: availableDates, loading, error, retry } = useFetchState(fetcher, "空き日程の取得");
  // SD-16: トリミング分岐で挿入される2ステップ分、以降のフロー全体の total を一貫させる
  const { current, total } = getStepProgress("dateSelect", isTrimming);

  const formatSelectedDate = (date: string): string => {
    return formatJapaneseDate(date);
  };

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={current} total={total} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-4">日付を選択</h2>
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
          ) : (
            <Calendar
              availableDates={availableDates ?? []}
              selectedDate={selectedDate}
              onSelect={onSelect}
              bookingWindow={bookingWindow}
            />
          )}

          {selectedDate ? (
            <div className="mt-4 p-3 bg-noah-teal bg-opacity-10 rounded-xl text-center text-sm text-noah-text">
              選択中: {formatSelectedDate(selectedDate)}
            </div>
          ) : null}
        </div>

        <div className="px-4 py-6">
          <PrimaryButton onClick={onNext} disabled={!selectedDate}>
            {EXPLICIT_PRIMARY_CTA_LABEL}
          </PrimaryButton>
        </div>
      </div>
    </div>
  );
}
