import { memo, useMemo, type ChangeEvent } from "react";
import { Scissors, Upload, X } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { HistoryFilterPanel } from "@/components/shared/HistoryFilterPanel";
import { LoadingFallback } from "@/components/shared/DataStates";
import { MasterLink } from "@/components/shared/MasterLink";
import { MasterSelectTrigger } from "@/components/shared/MasterSelectModal";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { C, ICON } from "@/lib/design-tokens";
import type { SortOrder } from "@/types";
import type { TrimmingFormData } from "@/types/trimming";

const BW_UNIT_ITEMS = (
  <>
    <SelectItem value="Kg">Kg</SelectItem>
    <SelectItem value="g">g</SelectItem>
  </>
);

export interface TrimmingMasterItem {
  id: string;
  name: string;
  price?: number;
  status?: string;
}

interface LeftColumnProps {
  formData: TrimmingFormData;
  courses: TrimmingMasterItem[];
  options: TrimmingMasterItem[];
  styleImagePreview: string | null;
  onFormChange: (updates: Partial<TrimmingFormData>) => void;
  onCourseModalOpen: () => void;
  onStyleImageChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onRemoveStyleImage: () => void;
  courseError?: string;
}

export const TrimmingLeftColumn = memo(function TrimmingLeftColumn({
  formData,
  courses,
  options,
  styleImagePreview,
  onFormChange,
  onCourseModalOpen,
  onStyleImageChange,
  onRemoveStyleImage,
  courseError,
}: LeftColumnProps) {
  const selectedCourse = courses.find((course) => course.id === formData.courseId);
  const optionIdSet = useMemo(() => new Set(formData.optionIds), [formData.optionIds]);

  return (
    <div className={`${C.bgWhite} rounded-lg shadow-sm border ${C.borderMedium} p-3 space-y-4`}>
      <div>
        <div className="flex items-center gap-2 mb-2">
          <Label className={`text-sm ${C.text60}`}>コース</Label>
          <MasterLink category="trimming_course" label="マスタ管理" />
        </div>
        <MasterSelectTrigger
          id="courseId"
          selectedItem={selectedCourse ? { name: selectedCourse.name, price: selectedCourse.price } : undefined}
          placeholder="コースを選択"
          icon={<Scissors className={ICON.action} />}
          onClick={onCourseModalOpen}
          variant="block"
        />
        <FormFieldError message={courseError} />
      </div>

      <div>
        <Label className={`text-sm ${C.text60} mb-2 block`}>スタイルの希望</Label>
        <Textarea
          value={formData.styleRequest}
          onChange={(event) => onFormChange({ styleRequest: event.target.value })}
          placeholder="スタイルの希望を入力..."
          className="min-h-[80px] text-sm"
        />
      </div>

      <div>
        <div className="flex items-center gap-2 mb-2">
          <Label className={`text-sm ${C.text60}`}>オプション</Label>
          <MasterLink category="trimming_option" label="マスタ管理" />
        </div>
        {options.length > 0 ? (
          <div className="space-y-2">
            {options.map((option) => (
              <div key={option.id} className="flex items-center gap-2">
                <Checkbox
                  id={`option-${option.id}`}
                  checked={optionIdSet.has(option.id)}
                  onCheckedChange={(checked) => {
                    if (checked) {
                      onFormChange({ optionIds: [...formData.optionIds, option.id] });
                    } else {
                      onFormChange({ optionIds: formData.optionIds.filter((id) => id !== option.id) });
                    }
                  }}
                />
                <label htmlFor={`option-${option.id}`} className={`text-sm ${C.text} cursor-pointer`}>
                  {option.name}
                </label>
                {option.price != null ? (
                  <span className={`text-xs ${C.text60} ml-auto`}>
                    ¥{option.price.toLocaleString()}
                  </span>
                ) : null}
              </div>
            ))}
          </div>
        ) : null}
      </div>

      <div>
        <Label className={`text-sm ${C.text60} mb-2 block`}>希望スタイル画像</Label>
        {styleImagePreview ? (
          <div className="relative">
            <img
              src={styleImagePreview}
              alt="Style preview"
              className={`w-full h-32 object-cover rounded-md border ${C.borderPrimary20}`}
            />
            <button
              type="button"
              onClick={onRemoveStyleImage}
              className={`absolute top-1 right-1 p-1 ${C.bgWhite} rounded-full shadow-sm ${C.hoverBgPage}`}
            >
              <X className={`${ICON.action} ${C.text}`} />
            </button>
          </div>
        ) : (
          <label className={`flex items-center justify-center w-full h-32 border-2 border-dashed ${C.borderMedium} rounded-md cursor-pointer ${C.hoverBgPage}`}>
            <div className="flex flex-col items-center">
              <Upload className={`${ICON.lg} ${C.text40} mb-1`} />
              <span className={`text-sm ${C.text60}`}>画像をアップロード</span>
            </div>
            <input type="file" accept="image/*" onChange={onStyleImageChange} className="hidden" />
          </label>
        )}
      </div>
    </div>
  );
});

