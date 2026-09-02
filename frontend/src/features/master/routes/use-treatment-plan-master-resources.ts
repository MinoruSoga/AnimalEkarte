import { useMemo } from "react";

import type { TreatmentItem } from "@/lib/transforms/treatment";
import { useGetAllConsultations, useCreateConsultation, useUpdateConsultation, useDeleteConsultation, useReorderConsultations } from "../api/consultations";
import { useGetAllExaminationTypes, useCreateExaminationType, useUpdateExaminationType, useDeleteExaminationType, useReorderExaminationTypes } from "../api/exam-types-master";
import { useGetAllProcedures, useCreateProcedure, useUpdateProcedure, useDeleteProcedure, useReorderProcedures } from "../api/procedures";
import { useGetAllVaccinesMaster, useCreateVaccineMaster, useUpdateVaccineMaster, useDeleteVaccineMaster, useReorderVaccinesMaster } from "../api/vaccines-master";
import { useGetAllCheckupTypes, useCreateCheckupType, useUpdateCheckupType, useDeleteCheckupType, useReorderCheckupTypes } from "../api/checkup-types";
import {
  buildTreatmentTabConfigs,
  type TreatmentPlanTabValue,
} from "./treatment-plan-master-model";

interface UseTreatmentPlanMasterResourcesOptions {
  canEdit: boolean;
  canEditCheckup: boolean;
  activeTab: TreatmentPlanTabValue;
  editTarget: TreatmentItem | "new" | null;
}

export function useTreatmentPlanMasterResources({
  canEdit,
  canEditCheckup,
  activeTab,
  editTarget,
}: UseTreatmentPlanMasterResourcesOptions) {
  const { data: consultationData } = useGetAllConsultations();
  const createConsultation = useCreateConsultation();
  const updateConsultation = useUpdateConsultation();
  const deleteConsultation = useDeleteConsultation();
  const reorderConsultations = useReorderConsultations();

  const { data: examinationData } = useGetAllExaminationTypes();
  const createExamination = useCreateExaminationType();
  const updateExamination = useUpdateExaminationType();
  const deleteExamination = useDeleteExaminationType();
  const reorderExaminations = useReorderExaminationTypes();

  const { data: procedureData } = useGetAllProcedures();
  const createProcedure = useCreateProcedure();
  const updateProcedure = useUpdateProcedure();
  const deleteProcedure = useDeleteProcedure();
  const reorderProcedures = useReorderProcedures();

  const { data: vaccineData } = useGetAllVaccinesMaster();
  const createVaccine = useCreateVaccineMaster();
  const updateVaccine = useUpdateVaccineMaster();
  const deleteVaccine = useDeleteVaccineMaster();
  const reorderVaccines = useReorderVaccinesMaster();

  const { data: checkupData } = useGetAllCheckupTypes();
  const createCheckup = useCreateCheckupType();
  const updateCheckup = useUpdateCheckupType();
  const deleteCheckup = useDeleteCheckupType();
  const reorderCheckups = useReorderCheckupTypes();

  const tabConfigs = useMemo(() => buildTreatmentTabConfigs({
    consultationData,
    examinationData,
    procedureData,
    vaccineData,
    checkupData,
    onReorderConsultations: (ids) => {
      if (!canEdit) return;
      reorderConsultations.mutate({ ids });
    },
    onReorderExaminations: (ids) => {
      if (!canEdit) return;
      reorderExaminations.mutate({ ids });
    },
    onReorderProcedures: (ids) => {
      if (!canEdit) return;
      reorderProcedures.mutate({ ids });
    },
    onReorderVaccines: (ids) => {
      if (!canEdit) return;
      reorderVaccines.mutate({ ids });
    },
    onReorderCheckups: (ids) => {
      if (!canEditCheckup) return;
      reorderCheckups.mutate({ ids });
    },
  }), [
    canEdit, canEditCheckup,
    consultationData, reorderConsultations,
    examinationData, reorderExaminations,
    procedureData, reorderProcedures,
    vaccineData, reorderVaccines,
    checkupData, reorderCheckups,
  ]);

  const selectedItem = editTarget !== null && editTarget !== "new" ? editTarget : null;
  const selectedExamination = activeTab === "examination" && selectedItem !== null
    ? examinationData?.find((item) => item.id === selectedItem.id)
    : undefined;
  const activeTabData = useMemo(
    () => tabConfigs[activeTab].data ?? [],
    [tabConfigs, activeTab],
  );
  const parentCandidates = useMemo(
    () =>
      activeTabData.filter(
        (item) => !item.parentId && item.isActive && item.id !== selectedItem?.id,
      ),
    [activeTabData, selectedItem?.id],
  );
  const selectedItemId = selectedItem?.id;
  const hasChildren = useMemo(
    () =>
      selectedItemId != null &&
      activeTabData.some((item) => item.parentId === selectedItemId),
    [activeTabData, selectedItemId],
  );

  const deleteMutationByTab = useMemo(() => ({
    consultation: deleteConsultation,
    examination: deleteExamination,
    procedure: deleteProcedure,
    vaccine: deleteVaccine,
    checkup: deleteCheckup,
  }), [deleteConsultation, deleteExamination, deleteProcedure, deleteVaccine, deleteCheckup]);

  return {
    tabConfigs,
    selectedItem,
    selectedExamination,
    parentCandidates,
    hasChildren,
    deleteMutationByTab,
    createConsultation,
    updateConsultation,
    createExamination,
    updateExamination,
    createProcedure,
    updateProcedure,
    createVaccine,
    updateVaccine,
    createCheckup,
    updateCheckup,
  };
}
