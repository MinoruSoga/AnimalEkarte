// React/Framework
import { useNavigate } from "react-router-dom";

// External
import { Settings, LogOut } from "lucide-react";

// Internal
import { Button } from "../../../components/ui/button";

// Relative
import { H_STYLES } from "../styles";

// Types
import type { Hospitalization } from "../../../types";

interface HospitalizationDetailActionsProps {
    hospitalization: Hospitalization;
    onDischargeClick: () => void;
}

export const HospitalizationDetailActions = ({ hospitalization, onDischargeClick }: HospitalizationDetailActionsProps) => {
    const navigate = useNavigate();

    return (
        <div className={`flex ${H_STYLES.gap.default}`}>
            {hospitalization.status !== "退院済" && (
                <Button 
                    variant="outline" 
                    className={`gap-2 text-red-600 hover:text-red-700 hover:bg-red-50 border-red-200 ${H_STYLES.button.action}`}
                    onClick={onDischargeClick}
                >
                    <LogOut className={H_STYLES.button.icon} />
                    退院処理
                </Button>
            )}
            <Button 
                variant="outline" 
                onClick={() => navigate(`/hospitalization/${hospitalization.id}/edit`)}
                className={`gap-2 ${H_STYLES.button.action}`}
            >
                <Settings className={H_STYLES.button.icon} />
                入院情報の編集
            </Button>
        </div>
    );
};
