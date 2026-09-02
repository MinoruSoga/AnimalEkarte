import { useCallback, useMemo, type TransitionStartFunction } from "react";

import { useMasterSave } from "../hooks/use-master-save";
import type { TreatmentFormData } from "../components/TreatmentItemSidePanel";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import type { CreateConsultationRequest, UpdateConsultationRequest } from "@/types/treatment";
import type { CreateExaminationTypeRequest as CreateExaminationRequest, UpdateExaminationTypeRequest as UpdateExaminationRequest } from "@/types/treatment";
import type { CreateProcedureRequest, UpdateProcedureRequest } from "@/types/treatment";
import type { CreateVaccineRequest as CreateVaccineMasterRequest, UpdateVaccineRequest as UpdateVaccineMasterRequest } from "@/types/treatment";
import type { CreateCheckupTypeRequest, UpdateCheckupTypeRequest } from "@/types/treatment";
import {
  buildCheckupCreateRequest,
  buildCheckupUpdateRequest,
  buildConsultationCreateRequest,
  buildConsultationUpdateRequest,
  buildExaminationCreateRequest,
  buildExaminationUpdateRequest,
  buildProcedureCreateRequest,
  buildProcedureUpdateRequest,
  buildVaccineCreateRequest,
  buildVaccineUpdateRequest,
  type TreatmentPlanTabValue,
} from "./treatment-plan-master-model";
import type { useTreatmentPlanMasterResources } from "./use-treatment-plan-master-resources";

type TreatmentPlanResources = ReturnType<typeof useTreatmentPlanMasterResources>;

interface UseTreatmentPlanMasterSavesOptions {
  editTarget: TreatmentItem | "new" | null;
  setEditTarget: (target: TreatmentItem | "new" | null) => void;
  startSaveTransition: TransitionStartFunction;
  canCreate: boolean;
  canEdit: boolean;
  canCreateCheckup: boolean;
  canEditCheckup: boolean;
  activeTab: TreatmentPlanTabValue;
  resources: TreatmentPlanResources;
}

export function useTreatmentPlanMasterSaves({
  editTarget,
  setEditTarget,
  startSaveTransition,
  canCreate,
  canEdit,
  canCreateCheckup,
  canEditCheckup,
  activeTab,
  resources,
}: UseTreatmentPlanMasterSavesOptions) {
  const minimalCrud = useMemo(() => ({
    editTarget,
    setEditTarget,
    startSaveTransition,
  }), [editTarget, setEditTarget, startSaveTransition]);

  const consultationSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateConsultationRequest, UpdateConsultationRequest>({
    crud: minimalCrud,
    createMutation: resources.createConsultation,
    updateMutation: resources.updateConsultation,
    permissions: { canCreate, canEdit },
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildConsultationCreateRequest,
    toUpdateRequest: buildConsultationUpdateRequest,
  });

  const examinationSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateExaminationRequest, UpdateExaminationRequest>({
    crud: minimalCrud,
    createMutation: resources.createExamination,
    updateMutation: resources.updateExamination,
    permissions: { canCreate, canEdit },
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildExaminationCreateRequest,
    toUpdateRequest: buildExaminationUpdateRequest,
  });

  const procedureSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateProcedureRequest, UpdateProcedureRequest>({
    crud: minimalCrud,
    createMutation: resources.createProcedure,
    updateMutation: resources.updateProcedure,
    permissions: { canCreate, canEdit },
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildProcedureCreateRequest,
    toUpdateRequest: buildProcedureUpdateRequest,
  });

  const vaccineSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateVaccineMasterRequest, UpdateVaccineMasterRequest>({
    crud: minimalCrud,
    createMutation: resources.createVaccine,
    updateMutation: resources.updateVaccine,
    permissions: { canCreate, canEdit },
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildVaccineCreateRequest,
    toUpdateRequest: buildVaccineUpdateRequest,
  });

  const checkupSave = useMasterSave<TreatmentItem, TreatmentFormData, CreateCheckupTypeRequest, UpdateCheckupTypeRequest>({
    crud: minimalCrud,
    createMutation: resources.createCheckup,
    updateMutation: resources.updateCheckup,
    permissions: { canCreate: canCreateCheckup, canEdit: canEditCheckup },
    validate: (data) => data.name.trim() ? null : "名称を入力してください",
    toCreateRequest: buildCheckupCreateRequest,
    toUpdateRequest: buildCheckupUpdateRequest,
  });

  const saveHooksByTab = useMemo(() => ({
    consultation: consultationSave,
    examination: examinationSave,
    procedure: procedureSave,
    vaccine: vaccineSave,
    checkup: checkupSave,
  }), [consultationSave, examinationSave, procedureSave, vaccineSave, checkupSave]);

  const handleSave = useCallback((data: TreatmentFormData) => {
    saveHooksByTab[activeTab].handleSave(data);
  }, [activeTab, saveHooksByTab]);

  return { handleSave };
}
