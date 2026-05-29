import type { Dispatch, ReactNode, SetStateAction } from "react";
import { Building2, CheckCircle2, MessageCircle, Shield } from "lucide-react";

import { Checkbox } from "@/components/ui/checkbox";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { PropertyRow } from "@/components/shared/SidePeek";
import { C, ICON, PALETTE, STYLE } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS } from "../constants/styles";
import type { ClinicSummary } from "../api/staffs";
import type { PermissionGroup } from "../api/permission-groups";
import type { ReservationType } from "../api/reservation-types";
import type { StaffFormData } from "./StaffSidePanelModel";

const STAFF_TYPE_SELECT_ITEMS = (
  <>
    <SelectItem value="doctor">医師</SelectItem>
    <SelectItem value="nurse">看護師</SelectItem>
    <SelectItem value="resource">設備</SelectItem>
  </>
);

interface StaffLineReservationSectionProps {
  formData: StaffFormData;
  setFormDataDirty: Dispatch<SetStateAction<StaffFormData>>;
}

export function StaffLineReservationSection({ formData, setFormDataDirty }: StaffLineReservationSectionProps) {
  return (
      <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
        <div className="flex items-center gap-1.5 mb-3">
          <MessageCircle className={ICON.smXs} style={{ color: PALETTE.lineGreen }} />
          <p className={`text-xs font-medium ${C.text50}`}>LINE予約設定</p>
        </div>

        <PropertyRow label="LINE表示名">
          <input
            type="text"
            className={MASTER_INPUT_CLASS}
            value={formData.reservationDisplayName}
            onChange={(event) => setFormDataDirty((prev) => ({ ...prev, reservationDisplayName: event.target.value }))}
            placeholder={formData.name || "空欄なら氏名を使用"}
          />
        </PropertyRow>

        <PropertyRow label="予約ページに表示">
          <Switch
            checked={formData.reservationVisible}
            onCheckedChange={(value) => setFormDataDirty((prev) => ({ ...prev, reservationVisible: value }))}
          />
        </PropertyRow>

        <PropertyRow label="スタッフ種別">
          <Select
            value={formData.staffType}
            onValueChange={(value) => setFormDataDirty((prev) => ({ ...prev, staffType: value }))}
          >
            <SelectTrigger className={STYLE.selectCompact}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>{STAFF_TYPE_SELECT_ITEMS}</SelectContent>
          </Select>
        </PropertyRow>

        <PropertyRow label="LINE説明文">
          <input
            type="text"
            className={MASTER_INPUT_CLASS}
            value={formData.reservationComment}
            onChange={(event) => setFormDataDirty((prev) => ({ ...prev, reservationComment: event.target.value }))}
            placeholder="LINE予約画面に表示する説明"
          />
        </PropertyRow>

        <PropertyRow label="画像URL">
          <input
            type="text"
            className={MASTER_INPUT_CLASS}
            value={formData.reservationImageUrl}
            onChange={(event) => setFormDataDirty((prev) => ({ ...prev, reservationImageUrl: event.target.value }))}
            placeholder="https://..."
          />
        </PropertyRow>
      </div>
  );
}

interface StaffExcludedReservationTypesSectionProps {
  activeReservationTypes: ReservationType[];
  allReservationTypes: ReservationType[];
  capableIdSet: Set<string>;
  isNew: boolean;
  onToggle: (reservationTypeId: string, checked: boolean) => void;
}

interface StaffCheckboxItem {
  id: string;
  name: string;
}

interface StaffCheckboxSectionProps<T extends StaffCheckboxItem> {
  title: string;
  icon: ReactNode;
  items: T[];
  selectedIdSet: Set<string>;
  isNew: boolean;
  newMessage: string;
  isEmpty: boolean;
  emptyMessage: string;
  onToggle: (id: string, checked: boolean) => void;
  renderLeading?: (item: T) => ReactNode;
}

function StaffCheckboxSection<T extends StaffCheckboxItem>({
  title,
  icon,
  items,
  selectedIdSet,
  isNew,
  newMessage,
  isEmpty,
  emptyMessage,
  onToggle,
  renderLeading,
}: StaffCheckboxSectionProps<T>) {
  return (
    <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
      <div className="flex items-center gap-1.5 mb-2">
        {icon}
        <p className={`text-xs font-medium ${C.text50}`}>{title}</p>
      </div>

      {isNew ? (
        <p className={`text-xs ${C.text50} pl-0.5`}>
          {newMessage}
        </p>
      ) : isEmpty ? (
        <p className={`text-xs ${C.text50} pl-0.5`}>
          {emptyMessage}
        </p>
      ) : (
        <div className="space-y-0.5">
          {items.map((item) => (
            <label
              key={item.id}
              className={`flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer ${C.hoverBgLight} transition-colors`}
            >
              <Checkbox
                checked={selectedIdSet.has(item.id)}
                onCheckedChange={(checked) =>
                  onToggle(item.id, checked === true)
                }
              />
              {renderLeading?.(item)}
              <span className="text-sm">{item.name}</span>
            </label>
          ))}
        </div>
      )}
    </div>
  );
}

