import type { Dispatch, SetStateAction } from "react";
import { Ban, Building2, MessageCircle, Shield } from "lucide-react";

import { Checkbox } from "@/components/ui/checkbox";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { PropertyRow } from "@/components/shared/SidePeek";
import { C, ICON, PALETTE, STYLE } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS } from "../constants/styles";
import type { ClinicSummary } from "../api/staffs";
import type { PermissionGroup } from "../api/permission-groups";
import type { ReservationType } from "../api/reservation-types";
import type { StaffFormData } from "./StaffSidePanel";

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
  excludedIdSet: Set<string>;
  isNew: boolean;
  onToggle: (reservationTypeId: string, checked: boolean) => void;
}

export function StaffExcludedReservationTypesSection({ activeReservationTypes, allReservationTypes, excludedIdSet, isNew, onToggle }: StaffExcludedReservationTypesSectionProps) {
  return (
      <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
        <div className="flex items-center gap-1.5 mb-2">
          <Ban className={`${ICON.xs} ${C.text50}`} />
          <p className={`text-xs font-medium ${C.text50}`}>対応不可コース</p>
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
          <div className="space-y-0.5">
            {activeReservationTypes.map((reservationType) => (
              <label
                key={reservationType.id}
                className={`flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer ${C.hoverBgLight} transition-colors`}
              >
                <Checkbox
                  checked={excludedIdSet.has(reservationType.id)}
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
      <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
        <div className="flex items-center gap-1.5 mb-2">
          <Building2 className={`${ICON.xs} ${C.text50}`} />
          <p className={`text-xs font-medium ${C.text50}`}>所属医院</p>
        </div>

        {isNew ? (
          <p className={`text-xs ${C.text50} pl-0.5`}>
            スタッフ登録後に所属医院を設定できます
          </p>
        ) : allClinics.length === 0 ? (
          <p className={`text-xs ${C.text50} pl-0.5`}>
            医院が登録されていません
          </p>
        ) : (
          <div className="space-y-0.5">
            {allClinics.map((clinic) => (
              <label
                key={clinic.id}
                className={`flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer ${C.hoverBgLight} transition-colors`}
              >
                <Checkbox
                  checked={clinicIdSet.has(clinic.id)}
                  onCheckedChange={(checked) =>
                    onToggle(clinic.id, checked === true)
                  }
                />
                <span className="text-sm">{clinic.name}</span>
              </label>
            ))}
          </div>
        )}
      </div>
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
      <div className={`mt-4 pt-4 ${STYLE.sectionDivider}`}>
        <div className="flex items-center gap-1.5 mb-2">
          <Shield className={`${ICON.xs} ${C.text50}`} />
          <p className={`text-xs font-medium ${C.text50}`}>権限グループ</p>
        </div>

        {isNew ? (
          <p className={`text-xs ${C.text50} pl-0.5`}>
            スタッフ登録後に権限グループを設定できます
          </p>
        ) : allGroups.length === 0 ? (
          <p className={`text-xs ${C.text50} pl-0.5`}>
            権限グループが登録されていません
          </p>
        ) : (
          <div className="space-y-0.5">
            {allGroups.map((group) => (
              <label
                key={group.id}
                className={`flex items-center gap-2.5 py-1.5 px-0.5 rounded cursor-pointer ${C.hoverBgLight} transition-colors`}
              >
                <Checkbox
                  checked={groupIdSet.has(group.id)}
                  onCheckedChange={(checked) =>
                    onToggle(group.id, checked === true)
                  }
                />
                <div
                  className={`${ICON.dotMd} rounded-full flex-shrink-0`}
                  style={{ backgroundColor: group.color ?? PALETTE.defaultGray }}
                />
                <span className="text-sm">{group.name}</span>
              </label>
            ))}
          </div>
        )}
      </div>
  );
}
