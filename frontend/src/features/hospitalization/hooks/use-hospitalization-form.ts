import { useState, useEffect, useActionState, useCallback, useLayoutEffect, useRef } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
import type { HospitalizationTreatmentPlan } from "@/types";
import type { HospitalizationFormData } from "../types";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
import { createHospitalization } from "../api/create-hospitalization";
import { updateHospitalization } from "../api/update-hospitalization";
import { useGetHospitalizationRaw } from "../api/get-hospitalization";
import { useGetTreatmentPlans } from "../api/get-treatment-plans";
import { calculateBillingTotals } from "@/lib/calculations";
import {
  buildCreateHospitalizationRequest,
  buildHospitalizationFormDataFromRecord,
  buildSelectedPetFromHospitalization,
  buildTreatmentPlansFromRecord,
  buildUpdateHospitalizationRequest,
  createEmptyTreatmentPlan,
  createInitialHospitalizationFormData,
  mergePetIntoHospitalizationFormData,
  updateTreatmentPlanField,
} from "./use-hospitalization-form-model";

export function useHospitalizationForm(id?: string, canSubmit = false) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  const petSelection = usePetSelection();
  const { selectedPets, setSelectedPets } = petSelection;

  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");

  const [formData, setFormData] = useState<HospitalizationFormData>(createInitialHospitalizationFormData);

  const handleFormDataChange = (updates: Partial<HospitalizationFormData>) => {
    setFormData((prev) => ({ ...prev, ...updates }));
  };

  // Create starts empty — only user-entered rows with content are POSTed after parent create.
  // Edit hydrates from GET /hospitalizations/:id/treatment-plans (read-only UI).
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

  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      const pet = selectedPetRef.current;
      if (!pet) {
        return { success: false, fieldErrors: { pet: "ペットを選択してください" }, timestamp: Date.now() };
      }

      if (pet.status === "死亡") {
        return {
          success: false,
          fieldErrors: {
            pet: isEdit
              ? "死亡したペットは入院情報を更新できません"
              : "死亡したペットは入院登録できません",
          },
          timestamp: Date.now(),
        };
      }
      try {
        const latestFormData = formDataRef.current;
        if (isEdit && id) {
          if (!isMutationAllowed()) return { success: false, timestamp: Date.now() };
          // Edit: parent fields only. Treatment plans are create-time snapshots (RO on this screen).
          await updateHospitalization(id, buildUpdateHospitalizationRequest(latestFormData));
          toast.success("入院情報を更新しました");
        } else {
          if (!isMutationAllowed()) return { success: false, timestamp: Date.now() };
          // Single success boundary: parent + nested treatment_plans in one BE TX (TASK-001).
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
    INITIAL_ACTION_STATE
  );

  const {
    data: hospitalizationData,
    isLoading,
    isError,
    error: hospitalizationError,
  } = useGetHospitalizationRaw(id);

  // Real wire: GET /hospitalizations/:id/treatment-plans (not embedded on detail).
  const {
    data: treatmentPlanWire,
    isSuccess: isTreatmentPlansSuccess,
  } = useGetTreatmentPlans(isEdit ? id : undefined);

  const hydratedHospitalizationId = useRef<number | undefined>(undefined);
  useEffect(() => {
    if (!hospitalizationData || hydratedHospitalizationId.current === hospitalizationData.id) return;
    hydratedHospitalizationId.current = hospitalizationData.id;
    setFormData((prev) => buildHospitalizationFormDataFromRecord(prev, hospitalizationData));
    const selectedPet = buildSelectedPetFromHospitalization(hospitalizationData);
    if (selectedPet) {
      selectedPetRef.current = selectedPet;
      setSelectedPets([selectedPet]);
    }
  }, [hospitalizationData, setSelectedPets]);

  const hydratedTreatmentPlansForId = useRef<string | undefined>(undefined);
  useEffect(() => {
    if (!isEdit || !id || !isTreatmentPlansSuccess || treatmentPlanWire === undefined) return;
    if (hydratedTreatmentPlansForId.current === id) return;
    hydratedTreatmentPlansForId.current = id;
    setTreatmentPlans(buildTreatmentPlansFromRecord(treatmentPlanWire));
  }, [isEdit, id, isTreatmentPlansSuccess, treatmentPlanWire]);

  useEffect(() => {
    if (isError) {
      handleApiError(hospitalizationError, "入院情報の取得");
    }
  }, [isError, hospitalizationError]);

  useEffect(() => {
    if (!petId || id) return;
    if (isPetLoading) return;
    if (petFromQuery) {
      selectedPetRef.current = petFromQuery;
      setSelectedPets([petFromQuery]);
    } else {
      toast.error("ペット情報の取得に失敗しました");
      navigate(paths.hospitalization.selectPet.getHref());
    }
  }, [petId, id, petFromQuery, isPetLoading, setSelectedPets, navigate]);

  const formDataWithPet = mergePetIntoHospitalizationFormData(formData, selectedPets);

  // rerender-functional-setstate: prev => 形式で treatmentPlans を deps から除外
  const addTreatmentPlan = useCallback(() => {
    setTreatmentPlans((prev) => [...prev, createEmptyTreatmentPlan()]);
  }, []);

  const removeTreatmentPlan = useCallback((planId: string) => {
    setTreatmentPlans((prev) => prev.filter((plan) => plan.id !== planId));
  }, []);

  const updateTreatmentPlan = useCallback((
    planId: string,
    field: keyof HospitalizationTreatmentPlan,
    value: string | number | boolean
  ) => {
    setTreatmentPlans((prev) =>
      prev.map((plan) => {
        if (plan.id === planId) {
          return updateTreatmentPlanField(plan, field, value);
        }
        return plan;
      })
    );
  }, []);

  const calculateTotals = () => {
    const billingItems = treatmentPlans.map((plan) => ({
      ...plan,
      isInsuranceApplicable: plan.is_insurance,
    }));
    // Bulk discount is not available on this screen (W-003); always pass 0.
    const result = calculateBillingTotals(billingItems, 0, 0);
    return {
      subtotalBeforeDiscount: result.subtotal,
      discountAmount: result.globalDiscountAmount,
      subtotalAfterDiscount: result.taxableAmount,
      consumptionTax: result.tax,
      total: result.total,
    };
  };

  return {
    isEdit,
    isLoading,
    isError,
    isSaving: isPending,
    formData: formDataWithPet,
    setFormData,
    treatmentPlans,
    addTreatmentPlan,
    removeTreatmentPlan,
    updateTreatmentPlan,
    calculateTotals,
    petSelection,
    formAction,
    formState,
    handleFormDataChange,
  };
}
