import { HeartPulse } from "lucide-react";

import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { Button } from "@/components/ui/button";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { ResourceMedicalRecords } from "@/types/generated/models";
import type { MedicalRecordFormGate } from "./medical-record-form-model";

export function MedicalRecordFormStatusView({
  gate,
  onBack,
}: {
  gate: MedicalRecordFormGate | { kind: "deceased-new" };
  onBack: () => void;
}) {
  if (gate.kind === "pet-loading") {
    return <LoadingFallback />;
  }
  if (gate.kind === "empty") {
    return null;
  }

  const title = gate.kind === "deceased-new" ? "カルテ入力" : "カルテ";
  const message =
    gate.kind === "not-found" || gate.kind === "missing-pet"
      ? "カルテが見つかりません"
      : gate.kind === "read-error"
        ? "カルテの取得に失敗しました"
        : gate.kind === "deceased-new"
          ? "死亡したペットは新規カルテを作成できません"
          : undefined;

  return (
    <PageLayout
      title={title}
      onBack={onBack}
      icon={<HeartPulse className={`${ICON.page} ${C.text}`} />}
      resource={gate.kind === "deceased-new" ? ResourceMedicalRecords : undefined}
      maxWidth={LAYOUT.pageContentMaxWidth.full}
    >
      {message ? (
        gate.kind === "read-error" ? (
          <div className="space-y-3">
            <ErrorFallback message={message} />
            {gate.retryRead ? (
              <Button type="button" variant="outline" size="sm" onClick={gate.retryRead}>
                再試行
              </Button>
            ) : null}
          </div>
        ) : (
          <ErrorFallback message={message} />
        )
      ) : (
        <LoadingFallback />
      )}
    </PageLayout>
  );
}
