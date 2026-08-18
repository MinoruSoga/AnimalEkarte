import { memo, useCallback, useEffect, useMemo, useState } from "react";
import { UserRound } from "lucide-react";
import { MasterSidePanel } from "@/components/shared/SidePeek";
import { LAYOUT } from "@/lib/design-tokens";
import {
  useGetStaffCapableReservationTypes,
  useGetStaffClinics,
  useGetStaffPermissionGroups,
  type ClinicSummary,
  type Staff,
} from "../api/staffs";
import type { Occupation } from "../api/occupations";
import type { PermissionGroup } from "../api/permission-groups";
import type { ReservationType } from "../api/reservation-types";
import { StaffBasicInfoSection } from "./StaffBasicInfoSection";
import { StaffLineReservationSection } from "./StaffLineReservationSection";
import { StaffClinicsSection, StaffExcludedReservationTypesSection, StaffPermissionGroupsSection } from "./StaffSidePanelSections";
import {
  staffToFormData,
  type StaffFormData,
} from "./staff-side-panel-model";
import { useEditableIdSelection } from "../hooks/use-editable-id-selection";

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
  onSaveCapableReservationTypes: (staffId: string, reservationTypeIds: string[]) => void;
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
  onSaveCapableReservationTypes,
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

  const { data: serverCapableIds } = useGetStaffCapableReservationTypes(staffId);
  const {
    ids: capableIds,
    idSet: capableIdSet,
    handleToggle: handleCapabilityToggle,
  } = useEditableIdSelection({ serverIds: serverCapableIds, markDirty });

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
      onSaveCapableReservationTypes(staffId, capableIds);
    }
    setIsDirty(false);
  }, [formData, isNew, staffId, groupIds, clinicIds, capableIds, onSave, onSaveGroups, onSaveClinics, onSaveCapableReservationTypes]);

  const handleTitleChange = useCallback((value: string) => {
    setFormDataDirty((prev) => ({ ...prev, name: value }));
    if (value.trim()) setNameError("");
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
      onSave={handleSave}
      onDelete={item !== null && onDeleteRequest ? () => onDeleteRequest(item) : undefined}
      icon={<UserRound className={LAYOUT.pageIcon.innerIcon} />}
      isDirty={isDirty}
      titleError={nameError}
      titleMaxLength={100}
      readOnly={readOnly}
    >
      <StaffBasicInfoSection
        item={item}
        isNew={isNew}
        formData={formData}
        setFormDataDirty={setFormDataDirty}
        allOccupations={allOccupations}
      />

      <StaffLineReservationSection
        formData={formData}
        setFormDataDirty={setFormDataDirty}
      />

      <StaffExcludedReservationTypesSection
        activeReservationTypes={activeReservationTypes}
        allReservationTypes={allReservationTypes}
        capableIdSet={capableIdSet}
        isNew={isNew}
        onToggle={handleCapabilityToggle}
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