function reservationCategoryLabel(category: string): string {
  switch (category) {
    case "trimming":
      return "トリミング";
    case "hospitalization":
      return "入院・ホテル";
    case "general":
    case "medical":
      return "診療";
    default:
      return "その他";
  }
}

export function StaffExcludedReservationTypesSection({
  activeReservationTypes,
  allReservationTypes,
  capableIdSet,
  isNew,
  onToggle,
}: StaffExcludedReservationTypesSectionProps) {
  const grouped = activeReservationTypes.reduce<Map<string, ReservationType[]>>(
    (acc, reservationType) => {
      const label = reservationCategoryLabel(reservationType.category);
      const group = acc.get(label) ?? [];
      group.push(reservationType);
      acc.set(label, group);
      return acc;
    },
    new Map(),
  );

  return (
    <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
      <div className="flex items-center gap-1.5 mb-2">
        <CheckCircle2 className={`${ICON.xs} ${C.text50}`} />
        <p className={`text-xs font-medium ${C.text50}`}>対応可能コース</p>
      </div>

      {isNew ? (
        <p className={`text-xs ${C.text50} pl-0.5`}>
          スタッフ登録後に設定できます
        </p>
      ) : allReservationTypes.length === 0 ? (
        <p className={`text-xs ${C.text50} pl-0.5`}>
          予約区分が登録されていません
        </p>
      ) : (
        <div className="space-y-3">
          {Array.from(grouped.entries()).map(([label, reservationTypes]) => (
            <div key={label} className="space-y-0.5">
              <p className={`text-[11px] font-medium ${C.text40} px-0.5`}>
                {label}
              </p>
              {reservationTypes.map((reservationType) => (
                <label
                  key={reservationType.id}
                  className={`flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer ${C.hoverBgLight} transition-colors`}
                >
                  <Checkbox
                    checked={capableIdSet.has(reservationType.id)}
                    onCheckedChange={(checked) =>
                      onToggle(reservationType.id, checked === true)
                    }
                  />
                  <span
                    className={`${ICON.dotMd} rounded-full shrink-0`}
                    style={{ backgroundColor: reservationType.color }}
                  />
                  <span className="text-sm">{reservationType.name}</span>
                </label>
              ))}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

interface StaffClinicsSectionProps {
  allClinics: ClinicSummary[];
  clinicIdSet: Set<string>;
  isNew: boolean;
  onToggle: (clinicId: string, checked: boolean) => void;
}

export function StaffClinicsSection({ allClinics, clinicIdSet, isNew, onToggle }: StaffClinicsSectionProps) {
  return (
    <StaffCheckboxSection
      title="所属医院"
      icon={<Building2 className={`${ICON.xs} ${C.text50}`} />}
      items={allClinics}
      selectedIdSet={clinicIdSet}
      isNew={isNew}
      newMessage="スタッフ登録後に所属医院を設定できます"
      isEmpty={allClinics.length === 0}
      emptyMessage="医院が登録されていません"
      onToggle={onToggle}
    />
  );
}

interface StaffPermissionGroupsSectionProps {
  allGroups: PermissionGroup[];
  groupIdSet: Set<string>;
  isNew: boolean;
  onToggle: (groupId: string, checked: boolean) => void;
}

export function StaffPermissionGroupsSection({ allGroups, groupIdSet, isNew, onToggle }: StaffPermissionGroupsSectionProps) {
  return (
    <StaffCheckboxSection
      title="権限グループ"
      icon={<Shield className={`${ICON.xs} ${C.text50}`} />}
      items={allGroups}
      selectedIdSet={groupIdSet}
      isNew={isNew}
      newMessage="スタッフ登録後に権限グループを設定できます"
      isEmpty={allGroups.length === 0}
      emptyMessage="権限グループが登録されていません"
      onToggle={onToggle}
      renderLeading={(group) => (
        <div
          className={`${ICON.dotMd} rounded-full flex-shrink-0`}
          style={{ backgroundColor: group.color ?? PALETTE.defaultGray }}
        />
      )}
    />
  );
}
