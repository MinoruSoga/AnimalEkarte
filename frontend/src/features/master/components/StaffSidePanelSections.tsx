import { useState } from "react";
import { Building2, CheckCircle2, Shield } from "lucide-react";

import { Checkbox } from "@/components/ui/checkbox";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { FieldHelp } from "@/components/shared/FieldHelp";
import { C, ICON, PALETTE, STYLE } from "@/lib/design-tokens";
import type { ClinicSummary } from "../api/staffs";
import type { PermissionGroup } from "../api/permission-groups";
import type { ReservationType } from "../api/reservation-types";
import { StaffCheckboxSection } from "./StaffCheckboxSection";

interface StaffExcludedReservationTypesSectionProps {
  activeReservationTypes: ReservationType[];
  allReservationTypes: ReservationType[];
  capableIdSet: Set<string>;
  isNew: boolean;
  onToggle: (reservationTypeId: string, checked: boolean) => void;
}

// 対応区分セクションの表示モード。capable view は「チェック=対応できる」、
// incapable view は「チェック=対応できない」。どちらも編集・保存するのは
// capableIds 1つであり、表示の切替自体は状態を書き換えない（EMR-250）。
type CapabilityView = "capable" | "incapable";

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
  const [view, setView] = useState<CapabilityView>("capable");
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

  const hasEditableList = !isNew && allReservationTypes.length > 0;

  return (
    <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
      <div className="flex items-center gap-1.5 mb-2">
        <CheckCircle2 className={`${ICON.xs} ${C.text50}`} />
        <p className={`text-xs font-medium ${C.text50}`}>対応区分</p>
        <FieldHelp
          label="対応区分"
          content="このスタッフが担当できる予約区分です。「対応可能」表示ではチェックした区分を担当でき、「対応不可」表示ではチェックした区分を担当できません。どちらの表示で編集しても同じ設定として保存されます。"
        />
        {hasEditableList ? (
          <ToggleGroup
            type="single"
            size="sm"
            value={view}
            onValueChange={(next) => {
              if (next === "capable" || next === "incapable") setView(next);
            }}
            aria-label="対応区分の表示切替"
            className="ml-auto"
          >
            <ToggleGroupItem value="capable">対応可能</ToggleGroupItem>
            <ToggleGroupItem value="incapable">対応不可</ToggleGroupItem>
          </ToggleGroup>
        ) : null}
      </div>

      {isNew ? (
        <p className={`text-xs ${C.text50} pl-0.5`}>スタッフ登録後に設定できます</p>
      ) : allReservationTypes.length === 0 ? (
        <p className={`text-xs ${C.text50} pl-0.5`}>予約区分が登録されていません</p>
      ) : (
        <div className="space-y-3">
          {Array.from(grouped.entries()).map(([label, reservationTypes]) => (
            <div key={label} className="space-y-0.5">
              <p className={`text-2xs font-medium ${C.text40} px-0.5`}>{label}</p>
              {reservationTypes.map((reservationType) => {
                const isCapable = capableIdSet.has(reservationType.id);
                return (
                  <label
                    key={reservationType.id}
                    className={`flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer min-h-11 ${C.hoverBgLight} transition-colors`}
                  >
                    <Checkbox
                      checked={view === "capable" ? isCapable : !isCapable}
                      onCheckedChange={(checked) =>
                        onToggle(
                          reservationType.id,
                          view === "capable" ? checked === true : checked !== true,
                        )
                      }
                    />
                    <span
                      className={`${ICON.dotMd} rounded-full shrink-0`}
                      style={{ backgroundColor: reservationType.color }}
                    />
                    <span className="text-sm">{reservationType.name}</span>
                  </label>
                );
              })}
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

export function StaffClinicsSection({
  allClinics,
  clinicIdSet,
  isNew,
  onToggle,
}: StaffClinicsSectionProps) {
  const clinicItems = allClinics
    .filter((clinic) => clinic.isActive || clinicIdSet.has(clinic.id))
    .map((clinic) => ({
      ...clinic,
      name: clinic.isActive ? clinic.name : `${clinic.name}（無効）`,
    }));
  return (
    <StaffCheckboxSection
      title="所属医院"
      icon={<Building2 className={`${ICON.xs} ${C.text50}`} />}
      items={clinicItems}
      checkedIdSet={clinicIdSet}
      isDisabledUntilSaved={isNew}
      disabledMessage="スタッフ登録後に所属医院を設定できます"
      emptyMessage="医院が登録されていません"
      description="このスタッフが所属する医院です。複数医院がある場合に所属先を選択します。"
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

export function StaffPermissionGroupsSection({
  allGroups,
  groupIdSet,
  isNew,
  onToggle,
}: StaffPermissionGroupsSectionProps) {
  return (
    <StaffCheckboxSection
      title="権限グループ"
      icon={<Shield className={`${ICON.xs} ${C.text50}`} />}
      items={allGroups}
      checkedIdSet={groupIdSet}
      isDisabledUntilSaved={isNew}
      disabledMessage="スタッフ登録後に権限グループを設定できます"
      emptyMessage="権限グループが登録されていません"
      description="このスタッフに付与する権限グループです。グループに設定された権限がそのまま適用されます。"
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
