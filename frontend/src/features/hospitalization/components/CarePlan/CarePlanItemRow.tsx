import { Edit2, Trash2, Utensils, Pill, ClipboardList, Stethoscope, CheckCircle2 } from "lucide-react";
import { Button } from "../../../../components/ui/button";
import { CarePlanItem } from "../../types";
import { H_STYLES } from "../../styles";

interface CarePlanItemProps {
    plan: CarePlanItem;
    onEdit: (plan: CarePlanItem) => void;
    onDelete: (id: string) => void;
}

export const CarePlanItemRow = ({ plan, onEdit, onDelete }: CarePlanItemProps) => {
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
            case 'food': return 'bg-orange-100 text-orange-600';
            case 'medicine': return 'bg-blue-100 text-blue-600';
            case 'treatment': return 'bg-purple-100 text-purple-600';
            default: return 'bg-gray-100 text-gray-600';
        }
    };

    return (
        <div className={`bg-white border border-[rgba(55,53,47,0.16)] rounded-md ${H_STYLES.padding.card} flex items-center justify-between shadow-sm`}>
            <div className="flex items-center gap-2 flex-1 min-w-0">
                <div className={`p-2 shrink-0 rounded ${getTypeColor(plan.type)}`}>
                    {getTypeIcon(plan.type)}
                </div>
                
                <div className={`flex flex-wrap items-center gap-x-2 gap-y-1 ${H_STYLES.text.base} text-[#37352F] leading-snug`}>
                    <span className="font-bold whitespace-nowrap">{plan.name}</span>
                    <span className={`${H_STYLES.text.sm} text-[#37352F]/60 px-2 border-l border-r border-gray-200`}>{getTypeLabel(plan.type)}</span>
                    <span className={`${H_STYLES.text.sm} bg-gray-50 px-2 py-0.5 rounded`}>{plan.description}</span>
                    
                    {plan.unitPrice && (
                        <span className={`${H_STYLES.text.sm} text-[#37352F]/60 font-mono`}>¥{plan.unitPrice.toLocaleString()}</span>
                    )}
                    
                    <div className="flex gap-1 ml-1">
                        {plan.timing.map(t => (
                            <span key={t} className={`${H_STYLES.text.sm} bg-gray-100 px-2 py-0.5 rounded text-gray-600`}>
                                {t === 'morning' ? '朝' : t === 'noon' ? '昼' : t === 'night' ? '夜' : t}
                            </span>
                        ))}
                    </div>
                </div>
            </div>

            <div className="flex items-center gap-2 ml-2 shrink-0">
                <div className={`w-2 h-2 rounded-full ${plan.status === 'active' ? 'bg-green-500' : 'bg-gray-300'}`} />
                <div className="flex gap-1">
                    <Button variant="ghost" size="sm" onClick={() => onEdit(plan)} className="h-9 w-9 p-0 bg-gray-50 hover:bg-gray-100">
                        <Edit2 className="h-4 w-4 text-[#37352F]/60 hover:text-[#37352F]" />
                    </Button>
                    <Button variant="ghost" size="sm" onClick={() => onDelete(plan.id)} className="h-9 w-9 p-0 bg-gray-50 hover:bg-red-50 hover:text-red-600">
                        <Trash2 className="h-4 w-4 text-[#37352F]/60 hover:text-red-600" />
                    </Button>
                </div>
            </div>
        </div>
    );
};
