import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import type { Clinic } from "../api/clinics";

interface ClinicDeleteDialogProps {
  pendingDelete: Clinic | null;
  isPending: boolean;
  onClose: () => void;
  onConfirm: () => void;
}

export function ClinicDeleteDialog({
  pendingDelete,
  isPending,
  onClose,
  onConfirm,
}: ClinicDeleteDialogProps) {
  return (
    <ConfirmDialog
      open={pendingDelete !== null}
      onClose={onClose}
      title="医院を削除しますか？"
      description={`「${pendingDelete?.name}」を削除します。この操作は取り消せません。`}
      confirmLabel="削除"
      variant="destructive"
      onConfirm={onConfirm}
      isPending={isPending}
    />
  );
}
