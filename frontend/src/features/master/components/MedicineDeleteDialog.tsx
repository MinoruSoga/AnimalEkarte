import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import type { Medicine } from "@/types";

interface MedicineDeleteDialogProps {
  open: boolean;
  medicine: Medicine | null;
  onClose: () => void;
  onConfirm: () => void;
}

export function MedicineDeleteDialog({
  open,
  medicine,
  onClose,
  onConfirm,
}: MedicineDeleteDialogProps) {
  return (
    <ConfirmDialog
      open={open}
      onClose={onClose}
      onConfirm={onConfirm}
      title="削除しますか？"
      description={`「${medicine?.name ?? ""}」を削除します。この操作は取り消せません。`}
      confirmLabel="削除する"
      cancelLabel="キャンセル"
      variant="destructive"
    />
  );
}
