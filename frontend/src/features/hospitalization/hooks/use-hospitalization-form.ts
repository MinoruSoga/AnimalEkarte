// React/Framework
import { useState, useActionState, useCallback, useLayoutEffect, useRef } from "react";
import { useSearchParams } from "react-router";

// External
import { toast } from "sonner";

// Internal
import { handleApiError } from "@/lib/handle-api-error";
import {
  isNonDisclosureReadStatus,
  resolveEntityReadResult,
  type EntityReadResult,
} from "@/lib/entity-read-result";
import { INITIAL_ACTION_STATE, type ActionState } from "@/types/form";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";

// Relative
import { createHospitalization } from "../api/create-hospitalization";
import { updateHospitalization } from "../api/update-hospitalization";
import { useGetHospitalizationRaw } from "../api/get-hospitalization";
import { useGetTreatmentPlans } from "../api/get-treatment-plans";
import {
  buildCreateHospitalizationRequest,
  buildUpdateHospitalizationRequest,
  calculateHospitalizationBillingTotals,
  createEmptyTreatmentPlan,
  createInitialHospitalizationFormData,
  hospitalizationSubmitFieldErrors,
  mergePetIntoHospitalizationFormData,
  updateTreatmentPlanField,
} from "./use-hospitalization-form-model";
import { useHospitalizationFormHydration } from "./use-hospitalization-form-helpers";

// Types
import type { HospitalizationTreatmentPlan } from "@/types";
import type { HospitalizationFormData } from "../types";
import type { BackendHospitalization } from "../api/types";

export function useHospitalizationForm(id?: string, canSubmit = false) {
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  const petSelection = usePetSelection();
  const { selectedPets, setSelectedPets } = petSelection;
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");
  const [formData, setFormData] = useState<HospitalizationFormData>(
    createInitialHospitalizationFormData,
  );
  // rerender-memo: functional setState で deps 空にし、呼び出し側 memo コンポーネントへ
  // 安定参照として渡せるようにする。
  const handleFormDataChange = useCallback((updates: Partial<HospitalizationFormData>) => {
    setFormData((prev) => ({ ...prev, ...updates }));
  }, []);
  const [treatmentPlans, setTreatmentPlans] = useState<HospitalizationTreatmentPlan[]>([]);

  const selectedPetRef = useRef(selectedPets[0]);
  useLayoutEffect(() => {
    selectedPetRef.current = selectedPets[0];
  }, [selectedPets]);
  const formDataRef = useRef(formData);
  useLayoutEffect(() => {
    formDataRef.current = formData;
  }, [formData]);
  const treatmentPlansRef = useRef(treatmentPlans);
  useLayoutEffect(() => {
    treatmentPlansRef.current = treatmentPlans;
  }, [treatmentPlans]);
  const canSubmitRef = useRef(canSubmit);
  useLayoutEffect(() => {
    canSubmitRef.current = canSubmit;
  }, [canSubmit]);
  const isMutationAllowed = useCallback(() => canSubmitRef.current === true, []);

  const {
    data: hospitalizationData,
    isLoading,
    isError,
    error: hospitalizationError,
    refetch: refetchHospitalization,
  } = useGetHospitalizationRaw(id);
  const entityRead: EntityReadResult<BackendHospitalization> = resolveEntityReadResult({
    id: isEdit ? id : undefined,
    data: hospitalizationData,
    isLoading,
    isError,
    error: hospitalizationError,
    refetch: refetchHospitalization,
  });
  const entityReadRef = useRef(entityRead);
  useLayoutEffect(() => {
    entityReadRef.current = entityRead;
  }, [entityRead]);

  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      const pet = selectedPetRef.current;
      const latestFormData = formDataRef.current;
      const fieldErrors = hospitalizationSubmitFieldErrors(pet, isEdit, latestFormData.cageId);
      if (fieldErrors) {
        return { success: false, fieldErrors, timestamp: Date.now() };
      }
      if (!pet) {
        return { success: false, timestamp: Date.now() };
      }

      try {
        if (isEdit && id) {
          if (entityReadRef.current.status !== "found") {
            return { success: false, timestamp: Date.now() };
          }
          if (!isMutationAllowed()) return { success: false, timestamp: Date.now() };
          await updateHospitalization(id, buildUpdateHospitalizationRequest(latestFormData));
          toast.success("入院情報を更新しました");
        } else {
          if (!isMutationAllowed()) return { success: false, timestamp: Date.now() };
          await createHospitalization(
            buildCreateHospitalizationRequest(latestFormData, pet, treatmentPlansRef.current),
          );
          toast.success("入院情報を登録しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "保存");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE,
  );

  const { data: treatmentPlanWire, isSuccess: isTreatmentPlansSuccess } = useGetTreatmentPlans(
    isEdit ? id : undefined,
  );

  useHospitalizationFormHydration({
    isEdit,
    id,
    petId,
    isPetLoading,
    petFromQuery,
    entityRead,
    treatmentPlanWire,
    isTreatmentPlansSuccess,
    selectedPetRef,
    setFormData,
    setSelectedPets,
    setTreatmentPlans,
  });

  const addTreatmentPlan = useCallback(() => {
    setTreatmentPlans((prev) => [...prev, createEmptyTreatmentPlan()]);
  }, []);
  const removeTreatmentPlan = useCallback((planId: string) => {
    setTreatmentPlans((prev) => prev.filter((plan) => plan.id !== planId));
  }, []);
  const updateTreatmentPlan = useCallback(
    (
      planId: string,
      field: keyof HospitalizationTreatmentPlan,
      value: string | number | boolean,
    ) => {
      setTreatmentPlans((prev) =>
        prev.map((plan) =>
          plan.id === planId ? updateTreatmentPlanField(plan, field, value) : plan,
        ),
      );
    },
    [],
  );

  return {
    isEdit,
    isLoading: entityRead.status === "loading",
    isError: entityRead.status === "error",
    entityRead,
    isReadLoading: entityRead.status === "loading",
    isReadNotFound: isNonDisclosureReadStatus(entityRead.status),
    isReadError: entityRead.status === "error",
    retryRead: entityRead.status === "error" ? entityRead.retry : undefined,
    isSaving: isPending,
    formData: mergePetIntoHospitalizationFormData(formData, selectedPets),
    setFormData,
    treatmentPlans,
    addTreatmentPlan,
    removeTreatmentPlan,
    updateTreatmentPlan,
    calculateTotals: () => calculateHospitalizationBillingTotals(treatmentPlans),
    petSelection,
    formAction,
    formState,
    handleFormDataChange,
  };
}
