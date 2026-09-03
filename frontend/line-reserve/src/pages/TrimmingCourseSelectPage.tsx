import { useCallback } from 'react';
import { liffApi } from '../api/liff-api';
import { ProgressDots } from '../components/ProgressDots';
import { ListItem } from '../components/ListItem';
import { BackButton } from '../components/BackButton';
import { AutoAdvanceHint } from '../components/AutoAdvanceHint';
import { useFetchState } from '@/shared-liff/use-fetch-state';
import { formatCurrency } from '@/lib/format/number';
import { getStepProgress } from '../lib/step-progress';

interface TrimmingCourseSelectPageProps {
  clinicId: string;
  idToken: string;
  onSelect: (courseId: number, courseName: string) => void;
  onBack: () => void;
}

function formatPrice(price: number | null): string {
  if (price === null) return '';
  return formatCurrency(price);
}

export function TrimmingCourseSelectPage({ clinicId, idToken, onSelect, onBack }: TrimmingCourseSelectPageProps) {
  const fetcher = useCallback(() => liffApi.getTrimmingCourses(clinicId, idToken), [clinicId, idToken]);
  // R-F22/R-F23: ステータス別メッセージ解決と再試行導線を共通フックに統合。
  const { data: courses, loading, error, retry } = useFetchState(fetcher, 'トリミングコースの取得');
  // SD-16: トリミングフロー内で一貫した total を使う（他の共有ページと同じ算出元）
  const { current, total } = getStepProgress('trimmingCourseSelect', true);

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={current} total={total} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-2">トリミングコースを選択</h2>
          <AutoAdvanceHint step="trimmingCourseSelect" />
        </div>

        <div className="flex-1">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-noah-text-sub">読み込み中...</div>
            </div>
          ) : error ? (
            <div className="px-4 py-8 text-center text-noah-danger">
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
            <div className="bg-white border-t border-gray-200">
              {(courses ?? []).map(course => (
                <ListItem
                  key={course.id}
                  onClick={() => onSelect(course.id, course.name)}
                  subtitle={formatPrice(course.price)}
                  description={course.description || undefined}
                >
                  {course.name}
                </ListItem>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
