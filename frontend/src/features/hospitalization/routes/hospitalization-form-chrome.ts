import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router";

import { paths } from "@/config/paths";
import { useGetMasterItems } from "@/hooks/use-master-items";
import { useGetStaffs } from "@/hooks/use-staffs";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { useGetHospitalizations } from "../api/get-hospitalizations";
import {
  selectHospitalizationDoctorStaffs,
  toHospitalizationHistoryItems,
} from "./hospitalization-form-model";
import type { MasterSelectItem } from "@/components/shared/MasterSelectModal";
import type { HospitalizationTreatmentPlan } from "@/types";
import type { HospitalizationFormData } from "../types";
import type { HospitalizationFormFieldsProps } from "./HospitalizationFormFields";

export function useHospitalizationFormChrome(input: {
  hospitalizationId: string | undefined;
  petId: string | null;
  locationFrom: string | undefined;
  isEdit: boolean;
  formData: HospitalizationFormData;
  formState: { success?: boolean; timestamp?: number; fieldErrors?: Record<string, string> };
  handleFormDataChangeRaw: (updates: Partial<HospitalizationFormData>) => void;
  calculateTotals: () => HospitalizationFormFieldsProps["totals"];
  selectedPet: HospitalizationFormFieldsProps["selectedPet"];
  canDelete: boolean | undefined;
  treatmentPlans: HospitalizationTreatmentPlan[];
}) {
  const navigate = useNavigate();
  const { data: cageItems } = useGetMasterItems("cage");
  const { data: staffs = [] } = useGetStaffs();
  const [staffModalOpen, setStaffModalOpen] = useState(false);
  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  useEffect(() => {
    if (input.formState.success) {
      markClean();
      navigate(input.locationFrom || paths.hospitalization.getHref());
    }
  }, [input.formState.success, input.formState.timestamp, navigate, markClean, input.locationFrom]);

  useEffect(() => {
    if (input.formState.fieldErrors && Object.keys(input.formState.fieldErrors).length > 0) {
      const firstErrorKey = Object.keys(input.formState.fieldErrors)[0];
      const element = document.getElementById(firstErrorKey);
      if (element) {
        element.focus();
        element.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }
  }, [input.formState.fieldErrors, input.formState.timestamp]);

  const totals = input.calculateTotals();
  const historyPetId = input.selectedPet?.id ?? input.petId ?? "";
  const { data: hospitalizationsResult, isLoading: isHistoryLoading } = useGetHospitalizations({
    petId: historyPetId || undefined,
    page: 1,
    limit: 100,
    statusFilter: "all",
  });
  const historyItems = useMemo(() => {
    if (!historyPetId) return [];
    return toHospitalizationHistoryItems(hospitalizationsResult?.data ?? []);
  }, [historyPetId, hospitalizationsResult?.data]);

  const handleBack = useCallback(() => {
    navigate(input.locationFrom || paths.hospitalization.getHref());
  }, [input.locationFrom, navigate]);

  const handleFormDataChangeRaw = input.handleFormDataChangeRaw;
  const handleFormChange = useCallback((updates: Partial<HospitalizationFormData>) => {
    markDirty();
    handleFormDataChangeRaw(updates);
  }, [markDirty, handleFormDataChangeRaw]);

  const doctorStaffItems = useMemo(
    () => selectHospitalizationDoctorStaffs(staffs, input.formData.doctorId),
    [staffs, input.formData.doctorId],
  );
  const handleSelectDoctor = useCallback((item: MasterSelectItem) => {
    handleFormChange({ doctorId: String(item.id), doctorName: item.name });
  }, [handleFormChange]);

  const hasChildTreatmentPlans = input.isEdit && input.treatmentPlans.length > 0;
  const canShowDelete = input.canDelete === true && !hasChildTreatmentPlans;

  useEffect(() => {
    if (!input.selectedPet && !input.isEdit && !input.petId) {
      navigate(paths.hospitalization.selectPet.getHref());
    }
  }, [input.selectedPet, input.isEdit, navigate, input.petId]);

  return {
    cageItems,
    staffModalOpen,
    setStaffModalOpen,
    isDirty,
    totals,
    isHistoryLoading,
    historyItems,
    handleBack,
    handleFormChange,
    doctorStaffItems,
    handleSelectDoctor,
    hasChildTreatmentPlans,
    canShowDelete,
  };
}
