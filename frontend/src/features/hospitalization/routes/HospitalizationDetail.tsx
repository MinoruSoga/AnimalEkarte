// React/Framework
import { useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router";

// Internal
import { LAYOUT } from "@/lib/design-tokens";
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { LoadingFallback, ErrorFallback } from "@/components/shared/DataStates";
import { paths } from "@/config/paths";
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

  if (isLoading) return <LoadingFallback />;
  if (isNotFound || !id) return <ErrorFallback message="入院情報が見つかりません" />;
  if (isError) return <ErrorFallback />;
  if (!hospitalization) return <ErrorFallback message="入院情報が見つかりません" />;

  return (
    <PageLayout
      title="入院詳細・カルテ"
      onBack={() => navigate(paths.hospitalization.getHref())}
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
