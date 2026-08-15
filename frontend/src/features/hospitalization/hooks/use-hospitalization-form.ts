import { useState, useEffect, useActionState, useCallback, useLayoutEffect, useRef } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import {
  isNonDisclosureReadStatus,
  resolveEntityReadResult,
  type EntityReadResult,
} from "@/lib/entity-read-result";
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
import type { BackendHospitalization } from "../api/types";
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

  const {
    data: hospitalizationData,
    isLoading,
    isError,
    error: hospitalizationError,
    refetch: refetchHospitalization,
  } = useGetHospitalizationRaw(id);

  // BUG-016: classify read failures; never fold into blank editable form
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

      // BUG-037: cage is required on create/update; care plan stays on detail screen (product SoT).
      const latestFormData = formDataRef.current;
      const cageId = latestFormData.cageId?.trim() ?? "";
      if (!cageId) {
        return {
          success: false,
          fieldErrors: { cage_id: "ケージ・個室を選択してください" },
          timestamp: Date.now(),
        };
      }

      try {
        if (isEdit && id) {
          if (entityReadRef.current.status !== "found") {
            return { success: false, timestamp: Date.now() };
          }
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

  // Real wire: GET /hospitalizations/:id/treatment-plans (not embedded on detail).
  const {
    data: treatmentPlanWire,
    isSuccess: isTreatmentPlansSuccess,
  } = useGetTreatmentPlans(isEdit ? id : undefined);

  const hydratedHospitalizationId = useRef<number | undefined>(undefined);
  useEffect(() => {
    // Only hydrate form model when the entity is actually found
    if (entityRead.status !== "found") return;
    const record = entityRead.data;
    if (hydratedHospitalizationId.current === record.id) return;
    hydratedHospitalizationId.current = record.id;
    setFormData((prev) => buildHospitalizationFormDataFromRecord(prev, record));
    const selectedPet = buildSelectedPetFromHospitalization(record);
    if (selectedPet) {
      selectedPetRef.current = selectedPet;
      setSelectedPets([selectedPet]);
    }
  }, [entityRead, setSelectedPets]);

  const hydratedTreatmentPlansForId = useRef<string | undefined>(undefined);
  useEffect(() => {
    if (!isEdit || !id || !isTreatmentPlansSuccess || treatmentPlanWire === undefined) return;
    if (entityRead.status !== "found") return;
    if (hydratedTreatmentPlansForId.current === id) return;
    hydratedTreatmentPlansForId.current = id;
    setTreatmentPlans(buildTreatmentPlansFromRecord(treatmentPlanWire));
  }, [isEdit, id, isTreatmentPlansSuccess, treatmentPlanWire, entityRead.status]);

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
    isLoading: entityRead.status === "loading",
    isError: entityRead.status === "error",
    entityRead,
    isReadLoading: entityRead.status === "loading",
    isReadNotFound: isNonDisclosureReadStatus(entityRead.status),
    isReadError: entityRead.status === "error",
    retryRead: entityRead.status === "error" ? entityRead.retry : undefined,
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
