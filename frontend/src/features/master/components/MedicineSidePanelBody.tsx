import { memo, useCallback, useEffect, useRef, useState } from "react";
import Pill from "lucide-react/dist/esm/icons/pill";
import { toast } from "sonner";

import { MasterSidePanel } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";
import type { Medicine } from "@/types";
import { MedicineCalculationTypePerWeight } from "@/types/generated/models";

import { MedicineDoseParamsEditor, type MedicineDoseParamsEditorHandle } from "./MedicineDoseParamsEditor";
import { medicineToFormData, type MedicineFormData } from "./medicine-side-panel-model";
import {
  MedicineBasicFlagsSection,
  MedicineDetailSection,
  MedicineDoseCalculationSection,
  MedicineParentCategorySection,
  MedicinePriceTaxSection,
} from "./MedicineSidePanelSections";

interface MedicineSidePanelBodyProps {
  selectedMedicine: Medicine | null;
  isCategory: boolean;
  defaultParentId?: string;
  categoryMedicines: Medicine[];
  onCloseEdit: () => void;
  onSave: (data: MedicineFormData) => Promise<boolean> | boolean;
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
  const doseParamsRef = useRef<MedicineDoseParamsEditorHandle>(null);

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(async () => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    let saved: boolean;
    if (selectedMedicine) {
      const doseOk = (await doseParamsRef.current?.saveFilled()) ?? true;
      if (!doseOk) {
        toast.error("投与量パラメータの入力内容を確認してください");
        return;
      }
      saved = Boolean(await onSave(formData));
    } else {
      const drafts = await doseParamsRef.current?.collectFilled();
      if (drafts === false) {
        toast.error("投与量パラメータの入力内容を確認してください");
        return;
      }
      saved = Boolean(await onSave({ ...formData, doseParamDrafts: drafts ?? [] }));
    }
    if (saved) {
      setIsDirty(false);
    }
  }, [formData, onSave, selectedMedicine]);

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
      onSave={handleAction}
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
      <MedicineDoseCalculationSection
        formData={formData}
        setFormDataDirty={setFormDataDirty}
      />
      {formData.calculationType === MedicineCalculationTypePerWeight ? (
        <MedicineDoseParamsEditor medicineId={selectedMedicine?.id ?? ""} ref={doseParamsRef} />
      ) : null}
    </MasterSidePanel>
  );
});
