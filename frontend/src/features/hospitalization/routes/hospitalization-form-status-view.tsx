import { FileText } from "lucide-react";

import { Button } from "@/components/ui/button";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { ResourceHospitalization } from "@/types/generated/models";
import type { HospitalizationFormGate } from "./hospitalization-form-model";

export function HospitalizationFormStatusView({
  gate,
  onBack,
}: {
  gate: HospitalizationFormGate | { kind: "new-deceased" };
  onBack: () => void;
}) {
  if (gate.kind === "new-pet-loading") return <LoadingFallback />;
  if (gate.kind === "new-no-pet") return null;

  if (gate.kind === "edit-loading") {
    return (
      <PageLayout
        title="入院"
        onBack={onBack}
        icon={<FileText className={`${ICON.page} ${C.text}`} />}
        resource={ResourceHospitalization}
        maxWidth={LAYOUT.pageContentMaxWidth.form}
      >
        <LoadingFallback />
      </PageLayout>
    );
  }
  if (gate.kind === "edit-not-found") {
    return (
      <PageLayout
        title="入院"
        onBack={onBack}
        icon={<FileText className={`${ICON.page} ${C.text}`} />}
        resource={ResourceHospitalization}
        maxWidth={LAYOUT.pageContentMaxWidth.form}
      >
        <ErrorFallback message="入院情報が見つかりません" />
      </PageLayout>
    );
  }
  if (gate.kind === "new-deceased") {
    return (
      <PageLayout
        title="入院登録"
        onBack={onBack}
        icon={<FileText className={`${ICON.page} ${C.text}`} />}
        resource={ResourceHospitalization}
        maxWidth={LAYOUT.pageContentMaxWidth.form}
      >
        <ErrorFallback message="死亡したペットは入院登録できません" />
      </PageLayout>
    );
  }
  return (
    <PageLayout
      title="入院"
      onBack={onBack}
      icon={<FileText className={`${ICON.page} ${C.text}`} />}
      resource={ResourceHospitalization}
      maxWidth={LAYOUT.pageContentMaxWidth.form}
    >
      <div className="space-y-3">
        <ErrorFallback message="入院情報の取得に失敗しました" />
        {gate.retryRead ? (
          <Button type="button" variant="outline" size="sm" onClick={gate.retryRead}>
            再試行
          </Button>
        ) : null}
      </div>
    </PageLayout>
  );
}