interface MiddleColumnProps {
  formData: TrimmingFormData;
  completedImagePreview: string | null;
  onFormChange: (updates: Partial<TrimmingFormData>) => void;
  onCompletedImageChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onRemoveCompletedImage: () => void;
}

export const TrimmingMiddleColumn = memo(function TrimmingMiddleColumn({
  formData,
  completedImagePreview,
  onFormChange,
  onCompletedImageChange,
  onRemoveCompletedImage,
}: MiddleColumnProps) {
  return (
    <div className={`${C.bgWhite} rounded-lg shadow-sm border ${C.borderMedium} p-3 space-y-4`}>
      <div>
        <Label className={`text-sm ${C.text60} mb-2 block`}>体重 (BW)</Label>
        <div className="flex gap-2">
          <NumberInput
            value={formData.bw}
            onChange={(value) => onFormChange({ bw: value })}
            placeholder="体重"
            className="flex-1 text-sm"
          />
          <Select
            value={formData.bwUnit}
            onValueChange={(value) => onFormChange({ bwUnit: value as "Kg" | "g" })}
          >
            <SelectTrigger className="w-[80px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {BW_UNIT_ITEMS}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div>
        <Label className={`text-sm ${C.text60} mb-2 block`}>体温 (BT)</Label>
        <NumberInput
          step={0.1}
          value={formData.bt}
          onChange={(value) => onFormChange({ bt: value })}
          placeholder="体温"
          suffix="℃"
          className="text-sm"
        />
      </div>

      <div>
        <Label className={`text-sm ${C.text60} mb-2 block`}>使用シャンプー</Label>
        <Input
          value={formData.usedShampoo}
          onChange={(event) => onFormChange({ usedShampoo: event.target.value })}
          placeholder="シャンプー名"
          className="text-sm"
        />
      </div>

      <div>
        <Label className={`text-sm ${C.text60} mb-2 block`}>使用リボン</Label>
        <Input
          value={formData.usedRibbon}
          onChange={(event) => onFormChange({ usedRibbon: event.target.value })}
          placeholder="リボン"
          className="text-sm"
        />
      </div>

      <div>
        <Label className={`text-sm ${C.text60} mb-2 block`}>備考</Label>
        <Textarea
          value={formData.remarks}
          onChange={(event) => onFormChange({ remarks: event.target.value })}
          placeholder="備考を入力..."
          className="min-h-[60px] text-sm"
        />
      </div>

      <div>
        <Label className={`text-sm ${C.text60} mb-2 block`}>完成画像</Label>
        {completedImagePreview ? (
          <div className="relative">
            <img
              src={completedImagePreview}
              alt="Completed preview"
              className={`w-full h-32 object-cover rounded-md border ${C.borderPrimary20}`}
            />
            <button
              type="button"
              onClick={onRemoveCompletedImage}
              className={`absolute top-1 right-1 p-1 ${C.bgWhite} rounded-full shadow-sm ${C.hoverBgPage}`}
            >
              <X className={`${ICON.action} ${C.text}`} />
            </button>
          </div>
        ) : (
          <label className={`flex items-center justify-center w-full h-32 border-2 border-dashed ${C.borderMedium} rounded-md cursor-pointer ${C.hoverBgPage}`}>
            <div className="flex flex-col items-center">
              <Upload className={`${ICON.lg} ${C.text40} mb-1`} />
              <span className={`text-sm ${C.text60}`}>画像をアップロード</span>
            </div>
            <input type="file" accept="image/*" onChange={onCompletedImageChange} className="hidden" />
          </label>
        )}
      </div>
    </div>
  );
});

