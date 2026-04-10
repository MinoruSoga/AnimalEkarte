import { useState, useEffect } from 'react';
import type { Staff } from '../types/models';
import { liffApi } from '../api/liff-api';
import { ProgressDots } from '../components/ProgressDots';
import { ListItem } from '../components/ListItem';
import { BackButton } from '../components/BackButton';

interface StaffSelectPageProps {
  clinicId: string;
  idToken: string;
  courseId: number;
  showNoStaffOption: boolean;
  onSelect: (staffId: number, staffName: string) => void;
  onBack: () => void;
}

export function StaffSelectPage({
  clinicId,
  idToken,
  courseId,
  showNoStaffOption,
  onSelect,
  onBack,
}: StaffSelectPageProps) {
  const [staffs, setStaffs] = useState<Staff[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    liffApi.getStaffs(clinicId, courseId, idToken)
      .then(data => {
        setStaffs(data);
        setError(null);
      })
      .catch(() => {
        setError('スタッフ一覧の取得に失敗しました');
      })
      .finally(() => {
        setLoading(false);
      });
  }, [clinicId, courseId, idToken]);

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={3} total={8} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-4">スタッフを選択</h2>
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
              {showNoStaffOption ? (
                <ListItem
                  onClick={() => onSelect(0, '指名なし')}
                  subtitle="担当スタッフはお任せします"
                >
                  指名なし
                </ListItem>
              ) : null}
              {staffs.map(staff => (
                <ListItem
                  key={staff.id}
                  onClick={() => onSelect(staff.id, staff.name)}
                  description={staff.reservation_comment || undefined}
                  imageUrl={staff.reservation_image_url || undefined}
                >
                  {staff.name}
                </ListItem>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
