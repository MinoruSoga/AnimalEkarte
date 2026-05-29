import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import type { DiagnosisName, DiagnosisType } from "../api/diagnosis";

interface DiagnosisDeleteDialogsProps {
  pendingTypeDelete: DiagnosisType | null;
  pendingNameDelete: DiagnosisName | null;
  onTypeDeleteCancel: () => void;
  onTypeDeleteConfirm: () => void;
  onNameDeleteCancel: () => void;
  onNameDeleteConfirm: () => void;
}

export function DiagnosisDeleteDialogs({
  pendingTypeDelete,
  pendingNameDelete,
  onTypeDeleteCancel,
  onTypeDeleteConfirm,
  onNameDeleteCancel,
  onNameDeleteConfirm,
}: DiagnosisDeleteDialogsProps) {
  return (
    <>
      <ConfirmDialog
        open={pendingTypeDelete !== null}
        onClose={onTypeDeleteCancel}
        title="診断カテゴリを削除しますか？"
        description={`「${pendingTypeDelete?.name}」を削除します。このカテゴリに属する診断名も影響を受けます。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={onTypeDeleteConfirm}
      />
      <ConfirmDialog
        open={pendingNameDelete !== null}
        onClose={onNameDeleteCancel}
        title="診断病名を削除しますか？"
        description={`「${pendingNameDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={onNameDeleteConfirm}
      />
    </>
  );
}
