import { memo, useCallback, useEffect, useState } from "react";
import Stethoscope from "lucide-react/dist/esm/icons/stethoscope";
import { MasterSidePanel, MoneyInput, PropertyInput, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { TaxRateSelector } from "@/components/shared/TaxRateSelector/TaxRateSelector";
import { TaxTypeSelector } from "@/components/shared/TaxTypeSelector/TaxTypeSelector";
import { SearchableSelect } from "@/components/ui/searchable-select";
import { C, LAYOUT } from "@/lib/design-tokens";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import type { TaxType } from "@/types/generated/models";

export type TreatmentFormData = {
  name: string;
  price: number;
  description: string;
  isActive: boolean;
  taxType: TaxType;
  taxRate: number;
  isNonInsurance: boolean;
  /** undefined = 未変更(Update時に送らない), "" = 親なし明示, "123" = 親指定 */
  parentId?: string;
};

const SELECT_TRIGGER_FULL = `h-[30px] text-base bg-transparent ${C.text} border-0 ${C.hoverBgLight} px-1.5 shadow-none rounded-xxs w-full`;

interface TreatmentItemSidePanelProps {
  item: TreatmentItem | null;
  /** 親候補: 現在タブ内の root 行のみ(自分自身除く). hasChildren=true の時は非表示. */
  parentCandidates: TreatmentItem[];
  /** true = 子を持つ root → parentId セレクタを非表示にして root 固定 */
  hasChildren: boolean;
  onClose: () => void;
  onSave: (data: TreatmentFormData) => void;
  onDeleteRequest?: () => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const TreatmentItemSidePanel = memo(function TreatmentItemSidePanel({
  item,
  parentCandidates,
  hasChildren,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: TreatmentItemSidePanelProps) {
  const [formData, setFormData] = useState<TreatmentFormData>(() => ({
    name: item?.name ?? "",
    price: item?.price ?? 0,
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
    taxType: (item?.taxType ?? "excluded") as TaxType,
    taxRate: item?.taxRate ?? 0.1,
    isNonInsurance: item?.isNonInsurance ?? false,
    parentId: undefined,
  }));
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
      isNew={item === null}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={onClose}
      action={readOnly ? undefined : handleAction}
      onDelete={!readOnly && item !== null && onDeleteRequest ? onDeleteRequest : undefined}
      icon={<Stethoscope className={LAYOUT.pageIcon.innerIcon} />}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={() => setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }))}
      />
      <MoneyInput
        value={formData.price}
        onChange={(value) => setFormDataDirty((prev) => ({ ...prev, price: value }))}
      />
      <PropertyRow label="課税区分">
        <TaxTypeSelector
          value={formData.taxType}
          onChange={(value) => setFormDataDirty((prev) => ({ ...prev, taxType: value }))}
        />
      </PropertyRow>
      <PropertyRow label="税率">
        <TaxRateSelector
          value={formData.taxRate}
          onChange={(value) => setFormDataDirty((prev) => ({ ...prev, taxRate: value }))}
        />
      </PropertyRow>
      <PropertyRow label="保険対象外">
        <button
          type="button"
          onClick={() => setFormDataDirty((prev) => ({ ...prev, isNonInsurance: !prev.isNonInsurance }))}
          aria-label="保険対象外を切り替え"
          className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-1.5 cursor-pointer text-sm ${formData.isNonInsurance ? C.textBrand : C.text50}`}
        >
          {formData.isNonInsurance ? "対象外" : "対象"}
        </button>
      </PropertyRow>
      {hasChildren ? (
        <PropertyRow label="親カテゴリ">
          <span className={`text-base ${C.text50}`}>
            子項目があるため変更できません
          </span>
        </PropertyRow>
      ) : (
        <PropertyRow label="親カテゴリ">
          {readOnly ? (
            <span className={`text-base ${C.text}`}>
              {item?.parentId != null
                ? (parentCandidates.find((c) => c.id === String(item.parentId))?.name ?? "不明")
                : "なし（ルート）"}
            </span>
          ) : (
            <SearchableSelect
              value={
                formData.parentId !== undefined
                  ? (formData.parentId || "__none__")
                  : (item?.parentId != null ? String(item.parentId) : "__none__")
              }
              onValueChange={(value) =>
                setFormDataDirty((prev) => ({
                  ...prev,
                  parentId: value === "__none__" ? "" : value,
                }))
              }
              options={[
                { value: "__none__", label: "なし（ルート）" },
                ...parentCandidates.map((c) => ({ value: c.id, label: c.name })),
              ]}
              searchPlaceholder="親カテゴリを検索..."
              className={SELECT_TRIGGER_FULL}
            />
          )}
        </PropertyRow>
      )}
      <PropertyRow label="備考">
        <PropertyInput
          value={formData.description}
          onChange={(value) => setFormDataDirty((prev) => ({ ...prev, description: value }))}
          placeholder="補足情報など"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});
