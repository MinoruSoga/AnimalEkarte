import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { useAuth } from "@/hooks/use-auth";
import { C } from "@/lib/design-tokens";
import { EmptyState } from "@/components/shared/DataStates";
import { useGetLineReservationSetting } from "../api/get-line-reservation-setting";
import { LineReservationSettingsForm } from "../components/LineReservationSettingsForm";

export function LineReservationSettings() {
  const { currentClinicId } = useAuth();
  const { data: setting, isLoading } = useGetLineReservationSetting(currentClinicId);

  return (
    <PageLayout title="基本設定">
      {isLoading ? (
        <div className={`text-sm ${C.textMuted} py-8 text-center`}>読み込み中...</div>
      ) : setting ? (
        <LineReservationSettingsForm setting={setting} clinicId={currentClinicId!} />
      ) : (
        <EmptyState message="設定データが見つかりません。" />
      )}
    </PageLayout>
  );
}
