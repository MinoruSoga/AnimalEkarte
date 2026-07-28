// React/Framework
import { useLayoutEffect, useRef } from "react";
import { useNavigate } from "react-router";

// External
import { Settings, LogOut, LogIn } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { PrimaryButton } from "@/components/shared/Form/PrimaryButton";
import { usePermission } from "@/hooks/use-permission";
import { HospitalizationStatusAdmitted } from "@/types/generated/models";

// Relative
import { H_STYLES } from "../styles";
import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { HOSPITALIZATION_STATUS } from "../constants";
import { useUpdateHospitalization } from "../api/update-hospitalization";

// Types
import type { Hospitalization } from "@/types";

interface HospitalizationDetailActionsProps {
    hospitalization: Hospitalization;
    onDischargeClick: () => void;
}

export function HospitalizationDetailActions({ hospitalization, onDischargeClick }: HospitalizationDetailActionsProps) {
    const navigate = useNavigate();
    const { canEdit } = usePermission("hospitalization");
    const canEditRef = useRef(canEdit);
    const petIsDeceasedRef = useRef(hospitalization.petIsDeceased);
    useLayoutEffect(() => {
        canEditRef.current = canEdit;
        petIsDeceasedRef.current = hospitalization.petIsDeceased;
    }, [canEdit, hospitalization.petIsDeceased]);
    const { mutateAsync: updateHospitalization, isPending: isCheckingIn } = useUpdateHospitalization();

    const isReserved = hospitalization.status === HOSPITALIZATION_STATUS.RESERVED;
    const isAdmitted = hospitalization.status === HOSPITALIZATION_STATUS.ACTIVE;
    const showCheckIn = canEdit && isReserved;
    const showDischarge = canEdit && isAdmitted;

    const handleCheckIn = async () => {
        if (canEditRef.current !== true || petIsDeceasedRef.current === true) return;
        try {
            await updateHospitalization({
                id: hospitalization.id,
                req: { status: HospitalizationStatusAdmitted },
            });
        } catch {
            // useUpdateHospitalization.onError → handleApiError 済み
        }
    };

    return (
        <div className={`flex ${H_STYLES.gap.default}`}>
            {showCheckIn ? (
                <PrimaryButton
                    colorVariant="primary"
                    className={`gap-2 ${H_STYLES.button.action}`}
                    onClick={handleCheckIn}
                    disabled={isCheckingIn}
                >
                    <LogIn className={H_STYLES.button.icon} />
                    チェックイン
                </PrimaryButton>
            ) : null}
            {showDischarge ? (
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
