import { memo, useCallback, useEffect, useMemo, useState, type ChangeEvent } from "react";
import { UserRound } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { MasterSidePanel, PropertyRow, StatusToggleButton } from "@/components/shared/SidePeek";
import { C, LAYOUT, STYLE } from "@/lib/design-tokens";
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
import { StaffClinicsSection, StaffExcludedReservationTypesSection, StaffLineReservationSection, StaffPermissionGroupsSection } from "./StaffSidePanelSections";

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

      <StaffLineReservationSection
        formData={formData}
        setFormDataDirty={setFormDataDirty}
      />

      <StaffExcludedReservationTypesSection
        activeReservationTypes={activeReservationTypes}
        allReservationTypes={allReservationTypes}
        excludedIdSet={excludedIdSet}
        isNew={isNew}
        onToggle={handleExcludedToggle}
      />

      <StaffClinicsSection
        allClinics={allClinics}
        clinicIdSet={clinicIdSet}
        isNew={isNew}
        onToggle={handleClinicToggle}
      />

      <StaffPermissionGroupsSection
        allGroups={allGroups}
        groupIdSet={groupIdSet}
        isNew={isNew}
        onToggle={handleGroupToggle}
      />
    </MasterSidePanel>
  );
});
