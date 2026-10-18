import { Button } from "@/components/ui/button";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { LAYOUT } from "@/lib/design-tokens";
import { ResourceHospitalization } from "@/types/generated/models";
import type { HospitalizationFormGate } from "./hospitalization-form-model";

export function HospitalizationFormStatusView({
  gate,
  onBack,
  savedHospitalizationType,
}: {
  gate: HospitalizationFormGate | { kind: "new-deceased" };
  onBack: () => void;
  /** EMR-412 R4: 保存済み区分。ロード中/不在の状態では推測せず汎用見出しにする。 */
  savedHospitalizationType?: "入院" | "ホテル";
}) {
  // EMR-412 R4: 保存済みが確認できるときだけ区分名を使う。それ以外は汎用「入院・ホテル編集」。
  const editTitle = savedHospitalizationType
    ? `${savedHospitalizationType}編集`
    : "入院・ホテル編集";
  if (gate.kind === "new-pet-loading") return <LoadingFallback />;
  if (gate.kind === "new-no-pet") return null;

  if (gate.kind === "edit-loading") {
    return (
      <PageLayout
        title={editTitle}
        onBack={onBack}
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
        title={editTitle}
        onBack={onBack}
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
        title="入院・ホテル登録"
        onBack={onBack}
        resource={ResourceHospitalization}
        maxWidth={LAYOUT.pageContentMaxWidth.form}
      >
        <ErrorFallback message="死亡したペットは入院・ホテル登録できません" />
      </PageLayout>
    );
  }
  return (
    <PageLayout
      title={editTitle}
      onBack={onBack}
      resource={ResourceHospitalization}
      maxWidth={LAYOUT.pageContentMaxWidth.form}
    >
      <div className="space-y-3">
        <ErrorFallback message="入院情報の取得に失敗しました" />
        {gate.retryRead ? (
          <Button type="button" variant="outline" onClick={gate.retryRead}>
            再試行
          </Button>
        ) : null}
      </div>
    </PageLayout>
  );
}
