// React/Framework
import { useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router";

// Internal
import { LAYOUT } from "@/lib/design-tokens";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { paths } from "@/config/paths";
import { useBackNavigation } from "@/hooks/use-back-navigation";
import { ResourceHospitalization } from "@/types/generated/models";

// Relative
import { DischargeAlertDialog } from "../components/DischargeAlertDialog";
import { HospitalizationDetailActions } from "../components/HospitalizationDetailActions";
import { HospitalizationExpandedView } from "../components/HospitalizationExpandedView";
import { HospitalizationTabbedView } from "../components/HospitalizationTabbedView";
import { useHospitalizationDetail } from "../hooks/use-hospitalization-detail";

export function HospitalizationDetail() {
  const { id } = useParams();
  const navigate = useNavigate();

  const { hospitalization, isLoading, isError, isNotFound, dischargeHospitalization } =
    useHospitalizationDetail(id);

  const [showDischargeDialog, setShowDischargeDialog] = useState(false);
  const handleBack = useBackNavigation(paths.hospitalization.getHref());

  // rerender-dependencies: hospitalization オブジェクトを deps に入れず、petId の primitive を抽出
  const hospitalizationPetId = hospitalization?.petId ? String(hospitalization.petId) : undefined;

  const handleDischargeConfirm = useCallback(
    async (navigateToAccounting: boolean) => {
      const result = await dischargeHospitalization(navigateToAccounting);
      if (result.success) {
        setShowDischargeDialog(false);
        if (navigateToAccounting && result.accountingId) {
          navigate(paths.accounting.detail.getHref(String(result.accountingId)));
        } else if (navigateToAccounting && hospitalizationPetId) {
          navigate(`${paths.accounting.new.getHref()}?petId=${hospitalizationPetId}`);
        } else {
          navigate(paths.hospitalization.getHref());
        }
      }
    },
    [dischargeHospitalization, navigate, hospitalizationPetId],
  );

  // EMR-412 N6: 読込中・見つからない時・取得失敗も PageLayout（タイトル＋戻る）で包む。
  // カルテ・トリミングの FormStatusPanels と同じ体裁に揃える。
  // 見つからない時は取得失敗より優先（旧実装と同じ優先順位を維持）。
  const isDetailNotFound = isNotFound || !id || (!isLoading && !isError && !hospitalization);
  if (isLoading || isError || isDetailNotFound || !hospitalization) {
    return (
      <PageLayout
        title="入院詳細・カルテ"
        onBack={handleBack}
        resource={ResourceHospitalization}
        maxWidth={LAYOUT.pageContentMaxWidth.detailWide}
      >
        {isLoading ? (
          <LoadingFallback />
        ) : (
          <ErrorFallback message={isDetailNotFound ? "入院情報が見つかりません" : undefined} />
        )}
      </PageLayout>
    );
  }

  // EMR-412 N2: 保存済みの区分名を詳細タイトルに使う（ホテル記録でも「入院詳細」にならない）。
  const detailTitle = `${hospitalization.hospitalizationType}詳細・カルテ`;

  return (
    <PageLayout
      title={detailTitle}
      onBack={handleBack}
      resource={ResourceHospitalization}
      headerAction={
        <HospitalizationDetailActions
          hospitalization={hospitalization}
          onDischargeClick={() => setShowDischargeDialog(true)}
        />
      }
      maxWidth={LAYOUT.pageContentMaxWidth.detailWide}
    >
      <div>
        <HospitalizationExpandedView hospitalization={hospitalization} />
        <HospitalizationTabbedView hospitalization={hospitalization} />
      </div>

      <DischargeAlertDialog
        open={showDischargeDialog}
        onOpenChange={setShowDischargeDialog}
        onConfirm={handleDischargeConfirm}
      />
    </PageLayout>
  );
}
