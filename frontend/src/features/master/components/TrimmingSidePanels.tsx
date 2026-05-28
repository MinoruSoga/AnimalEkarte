import { memo, useCallback, useEffect, useState, type ChangeEvent } from "react";
import { Scissors } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { MasterSidePanel, PropertyInput, PropertyRow } from "@/components/shared/SidePeek";
import { NotionStatusPill } from "@/components/shared/StatusPill/NotionStatusPill";
import { C, LAYOUT, STYLE } from "@/lib/design-tokens";
import {
  TARGET_SIZE_OPTIONS,
  type TargetSize,
  type TrimmingCourse,
  type TrimmingOption,
} from "../api/trimming";
import { CombinablePill } from "./TrimmingTabs";

const TARGET_SIZE_SELECT_ITEMS = [
  <SelectItem key="__none__" value="__none__">指定なし</SelectItem>,
  ...TARGET_SIZE_OPTIONS.map((option) => (
    <SelectItem key={option.value} value={option.value}>
      {option.label}
    </SelectItem>
  )),
];

export interface CourseFormData {
  name: string;
  price: string;
  targetSize: TargetSize | "";
  duration: string;
  description: string;
  isActive: boolean;
}

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
  const [formData, setFormData] = useState<CourseFormData>(() => ({
    name: item?.name ?? "",
    price: item?.price != null ? String(item.price) : "",
    targetSize: item?.targetSize ?? "",
    duration: item?.duration != null ? String(item.duration) : "",
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
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

  const handleToggleStatus = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleTargetSizeChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({
      ...prev,
      targetSize: value === "__none__" ? "" : (value as TargetSize),
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
          className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <NotionStatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow label="対象サイズ">
        <Select
          value={formData.targetSize || "__none__"}
          onValueChange={handleTargetSizeChange}
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>{TARGET_SIZE_SELECT_ITEMS}</SelectContent>
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

export interface OptionFormData {
  name: string;
  price: string;
  duration: string;
  combinable: boolean;
  description: string;
  isActive: boolean;
}

interface TrimmingOptionSidePanelProps {
  item: TrimmingOption | null;
  onClose: () => void;
  onSave: (data: OptionFormData) => void;
  onDeleteRequest?: (item: TrimmingOption) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const TrimmingOptionSidePanel = memo(function TrimmingOptionSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
}: TrimmingOptionSidePanelProps) {
  const [formData, setFormData] = useState<OptionFormData>(() => ({
    name: item?.name ?? "",
    price: item?.price != null ? String(item.price) : "",
    duration: item?.duration != null ? String(item.duration) : "",
    combinable: item?.combinable ?? true,
    description: item?.description ?? "",
    isActive: item?.isActive ?? true,
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

  const handleToggleStatus = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleDurationChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, duration: value }));
  }, [setFormDataDirty]);

  const handleToggleCombinability = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, combinable: !prev.combinable }));
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
          className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <NotionStatusPill isActive={formData.isActive} />
        </button>
      </PropertyRow>

      <PropertyRow label="所要時間(分)">
        <PropertyInput
          type="number"
          value={formData.duration}
          onChange={handleDurationChange}
          placeholder="30"
        />
      </PropertyRow>

      <PropertyRow label="組合せ可否">
        <button
          type="button"
          onClick={handleToggleCombinability}
          className={`inline-flex items-center rounded-[3px] ${C.hoverBgLight} transition-colors py-0.5 px-0.5 cursor-pointer`}
        >
          <CombinablePill combinable={formData.combinable} />
        </button>
      </PropertyRow>

      <PropertyRow label="単価(税込)">
        <div className="flex items-center gap-1">
          <span className={`text-base ${C.text65} select-none`}>¥</span>
          <input
            type="number"
            min={0}
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
