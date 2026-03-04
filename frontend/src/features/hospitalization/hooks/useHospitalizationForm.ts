import { useState, useEffect } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { toast } from "sonner";
import type { TreatmentPlan } from "@/types";
import type { HospitalizationFormData } from "../types";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { getPet } from "@/features/pets/api/get-pet";
import { axios } from "@/lib/axios";
import { createHospitalization, updateHospitalization } from "../api";
import type { BackendHospitalization } from "../api/types";

export function useHospitalizationForm(id?: string, onSuccess?: () => void) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  const petSelection = usePetSelection();
  const { selectedPets, setSelectedPets } = petSelection;

  const [formData, setFormData] = useState<HospitalizationFormData>({
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
    memo: "",
    ownerRequest: "",
    staffNotes: "",
    cageId: "",
  });

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

  useEffect(() => {
    if (id) {
      const loadHospitalization = async () => {
        try {
          const { data } = await axios.get<BackendHospitalization>(
            `/v1/hospitalizations/${id}`
          );
          setFormData((prev) => ({
            ...prev,
            hospitalizationType:
              data.type === "入院" ? "入院" : "ホテル",
            cageId: data.cage_id ?? "",
            displayDate: data.start_date,
            memo: data.memo ?? "",
            ownerRequest: data.owner_request ?? "",
            staffNotes: data.staff_notes ?? "",
          }));
          // ペット情報を復元
          if (data.pet && data.owner_id) {
            setSelectedPets([
              {
                id: data.pet_id,
                ownerId: data.owner_id,
                ownerName: data.owner?.name ?? "",
                name: data.pet.name,
                species: data.pet.species,
                breed: data.pet.breed,
                gender: data.pet.gender,
              },
            ]);
          }
        } catch {
          toast.error("入院情報の取得に失敗しました");
        }
      };
      loadHospitalization();
    } else if (petId) {
      // petId が URL から来た場合は Pet API で取得
      const loadPet = async () => {
        try {
          const pet = await getPet(petId);
          setSelectedPets([pet]);
        } catch {
          toast.error("ペット情報の取得に失敗しました");
          navigate("/hospitalization/select-pet");
        }
      };
      loadPet();
    }
  }, [id, petId, setSelectedPets, navigate]);

  // ペット選択情報をフォームデータにマージ
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
    const subtotalBeforeDiscount = treatmentPlans.reduce(
      (sum, plan) => sum + plan.subtotal,
      0
    );
    const discountAmount = globalDiscountAmount;
    const subtotalAfterDiscount = subtotalBeforeDiscount - discountAmount;
    const consumptionTax = Math.floor(subtotalAfterDiscount * 0.1);
    const total = subtotalAfterDiscount + consumptionTax;

    return {
      subtotalBeforeDiscount,
      discountAmount,
      subtotalAfterDiscount,
      consumptionTax,
      total,
    };
  };

  const handleSave = async () => {
    if (!selectedPets.length) {
      toast.error("ペットを選択してください");
      return;
    }

    const pet = selectedPets[0];
    const today = new Date().toISOString().split("T")[0];
    const endDate = new Date(Date.now() + 7 * 86400000)
      .toISOString()
      .split("T")[0];

    try {
      if (isEdit && id) {
        await updateHospitalization(id, {
          type: formData.hospitalizationType === "入院" ? "入院" : "ホテル",
          owner_request: formData.ownerRequest,
          staff_notes: formData.staffNotes,
          memo: formData.memo,
          cage_id: formData.cageId || undefined,
        });
        toast.success("入院情報を更新しました");
      } else {
        await createHospitalization({
          pet_id: pet.id,
          owner_id: pet.ownerId,
          type: formData.hospitalizationType === "入院" ? "入院" : "ホテル",
          start_date: formData.displayDate || today,
          end_date: endDate,
          owner_request: formData.ownerRequest,
          staff_notes: formData.staffNotes,
          memo: formData.memo,
          cage_id: formData.cageId || undefined,
        });
        toast.success("入院情報を登録しました");
      }

      if (onSuccess) {
        onSuccess();
      } else {
        setTimeout(() => {
          if (location.state?.from) {
            navigate(location.state.from);
          } else {
            navigate("/hospitalization");
          }
        }, 500);
      }
    } catch {
      toast.error("保存に失敗しました");
    }
  };

  return {
    isEdit,
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
    handleSave,
    handleFormDataChange,
  };
}
