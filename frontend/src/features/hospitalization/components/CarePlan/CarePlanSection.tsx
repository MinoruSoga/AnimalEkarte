// React/Framework
import { lazy, memo, Suspense, useCallback, useMemo, useState } from "react";

// External
import { Plus } from "lucide-react";

// Internal
import { Button } from "@/components/ui/button";
import { usePermission } from "@/features/auth/hooks/use-permission";

// Relative
import { CarePlanItemRow } from "@/features/hospitalization/components/CarePlan/CarePlanItemRow";
import { H_STYLES } from "@/features/hospitalization/styles";
import { C } from "@/lib/design-tokens";

const CarePlanDialog = lazy(() =>
  import("@/features/hospitalization/components/CarePlan/CarePlanDialog").then((m) => ({ default: m.CarePlanDialog }))
);

// Types
import type { CarePlanItem, CreateCarePlanDTO, UpdateCarePlanDTO } from "@/features/hospitalization/types";

interface CarePlanSectionProps {
    plans: CarePlanItem[];
    onAdd: (plan: CreateCarePlanDTO) => void;
    onUpdate: (id: string, plan: UpdateCarePlanDTO) => void;
    onDelete: (id: string) => void;
}

export const CarePlanSection = memo(function CarePlanSection({ plans, onAdd, onUpdate, onDelete }: CarePlanSectionProps) {
    const { canCreate, canDelete } = usePermission("hospitalization");
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingItem, setEditingItem] = useState<CarePlanItem | undefined>(undefined);

    const handleOpenCreate = useCallback(() => {
        setEditingItem(undefined);
        setIsDialogOpen(true);
    }, []);

    const handleOpenEdit = useCallback((item: CarePlanItem) => {
        setEditingItem(item);
        setIsDialogOpen(true);
    }, []);

    const planRows = useMemo(() =>
        plans.map(plan => (
            <CarePlanItemRow
                key={plan.id}
                plan={plan}
                onEdit={handleOpenEdit}
                onDelete={canDelete ? onDelete : undefined}
            />
        )),
        [plans, handleOpenEdit, onDelete, canDelete]
    );

    return (
        <div className={H_STYLES.layout.section_mb}>
            <div className="flex items-center justify-between mb-2">
                <h3 className={`${H_STYLES.text.lg} ${C.text}`}>入院治療プラン</h3>
                {canCreate ? (
                    <Button variant="primary" onClick={handleOpenCreate} className={`gap-2 ${H_STYLES.button.action}`}>
                        <Plus className={H_STYLES.button.icon} />
                        プラン追加
                    </Button>
                ) : null}
            </div>

            <div className={`flex flex-col ${H_STYLES.gap.tight}`}>
                {planRows}
            </div>

            <Suspense fallback={null}>
              <CarePlanDialog
                  open={isDialogOpen}
                  onOpenChange={setIsDialogOpen}
                  editingPlan={editingItem}
                  onCreate={onAdd}
                  onUpdate={onUpdate}
              />
            </Suspense>
        </div>
    );
});
