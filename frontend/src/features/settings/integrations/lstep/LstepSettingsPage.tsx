import { usePermission } from "@/hooks/use-permission";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { ResourceHospitalSettings } from "@/types/generated/models";
import { C } from "@/lib/design-tokens";
import { LstepSettingsForm } from "./LstepSettingsForm";
import { TriggerPrioritySection } from "./TriggerPrioritySection";
import { LstepTagCodeMappingsSection } from "./LstepTagCodeMappingsSection";

export function LstepSettingsPage() {
  const { canEdit } = usePermission("hospital-settings");

  if (!canEdit) {
    return (
      <PageLayout title="Lステップ連携設定" resource={ResourceHospitalSettings}>
        <div className={`flex items-center justify-center py-16 text-sm ${C.text50}`}>
          この画面を表示する権限がありません。
        </div>
      </PageLayout>
    );
  }

  return (
    <PageLayout title="Lステップ連携設定" resource={ResourceHospitalSettings}>
      <LstepSettingsForm />
      <TriggerPrioritySection />
      <LstepTagCodeMappingsSection />
    </PageLayout>
  );
}
