import { ConfirmDialog } from "@/components/shared/ConfirmDialog/ConfirmDialog";
import type { TrimmingCourse, TrimmingOption } from "../api/trimming";

interface TrimmingDeleteDialogsProps {
  pendingCourseDelete: TrimmingCourse | null;
  pendingOptionDelete: TrimmingOption | null;
  onCourseDeleteCancel: () => void;
  onCourseDeleteConfirm: () => void;
  onOptionDeleteCancel: () => void;
  onOptionDeleteConfirm: () => void;
}

export function TrimmingDeleteDialogs({
  pendingCourseDelete,
  pendingOptionDelete,
  onCourseDeleteCancel,
  onCourseDeleteConfirm,
  onOptionDeleteCancel,
  onOptionDeleteConfirm,
}: TrimmingDeleteDialogsProps) {
  return (
    <>
      <ConfirmDialog
        open={pendingCourseDelete !== null}
        onClose={onCourseDeleteCancel}
        title="トリミングコースを削除しますか？"
        description={`「${pendingCourseDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={onCourseDeleteConfirm}
      />
      <ConfirmDialog
        open={pendingOptionDelete !== null}
        onClose={onOptionDeleteCancel}
        title="トリミングオプションを削除しますか？"
        description={`「${pendingOptionDelete?.name}」を削除します。この操作は取り消せません。`}
        confirmLabel="削除"
        variant="destructive"
        onConfirm={onOptionDeleteConfirm}
      />
    </>
  );
}
