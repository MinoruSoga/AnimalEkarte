import { useCallback, useMemo, type TransitionStartFunction } from "react";

import { useMasterSave } from "../hooks/use-master-save";
import type { TreatmentFormData } from "../components/TreatmentItemSidePanel";
import type { TreatmentItem } from "@/lib/transforms/treatment";
import type { CreateConsultationRequest, UpdateConsultationRequest } from "@/types/treatment";
import type {
  CreateExaminationTypeRequest as CreateExaminationRequest,
  UpdateExaminationTypeRequest as UpdateExaminationRequest,
} from "@/types/treatment";
import type { CreateProcedureRequest, UpdateProcedureRequest } from "@/types/treatment";
import type {
  CreateVaccineRequest as CreateVaccineMasterRequest,
  UpdateVaccineRequest as UpdateVaccineMasterRequest,
} from "@/types/treatment";
import type { CreateCheckupTypeRequest, UpdateCheckupTypeRequest } from "@/types/treatment";
import {
  applyItemInsuranceForCreate,
  applyItemInsuranceForUpdate,
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
} from "../routes/treatment-plan-master-model";
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
  const minimalCrud = useMemo(
    () => ({
      editTarget,
      setEditTarget,
      startSaveTransition,
    }),
    [editTarget, setEditTarget, startSaveTransition],
  );

  const consultationSave = useMasterSave<
    TreatmentItem,
    TreatmentFormData,
    CreateConsultationRequest,
    UpdateConsultationRequest
  >({
    crud: minimalCrud,
    createMutation: resources.createConsultation,
    updateMutation: resources.updateConsultation,
    permissions: { canCreate, canEdit },
    validate: (data) => (data.name.trim() ? null : "名称を入力してください"),
    // EMR-454: 診察は is_non_insurance が nullable。保存時だけ保険設定を明示的に足す。
    toCreateRequest: (data) =>
      applyItemInsuranceForCreate(buildConsultationCreateRequest(data), data.isNonInsurance),
    toUpdateRequest: (data) =>
      applyItemInsuranceForUpdate(
        buildConsultationUpdateRequest(data),
        data.insuranceChanged ? data.isNonInsurance : undefined,
      ),
  });

  const examinationSave = useMasterSave<
    TreatmentItem,
    TreatmentFormData,
    CreateExaminationRequest,
    UpdateExaminationRequest
  >({
    crud: minimalCrud,
    createMutation: resources.createExamination,
    updateMutation: resources.updateExamination,
    permissions: { canCreate, canEdit },
    validate: (data) => (data.name.trim() ? null : "名称を入力してください"),
    toCreateRequest: buildExaminationCreateRequest,
    toUpdateRequest: buildExaminationUpdateRequest,
  });

  const procedureSave = useMasterSave<
    TreatmentItem,
    TreatmentFormData,
    CreateProcedureRequest,
    UpdateProcedureRequest
  >({
    crud: minimalCrud,
    createMutation: resources.createProcedure,
    updateMutation: resources.updateProcedure,
    permissions: { canCreate, canEdit },
    validate: (data) => (data.name.trim() ? null : "名称を入力してください"),
    // EMR-454: 処置も診察と同じ三状態規則で保存する（未指定/null は送らない）。
    toCreateRequest: (data) =>
      applyItemInsuranceForCreate(buildProcedureCreateRequest(data), data.isNonInsurance),
    toUpdateRequest: (data) =>
      applyItemInsuranceForUpdate(
        buildProcedureUpdateRequest(data),
        data.insuranceChanged ? data.isNonInsurance : undefined,
      ),
  });

  const vaccineSave = useMasterSave<
    TreatmentItem,
    TreatmentFormData,
    CreateVaccineMasterRequest,
    UpdateVaccineMasterRequest
  >({
    crud: minimalCrud,
    createMutation: resources.createVaccine,
    updateMutation: resources.updateVaccine,
    permissions: { canCreate, canEdit },
    validate: (data) => (data.name.trim() ? null : "名称を入力してください"),
    toCreateRequest: (data) =>
      applyItemInsuranceForCreate(buildVaccineCreateRequest(data), data.isNonInsurance),
    toUpdateRequest: (data) =>
      applyItemInsuranceForUpdate(
        buildVaccineUpdateRequest(data),
        data.insuranceChanged ? data.isNonInsurance : undefined,
      ),
  });

  const checkupSave = useMasterSave<
    TreatmentItem,
    TreatmentFormData,
    CreateCheckupTypeRequest,
    UpdateCheckupTypeRequest
  >({
    crud: minimalCrud,
    createMutation: resources.createCheckup,
    updateMutation: resources.updateCheckup,
    permissions: { canCreate: canCreateCheckup, canEdit: canEditCheckup },
    validate: (data) => {
      if (!data.name.trim()) return "名称を入力してください";
      if (
        editTarget !== "new" &&
        (typeof data.version !== "number" ||
          !Number.isSafeInteger(data.version) ||
          data.version < 1)
      ) {
        return "編集開始時の版を取得できません。再読込してください";
      }
      return null;
    },
    toCreateRequest: (data) =>
      applyItemInsuranceForCreate(buildCheckupCreateRequest(data), data.isNonInsurance),
    toUpdateRequest: (data) =>
      applyItemInsuranceForUpdate(
        buildCheckupUpdateRequest(data),
        data.insuranceChanged ? data.isNonInsurance : undefined,
      ),
  });

  const saveHooksByTab = useMemo(
    () => ({
      consultation: consultationSave,
      examination: examinationSave,
      procedure: procedureSave,
      vaccine: vaccineSave,
      checkup: checkupSave,
    }),
    [consultationSave, examinationSave, procedureSave, vaccineSave, checkupSave],
  );

  const handleSave = useCallback(
    (data: TreatmentFormData) => {
      return saveHooksByTab[activeTab].handleSave(data);
    },
    [activeTab, saveHooksByTab],
  );

  return { handleSave };
}
