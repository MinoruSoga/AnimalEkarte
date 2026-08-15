import { memo, useCallback, useEffect, useState, type ChangeEvent } from "react";
import { Scissors } from "lucide-react";

import { MasterSidePanel, PropertyInput, PropertyRow } from "@/components/shared/SidePeek";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { C, LAYOUT } from "@/lib/design-tokens";

import type { TrimmingOption } from "../api/trimming";
import { CombinablePill } from "./TrimmingTabRows";
import {
  trimmingOptionToFormData,
  type OptionFormData,
} from "./trimming-side-panel-model";

interface TrimmingOptionSidePanelProps {
  item: TrimmingOption | null;
  onClose: () => void;
  onSave: (data: OptionFormData) => void;
  onDeleteRequest?: (item: TrimmingOption) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const TrimmingOptionSidePanel = memo(function TrimmingOptionSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: TrimmingOptionSidePanelProps) {
  const [formData, setFormData] = useState<OptionFormData>(() => trimmingOptionToFormData(item));
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

  const handleToggleStatus = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleDurationChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, duration: value }));
  }, [setFormDataDirty]);

  const handleToggleCombinability = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, combinable: !prev.combinable }));
  }, [setFormDataDirty]);

  const handlePriceChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, price: event.target.value }));
  }, [setFormDataDirty]);

  const handleDescriptionChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, description: value }));
  }, [setFormDataDirty]);

  return (
    <MasterSidePanel
      isNew={item === null}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={onClose}
      action={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Scissors className={LAYOUT.pageIcon.innerIcon} />}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <PropertyRow label="ステータス">
        <button
          type="button"
          onClick={handleToggleStatus}
          className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <StatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow label="所要時間(分)">
        <PropertyInput
          type="number"
          value={formData.duration}
          onChange={handleDurationChange}
          placeholder="30"
        />
      </PropertyRow>

      <PropertyRow label="組合せ可否">
        <button
          type="button"
          onClick={handleToggleCombinability}
          className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <CombinablePill combinable={formData.combinable} />
        </button>
      </PropertyRow>

      <PropertyRow label="単価(税込)">
        <div className="flex items-center gap-1">
          <span className={`text-base ${C.text65} select-none`}>¥</span>
          <input
            type="number"
            min={0}
            aria-label="単価(税込)"
            className={`w-32 bg-transparent text-base ${C.text} outline-none border-none ${LAYOUT.inputCompact} ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
            value={formData.price}
            onChange={handlePriceChange}
            placeholder="0"
          />
        </div>
      </PropertyRow>

      <PropertyRow label="備考">
        <PropertyInput
          value={formData.description}
          onChange={handleDescriptionChange}
          placeholder="補足情報など"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});
