import { memo, useState, useCallback } from "react";
import { Layers } from "lucide-react";
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, LAYOUT, PALETTE } from "@/lib/design-tokens";
import type { ReservationTypeGroup } from "@/features/master/api/reservation-type-groups";

export interface GroupFormData {
  name: string;
  color: string;
  isActive: boolean;
}

export const GroupSidePanel = memo(function GroupSidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly,
}: {
  item: ReservationTypeGroup | null;
  onClose: () => void;
  onSave: (d: GroupFormData) => void;
  onDeleteRequest?: (i: ReservationTypeGroup) => void;
  readOnly?: boolean;
}) {
  const [f, setF] = useState<GroupFormData>(() => ({
    name: item?.name ?? "",
    color: item?.color ?? PALETTE.pickerDefaultBlue,
    isActive: item?.isActive ?? true,
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  const handleTitleChange = useCallback((v: string) => {
    setF((p) => ({ ...p, name: v }));
    setIsDirty(true);
    if (v.trim()) setNameError("");
  }, []);

  const handleColorPickerChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setF((p) => ({ ...p, color: e.target.value }));
    setIsDirty(true);
  }, []);

  const handleColorInputChange = useCallback((v: string) => {
    setF((p) => ({ ...p, color: v }));
    setIsDirty(true);
  }, []);

  const handleToggleActive = useCallback(() => {
    setF((p) => ({ ...p, isActive: !p.isActive }));
    setIsDirty(true);
  }, []);

  const handleAction = useCallback(() => {
    if (!f.name.trim()) { setNameError("名称を入力してください"); return; }
    setNameError("");
    onSave(f);
    setIsDirty(false);
  }, [f, onSave]);

  const handleClose = useCallback(() => { setIsDirty(false); onClose(); }, [onClose]);

  return (
    <MasterSidePanel isNew={item === null} title={f.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Layers className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty} titleError={nameError} titleMaxLength={100} readOnly={readOnly}>
      <StatusToggleButton isActive={f.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="カラー">
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center gap-2">
            <input type="color" value={f.color} onChange={handleColorPickerChange}
              className="w-7 h-7 rounded cursor-pointer border-0 bg-transparent p-0" />
            <PropertyInput value={f.color} onChange={handleColorInputChange} placeholder="#3B82F6" />
          </div>
          <p className={`text-xs ${C.text40}`}>
            予約管理カレンダーでこのグループに属する区分の予約枠を色別表示します。
          </p>
        </div>
      </PropertyRow>
    </MasterSidePanel>
  );
});
