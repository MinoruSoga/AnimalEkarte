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
import {
  staffToFormData,
  type StaffFormData,
} from "./StaffSidePanelModel";

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

function useEditableIdSelection({
  serverIds,
  markDirty,
}: {
  serverIds: string[] | undefined;
  markDirty: () => void;
}) {
  const [userEditedIds, setUserEditedIds] = useState<string[] | null>(null);

  const ids = useMemo(
    () => userEditedIds ?? serverIds ?? [],
    [userEditedIds, serverIds],
  );
  const idSet = useMemo(() => new Set(ids), [ids]);

  const handleToggle = useCallback(
    (id: string, checked: boolean) => {
      setUserEditedIds((prev) => {
        const current = prev ?? serverIds ?? [];
        return checked ? [...current, id] : current.filter((currentId) => currentId !== id);
      });
      markDirty();
    },
    [serverIds, markDirty],
  );

  return { ids, idSet, handleToggle };
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

  const [formData, setFormData] = useState<StaffFormData>(() => staffToFormData(item));
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
  const {
    ids: groupIds,
    idSet: groupIdSet,
    handleToggle: handleGroupToggle,
  } = useEditableIdSelection({ serverIds: serverGroupIds, markDirty });

  const { data: serverClinicIds } = useGetStaffClinics(staffId);
  const {
    ids: clinicIds,
    idSet: clinicIdSet,
    handleToggle: handleClinicToggle,
  } = useEditableIdSelection({ serverIds: serverClinicIds, markDirty });

  const { data: serverExcludedIds } = useGetStaffExcludedReservationTypes(staffId);
  const {
    ids: excludedIds,
    idSet: excludedIdSet,
    handleToggle: handleExcludedToggle,
  } = useEditableIdSelection({ serverIds: serverExcludedIds, markDirty });

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
