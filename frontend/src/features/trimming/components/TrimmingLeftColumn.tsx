import { memo, useMemo } from "react";
import { Scissors } from "lucide-react";

import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { MasterLink } from "@/components/shared/MasterLink";
import { MasterSelectTrigger } from "@/components/shared/MasterSelectModal";
import { C, ICON } from "@/lib/design-tokens";

import type { TrimmingLeftColumnProps } from "./trimming-form-column-types";
import { TrimmingImageUploadField } from "./TrimmingImageUploadField";

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
}: TrimmingLeftColumnProps) {
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
