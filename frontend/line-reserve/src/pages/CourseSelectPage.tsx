import { useCallback } from 'react';
import { liffApi } from '../api/liff-api';
import { ProgressDots } from '../components/ProgressDots';
import { ListItem } from '../components/ListItem';
import { BackButton } from '../components/BackButton';
import { AutoAdvanceHint } from '../components/AutoAdvanceHint';
import { useFetchState } from '@/shared-liff/use-fetch-state';

interface CourseSelectPageProps {
  clinicId: string;
  idToken: string;
  onSelect: (courseId: number, courseName: string, category?: 'general' | 'trimming') => void;
  onBack: () => void;
}

export function CourseSelectPage({ clinicId, idToken, onSelect, onBack }: CourseSelectPageProps) {
  const fetcher = useCallback(() => liffApi.getCourses(clinicId, idToken), [clinicId, idToken]);
  // R-F22/R-F23: ステータス別メッセージ解決と再試行導線を共通フックに統合。
  const { data: courses, loading, error, retry } = useFetchState(fetcher, 'コースの取得');

  const formatDuration = (minutes: number | undefined): string => {
    if (!minutes) return '';
    if (minutes < 60) return `${minutes}分`;
    const h = Math.floor(minutes / 60);
    const m = minutes % 60;
    return m > 0 ? `${h}時間${m}分` : `${h}時間`;
  };

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={2} total={8} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-2">コースを選択</h2>
          <AutoAdvanceHint step="courseSelect" />
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
          ) : (
            <div className="bg-white border-t border-gray-200">
              {(courses ?? []).map(course => (
                <ListItem
                  key={course.id}
                  onClick={() => onSelect(course.id, course.name, course.category)}
                  subtitle={formatDuration(course.duration_minutes)}
                  description={course.reservation_comment || undefined}
                  imageUrl={course.reservation_image_url || undefined}
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
