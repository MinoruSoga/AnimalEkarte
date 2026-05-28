import { memo, useCallback, useEffect, useMemo, useState, type ChangeEvent } from "react";
import { Ban, Building2, MessageCircle, Shield, UserRound } from "lucide-react";
import { Checkbox } from "@/components/ui/checkbox";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { MasterSidePanel, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { C, ICON, LAYOUT, PALETTE, STYLE } from "@/lib/design-tokens";
import { MASTER_INPUT_CLASS } from "../constants/styles";
import {
  useGetStaffClinics,
  useGetStaffExcludedReservationTypes,
  useGetStaffPermissionGroups,
  type ClinicSummary,
  type Staff,
} from "../api/staffs";
import type { Occupation } from "../api/occupations";
import type { PermissionGroup } from "../api/permission-groups";
import type { ReservationType } from "../api/reservation-types";

const STAFF_TYPE_SELECT_ITEMS = (
  <>
    <SelectItem value="doctor">医師</SelectItem>
    <SelectItem value="nurse">看護師</SelectItem>
    <SelectItem value="resource">設備</SelectItem>
  </>
);

export interface StaffFormData {
  name: string;
  jobTitleId: string | null;
  licenseNumber: string;
  isActive: boolean;
  email: string;
  password: string;
  staffType: string;
  reservationDisplayName: string;
  reservationVisible: boolean;
  reservationComment: string;
  reservationImageUrl: string;
}

interface StaffSidePanelProps {
  item: Staff | null;
  onClose: () => void;
  onSave: (data: StaffFormData) => void;
  onDeleteRequest?: (item: Staff) => void;
  readOnly?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
  allOccupations: Occupation[];
  allGroups: PermissionGroup[];
  onSaveGroups: (staffId: string, groupIds: string[]) => void;
  allClinics: ClinicSummary[];
  onSaveClinics: (staffId: string, clinicIds: string[]) => void;
  allReservationTypes: ReservationType[];
  onSaveExcludedReservationTypes: (staffId: string, reservationTypeIds: string[]) => void;
}

export const StaffSidePanel = memo(function StaffSidePanel({
  item,
  onClose,
  onSave,
  onDeleteRequest,
  readOnly,
  onDirtyChange,
  allOccupations,
  allGroups,
  onSaveGroups,
  allClinics,
  onSaveClinics,
  allReservationTypes,
  onSaveExcludedReservationTypes,
}: StaffSidePanelProps) {
  const isNew = item === null;
  const staffId = item?.id ?? null;

  const [formData, setFormData] = useState<StaffFormData>(() => ({
    name: item?.name ?? "",
    jobTitleId: item?.occupationId ?? null,
    licenseNumber: item?.licenseNumber ?? "",
    isActive: item?.isActive ?? true,
    email: item?.email ?? "",
    password: "",
    staffType: item?.staffType ?? "doctor",
    reservationDisplayName: item?.reservationDisplayName ?? "",
    reservationVisible: item?.reservationVisible ?? true,
    reservationComment: item?.reservationComment ?? "",
    reservationImageUrl: item?.reservationImageUrl ?? "",
  }));
  const [isDirty, setIsDirty] = useState(false);
  const [nameError, setNameError] = useState("");

  useEffect(() => {
    onDirtyChange?.(isDirty);
  }, [isDirty, onDirtyChange]);

  const markDirty = useCallback(() => setIsDirty(true), []);
  const setFormDataDirty = useCallback<typeof setFormData>((updater) => {
    setFormData(updater);
    markDirty();
  }, [markDirty]);

  const occupationSelectItems = useMemo(
    () =>
      allOccupations
        .filter((occupation) => occupation.isActive)
        .map((occupation) => (
          <SelectItem key={occupation.id} value={occupation.id}>{occupation.name}</SelectItem>
        )),
    [allOccupations],
  );

  const activeReservationTypes = useMemo(
    () => allReservationTypes.filter((reservationType) => reservationType.isActive),
    [allReservationTypes],
  );

  const { data: serverGroupIds } = useGetStaffPermissionGroups(staffId);
  const [userEditedGroupIds, setUserEditedGroupIds] = useState<string[] | null>(null);

  const groupIds = useMemo(
    () => userEditedGroupIds ?? serverGroupIds ?? [],
    [userEditedGroupIds, serverGroupIds],
  );
  const groupIdSet = useMemo(() => new Set(groupIds), [groupIds]);

  const { data: serverClinicIds } = useGetStaffClinics(staffId);
  const [userEditedClinicIds, setUserEditedClinicIds] = useState<string[] | null>(null);

  const clinicIds = useMemo(
    () => userEditedClinicIds ?? serverClinicIds ?? [],
    [userEditedClinicIds, serverClinicIds],
  );
  const clinicIdSet = useMemo(() => new Set(clinicIds), [clinicIds]);

  const { data: serverExcludedIds } = useGetStaffExcludedReservationTypes(staffId);
  const [userEditedExcludedIds, setUserEditedExcludedIds] = useState<string[] | null>(null);

  const excludedIds = useMemo(
    () => userEditedExcludedIds ?? serverExcludedIds ?? [],
    [userEditedExcludedIds, serverExcludedIds],
  );
  const excludedIdSet = useMemo(() => new Set(excludedIds), [excludedIds]);

  const handleExcludedToggle = useCallback(
    (reservationTypeId: string, checked: boolean) => {
      setUserEditedExcludedIds((prev) => {
        const current = prev ?? serverExcludedIds ?? [];
        return checked ? [...current, reservationTypeId] : current.filter((id) => id !== reservationTypeId);
      });
      markDirty();
    },
    [serverExcludedIds, markDirty],
  );

  const handleClinicToggle = useCallback(
    (clinicId: string, checked: boolean) => {
      setUserEditedClinicIds((prev) => {
        const current = prev ?? serverClinicIds ?? [];
        return checked ? [...current, clinicId] : current.filter((id) => id !== clinicId);
      });
      markDirty();
    },
    [serverClinicIds, markDirty],
  );

  const handleGroupToggle = useCallback(
    (groupId: string, checked: boolean) => {
      setUserEditedGroupIds((prev) => {
        const current = prev ?? serverGroupIds ?? [];
        return checked ? [...current, groupId] : current.filter((id) => id !== groupId);
      });
      markDirty();
    },
    [serverGroupIds, markDirty],
  );

  const handleSave = useCallback(() => {
    if (!formData.name.trim()) {
      setNameError("氏名を入力してください");
      return;
    }
    setNameError("");
    onSave(formData);
    if (!isNew && staffId) {
      onSaveGroups(staffId, groupIds);
      onSaveClinics(staffId, clinicIds);
      onSaveExcludedReservationTypes(staffId, excludedIds);
    }
    setIsDirty(false);
  }, [formData, isNew, staffId, groupIds, clinicIds, excludedIds, onSave, onSaveGroups, onSaveClinics, onSaveExcludedReservationTypes]);

  const handleTitleChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: value }));
    if (value.trim()) setNameError("");
  }, [setFormDataDirty]);

  const handleToggleActive = useCallback(() => {
    setFormDataDirty((prev) => ({ ...prev, isActive: !prev.isActive }));
  }, [setFormDataDirty]);

  const handleOccupationChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, jobTitleId: value }));
  }, [setFormDataDirty]);

  const handleLicenseNumberChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, licenseNumber: event.target.value }));
  }, [setFormDataDirty]);

  const handleEmailChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, email: event.target.value }));
  }, [setFormDataDirty]);

  const handlePasswordChange = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    setFormDataDirty((prev) => ({ ...prev, password: event.target.value }));
  }, [setFormDataDirty]);

  const handleClose = useCallback(() => {
    setIsDirty(false);
    onClose();
  }, [onClose]);

  return (
    <MasterSidePanel
      isNew={isNew}
      title={formData.name}
      onTitleChange={handleTitleChange}
      onClose={handleClose}
      action={handleSave}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<UserRound className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StatusToggleButton
        isActive={formData.isActive}
        onToggle={handleToggleActive}
      />

      <PropertyRow label="職種">
        <Select
          value={formData.jobTitleId ?? undefined}
          onValueChange={handleOccupationChange}
        >
          <SelectTrigger className={STYLE.selectCompact}>
            <SelectValue placeholder="選択" />
          </SelectTrigger>
          <SelectContent>{occupationSelectItems}</SelectContent>
        </Select>
      </PropertyRow>

      <PropertyRow label="資格番号">
        <input
          type="text"
          className={MASTER_INPUT_CLASS}
          value={formData.licenseNumber}
          onChange={handleLicenseNumberChange}
          placeholder="空"
        />
      </PropertyRow>

      {isNew ? (
        <>
          <PropertyRow label="メールアドレス">
            <input
              type="email"
              className={MASTER_INPUT_CLASS}
              value={formData.email}
              onChange={handleEmailChange}
              placeholder="例: staff@clinic.com"
            />
          </PropertyRow>
          <PropertyRow label="パスワード">
            <input
              type="password"
              className={MASTER_INPUT_CLASS}
              value={formData.password}
              onChange={handlePasswordChange}
              placeholder="8文字以上"
            />
          </PropertyRow>
        </>
      ) : (
        <>
          <PropertyRow label="メールアドレス">
            <span className={`text-sm ${C.text65}`}>{item?.email || "未設定"}</span>
          </PropertyRow>
          <PropertyRow label="パスワード">
            <input
              type="password"
              className={MASTER_INPUT_CLASS}
              value={formData.password}
              onChange={handlePasswordChange}
              placeholder="変更する場合のみ入力"
            />
          </PropertyRow>
        </>
      )}

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
                    handleExcludedToggle(reservationType.id, checked === true)
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
                    handleClinicToggle(clinic.id, checked === true)
                  }
                />
                <span className="text-sm">{clinic.name}</span>
              </label>
            ))}
          </div>
        )}
      </div>

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
                    handleGroupToggle(group.id, checked === true)
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
    </MasterSidePanel>
  );
});
