import { useState, useEffect, useActionState } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";
import type { TreatmentPlan, Pet } from "@/types";
import type { HospitalizationFormData } from "../types";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
import { createHospitalization } from "../api/create-hospitalization";
import { updateHospitalization } from "../api/update-hospitalization";
import { useGetHospitalizationRaw } from "../api/get-hospitalization";
import { calculateBillingTotals } from "@/lib/calculations";

const MS_PER_DAY = 86_400_000;
const DEFAULT_HOSPITALIZATION_DAYS = 7;

export function useHospitalizationForm(id?: string, _onSuccess?: () => void) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  const petSelection = usePetSelection();
  const { selectedPets, setSelectedPets } = petSelection;

  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");

  const [formData, setFormData] = useState<HospitalizationFormData>(() => ({
    hospitalizationType: "入院",
    ownerName: "",
    species: "",
    petName: "",
    petNumber: "",
    petInsurance: "",
    petDetails: "",
    visit: "診療",
    nextVisit: "",
    weight: "",
    displayDate: "",
    endDate: new Date(Date.now() + DEFAULT_HOSPITALIZATION_DAYS * MS_PER_DAY).toISOString().split("T")[0],
    memo: "",
    ownerRequest: "",
    staffNotes: "",
    cageId: "",
  }));

  const handleFormDataChange = (updates: Partial<HospitalizationFormData>) => {
    setFormData((prev) => ({ ...prev, ...updates }));
  };

  const [treatmentPlans, setTreatmentPlans] = useState<TreatmentPlan[]>([
    {
      id: "1",
      treatmentContent: "adm rate",
      memo: "入院料1日分",
      insurance: true,
      unitPrice: 990,
      quantity: 1,
      discount: 0,
      discountAmount: 0,
      subtotal: 990,
    },
    {
      id: "2",
      treatmentContent: "PCG/SC ~15kg",
      memo: "",
      insurance: false,
      unitPrice: 990,
      quantity: 1,
      discount: 0,
      discountAmount: 0,
      subtotal: 990,
    },
  ]);

  const [globalDiscount, setGlobalDiscount] = useState(0);
  const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);

  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      if (!selectedPets.length) {
        toast.error("ペットを選択してください");
        return { success: false, timestamp: Date.now() };
      }

      const pet = selectedPets[0];
      const today = new Date().toISOString().split("T")[0];

      try {
        if (isEdit && id) {
          await updateHospitalization(id, {
            hospitalization_type: formData.hospitalizationType === "入院" ? "hospitalization" : "hotel",
            owner_request: formData.ownerRequest,
            staff_notes: formData.staffNotes,
            memo: formData.memo,
            cage_id: formData.cageId || undefined,
          });
          toast.success("入院情報を更新しました");
        } else {
          const startISO = (formData.displayDate || today) + "T00:00:00Z";
          const endISO = (formData.endDate || new Date(Date.now() + DEFAULT_HOSPITALIZATION_DAYS * MS_PER_DAY).toISOString().split("T")[0]) + "T00:00:00Z";
          await createHospitalization({
            pet_id: pet.id,
            owner_id: pet.ownerId || "",
            hospitalization_type: formData.hospitalizationType === "入院" ? "hospitalization" : "hotel",
            start_date: startISO,
            end_date: endISO,
            owner_request: formData.ownerRequest,
            staff_notes: formData.staffNotes,
            memo: formData.memo,
            cage_id: formData.cageId || undefined,
          });
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

  useEffect(() => {
    if (!hospitalizationData) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- intentional: initialize form from API data once loaded; functional setState avoids stale closure
    setFormData((prev) => ({
      ...prev,
      hospitalizationType:
        hospitalizationData.hospitalization_type === "hospitalization"
          ? "入院"
          : "ホテル",
      cageId: hospitalizationData.cage_id
        ? String(hospitalizationData.cage_id)
        : "",
      displayDate: hospitalizationData.start_date,
      endDate: hospitalizationData.end_date
        ? hospitalizationData.end_date.split("T")[0]
        : new Date(Date.now() + DEFAULT_HOSPITALIZATION_DAYS * MS_PER_DAY).toISOString().split("T")[0],
      memo: hospitalizationData.memo ?? "",
      ownerRequest: hospitalizationData.owner_request ?? "",
      staffNotes: hospitalizationData.staff_notes ?? "",
    }));
    if (hospitalizationData.pet && hospitalizationData.owner_id) {
      setSelectedPets([
        {
          id: String(hospitalizationData.pet_id),
          ownerId: String(hospitalizationData.owner_id),
          ownerName: hospitalizationData.owner?.owner_name ?? "",
          name: hospitalizationData.pet.name,
          species: hospitalizationData.pet.animal_species?.name ?? "",
          breed: hospitalizationData.pet.breed,
          gender: hospitalizationData.pet.gender,
        } as Pet,
      ]);
    }
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

  const formDataWithPet =
    selectedPets.length > 0
      ? {
          ...formData,
          ownerName: selectedPets[0].ownerName,
          petName: selectedPets[0].name,
          petNumber: selectedPets[0].id,
          species: selectedPets[0].species,
          weight: selectedPets[0].weight ? `${selectedPets[0].weight}kg` : "",
        }
      : formData;

  const addTreatmentPlan = () => {
    const newPlan: TreatmentPlan = {
      id: Date.now().toString(),
      treatmentContent: "",
      memo: "",
      insurance: false,
      unitPrice: 0,
      quantity: 1,
      discount: 0,
      discountAmount: 0,
      subtotal: 0,
    };
    setTreatmentPlans([...treatmentPlans, newPlan]);
  };

  const removeTreatmentPlan = (planId: string) => {
    setTreatmentPlans(treatmentPlans.filter((plan) => plan.id !== planId));
  };

  const updateTreatmentPlan = (
    planId: string,
    field: keyof TreatmentPlan,
    value: string | number | boolean
  ) => {
    setTreatmentPlans(
      treatmentPlans.map((plan) => {
        if (plan.id === planId) {
          const updated = { ...plan, [field]: value };
          if (
            field === "unitPrice" ||
            field === "quantity" ||
            field === "discount"
          ) {
            const unitPrice = (
              field === "unitPrice" ? value : plan.unitPrice
            ) as number;
            const quantity = (
              field === "quantity" ? value : plan.quantity
            ) as number;
            const discount = (
              field === "discount" ? value : plan.discount
            ) as number;
            const baseAmount = unitPrice * quantity;
            updated.discountAmount = Math.floor(baseAmount * (discount / 100));
            updated.subtotal = baseAmount - updated.discountAmount;
          }
          return updated;
        }
        return plan;
      })
    );
  };

  const calculateTotals = () => {
    const result = calculateBillingTotals(treatmentPlans, globalDiscount, globalDiscountAmount);
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
