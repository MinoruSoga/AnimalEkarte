import { useEffect, useState } from "react";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import type { ExaminationTypeMaster } from "../api/exam-types-master";
import { ExamTypeFieldsEditor } from "./ExamTypeFieldsEditor";
import {
  TreatmentItemSidePanel,
  type TreatmentFormData,
} from "./TreatmentItemSidePanel";

interface TreatmentPlanSidePanelHostProps {
  editTarget: TreatmentItem | "new" | null;
  selectedItem: TreatmentItem | null;
  parentCandidates: TreatmentItem[];
  hasChildren: boolean;
  canDelete: boolean;
  canCreate: boolean;
  canEdit: boolean;
  examinationType?: ExaminationTypeMaster;
  /** true when active tab is procedure — show anesthesia control (BUG-028) */
  showAnesthesia?: boolean;
  onClose: () => void;
  onSave: (data: TreatmentFormData) => Promise<boolean> | boolean;
  onDeleteRequest: () => void;
  onDirtyChange: (dirty: boolean) => void;
}

export function TreatmentPlanSidePanelHost(
  props: TreatmentPlanSidePanelHostProps,
) {
  const { editTarget, examinationType } = props;
  if (editTarget === null) return null;
  const editTargetKey = editTarget === "new" ? "new" : editTarget.id;
  return (
    <TreatmentPlanSidePanelHostContent
      key={`${editTargetKey}:${examinationType?.id ?? "none"}`}
      {...props}
    />
  );
}

function TreatmentPlanSidePanelHostContent({
  editTarget,
  selectedItem,
  parentCandidates,
  hasChildren,
  canDelete,
  canCreate,
  canEdit,
  examinationType,
  showAnesthesia = false,
  onClose,
  onSave,
  onDeleteRequest,
  onDirtyChange,
}: TreatmentPlanSidePanelHostProps) {
  const [parentDirty, setParentDirty] = useState(false);
  const [fieldDirty, setFieldDirty] = useState(false);

  useEffect(() => {
    onDirtyChange(parentDirty || fieldDirty);
  }, [fieldDirty, onDirtyChange, parentDirty]);

  useEffect(() => {
    return () => onDirtyChange(false);
  }, [onDirtyChange]);

  if (editTarget === null) return null;

  return (
    <TreatmentItemSidePanel
      key={selectedItem ? String(selectedItem.id) : "new-item"}
      item={selectedItem}
      parentCandidates={parentCandidates}
      hasChildren={hasChildren}
      onClose={onClose}
      onSave={onSave}
      onDeleteRequest={canDelete ? onDeleteRequest : undefined}
      readOnly={!canEdit}
      onDirtyChange={setParentDirty}
      showAnesthesia={showAnesthesia}
      details={examinationType ? (
        <ExamTypeFieldsEditor
          examType={examinationType}
          canCreate={canCreate}
          canEdit={canEdit}
          canDelete={canDelete}
          onDirtyChange={setFieldDirty}
        />
      ) : null}
    />
  );
}
