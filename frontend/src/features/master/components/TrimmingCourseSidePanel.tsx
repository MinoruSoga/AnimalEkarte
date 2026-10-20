import { memo, useCallback, useEffect, useRef, useTransition, useState } from "react";
import { Scissors } from "lucide-react";

import {
  MasterSidePanel,
  MoneyInput,
  PropertyInput,
  PropertyRow,
} from "@/components/shared/SidePeek";
import { NumberInput } from "@/components/shared/NumberInput/NumberInput";
import { StatusPill } from "@/components/shared/StatusPill/StatusPill";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FIELD_DESCRIPTIONS } from "@/constants/field-descriptions";
import { useCourseImageDrafts } from "../hooks/use-course-image-drafts";
import { useCurrentClinicName } from "@/hooks/use-current-clinic-name";
import { C, LAYOUT, STYLE } from "@/lib/design-tokens";

import { TARGET_SIZE_OPTIONS, type TargetSize, type TrimmingCourse } from "../api/trimming";
import { useGetTrimmingCourseTypes } from "../api/trimming-course-type";
import {
  COURSE_TYPE_EMPTY_VALUE,
  TARGET_SIZE_EMPTY_VALUE,
  trimmingCourseToFormData,
  type CourseFormData,
} from "../lib/trimming-side-panel-model";
import {
  TrimmingCourseImageField,
  TrimmingCourseWeekdayField,
} from "./TrimmingCourseAttributeFields";

