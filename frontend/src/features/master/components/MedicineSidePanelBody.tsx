import { memo, useCallback, useEffect, useState } from "react";
import Pill from "lucide-react/dist/esm/icons/pill";

import { MasterSidePanel } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";
import type { Medicine } from "@/types";

import { medicineToFormData, type MedicineFormData } from "./medicine-side-panel-model";
import {
  MedicineBasicFlagsSection,
  MedicineDetailSection,
  MedicineParentCategorySection,
  MedicinePriceTaxSection,
} from "./MedicineSidePanelSections";

interface MedicineSidePanelBodyProps {
  selectedMedicine: Medicine | null;
  isCategory: boolean;
  defaultParentId?: string;
  categoryMedicines: Medicine[];
  onCloseEdit: () => void;
  onSave: (data: MedicineFormData) => void;
  onDeleteRequest: () => void;
  readOnly?: boolean;
  canDelete?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const MedicineSidePanelBody = memo(function MedicineSidePanelBody({
  selectedMedicine,
  isCategory,
  defaultParentId,
  categoryMedicines,
  onCloseEdit,
  onSave,
  onDeleteRequest,
  readOnly,
  canDelete,
  onDirtyChange,
}: MedicineSidePanelBodyProps) {
  const [formData, setFormData] = useState<MedicineFormData>(() =>
    medicineToFormData(selectedMedicine, defaultParentId),
  );
  const [nameError, setNameError] = useState("");
  const [isDirty, setIsDirty] = useState(false);

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(formData);
    setIsDirty(false);
  }, [formData, onSave]);

  const handleTitleChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: value }));
    if (value.trim()) setNameError("");
  }, [setFormDataDirty]);

  return (
    <MasterSidePanel
      isNew={!selectedMedicine}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={onCloseEdit}
      action={handleAction}
      onDelete={selectedMedicine && canDelete ? onDeleteRequest : undefined}
      icon={<Pill className={LAYOUT.pageIcon.innerIcon} />}
      titlePlaceholder="薬品名"
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <MedicineParentCategorySection
        formData={formData}
        isCategory={isCategory}
        categoryMedicines={categoryMedicines}
        setFormDataDirty={setFormDataDirty}
      />
      <MedicinePriceTaxSection
        formData={formData}
        isCategory={isCategory}
        setFormDataDirty={setFormDataDirty}
      />
      <MedicineBasicFlagsSection
        formData={formData}
        setFormDataDirty={setFormDataDirty}
      />
      <MedicineDetailSection
        formData={formData}
        setFormDataDirty={setFormDataDirty}
      />
    </MasterSidePanel>
  );
});
