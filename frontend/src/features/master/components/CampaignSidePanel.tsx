import { memo, useCallback, useMemo, useState } from "react";
import { Tag } from "lucide-react";
import { toast } from "sonner";

import { MasterSidePanel, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { FormFieldError } from "@/components/shared/FormFieldError";
import { Pagination } from "@/components/shared/Pagination";
import { usePagination } from "@/hooks/use-pagination";
import { Input } from "@/components/ui/input";
import { DatePicker } from "@/components/shared/DatePicker";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
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
  { value: "house_call", label: "往診" },
];

/** Returns a new selection array with `value` toggled, without mutating the input array. */
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
  // EMR-319: 必須未入力は片側だけを invalid にする。日付前後エラーはペア課題のため
  // missing を空にして両欄 invalid のまま残す（LSTEP と同じ per-field 契約）。
  const [periodMissing, setPeriodMissing] = useState<ReadonlySet<"start" | "end">>(new Set());
  // 全欠落が解消した時点の periodError は必ず missing 由来（order エラーは
  // missing 空でのみ成立）なので、最後の欠落解除でメッセージも消す。
  const clearPeriodMissing = useCallback(
    (key: "start" | "end") => {
      if (!periodMissing.has(key)) return;
      const next = new Set(periodMissing);
      next.delete(key);
      setPeriodMissing(next);
      if (next.size === 0) setPeriodError("");
    },
    [periodMissing],
  );

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
      // PR#614 R2-3/R3-P2: 名称と期間は独立した必須群。同一パスで全て評価し、
      // 一方の失敗が他方の正常解除（エラークリア）をスキップしないよう、
      // 各群のエラー状態を独立に更新してから全体の成否を返す。
      let valid = true;
      if (!data.name.trim()) {
        setNameError("名称を入力してください");
        toast.error("名称を入力してください");
        valid = false;
      } else {
        setNameError("");
      }
      if (!data.startDate || !data.endDate) {
        const missing = new Set<"start" | "end">();
        if (!data.startDate) missing.add("start");
        if (!data.endDate) missing.add("end");
        setPeriodMissing(missing);
        setPeriodError("開始日・終了日を入力してください");
        toast.error("開始日・終了日を入力してください");
        valid = false;
      } else if (data.endDate < data.startDate) {
        setPeriodMissing(new Set());
        setPeriodError("終了日は開始日以降にしてください");
        toast.error("終了日は開始日以降にしてください");
        valid = false;
      } else {
        setPeriodError("");
        setPeriodMissing(new Set());
      }
      return valid;
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
  const pagination = usePagination(filteredMerchandise, {
    pageSize: 50,
    resetKey: merchandiseSearch,
  });

  // 日付前後エラー（missing 空・periodError 有）はペア課題として両欄 invalid。
  const periodOrderError = periodError !== "" && periodMissing.size === 0;
  const startDateInvalid = periodMissing.has("start") || periodOrderError;
  const endDateInvalid = periodMissing.has("end") || periodOrderError;

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose, setIsDirty]);

  return (
    <MasterSidePanel
      tabletOverlay
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
      titleDescription="キャンペーンの名称です。キャンペーン一覧や会計時の適用確認で使われます。"
      readOnly={readOnly}
    >
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow
        label="開始日"
        description="キャンペーンの適用開始日です。この日以降の会計に割引が適用されます。"
      >
        <DatePicker
          ariaLabel="開始日"
          value={formData.startDate}
          onChange={(v) => {
            setFormDataDirty((prev) => ({ ...prev, startDate: v }));
            clearPeriodMissing("start");
          }}
          ariaInvalid={startDateInvalid}
          ariaDescribedBy={startDateInvalid ? "campaign-period-error" : undefined}
          disabled={readOnly}
        />
      </PropertyRow>
      <PropertyRow
        label="終了日"
        description="キャンペーンの適用終了日です。この日を過ぎると割引は適用されません。"
      >
        <DatePicker
          ariaLabel="終了日"
          value={formData.endDate}
          onChange={(v) => {
            setFormDataDirty((prev) => ({ ...prev, endDate: v }));
            clearPeriodMissing("end");
          }}
          ariaInvalid={endDateInvalid}
          ariaDescribedBy={endDateInvalid ? "campaign-period-error" : undefined}
          disabled={readOnly}
        />
      </PropertyRow>
      <FormFieldError id="campaign-period-error" message={periodError} />
      <PropertyRow
        label="割引種別"
        description="割引の指定方法です。「割引率(%)」は対象価格からの割合、「割引額(円)」は固定金額を割り引きます。"
      >
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
      <PropertyRow
        label={formData.discountType === "rate" ? "割引率(%)" : "割引額(円)"}
        description={
          formData.discountType === "rate"
            ? "対象品目の価格から割り引く割合（%）です。"
            : "対象品目の価格から割り引く金額（円）です。"
        }
      >
        <NumberInput
          aria-label={formData.discountType === "rate" ? "割引率(%)" : "割引額(円)"}
          min={0}
          max={formData.discountType === "rate" ? 100 : undefined}
          value={formData.discountValue}
          disabled={readOnly}
          onChange={(v) =>
            setFormDataDirty((prev) => ({
              ...prev,
              discountValue: Math.max(0, Number(v) || 0),
            }))
          }
        />
      </PropertyRow>
      <PropertyRow
        label="対象カテゴリ"
        description="割引を適用する項目カテゴリを選択します。選択したカテゴリの会計に割引が適用されます。"
      >
        <div className="grid w-full grid-cols-1 gap-2">
          {CATEGORY_OPTIONS.map((o) => (
            <label key={o.value} className="flex items-center gap-2 text-sm min-h-11">
              <Checkbox
                checked={formData.targetCategories.includes(o.value)}
                disabled={readOnly}
                onCheckedChange={() => toggleCategory(o.value)}
                touchTarget
              />
              {o.label}
            </label>
          ))}
        </div>
      </PropertyRow>
      <PropertyRow
        label="対象商品"
        description="割引を適用する個別の品目を選択します。カテゴリ指定に加えて個別指定できます。"
      >
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
              pagination.paginatedData.map((mItem) => (
                <label key={mItem.id} className="flex items-center gap-2 text-sm min-h-11">
                  <Checkbox
                    checked={formData.targetItemIds.includes(mItem.id)}
                    disabled={readOnly}
                    onCheckedChange={() => toggleItem(mItem.id)}
                    touchTarget
                  />
                  {mItem.name}
                </label>
              ))
            )}
          </div>
          <Pagination
            currentPage={pagination.currentPage}
            totalPages={pagination.totalPages}
            totalCount={pagination.totalCount}
            startIndex={pagination.startIndex}
            endIndex={pagination.endIndex}
            onPageChange={pagination.goToPage}
            onPrev={pagination.prevPage}
            onNext={pagination.nextPage}
          />
        </div>
      </PropertyRow>
    </MasterSidePanel>
  );
});
