import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import type { TreatmentItem } from "@/lib/transforms/treatment";

interface TreatmentPlanDeleteDialogProps {
  entityLabel: string;
  pendingDelete: TreatmentItem | null;
  onClose: () => void;
  onConfirm: () => void;
}

export function TreatmentPlanDeleteDialog({
  entityLabel,
  pendingDelete,
  onClose,
  onConfirm,
}: TreatmentPlanDeleteDialogProps) {
  return (
    <ConfirmDialog
      open={pendingDelete !== null}
      onClose={onClose}
      title={`${entityLabel}を削除しますか？`}
      description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
      confirmLabel="削除"
      variant="destructive"
      onConfirm={onConfirm}
    />
  );
}