interface TrimmingCourseSidePanelProps {
  item: TrimmingCourse | null;
  refreshedImageUrl?: string | null;
  onClose: () => void;
  onSave: (data: CourseFormData) => Promise<boolean> | boolean;
  onDeleteRequest?: (item: TrimmingCourse) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const TrimmingCourseSidePanel = memo(function TrimmingCourseSidePanel({
  item,
  refreshedImageUrl,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: TrimmingCourseSidePanelProps) {
  const [formData, setFormData] = useState<CourseFormData>(() => trimmingCourseToFormData(item));
  const [nameError, setNameError] = useState("");
  const [isDirty, setIsDirty] = useState(false);
  const [isImageUploading, setIsImageUploading] = useState(false);
  const [isSaving, startSaving] = useTransition();
  const savePending = useRef(false);
  const { register, discard, protectSubmitted } = useCourseImageDrafts();
  const clinicName = useCurrentClinicName();
  const { data: courseTypes = [] } = useGetTrimmingCourseTypes();
  const activeCourseTypes = courseTypes.filter((t) => t.isActive || t.id === formData.courseTypeId);

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(async () => {
    // 画像 upload 中は保存しない（成功 key を受け取る前の送信で画像を失わないため）
    if (isImageUploading || savePending.current) return;
    if (!formData.name.trim()) {
      setNameError("名称を入力してください");
      return;
    }
    setNameError("");
    savePending.current = true;
    protectSubmitted(formData.imageObjectKey);
    startSaving(async () => {
      try {
        const saved = await onSave(formData);
        if (saved) {
          setIsDirty(false);
          onDirtyChange?.(false);
        }
      } finally {
        savePending.current = false;
      }
    });
  }, [formData, isImageUploading, onSave, onDirtyChange, protectSubmitted]);

  const handleTitleChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, name: value }));
      if (value.trim()) setNameError("");
    },
    [setFormDataDirty],
  );

  const handleToggleStatus = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleTargetSizeChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({
        ...prev,
        targetSize: value === TARGET_SIZE_EMPTY_VALUE ? "" : (value as TargetSize),
      }));
    },
    [setFormDataDirty],
  );

  const handleCourseTypeChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({
        ...prev,
        courseTypeId: value === COURSE_TYPE_EMPTY_VALUE ? "" : value,
      }));
    },
    [setFormDataDirty],
  );

  const handleDurationChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, duration: value }));
    },
    [setFormDataDirty],
  );

  const handleShortNameChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, shortName: value, shortNameChanged: true }));
    },
    [setFormDataDirty],
  );

  const handleWeekdaysChange = useCallback(
    (allowedWeekdays: number[]) => {
      setFormDataDirty((prev) => ({ ...prev, allowedWeekdays, weekdaysChanged: true }));
    },
    [setFormDataDirty],
  );

  const handleImageUploaded = useCallback(
    (imageObjectKey: string, imagePreviewUrl: string) => {
      if (!register(imageObjectKey)) return;
      discard(formData.imageObjectKey);
      setFormDataDirty((prev) => ({
        ...prev,
        imageObjectKey,
        imagePreviewUrl,
        imageChanged: true,
      }));
    },
    [setFormDataDirty, register, discard, formData.imageObjectKey],
  );

  const handleImageRemove = useCallback(() => {
    if (savePending.current) return;
    discard(formData.imageObjectKey);
    setFormDataDirty((prev) => ({
      ...prev,
      imageObjectKey: "",
      imagePreviewUrl: "",
      imageChanged: true,
    }));
  }, [setFormDataDirty, discard, formData.imageObjectKey]);

  const handleDescriptionChange = useCallback(
    (value: string) => {
      setFormDataDirty((prev) => ({ ...prev, description: value }));
    },
    [setFormDataDirty],
  );

  return (
    <MasterSidePanel
      tabletOverlay
      isNew={item === null}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={onClose}
      isPending={isSaving || isImageUploading}
      action={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Scissors className={LAYOUT.pageIcon.innerIcon} />}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <PropertyRow label="ステータス" description={FIELD_DESCRIPTIONS.status}>
        <button
          type="button"
          onClick={handleToggleStatus}
          className={`inline-flex min-h-11 items-center rounded-xxs ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <StatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow
        label="短縮名"
        description="予約表などで名称の代わりに使う短い表示名です。任意で、未登録でもコースは使えます。"
      >
        <PropertyInput
          value={formData.shortName}
          onChange={handleShortNameChange}
          placeholder="未登録"
          ariaLabel="短縮名"
        />
      </PropertyRow>

      <PropertyRow
        label="利用医院"
        description="このコースはここで選択中の医院のマスタとして登録され、この医院でだけ使われます。"
      >
        <span className={`px-1.5 py-0.5 text-sm ${C.text}`}>{clinicName || "選択中の医院"}</span>
      </PropertyRow>

      <PropertyRow
        label="コース種別"
        description="このコースが属するコース種別です。トリミングメニューの分類に使われます。"
      >
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

      <PropertyRow
        label="対象サイズ"
        description="このコースの対象となる犬のサイズです。予約・施術時の案内の目安に使われます。"
      >
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

      <PropertyRow
        label="所要時間(分)"
        description="このコースの施術にかかる目安時間（分）です。予約枠の見通しに使われます。"
      >
        <NumberInput value={formData.duration} onChange={handleDurationChange} placeholder="90" />
      </PropertyRow>

      <MoneyInput
        min={0}
        stringValue={formData.price}
        onStringChange={(price) => setFormDataDirty((prev) => ({ ...prev, price }))}
        description={FIELD_DESCRIPTIONS.unitPrice}
      />

      <TrimmingCourseWeekdayField
        value={formData.allowedWeekdays}
        onChange={handleWeekdaysChange}
        readOnly={readOnly}
      />

      <TrimmingCourseImageField
        currentKey={formData.imageObjectKey}
        previewUrl={
          !formData.imageChanged && refreshedImageUrl != null
            ? refreshedImageUrl
            : formData.imagePreviewUrl
        }
        hasImage={formData.imageObjectKey !== ""}
        onUploaded={handleImageUploaded}
        onRemove={handleImageRemove}
        onUploadingChange={(uploading) => {
          setIsImageUploading(uploading);
          if (uploading) setIsDirty(true);
        }}
        readOnly={readOnly || isSaving}
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
