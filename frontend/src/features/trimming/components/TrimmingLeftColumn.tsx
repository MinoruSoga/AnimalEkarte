import { memo, useMemo } from "react";
import { Scissors } from "lucide-react";

import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { MasterLink } from "@/components/shared/MasterLink";
import { MasterSelectTrigger } from "@/components/shared/MasterSelectModal/MasterSelectTrigger";
import { C, ICON } from "@/lib/design-tokens";
import { formatCurrency } from "@/lib/format/number";

import type { TrimmingLeftColumnProps } from "./trimming-form-column-types";
import { TrimmingImageUploadField } from "./TrimmingImageUploadField";

const INITIAL_STATUS_VALUES = ["in_consultation", "pending"] as const;
type InitialStatus = (typeof INITIAL_STATUS_VALUES)[number];
const isInitialStatus = (value: string): value is InitialStatus =>
  (INITIAL_STATUS_VALUES as readonly string[]).includes(value);

const INITIAL_STATUS_ITEMS = (
  <>
    <SelectItem value="in_consultation">進行中</SelectItem>
    <SelectItem value="pending">予約</SelectItem>
  </>
);

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
  showInitialStatusSelector,
}: TrimmingLeftColumnProps) {
  const selectedCourse = courses.find((course) => course.id === formData.courseId);
  const optionIdSet = useMemo(() => new Set(formData.optionIds), [formData.optionIds]);

  return (
    <div className={`${C.bgWhite} rounded-lg border ${C.borderMedium} p-3 space-y-4`}>
      {showInitialStatusSelector ? (
        <div>
          <Label className={`text-sm ${C.text60} mb-2 block`}>登録時ステータス</Label>
          <Select
            value={formData.initialStatus}
            onValueChange={(value) => {
              if (isInitialStatus(value)) onFormChange({ initialStatus: value });
            }}
          >
            <SelectTrigger className="w-[160px]" aria-label="登録時ステータスを選択">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {INITIAL_STATUS_ITEMS}
            </SelectContent>
          </Select>
        </div>
      ) : null}

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
        <Label htmlFor="trimming-style-request" className={`text-sm ${C.text60} mb-2 block`}>
          スタイルの希望
        </Label>
        <Textarea
          id="trimming-style-request"
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
                  aria-label={option.name}
                  touchTarget
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
                    {formatCurrency(option.price)}
                  </span>
                ) : null}
              </div>
            ))}
          </div>
        ) : null}
      </div>

      <TrimmingImageUploadField
        label="希望スタイル画像"
        preview={styleImagePreview}
        previewAlt="Style preview"
        onImageChange={onStyleImageChange}
        onRemoveImage={onRemoveStyleImage}
      />
    </div>
  );
});
