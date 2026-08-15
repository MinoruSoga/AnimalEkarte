import { memo, useCallback, useEffect, useMemo, useState } from "react";
import { Tag } from "lucide-react";

import { MasterSidePanel, StatusToggleButton } from "@/components/shared/SidePeek";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { C, LAYOUT } from "@/lib/design-tokens";
import { normalizeKana } from "@/lib/normalize-kana";

import type { Campaign, CampaignDiscountType } from "../api/campaign";
import { useGetAllMerchandiseItems } from "../api/merchandise-items";
import { campaignToFormData, type CampaignFormData } from "./campaign-side-panel-model";

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
  return values.includes(value)
    ? values.filter((item) => item !== value)
    : [...values, value];
}

interface CampaignSidePanelProps {
  item: Campaign | null;
  onClose: () => void;
  onSave: (data: CampaignFormData) => void;
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
  const [formData, setFormData] = useState<CampaignFormData>(() => campaignToFormData(item));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");
  const [periodError, setPeriodError] = useState("");

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

  const toggleCategory = useCallback((cat: string) => {
    setFormDataDirty((prev) => ({
      ...prev,
      targetCategories: toggleSelection(prev.targetCategories, cat),
    }));
  }, [setFormDataDirty]);

  const toggleItem = useCallback((id: string) => {
    setFormDataDirty((prev) => ({
      ...prev,
      targetItemIds: toggleSelection(prev.targetItemIds, id),
    }));
  }, [setFormDataDirty]);

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

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    if (!formData.startDate || !formData.endDate) {
      setPeriodError("開始日・終了日を入力してください");
      return;
    }
    if (formData.endDate < formData.startDate) {
      setPeriodError("終了日は開始日以降にしてください");
      return;
    }
    setNameError("");
    setPeriodError("");
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
      action={handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Tag className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <div className="space-y-4">
        {/* 期間 */}
        <div className="grid w-full grid-cols-1 gap-3 sm:grid-cols-2">
          <div className="space-y-1">
            <Label>開始日</Label>
            <Input
              type="date"
              value={formData.startDate}
              disabled={readOnly}
              onChange={(e) => setFormDataDirty((prev) => ({ ...prev, startDate: e.target.value }))}
            />
          </div>
          <div className="space-y-1">
            <Label>終了日</Label>
            <Input
              type="date"
              value={formData.endDate}
              disabled={readOnly}
              onChange={(e) => setFormDataDirty((prev) => ({ ...prev, endDate: e.target.value }))}
            />
          </div>
        </div>
        {periodError ? <p className={`text-xs ${C.danger}`}>{periodError}</p> : null}

        {/* 割引種別・値 */}
        <div className="grid w-full grid-cols-1 gap-3 sm:grid-cols-2">
          <div className="space-y-1">
            <Label>割引種別</Label>
            <Select
              value={formData.discountType}
              disabled={readOnly}
              onValueChange={(v) => setFormDataDirty((prev) => ({ ...prev, discountType: v as CampaignDiscountType }))}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="rate">割引率(%)</SelectItem>
                <SelectItem value="amount">割引額(円)</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1">
            <Label>{formData.discountType === "rate" ? "割引率(%)" : "割引額(円)"}</Label>
            <Input
              type="number"
              min={0}
              max={formData.discountType === "rate" ? 100 : undefined}
              value={formData.discountValue}
              disabled={readOnly}
              onChange={(e) => setFormDataDirty((prev) => ({ ...prev, discountValue: Math.max(0, Number(e.target.value) || 0) }))}
            />
          </div>
        </div>

        {/* 対象カテゴリ（Q1=D カテゴリ単位指定） */}
        <div className="space-y-2">
          <Label>対象カテゴリ</Label>
          <div className="grid w-full grid-cols-1 gap-2 sm:grid-cols-2">
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
        </div>

        {/* 対象商品（Q1=D 個別商品指定） */}
        <div className="space-y-2">
          <Label>対象商品（個別指定）</Label>
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

        <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      </div>
    </MasterSidePanel>
  );
});
