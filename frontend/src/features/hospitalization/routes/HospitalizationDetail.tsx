// React/Framework
import { useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router";

// Internal
import { PageLayout } from "@/components/shared/PageLayout/PageLayout";

// Relative
import { DischargeAlertDialog } from "../components/DischargeAlertDialog";
import { HospitalizationDetailActions } from "../components/HospitalizationDetailActions";
import { HospitalizationExpandedView } from "../components/HospitalizationExpandedView";
import { HospitalizationTabbedView } from "../components/HospitalizationTabbedView";
import { useHospitalizationDetail } from "../hooks/use-hospitalization-detail";
import { paths } from "@/config/paths";

export function HospitalizationDetail() {
    const { id } = useParams();
    const navigate = useNavigate();

    const {
        hospitalization,
        plans,
        records,
        isLoading,
        handleAddPlan,
        handleUpdatePlan,
        handleDeletePlan,
        handleAddVital,
        handleAddLog,
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
        return <div className="p-8 text-center text-gray-500">読み込み中...</div>;
    }

    const commonProps = {
        hospitalization,
        plans,
        records,
        onAddPlan: handleAddPlan,
        onUpdatePlan: handleUpdatePlan,
        onDeletePlan: handleDeletePlan,
        onAddVital: handleAddVital,
        onAddLog: handleAddLog
    };

    return (
        <PageLayout
            title="入院詳細・カルテ"
            onBack={() => navigate(paths.hospitalization.getHref())}
            headerAction={
                <HospitalizationDetailActions 
                    hospitalization={hospitalization} 
                    onDischargeClick={() => setShowDischargeDialog(true)} 
                />
            }
            maxWidth="max-w-[1600px]"
        >
            <div>
                <HospitalizationExpandedView {...commonProps} />
                <HospitalizationTabbedView {...commonProps} />
            </div>

            <DischargeAlertDialog 
                open={showDischargeDialog} 
                onOpenChange={setShowDischargeDialog}
                onConfirm={handleDischargeConfirm}
            />
        </PageLayout>
    );
}
