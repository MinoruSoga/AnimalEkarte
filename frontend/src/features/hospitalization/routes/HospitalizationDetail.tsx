// React/Framework
import { useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";
import { C } from "@/lib/design-tokens";

// Relative
import { DischargeAlertDialog } from "../components/DischargeAlertDialog";
import { HospitalizationDetailActions } from "../components/HospitalizationDetailActions";
import { HospitalizationExpandedView } from "../components/HospitalizationExpandedView";
import { HospitalizationTabbedView } from "../components/HospitalizationTabbedView";
import { useHospitalizationDetail } from "../hooks/use-hospitalization-detail";
import { paths } from "@/config/paths";
import { ResourceHospitalization } from "@/types/generated/models";

export function HospitalizationDetail() {
    const { id } = useParams();
    const navigate = useNavigate();

    const {
        hospitalization,
        isLoading,
        dischargeHospitalization
    } = useHospitalizationDetail(id);

    const [showDischargeDialog, setShowDischargeDialog] = useState(false);

    const handleDischargeConfirm = useCallback(async (navigateToAccounting: boolean) => {
        const result = await dischargeHospitalization(navigateToAccounting);
        if (result.success) {
            setShowDischargeDialog(false);
            if (navigateToAccounting && result.accountingId) {
                navigate(paths.accounting.detail.getHref(String(result.accountingId)));
            } else if (navigateToAccounting && hospitalization?.petId) {
                navigate(`${paths.accounting.new.getHref()}?petId=${hospitalization.petId}`);
            } else {
                navigate(paths.hospitalization.getHref());
            }
        }
    }, [dischargeHospitalization, navigate, hospitalization]);

    if (isLoading || !hospitalization) {
        return <div className={`p-8 text-center ${C.textStatusGray}`}>読み込み中...</div>;
    }

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
            maxWidth="max-w-[1600px]"
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
