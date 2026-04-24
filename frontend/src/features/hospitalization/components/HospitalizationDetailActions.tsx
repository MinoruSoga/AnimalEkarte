// React/Framework
import { useNavigate } from "react-router";

// External
import { Settings, LogOut } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { usePermission } from "@/hooks/use-permission";

// Relative
import { H_STYLES } from "../styles";
import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";

// Types
import type { Hospitalization } from "@/types";

interface HospitalizationDetailActionsProps {
    hospitalization: Hospitalization;
    onDischargeClick: () => void;
}

export function HospitalizationDetailActions({ hospitalization, onDischargeClick }: HospitalizationDetailActionsProps) {
    const navigate = useNavigate();
    const { canEdit, canDelete } = usePermission("hospitalization");

    return (
        <div className={`flex ${H_STYLES.gap.default}`}>
            {canDelete && hospitalization.status !== "退院済" ? (
                <Button
                    variant="ghost-danger"
                    className={`gap-2 border ${C.borderDanger20} ${H_STYLES.button.action}`}
                    onClick={onDischargeClick}
                >
                    <LogOut className={H_STYLES.button.icon} />
                    退院処理
                </Button>
            ) : null}
            {canEdit ? (
                <Button
                    variant="outline"
                    onClick={() => navigate(paths.hospitalization.edit.getHref(hospitalization.id))}
                    className={`gap-2 ${H_STYLES.button.action}`}
                >
                    <Settings className={H_STYLES.button.icon} />
                    入院情報の編集
                </Button>
            ) : null}
        </div>
    );
}
