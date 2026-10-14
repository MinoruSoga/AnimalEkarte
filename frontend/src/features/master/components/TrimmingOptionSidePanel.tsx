import { memo, useCallback, useEffect, useState } from "react";
import { Scissors } from "lucide-react";

import {
  MasterSidePanel,
  MoneyInput,
  PropertyInput,
  PropertyRow,
} from "@/components/shared/SidePeek";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { FIELD_DESCRIPTIONS } from "@/constants/field-descriptions";
import { C, LAYOUT } from "@/lib/design-tokens";

import type { TrimmingOption } from "../api/trimming";
import { CombinablePill } from "./TrimmingTabRows";
import { trimmingOptionToFormData, type OptionFormData } from "../lib/trimming-side-panel-model";

interface TrimmingOptionSidePanelProps {
  item: TrimmingOption | null;
  onClose: () => void;
  onSave: (data: OptionFormData) => Promise<boolean> | boolean;
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

  const handleAction = useCallback(async () => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    const saved = await onSave(formData);
    if (saved) {
      setIsDirty(false);
      onDirtyChange?.(false);
    }
  }, [formData, onSave, onDirtyChange]);

  const handleTitleChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, name: value }));
      if (value.trim()) setNameError("");
    },
    [setFormDataDirty],
  );

  const handleToggleStatus = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleDurationChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, duration: value }));
    },
    [setFormDataDirty],
  );

  const handleToggleCombinability = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, combinable: !prev.combinable }));
  }, [setFormDataDirty]);

  const handleDescriptionChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, description: value }));
    },
    [setFormDataDirty],
  );

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
      <PropertyRow label="ステータス" description={FIELD_DESCRIPTIONS.status}>
        <button
          type="button"
          onClick={handleToggleStatus}
          className={`inline-flex min-h-11 items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <StatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow
        label="所要時間(分)"
        description="このオプションを追加した場合にかかる目安時間（分）です。"
      >
        <NumberInput value={formData.duration} onChange={handleDurationChange} placeholder="30" />
      </PropertyRow>

      <PropertyRow
        label="組合せ可否"
        description="このオプションを他のオプションと同時に選択できるかどうかです。不可にすると単独でのみ選択できます。"
      >
        <button
          type="button"
          onClick={handleToggleCombinability}
          className={`inline-flex min-h-11 items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <CombinablePill combinable={formData.combinable} />
        </button>
      </PropertyRow>

      <MoneyInput
        min={0}
        stringValue={formData.price}
        onStringChange={(price) => setFormDataDirty((prev) => ({ ...prev, price }))}
        description={FIELD_DESCRIPTIONS.unitPrice}
      />

      <PropertyRow label="備考" description={FIELD_DESCRIPTIONS.note}>
        <PropertyInput
          value={formData.description}
          onChange={handleDescriptionChange}
          placeholder="補足情報など"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});
