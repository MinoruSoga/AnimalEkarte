import { useCallback } from 'react';
import { liffApi } from '../api/liff-api';
import { ProgressDots } from '../components/ProgressDots';
import { ListItem } from '../components/ListItem';
import { BackButton } from '../components/BackButton';
import { AutoAdvanceHint } from '../components/AutoAdvanceHint';
import { useFetchState } from '@/shared-liff/use-fetch-state';
import { getStepProgress } from '../lib/step-progress';

interface StaffSelectPageProps {
  clinicId: string;
  idToken: string;
  courseId: number;
  showNoStaffOption: boolean;
  isTrimming: boolean;
  onSelect: (staffId: number, staffName: string) => void;
  onBack: () => void;
}

export function StaffSelectPage({
  clinicId,
  idToken,
  courseId,
  showNoStaffOption,
  isTrimming,
  onSelect,
  onBack,
}: StaffSelectPageProps) {
  const fetcher = useCallback(
    () => liffApi.getStaffs(clinicId, courseId, idToken),
    [clinicId, courseId, idToken],
  );
  // R-F22/R-F23: ステータス別メッセージ解決と再試行導線を共通フックに統合。
  const { data: staffs, loading, error, retry } = useFetchState(fetcher, 'スタッフ一覧の取得');
  // SD-16: トリミング分岐で挿入される2ステップ分、以降のフロー全体の total を一貫させる
  const { current, total } = getStepProgress('staffSelect', isTrimming);

  return (
    <div className="min-h-screen bg-noah-teal-light flex flex-col">
      <div className="max-w-md mx-auto w-full flex flex-col flex-1">
        <ProgressDots current={current} total={total} />

        <div className="px-4">
          <BackButton onClick={onBack} />
          <h2 className="text-lg font-bold text-noah-teal-dark mb-2">スタッフを選択</h2>
          <AutoAdvanceHint step="staffSelect" />
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
              {showNoStaffOption ? (
                <ListItem
                  onClick={() => onSelect(0, '指名なし')}
                  subtitle="担当スタッフはお任せします"
                >
                  指名なし
                </ListItem>
              ) : null}
              {(staffs ?? []).map(staff => (
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
