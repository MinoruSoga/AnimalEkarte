import { Building2, CheckCircle2, Shield } from "lucide-react";

import { Checkbox } from "@/components/ui/checkbox";
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
              <p className={`text-2xs font-medium ${C.text40} px-0.5`}>
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
      checkedIdSet={clinicIdSet}
      isDisabledUntilSaved={isNew}
      disabledMessage="スタッフ登録後に所属医院を設定できます"
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
      checkedIdSet={groupIdSet}
      isDisabledUntilSaved={isNew}
      disabledMessage="スタッフ登録後に権限グループを設定できます"
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
