import { memo, useCallback, useState } from "react";
import { Building2 } from "lucide-react";

import {
  MoneyInput,
  MasterSidePanel,
  PropertyInput,
  PropertyRow,
  StatusToggleButton,
} from "@/components/shared/SidePeek";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FIELD_DESCRIPTIONS } from "@/constants/field-descriptions";
import { LAYOUT, STYLE } from "@/lib/design-tokens";

import type { Cage, CageSize, CageType } from "../api/cages";
import { useMasterSidePanelForm } from "../hooks/use-master-side-panel-form";
import {
  CAGE_SIZE_OPTIONS,
  CAGE_TYPE_OPTIONS,
  cageToFormData,
  type CageFormData,
} from "../lib/cage-side-panel-model";

interface CageSidePanelProps {
  item: Cage | null;
  onClose: () => void;
  onSave: (data: CageFormData) => Promise<boolean> | boolean;
  onDeleteRequest?: (item: Cage) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const CageSidePanel = memo(function CageSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: CageSidePanelProps) {
  const [nameError, setNameError] = useState("");

  const {
    formData,
    setFormData: setFormDataDirty,
    isDirty,
    setIsDirty,
    handleAction,
  } = useMasterSidePanelForm<CageFormData>({
    initialFormData: cageToFormData(item),
    onSave,
    onDirtyChange,
    validate: (data) => {
      if (!data.name.trim()) {
        setNameError("名称を入力してください");
        return false;
      }
      setNameError("");
      return true;
    },
  });

  const handleTitleChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, name: value }));
      if (value.trim()) setNameError("");
    },
    [setFormDataDirty],
  );

  const handleCageTypeChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, cageType: value as CageType }));
    },
    [setFormDataDirty],
  );

  const handleCageSizeChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, cageSize: value as CageSize }));
    },
    [setFormDataDirty],
  );

  const handlePriceChange = useCallback(
    (value: number) => {
      setFormDataDirty((prev) => ({ ...prev, price: value }));
    },
    [setFormDataDirty],
  );

  const handleDescriptionChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, description: value }));
    },
    [setFormDataDirty],
  );

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose, setIsDirty]);

  return (
    <MasterSidePanel
      isNew={item === null}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      onSave={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Building2 className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow
        label="エリア"
        description="ケージが属するエリア種別（ICU・犬舎・猫舎・汎用）です。入院管理でケージを分類するために使われます。"
      >
        <Select value={formData.cageType} onValueChange={handleCageTypeChange}>
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>
            {CAGE_TYPE_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </PropertyRow>
      <PropertyRow
        label="サイズ"
        description="ケージのサイズ区分です。入院する動物の体格に合ったケージを選ぶ目安になります。"
      >
        <Select value={formData.cageSize} onValueChange={handleCageSizeChange}>
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>
            {CAGE_SIZE_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </PropertyRow>
      <MoneyInput
        value={formData.price}
        onChange={handlePriceChange}
        description="このケージを利用した場合の料金（税込）です。入院料金の計算に使われます。"
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
