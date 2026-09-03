import { memo, useCallback, useMemo, useState } from "react";
import { Tag } from "lucide-react";
import { toast } from "sonner";

import { MasterSidePanel, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { C, LAYOUT } from "@/lib/design-tokens";
import { normalizeKana } from "@/lib/normalize-kana";

import type { Campaign, CampaignDiscountType } from "../api/campaign";
import { useGetAllMerchandiseItems } from "../api/merchandise-items";
import { useMasterSidePanelForm } from "../hooks/use-master-side-panel-form";
import { campaignToFormData, type CampaignFormData } from "../lib/campaign-side-panel-model";

// item_category に対応する対象カテゴリの選択肢
const CATEGORY_OPTIONS: { value: string; label: string }[] = [
  { value: "food", label: "フード" },
  { value: "goods", label: "物販" },
  { value: "medicine", label: "処方" },
  { value: "examination", label: "診察" },
  { value: "test", label: "検査" },
  { value: "procedure", label: "処置" },
  { value: "surgery", label: "手術" },
  { value: "vaccine", label: "ワクチン" },
  { value: "trimming", label: "トリミング" },
  { value: "hotel", label: "ホテル" },
  { value: "training", label: "しつけ" },
  { value: "other", label: "その他" },
];

function toggleSelection(values: string[], value: string): string[] {
  return values.includes(value) ? values.filter((item) => item !== value) : [...values, value];
}

interface CampaignSidePanelProps {
  item: Campaign | null;
  onClose: () => void;
  onSave: (data: CampaignFormData) => Promise<boolean> | boolean;
  onDeleteRequest?: (item: Campaign) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const CampaignSidePanel = memo(function CampaignSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: CampaignSidePanelProps) {
  const [nameError, setNameError] = useState("");
  const [periodError, setPeriodError] = useState("");

  const {
    formData,
    setFormData: setFormDataDirty,
    isDirty,
    setIsDirty,
    handleAction,
  } = useMasterSidePanelForm<CampaignFormData>({
    initialFormData: campaignToFormData(item),
    onSave,
    onDirtyChange,
    validate: (data) => {
      if (!data.name.trim()) {
        setNameError("名称を入力してください");
        toast.error("名称を入力してください");
        return false;
      }
      if (!data.startDate || !data.endDate) {
        setPeriodError("開始日・終了日を入力してください");
        toast.error("開始日・終了日を入力してください");
        return false;
      }
      if (data.endDate < data.startDate) {
        setPeriodError("終了日は開始日以降にしてください");
        toast.error("終了日は開始日以降にしてください");
        return false;
      }
      setNameError("");
      setPeriodError("");
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

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const toggleCategory = useCallback(
    (cat: string) => {
      setFormDataDirty((prev) => ({
        ...prev,
        targetCategories: toggleSelection(prev.targetCategories, cat),
      }));
    },
    [setFormDataDirty],
  );

  const toggleItem = useCallback(
    (id: string) => {
      setFormDataDirty((prev) => ({
        ...prev,
        targetItemIds: toggleSelection(prev.targetItemIds, id),
      }));
    },
    [setFormDataDirty],
  );

  const { data: merchandiseItems = [] } = useGetAllMerchandiseItems();
  const [merchandiseSearch, setMerchandiseSearch] = useState("");
  const filteredMerchandise = useMemo(() => {
    let result = merchandiseItems.filter((i) => i.isActive);
    if (merchandiseSearch) {
      const lower = normalizeKana(merchandiseSearch).toLowerCase();
      result = result.filter((i) => normalizeKana(i.name).toLowerCase().includes(lower));
    }
    return result;
  }, [merchandiseItems, merchandiseSearch]);

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
      onSave={handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Tag className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="開始日">
        <Input
          type="date"
          aria-label="開始日"
          value={formData.startDate}
          disabled={readOnly}
          onChange={(e) => setFormDataDirty((prev) => ({ ...prev, startDate: e.target.value }))}
        />
      </PropertyRow>
      <PropertyRow label="終了日">
        <Input
          type="date"
          aria-label="終了日"
          value={formData.endDate}
          disabled={readOnly}
          onChange={(e) => setFormDataDirty((prev) => ({ ...prev, endDate: e.target.value }))}
        />
      </PropertyRow>
      {periodError ? <p className={`text-xs ${C.danger}`}>{periodError}</p> : null}
      <PropertyRow label="割引種別">
        <Select
          value={formData.discountType}
          disabled={readOnly}
          onValueChange={(v) =>
            setFormDataDirty((prev) => ({ ...prev, discountType: v as CampaignDiscountType }))
          }
        >
          <SelectTrigger aria-label="割引種別">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="rate">割引率(%)</SelectItem>
            <SelectItem value="amount">割引額(円)</SelectItem>
          </SelectContent>
        </Select>
      </PropertyRow>
      <PropertyRow label={formData.discountType === "rate" ? "割引率(%)" : "割引額(円)"}>
        <Input
          type="number"
          min={0}
          max={formData.discountType === "rate" ? 100 : undefined}
          aria-label={formData.discountType === "rate" ? "割引率(%)" : "割引額(円)"}
          value={formData.discountValue}
          disabled={readOnly}
          onChange={(e) =>
            setFormDataDirty((prev) => ({
              ...prev,
              discountValue: Math.max(0, Number(e.target.value) || 0),
            }))
          }
        />
      </PropertyRow>
      <PropertyRow label="対象カテゴリ">
        <div className="grid w-full grid-cols-1 gap-2">
          {CATEGORY_OPTIONS.map((o) => (
            <label key={o.value} className="flex items-center gap-2 text-sm">
              <Checkbox
                checked={formData.targetCategories.includes(o.value)}
                disabled={readOnly}
                onCheckedChange={() => toggleCategory(o.value)}
              />
              {o.label}
            </label>
          ))}
        </div>
      </PropertyRow>
      <PropertyRow label="対象商品">
        <div className="w-full space-y-2">
          <Input
            placeholder="商品名で検索..."
            value={merchandiseSearch}
            disabled={readOnly}
            onChange={(e) => setMerchandiseSearch(e.target.value)}
          />
          <div className="max-h-48 space-y-1 overflow-y-auto rounded border p-2">
            {filteredMerchandise.length === 0 ? (
              <p className={`text-xs ${C.text50}`}>商品がありません</p>
            ) : (
              filteredMerchandise.map((mItem) => (
                <label key={mItem.id} className="flex items-center gap-2 text-sm">
                  <Checkbox
                    checked={formData.targetItemIds.includes(mItem.id)}
                    disabled={readOnly}
                    onCheckedChange={() => toggleItem(mItem.id)}
                  />
                  {mItem.name}
                </label>
              ))
            )}
          </div>
        </div>
      </PropertyRow>
    </MasterSidePanel>
  );
});
