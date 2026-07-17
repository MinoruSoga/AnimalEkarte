import { memo, useState, useCallback, useEffect } from "react";
import { Layers } from "lucide-react";
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, LAYOUT, PALETTE } from "@/lib/design-tokens";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";

export interface GroupFormData {
  name: string;
  color: string;
  isActive: boolean;
}

export const GroupSidePanel = memo(function GroupSidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly, onDirtyChange,
}: {
  item: ReservationTypeGroup | null;
  onClose: () => void;
  onSave: (d: GroupFormData) => void;
  onDeleteRequest?: (i: ReservationTypeGroup) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}) {
  const [formData, setFormData] = useState<GroupFormData>(() => ({
    name: item?.name ?? "",
    color: item?.color ?? PALETTE.pickerDefaultBlue,
    isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  // BUG-380
  useEffect(() => { onDirtyChange?.(isDirty); }, [isDirty, onDirtyChange]);
  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    setIsDirty(true);
  }, []);

  const handleTitleChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: v }));
    if (v.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleColorPickerChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, color: e.target.value }));
  }, [setFormDataDirty]);

  const handleColorInputChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, color: v }));
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleAction = useCallback(() => {
    if (!formData.name.trim()) { setNameError("名称を入力してください"); return; }
    setNameError("");
    onSave(formData);
    setIsDirty(false);
  }, [formData, onSave]);

  const handleClose = useCallback(() => { setIsDirty(false); onClose(); }, [onClose]);

  return (
    <MasterSidePanel isNew={item === null} title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Layers className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty} titleError={nameError} titleMaxLength={100} readOnly={readOnly}>
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="カラー">
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center gap-2">
            <input type="color" aria-label="カラー" value={formData.color} onChange={handleColorPickerChange}
              className={LAYOUT.colorInputSmall} />
            <PropertyInput value={formData.color} onChange={handleColorInputChange} placeholder={PALETTE.pickerDefaultBlue} />
          </div>
          <p className={`text-xs ${C.text40}`}>
            予約管理カレンダーでこのグループに属する区分の予約枠を色別表示します。
          </p>
        </div>
      </PropertyRow>
    </MasterSidePanel>
  );
});
