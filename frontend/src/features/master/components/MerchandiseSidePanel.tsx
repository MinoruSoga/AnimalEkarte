import { memo, useCallback, useState } from "react";
import { ShoppingBag } from "lucide-react";

import {
  MoneyInput,
  MasterSidePanel,
  PropertyRow,
  StatusToggleButton,
} from "@/components/shared/SidePeek";
import { TaxRateSelector } from "@/components/shared/TaxRateSelector/TaxRateSelector";
import { TaxTypeSelector } from "@/components/shared/TaxTypeSelector/TaxTypeSelector";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FIELD_DESCRIPTIONS } from "@/constants/field-descriptions";
import { LAYOUT, STYLE } from "@/lib/design-tokens";

import type { FrontendMerchandiseItem } from "../api/merchandise-items";
import { useMasterSidePanelForm } from "../hooks/use-master-side-panel-form";
import {
  MERCHANDISE_CATEGORY_OPTIONS,
  merchandiseToFormData,
  type MerchandiseFormData,
} from "../lib/merchandise-side-panel-model";

interface MerchandiseSidePanelProps {
  item: FrontendMerchandiseItem | null;
  onClose: () => void;
  onSave: (data: MerchandiseFormData) => Promise<boolean> | boolean;
  onDeleteRequest?: (item: FrontendMerchandiseItem) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const MerchandiseSidePanel = memo(function MerchandiseSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: MerchandiseSidePanelProps) {
  const [nameError, setNameError] = useState("");

  const {
    formData,
    setFormData: setFormDataDirty,
    isDirty,
    setIsDirty,
    handleAction,
  } = useMasterSidePanelForm<MerchandiseFormData>({
    initialFormData: merchandiseToFormData(item),
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

  const handleCategoryChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, category: value }));
    },
    [setFormDataDirty],
  );

  const handleUnitPriceChange = useCallback(
    (value: number) => {
      setFormDataDirty((prev) => ({ ...prev, unitPrice: value }));
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
      action={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<ShoppingBag className={LAYOUT.pageIcon.innerIcon} />}
      titlePlaceholder="品目名"
      titleDescription="販売する品目の名称です。物販登録や会計時の選択肢に表示されます。"
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow
        label="カテゴリ"
        description="この品目の分類です。物販一覧の整理や会計時の分類に使われます。"
      >
        <Select value={formData.category} onValueChange={handleCategoryChange}>
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>
            {MERCHANDISE_CATEGORY_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </PropertyRow>
      <MoneyInput
        value={formData.unitPrice}
        onChange={handleUnitPriceChange}
        description={FIELD_DESCRIPTIONS.unitPrice}
      />
      <PropertyRow label="課税区分" description={FIELD_DESCRIPTIONS.taxCategory}>
        <TaxTypeSelector
          value={formData.taxType}
          onChange={(value) => setFormDataDirty((prev) => ({ ...prev, taxType: value }))}
        />
      </PropertyRow>
      <PropertyRow label="税率" description={FIELD_DESCRIPTIONS.taxRate}>
        <TaxRateSelector
          value={formData.taxRate}
          onChange={(value) => setFormDataDirty((prev) => ({ ...prev, taxRate: value }))}
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});
