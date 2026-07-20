import { memo, useCallback, useEffect, useState, type ChangeEvent } from "react";
import { Scissors } from "lucide-react";

import { MasterSidePanel, PropertyInput, PropertyRow } from "@/components/shared/SidePeek";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { C, LAYOUT, STYLE } from "@/lib/design-tokens";

import {
  TARGET_SIZE_OPTIONS,
  type TargetSize,
  type TrimmingCourse,
} from "../api/trimming";
import { useGetTrimmingCourseTypes } from "../api/trimming-course-type";
import {
  COURSE_TYPE_EMPTY_VALUE,
  TARGET_SIZE_EMPTY_VALUE,
  trimmingCourseToFormData,
  type CourseFormData,
} from "./trimming-side-panel-model";

interface TrimmingCourseSidePanelProps {
  item: TrimmingCourse | null;
  onClose: () => void;
  onSave: (data: CourseFormData) => void;
  onDeleteRequest?: (item: TrimmingCourse) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const TrimmingCourseSidePanel = memo(function TrimmingCourseSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: TrimmingCourseSidePanelProps) {
  const [formData, setFormData] = useState<CourseFormData>(() => trimmingCourseToFormData(item));
  const [nameError, setNameError] = useState("");
  const [isDirty, setIsDirty] = useState(false);
  const { data: courseTypes = [] } = useGetTrimmingCourseTypes();
  const activeCourseTypes = courseTypes.filter((t) => t.isActive || t.id === formData.courseTypeId);

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

  const handleToggleStatus = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleTargetSizeChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({
      ...prev,
      targetSize: value === TARGET_SIZE_EMPTY_VALUE ? "" : (value as TargetSize),
    }));
  }, [setFormDataDirty]);

  const handleCourseTypeChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({
      ...prev,
      courseTypeId: value === COURSE_TYPE_EMPTY_VALUE ? "" : value,
    }));
  }, [setFormDataDirty]);

  const handleDurationChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, duration: value }));
  }, [setFormDataDirty]);

  const handlePriceChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, price: event.target.value }));
  }, [setFormDataDirty]);

  const handleDescriptionChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, description: value }));
  }, [setFormDataDirty]);

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
      <PropertyRow label="ステータス">
        <button
          type="button"
          onClick={handleToggleStatus}
          className={`inline-flex items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <StatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow label="コース種別">
        <Select
          value={formData.courseTypeId || COURSE_TYPE_EMPTY_VALUE}
          onValueChange={handleCourseTypeChange}
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={COURSE_TYPE_EMPTY_VALUE}>指定なし</SelectItem>
            {activeCourseTypes.map((t) => (
              <SelectItem key={t.id} value={t.id}>
                {t.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </PropertyRow>

      <PropertyRow label="対象サイズ">
        <Select
          value={formData.targetSize || TARGET_SIZE_EMPTY_VALUE}
          onValueChange={handleTargetSizeChange}
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={TARGET_SIZE_EMPTY_VALUE}>指定なし</SelectItem>
            {TARGET_SIZE_OPTIONS.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </PropertyRow>

      <PropertyRow label="所要時間(分)">
        <PropertyInput
          type="number"
          value={formData.duration}
          onChange={handleDurationChange}
          placeholder="90"
        />
      </PropertyRow>

      <PropertyRow label="単価(税込)">
        <div className="flex items-center gap-1">
          <span className={`text-base ${C.text65} select-none`}>¥</span>
          <input
            type="number"
            min={0}
            aria-label="単価(税込)"
            className={`w-32 bg-transparent text-base ${C.text} outline-none border-none ${LAYOUT.inputCompact} ${C.hoverBgLight} ${C.focusBgLight} transition-colors ${C.textPlaceholder}`}
            value={formData.price}
            onChange={handlePriceChange}
            placeholder="0"
          />
        </div>
      </PropertyRow>

      <PropertyRow label="備考">
        <PropertyInput
          value={formData.description}
          onChange={handleDescriptionChange}
          placeholder="補足情報など"
        />
      </PropertyRow>
    </MasterSidePanel>
  );
});
