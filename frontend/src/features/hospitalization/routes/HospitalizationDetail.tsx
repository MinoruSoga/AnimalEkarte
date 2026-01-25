// React/Framework
import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";

// Internal
import { PageLayout } from "../../../components/shared/PageLayout";

// Relative
import {
    DischargeAlertDialog,
    HospitalizationDetailActions,
    HospitalizationDesktopLayout,
    HospitalizationMobileLayout
} from "../components";
import { useHospitalizationDetail } from "../hooks/useHospitalizationDetail";

export const HospitalizationDetail = () => {
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

    if (isLoading || !hospitalization) {
        return <div className="p-8 text-center text-gray-500">読み込み中...</div>;
    }

    const handleDischargeConfirm = async () => {
        const success = await dischargeHospitalization();
        if (success) {
            setShowDischargeDialog(false);
            navigate("/hospitalization");
        }
    };

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
            onBack={() => navigate("/hospitalization")}
            headerAction={
                <HospitalizationDetailActions 
                    hospitalization={hospitalization} 
                    onDischargeClick={() => setShowDischargeDialog(true)} 
                />
            }
            maxWidth="max-w-[1600px]"
        >
            <div>
                <HospitalizationDesktopLayout {...commonProps} />
                <HospitalizationMobileLayout {...commonProps} />
            </div>

            <DischargeAlertDialog 
                open={showDischargeDialog} 
                onOpenChange={setShowDischargeDialog}
                onConfirm={handleDischargeConfirm}
            />
        </PageLayout>
    );
};
