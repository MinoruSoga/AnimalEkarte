import type { DiagnosisName, DiagnosisType } from "../api/diagnosis";
import type {
  DiagnosisNameFormData,
  DiagnosisTypeFormData,
} from "./diagnosis-side-panel-model";
import { DiagnosisNameSidePanel } from "./DiagnosisNameSidePanel";
import { DiagnosisTypeSidePanel } from "./DiagnosisTypeSidePanel";
import type { DiagnosisTabValue } from "../routes/diagnosis-settings-model";

interface DiagnosisSettingsSidePanelsProps {
  activeTab: DiagnosisTabValue;
  typeEditTarget: DiagnosisType | "new" | null;
  typePanelItem: DiagnosisType | null;
  nameEditTarget: DiagnosisName | "new" | null;
  namePanelItem: DiagnosisName | null;
  categories: DiagnosisType[];
  canDelete: boolean;
  canEdit: boolean;
  onTypeClose: () => void;
  onTypeSave: (data: DiagnosisTypeFormData) => Promise<boolean> | boolean;
  onTypeDeleteRequest: (item: DiagnosisType) => void;
  onNameClose: () => void;
  onNameSave: (data: DiagnosisNameFormData) => Promise<boolean> | boolean;
  onNameDeleteRequest: (item: DiagnosisName) => void;
  onDirtyChange: (dirty: boolean) => void;
}

export function DiagnosisSettingsSidePanels({
  activeTab,
  typeEditTarget,
  typePanelItem,
  nameEditTarget,
  namePanelItem,
  categories,
  canDelete,
  canEdit,
  onTypeClose,
  onTypeSave,
  onTypeDeleteRequest,
  onNameClose,
  onNameSave,
  onNameDeleteRequest,
  onDirtyChange,
}: DiagnosisSettingsSidePanelsProps) {
  return (
    <>
      {activeTab === "diagnosis_type" && typeEditTarget !== null ? (
        <DiagnosisTypeSidePanel
          key={typePanelItem ? String(typePanelItem.id) : "new-category"}
          item={typePanelItem}
          onClose={onTypeClose}
          onSave={onTypeSave}
          onDeleteRequest={canDelete ? onTypeDeleteRequest : undefined}
          readOnly={!canEdit}
          onDirtyChange={onDirtyChange}
        />
      ) : null}
      {activeTab === "diagnosis_name" && nameEditTarget !== null ? (
        <DiagnosisNameSidePanel
          key={namePanelItem ? String(namePanelItem.id) : "new-name"}
          item={namePanelItem}
          categories={categories}
          onClose={onNameClose}
          onSave={onNameSave}
          onDeleteRequest={canDelete ? onNameDeleteRequest : undefined}
          readOnly={!canEdit}
          onDirtyChange={onDirtyChange}
        />
      ) : null}
    </>
  );
}
