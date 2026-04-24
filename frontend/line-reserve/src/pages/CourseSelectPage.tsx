import { useState, useEffect } from 'react';
import type { Course } from '../types/models';
import { liffApi } from '../api/liff-api';
import { ProgressDots } from '../components/ProgressDots';
import { ListItem } from '../components/ListItem';
import { BackButton } from '../components/BackButton';

interface CourseSelectPageProps {
  clinicId: string;
  idToken: string;
  onSelect: (courseId: number, courseName: string, category?: 'general' | 'trimming') => void;
  onBack: () => void;
}

export function CourseSelectPage({ clinicId, idToken, onSelect, onBack }: CourseSelectPageProps) {
  const [courses, setCourses] = useState<Course[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    liffApi.getCourses(clinicId, idToken)
      .then(data => {
        setCourses(data);
        setError(null);
      })
      .catch(() => {
        setError('コースの取得に失敗しました');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [clinicId, idToken]);

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
          <h2 className="text-lg font-bold text-noah-teal-dark mb-4">コースを選択</h2>
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
