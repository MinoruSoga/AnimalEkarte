import { ICON, C, BADGE } from "@/lib/design-tokens";
import { memo } from "react";
// External
import { Edit2, Utensils, Pill, ClipboardList, Stethoscope, CheckCircle2 } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { DeleteIconButton } from "@/components/shared/DeleteIconButton/DeleteIconButton";

// Relative
import { H_STYLES } from "@/features/hospitalization/styles";

// Types
import type { CarePlanItem } from "@/features/hospitalization/types";

interface CarePlanItemProps {
    plan: CarePlanItem;
    onEdit: (plan: CarePlanItem) => void;
    onDelete?: (id: string) => void;
}

export const CarePlanItemRow = memo(function CarePlanItemRow({ plan, onEdit, onDelete }: CarePlanItemProps) {
    const getTypeIcon = (type: string) => {
        switch (type) {
            case "food": return <Utensils className={H_STYLES.button.icon} />;
            case "medicine": return <Pill className={H_STYLES.button.icon} />;
            case "instruction": return <ClipboardList className={H_STYLES.button.icon} />;
            case "treatment": return <Stethoscope className={H_STYLES.button.icon} />;
            default: return <CheckCircle2 className={H_STYLES.button.icon} />;
        }
    };

    const getTypeLabel = (type: string) => {
        switch (type) {
            case "food": return "食事";
            case "medicine": return "投薬";
            case "instruction": return "処置・指示";
            case "item": return "持ち物・その他";
            case "treatment": return "処置・検査";
            default: return type;
        }
    };

    const getTypeColor = (type: string) => {
        switch (type) {
            case 'food': return BADGE.orangeNoBorder;
            case 'medicine': return BADGE.blueNoBorder;
            case 'treatment': return BADGE.purpleNoBorder;
            default: return BADGE.grayNoBorder;
        }
    };

    return (
        <div className={`bg-white border ${C.borderMedium} rounded-md ${H_STYLES.padding.card} flex items-center justify-between shadow-sm`}>
            <div className="flex items-center gap-2 flex-1 min-w-0">
                <div className={`p-2 shrink-0 rounded ${getTypeColor(plan.type)}`}>
                    {getTypeIcon(plan.type)}
                </div>
                
                <div className={`flex flex-wrap items-center gap-x-2 gap-y-1 ${H_STYLES.text.base} ${C.text} leading-snug`}>
                    <span className="font-bold whitespace-nowrap">{plan.name}</span>
                    <span className={`${H_STYLES.text.sm} ${C.text60} px-2 border-l border-r ${C.borderMedium}`}>{getTypeLabel(plan.type)}</span>
                    <span className={`${H_STYLES.text.sm} ${C.bgPage} px-2 py-0.5 rounded`}>{plan.description}</span>
                    
                    {plan.unitPrice ? (
                        <span className={`${H_STYLES.text.sm} ${C.text60} font-mono`}>¥{plan.unitPrice.toLocaleString()}</span>
                    ) : null}
                    
                    <div className="flex gap-1 ml-1">
                        {plan.timing.map(t => (
                            <span key={t} className={`${H_STYLES.text.sm} ${C.bgStatusGray} px-2 py-0.5 rounded ${C.textStatusGray}`}>
                                {t === 'morning' ? '朝' : t === 'noon' ? '昼' : t === 'night' ? '夜' : t}
                            </span>
                        ))}
                    </div>
                </div>
            </div>

            <div className="flex items-center gap-2 ml-2 shrink-0">
                <div className={`w-2 h-2 rounded-full ${plan.status === 'active' ? C.bgStatusGreenDot : C.bgInactive}`} />
                <div className="flex gap-1">
                    <Button variant="ghost" size="sm" onClick={() => onEdit(plan)} className={`h-9 w-9 p-0 ${C.bgPage} ${C.hoverBgPage}`}>
                        <Edit2 className={`${ICON.action} ${C.text60} ${C.hoverText}`} />
                    </Button>
                    {onDelete !== undefined ? (
                        <DeleteIconButton onClick={() => onDelete(plan.id)} />
                    ) : null}
                </div>
            </div>
        </div>
    );
});
