import { memo, useState, useCallback } from "react";
import { Activity, MessageCircle } from "lucide-react";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PropertyRow, StatusToggleButton, PropertyInput, MasterSidePanel } from "@/components/shared/SidePeek";
import { C, ICON, LAYOUT, PALETTE, STYLE } from "@/lib/design-tokens";
import { ReservationTypeAvailableSlotsSection } from "./ReservationTypeAvailableSlotsSection";
import { ReservationTypeUnavailableTimesSection } from "./ReservationTypeUnavailableTimesSection";
import { ReservationTypeOccupationsSection } from "./ReservationTypeOccupationsSection";
import type { ReservationType } from "../api/reservation-types";
import { useMasterSidePanelForm } from "../hooks/use-master-side-panel-form";

// ── 静的 SelectItem JSX (rendering-hoist-jsx) ──────────────────
const RESERVATION_DAY_OPTION_ITEMS = (
  <>
    <SelectItem value="none">制限なし</SelectItem>
    <SelectItem value="weekday">平日のみ</SelectItem>
    <SelectItem value="saturday">土曜含む</SelectItem>
    <SelectItem value="anyday">毎日</SelectItem>
  </>
);

export interface CategoryFormData {
  name: string;
  description: string;
  isActive: boolean;
  groupId: string | undefined;
  reservationDisplayName: string;
  durationMinutes: number;
  shortName: string;
  reservationVisible: boolean;
  reservationComment: string;
  reservationImageUrl: string;
  showShortName: boolean;
  reservationDayOption: string;
  isInternal: boolean;
}

interface GroupOption { id: string; name: string; color: string; }

