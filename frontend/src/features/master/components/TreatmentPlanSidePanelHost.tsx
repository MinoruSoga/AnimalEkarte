import type { TreatmentItem } from "@/lib/transforms/treatment";
import {
  TreatmentItemSidePanel,
  type TreatmentFormData,
} from "./TreatmentItemSidePanel";

interface TreatmentPlanSidePanelHostProps {
  editTarget: TreatmentItem | "new" | null;
  selectedItem: TreatmentItem | null;
  canDelete: boolean;
  canEdit: boolean;
  onClose: () => void;
  onSave: (data: TreatmentFormData) => void;
  onDeleteRequest: () => void;
  onDirtyChange: (dirty: boolean) => void;
}

export function TreatmentPlanSidePanelHost({
  editTarget,
  selectedItem,
  canDelete,
  canEdit,
  onClose,
  onSave,
  onDeleteRequest,
  onDirtyChange,
}: TreatmentPlanSidePanelHostProps) {
  if (editTarget === null) return null;

  return (
    <TreatmentItemSidePanel
      key={selectedItem ? String(selectedItem.id) : "new-item"}
      item={selectedItem}
      onClose={onClose}
      onSave={onSave}
      onDeleteRequest={canDelete ? onDeleteRequest : undefined}
      readOnly={!canEdit}
      onDirtyChange={onDirtyChange}
    />
  );
}
