import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import type { ReservationType } from "../api/reservation-types";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";

interface ReservationTypeDeleteDialogsProps {
  pendingGroupDelete: ReservationTypeGroup | null;
  pendingCategoryDelete: ReservationType | null;
  onGroupDeleteCancel: () => void;
  onGroupDeleteConfirm: () => void;
  onCategoryDeleteCancel: () => void;
  onCategoryDeleteConfirm: () => void;
}

export function ReservationTypeDeleteDialogs({
  pendingGroupDelete,
  pendingCategoryDelete,
  onGroupDeleteCancel,
  onGroupDeleteConfirm,
  onCategoryDeleteCancel,
  onCategoryDeleteConfirm,
}: ReservationTypeDeleteDialogsProps) {
  return (
    <>
      <ConfirmDialog
        open={pendingGroupDelete !== null}
        onClose={onGroupDeleteCancel}
        title="グループを削除しますか？"
        description={`「${pendingGroupDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={onGroupDeleteConfirm}
      />
      <ConfirmDialog
        open={pendingCategoryDelete !== null}
        onClose={onCategoryDeleteCancel}
        title="予約区分を削除しますか？"
        description={`「${pendingCategoryDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={onCategoryDeleteConfirm}
      />
    </>
  );
}
