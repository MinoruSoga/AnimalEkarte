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
import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";

// Relative
import { H_STYLES } from "../lib/styles";
import { HOSPITALIZATION_DECEASED_BLOCK_MESSAGE, HOSPITALIZATION_STATUS } from "../constants";
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
    const petIsDeceased = hospitalization.petIsDeceased === true;
    // 臨床安全境界1: 死亡ペットは render 側でも要素を出さない（callback 側の拒否は handleCheckIn 内で維持）。
    const showCheckIn = canEdit && isReserved && !petIsDeceased;
    const showDischarge = canEdit && isAdmitted;
    const checkInBlockedByDeath = canEdit && isReserved && petIsDeceased;

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
        <div className={`flex items-center ${H_STYLES.gap.default}`}>
            {checkInBlockedByDeath ? (
                <span role="status" className={`text-sm ${C.text50}`}>
                    {HOSPITALIZATION_DECEASED_BLOCK_MESSAGE.CHECK_IN}
                </span>
            ) : null}
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
