import { useState, useEffect, useActionState, useCallback, useRef } from "react";
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
import { calculateBillingTotals } from "@/lib/calculations";
import {
  DEFAULT_TREATMENT_PLANS,
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

export function useHospitalizationForm(id?: string, _onSuccess?: () => void) {
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

  const [treatmentPlans, setTreatmentPlans] = useState<HospitalizationTreatmentPlan[]>(
    isEdit ? [] : DEFAULT_TREATMENT_PLANS,
  );

  const [globalDiscount, setGlobalDiscount] = useState(0);
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);

  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      if (!selectedPets.length) {
        return { success: false, fieldErrors: { pet: "ペットを選択してください" }, timestamp: Date.now() };
      }

      const pet = selectedPets[0];
      try {
        if (isEdit && id) {
          await updateHospitalization(id, buildUpdateHospitalizationRequest(formData));
          toast.success("入院情報を更新しました");
        } else {
          await createHospitalization(buildCreateHospitalizationRequest(formData, pet));
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

  const hydratedHospitalizationId = useRef<number | undefined>(undefined);
  useEffect(() => {
    if (!hospitalizationData || hydratedHospitalizationId.current === hospitalizationData.id) return;
    hydratedHospitalizationId.current = hospitalizationData.id;
    setFormData((prev) => buildHospitalizationFormDataFromRecord(prev, hospitalizationData));
    setTreatmentPlans(buildTreatmentPlansFromRecord(hospitalizationData));
    const selectedPet = buildSelectedPetFromHospitalization(hospitalizationData);
    if (selectedPet) setSelectedPets([selectedPet]);
  }, [hospitalizationData, setSelectedPets]);

  useEffect(() => {
    if (isError) {
      handleApiError(hospitalizationError, "入院情報の取得");
    }
  }, [isError, hospitalizationError]);

  useEffect(() => {
    if (!petId || id) return;
    if (isPetLoading) return;
    if (petFromQuery) {
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
    const result = calculateBillingTotals(billingItems, globalDiscount, globalDiscountAmount);
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
    globalDiscount,
    setGlobalDiscount,
    globalDiscountAmount,
    setGlobalDiscountAmount,
    calculateTotals,
    petSelection,
    formAction,
    formState,
    handleFormDataChange,
  };
}
