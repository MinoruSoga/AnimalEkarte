import { useState, useEffect } from 'react';
import type { TrimmingCourse } from '../types/models';
import { liffApi } from '../api/liff-api';
import { ProgressDots } from '../components/ProgressDots';
import { ListItem } from '../components/ListItem';
import { BackButton } from '../components/BackButton';

interface TrimmingCourseSelectPageProps {
  clinicId: string;
  idToken: string;
  onSelect: (courseId: number, courseName: string) => void;
  onBack: () => void;
}

function formatPrice(price: number | null): string {
  if (price === null) return '';
  return `¥${price.toLocaleString()}`;
}

export function TrimmingCourseSelectPage({ clinicId, idToken, onSelect, onBack }: TrimmingCourseSelectPageProps) {
  const [courses, setCourses] = useState<TrimmingCourse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    liffApi.getTrimmingCourses(clinicId, idToken)
      .then(data => {
        setCourses(data);
        setError(null);
      })
      .catch(() => {
        setError('トリミングコースの取得に失敗しました');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [clinicId, idToken]);

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={3} total={9} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-4">トリミングコースを選択</h2>
        </div>

        <div className="flex-1">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-noah-text-sub">読み込み中...</div>
            </div>
          ) : error ? (
            <div className="px-4 py-8 text-center text-red-500">{error}</div>
          ) : (
            <div className="bg-white border-t border-gray-200">
              {courses.map(course => (
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
