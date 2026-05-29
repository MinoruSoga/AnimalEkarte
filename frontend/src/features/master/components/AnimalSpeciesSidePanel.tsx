import { memo, useCallback, useEffect, useState } from "react";
import PawPrint from "lucide-react/dist/esm/icons/paw-print";

import { MasterSidePanel, StatusToggleButton } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";

import type { AnimalSpecies } from "../api/animal-species";
import {
  animalSpeciesToFormData,
  type AnimalSpeciesFormData,
} from "./AnimalSpeciesSidePanelModel";

interface AnimalSpeciesSidePanelProps {
  item: AnimalSpecies | null;
  onClose: () => void;
  onSave: (data: AnimalSpeciesFormData) => void;
  onDeleteRequest?: (item: AnimalSpecies) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const AnimalSpeciesSidePanel = memo(function AnimalSpeciesSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: AnimalSpeciesSidePanelProps) {
  const [formData, setFormData] = useState<AnimalSpeciesFormData>(() =>
    animalSpeciesToFormData(item),
  );
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleTitleChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: value }));
    if (value.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    onSave(formData);
    setIsDirty(false);
  }, [formData, onSave]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel
      isNew={item === null}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<PawPrint className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
    </MasterSidePanel>
  );
});
