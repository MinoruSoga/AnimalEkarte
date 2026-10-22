import { memo, useCallback, useEffect, useRef, useState } from "react";
import Pill from "lucide-react/dist/esm/icons/pill";
import { toast } from "sonner";

import { MasterSidePanel } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";
import type { Medicine } from "@/types";
import { MedicineCalculationTypePerWeight } from "@/types/generated/models";

import {
  MedicineDoseParamsEditor,
  type MedicineDoseParamsEditorHandle,
} from "./MedicineDoseParamsEditor";
import {
  getMedicineStrengthError,
  medicineToFormData,
  type MedicineFormData,
} from "../lib/medicine-side-panel-model";
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
  const [strengthError, setStrengthError] = useState("");
  const [isDirty, setIsDirty] = useState(false);
  const doseParamsRef = useRef<MedicineDoseParamsEditorHandle>(null);

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  // 計算方式・含量を編集したら含量エラーを解除する（再保存時に再検査する）。
  const setDoseFormDataDirty = useCallback<typeof setFormData>(
    (updater) => {
      setFormDataDirty(updater);
      setStrengthError("");
    },
    [setFormDataDirty],
  );

  const handleAction = useCallback(async () => {
    // EMR-447 Q4: 必須含量の空欄は当該欄に理由を出し、送信・成功通知・close をしない（入力は保持）。
    const nameMissing = !formData.name.trim();
    const nextStrengthError = getMedicineStrengthError(formData) ?? "";
    setNameError(nameMissing ? "名称を入力してください" : "");
    setStrengthError(nextStrengthError);
    if (nameMissing || nextStrengthError) {
      return;
    }
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

  const handleTitleChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, name: value }));
      if (value.trim()) setNameError("");
    },
    [setFormDataDirty],
  );

  return (
    <MasterSidePanel
      tabletOverlay
      isNew={!selectedMedicine}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={onCloseEdit}
      onSave={handleAction}
      onDelete={selectedMedicine && canDelete ? onDeleteRequest : undefined}
      icon={<Pill className={LAYOUT.pageIcon.innerIcon} />}
      titlePlaceholder="薬品名"
      titleDescription="薬品の名称です。処方登録や薬品選択肢に表示されます。カテゴリの場合は分類名です。"
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
      <MedicineBasicFlagsSection formData={formData} setFormDataDirty={setFormDataDirty} />
      <MedicineDetailSection formData={formData} setFormDataDirty={setFormDataDirty} />
      <MedicineDoseCalculationSection
        formData={formData}
        setFormDataDirty={setDoseFormDataDirty}
        strengthError={strengthError}
      />
      {formData.calculationType === MedicineCalculationTypePerWeight ? (
        <MedicineDoseParamsEditor medicineId={selectedMedicine?.id ?? ""} ref={doseParamsRef} />
      ) : null}
    </MasterSidePanel>
  );
});
