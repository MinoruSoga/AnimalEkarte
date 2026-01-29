// React/Framework
import { useState } from "react";

// External
import { Plus } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";

// Relative
import { CarePlanItemRow } from "./CarePlanItemRow";
import { CarePlanDialog } from "./CarePlanDialog";
import { H_STYLES } from "../../styles";

// Types
import type { CarePlanItem, CreateCarePlanDTO, UpdateCarePlanDTO } from "../../types";

interface CarePlanSectionProps {
    plans: CarePlanItem[];
    onAdd: (plan: CreateCarePlanDTO) => void;
    onUpdate: (id: string, plan: UpdateCarePlanDTO) => void;
    onDelete: (id: string) => void;
}

export const CarePlanSection = ({ plans, onAdd, onUpdate, onDelete }: CarePlanSectionProps) => {
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingItem, setEditingItem] = useState<CarePlanItem | undefined>(undefined);

    const handleOpenCreate = () => {
        setEditingItem(undefined);
        setIsDialogOpen(true);
    };

    const handleOpenEdit = (item: CarePlanItem) => {
        setEditingItem(item);
        setIsDialogOpen(true);
    };

    return (
        <div className={H_STYLES.layout.section_mb}>
            <div className="flex items-center justify-between mb-2">
                <h3 className={`${H_STYLES.text.lg} text-[#37352F]`}>入院治療プラン</h3>
                <Button onClick={handleOpenCreate} className={`gap-2 bg-[#2EAADC] hover:bg-[#2EAADC]/90 text-white ${H_STYLES.button.action}`}>
                    <Plus className={H_STYLES.button.icon} />
                    プラン追加
                </Button>
            </div>

            <div className={`flex flex-col ${H_STYLES.gap.tight}`}>
                {plans.map(plan => (
                    <CarePlanItemRow
                        key={plan.id}
                        plan={plan}
                        onEdit={handleOpenEdit}
                        onDelete={onDelete}
                    />
                ))}
            </div>

            <CarePlanDialog
                open={isDialogOpen}
                onOpenChange={setIsDialogOpen}
                editingPlan={editingItem}
                onCreate={onAdd}
                onUpdate={onUpdate}
            />
        </div>
    );
};