export interface TrimmingHistoryItem {
  id: string;
  date: string;
  styleRequest: string;
  courseId: string;
  optionIds: string[];
  usedShampoo: string;
  usedRibbon: string;
  remarks: string;
  staff: string;
}

interface RightColumnProps {
  sortedHistory: TrimmingHistoryItem[];
  isHistoryLoading: boolean;
  historySearchTerm: string;
  historySortOrder: SortOrder;
  historyDateRange: { from: string; to: string };
  onSearchTermChange: (value: string) => void;
  onSortOrderChange: (value: SortOrder) => void;
  onClear: () => void;
  onFilterStartDateChange: (value: string) => void;
  onFilterEndDateChange: (value: string) => void;
  onHistoryClick: (updates: Partial<TrimmingFormData>) => void;
}

export const TrimmingRightColumn = memo(function TrimmingRightColumn({
  sortedHistory,
  isHistoryLoading,
  historySearchTerm,
  historySortOrder,
  historyDateRange,
  onSearchTermChange,
  onSortOrderChange,
  onClear,
  onFilterStartDateChange,
  onFilterEndDateChange,
  onHistoryClick,
}: RightColumnProps) {
  const historyCards = useMemo(
    () =>
      sortedHistory.map((history) => (
        <button
          key={history.id}
          type="button"
          className={`block w-full text-left p-3 border ${C.borderMedium} rounded-lg ${C.bgWhite} ${C.hoverBgPage} transition-colors cursor-pointer`}
          onClick={() =>
            onHistoryClick({
              courseId: history.courseId,
              optionIds: history.optionIds,
              styleRequest: history.styleRequest,
              usedShampoo: history.usedShampoo,
              usedRibbon: history.usedRibbon,
              remarks: history.remarks,
              staffName: history.staff,
            })
          }
        >
          <div className="flex items-start justify-between">
            <div className="flex-1 min-w-0">
              <div className={`text-xs ${C.text60} mb-1`}>{history.date}</div>
              <div className={`text-sm ${C.text} font-medium truncate`}>{history.styleRequest}</div>
              <div className={`text-xs ${C.text60} mt-1`}>{history.staff}</div>
            </div>
          </div>
        </button>
      )),
    [sortedHistory, onHistoryClick]
  );

  return (
    <div className={`${C.bgWhite} rounded-lg shadow-sm border ${C.borderMedium} p-3 space-y-4`}>
      <div>
        <Label className={`text-sm ${C.text60} mb-2 block`}>施術履歴</Label>
        <HistoryFilterPanel
          searchTerm={historySearchTerm}
          onSearchTermChange={onSearchTermChange}
          sortOrder={historySortOrder}
          onSortOrderChange={onSortOrderChange}
          onClear={onClear}
          showDateRange={true}
          filterStartDate={historyDateRange.from}
          onFilterStartDateChange={onFilterStartDateChange}
          filterEndDate={historyDateRange.to}
          onFilterEndDateChange={onFilterEndDateChange}
          searchPlaceholder="スタイル希望で検索..."
        />
      </div>

      <div className="space-y-2 max-h-[600px] overflow-y-auto">
        {isHistoryLoading ? (
          <LoadingFallback />
        ) : sortedHistory.length === 0 ? (
          <div className={`text-center py-8 text-sm ${C.text40}`}>
            施術履歴がありません
          </div>
        ) : (
          historyCards
        )}
      </div>
    </div>
  );
});