export const CategorySidePanel = memo(function CategorySidePanel({
  item, onClose, onSave, onDeleteRequest, readOnly, groups, defaultGroupId, onDirtyChange,
}: {
  item: ReservationType | null;
  onClose: () => void;
  onSave: (d: CategoryFormData) => Promise<boolean> | boolean;
  onDeleteRequest?: (i: ReservationType) => void;
  readOnly?: boolean;
  groups: GroupOption[];
  defaultGroupId?: string;
  onDirtyChange?: (dirty: boolean) => void;
}) {
  const [nameError, setNameError] = useState("");

  const { formData, setFormData: setFormDataDirty, isDirty, setIsDirty, handleAction } =
    useMasterSidePanelForm<CategoryFormData>({
      initialFormData: {
        name: item?.name ?? "",
        description: item?.description ?? "",
        isActive: item?.isActive ?? true,
        groupId: item?.groupId ?? defaultGroupId,
        reservationDisplayName: item?.reservationDisplayName ?? "",
        durationMinutes: item?.durationMinutes ?? 15,
        shortName: item?.shortName ?? "",
        reservationVisible: item?.reservationVisible ?? true,
        reservationComment: item?.reservationComment ?? "",
        reservationImageUrl: item?.reservationImageUrl ?? "",
        showShortName: item?.showShortName ?? false,
        reservationDayOption: item?.reservationDayOption ?? "none",
        isInternal: item?.isInternal ?? false,
      },
      onSave,
      onDirtyChange,
      validate: (data) => {
        if (!data.name.trim()) {
          setNameError("名称を入力してください");
          return false;
        }
        setNameError("");
        return true;
      },
    });

  const handleTitleChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: v }));
    if (v.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleDescriptionChange = useCallback((v: string) => {
    setFormDataDirty((prev) => ({ ...prev, description: v }));
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleClose = useCallback(() => { setIsDirty(false); onClose(); }, [onClose, setIsDirty]);

  return (
    <MasterSidePanel isNew={item === null} title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={readOnly ? undefined : handleAction}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<Activity className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty} titleError={nameError} titleMaxLength={100} readOnly={readOnly}>
      <StatusToggleButton isActive={formData.isActive} onToggle={handleToggleActive} />
      <PropertyRow label="グループ">
        <Select
          value={formData.groupId ?? "none"}
          onValueChange={(v) => setFormDataDirty((prev) => ({ ...prev, groupId: v === "none" ? undefined : v }))}
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="グループを選択" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="none">未分類</SelectItem>
            {groups.map((g) => (
              <SelectItem key={g.id} value={g.id}>
                <div className="flex items-center gap-2">
                  <span className={`${ICON.dotMd} rounded-full shrink-0`} style={{ backgroundColor: g.color }} />
                  {g.name}
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </PropertyRow>
      <PropertyRow label="備考">
        <PropertyInput value={formData.description} onChange={handleDescriptionChange} placeholder="補足情報など" />
      </PropertyRow>

      <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
        <div className="flex items-center gap-1.5 mb-3">
          <MessageCircle className={ICON.smXs} style={{ color: PALETTE.lineGreen }} />
          <p className={`text-xs font-medium ${C.text50}`}>LINE予約設定</p>
        </div>
        <PropertyRow label="LINE表示名">
          <PropertyInput value={formData.reservationDisplayName}
            onChange={(v) => setFormDataDirty((prev) => ({ ...prev, reservationDisplayName: v }))}
            placeholder={formData.name || "空欄なら名称を使用"} />
        </PropertyRow>
        <PropertyRow label="予約ページに表示">
          <Switch checked={formData.reservationVisible}
            onCheckedChange={(v) => setFormDataDirty((prev) => ({ ...prev, reservationVisible: v }))} />
        </PropertyRow>
        <PropertyRow label="内部サービス">
          <Switch checked={formData.isInternal}
            onCheckedChange={(v) => setFormDataDirty((prev) => ({ ...prev, isInternal: v }))} />
        </PropertyRow>
        <PropertyRow label="所要時間（分）">
          <input type="number" min={5} max={480}
            aria-label="所要時間（分）"
            className={`w-20 border ${C.borderMedium} ${LAYOUT.inputStandard} text-base ${C.text}`}
            value={formData.durationMinutes}
            onChange={(e) => setFormDataDirty((prev) => ({ ...prev, durationMinutes: Number(e.target.value) || 15 }))} />
        </PropertyRow>
        <PropertyRow label="略称">
          <PropertyInput value={formData.shortName}
            onChange={(v) => setFormDataDirty((prev) => ({ ...prev, shortName: v }))}
            placeholder="LINE表示用の略称" />
        </PropertyRow>
        <PropertyRow label="略称を使用">
          <Switch checked={formData.showShortName}
            onCheckedChange={(v) => setFormDataDirty((prev) => ({ ...prev, showShortName: v }))} />
        </PropertyRow>
        <PropertyRow label="画像URL">
          <PropertyInput value={formData.reservationImageUrl}
            onChange={(v) => setFormDataDirty((prev) => ({ ...prev, reservationImageUrl: v }))}
            placeholder="https://..." />
        </PropertyRow>
        <PropertyRow label="予約可能曜日">
          <Select value={formData.reservationDayOption}
            onValueChange={(v) => setFormDataDirty((prev) => ({ ...prev, reservationDayOption: v }))}>
            <SelectTrigger className={STYLE.selectCompact}><SelectValue /></SelectTrigger>
            <SelectContent>{RESERVATION_DAY_OPTION_ITEMS}</SelectContent>
          </Select>
        </PropertyRow>
        <PropertyRow label="LINE説明文">
          <PropertyInput value={formData.reservationComment}
            onChange={(v) => setFormDataDirty((prev) => ({ ...prev, reservationComment: v }))}
            placeholder="LINE予約画面に表示する説明" />
        </PropertyRow>
      </div>

      {item !== null ? (
        <>
          {item.isLeaf ? (
            <>
              <ReservationTypeAvailableSlotsSection
                clinicId={item.clinicId}
                reservationTypeId={item.id}
              />
              <ReservationTypeUnavailableTimesSection
                clinicId={item.clinicId}
                reservationTypeId={item.id}
              />
            </>
          ) : (
            <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
              <p className={`text-xs ${C.text40}`}>
                子予約区分ごとに予約枠を設定してください
              </p>
            </div>
          )}
          <ReservationTypeOccupationsSection
            clinicId={item.clinicId}
            reservationTypeId={item.id}
          />
        </>
      ) : null}
    </MasterSidePanel>
  );
});
