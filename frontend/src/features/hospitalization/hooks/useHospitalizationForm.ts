import { useState, useEffect } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { toast } from "sonner";
import { TreatmentPlan } from "@/types";
import { HospitalizationFormData } from "../types";
import { MOCK_PETS } from "@/config/mock-data";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { createHospitalization, updateHospitalization, getHospitalization } from "../api";

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
      // Load existing hospitalization
      getHospitalization(id).then(data => {
        const h = data.hospitalization;
        if (h) {
          setFormData(prev => ({
            ...prev,
            hospitalizationType: h.hospitalizationType,
            ownerName: h.ownerName,
            petName: h.petName,
            species: h.species,
            cageId: h.cageId || "",
            displayDate: h.startDate,
          }));

          const pet = MOCK_PETS.find(p => p.name === h.petName && p.ownerName === h.ownerName);
          if (pet) {
            setSelectedPets([pet]);
          }
        }
      }).catch(() => {
        toast.error("入院情報の取得に失敗しました");
      });
    } else {
      if (petId) {
        const foundPet = MOCK_PETS.find(p => p.id === petId);
        if (foundPet) {
          setSelectedPets([foundPet]);
        } else {
          navigate("/hospitalization/select-pet");
        }
      }
    }
  }, [id, petId, setSelectedPets, navigate]);

  // Derive form data with pet info at render time (no setState-in-useEffect)
  const formDataWithPet = selectedPets.length > 0
    ? {
        ...formData,
        ownerName: selectedPets[0].ownerName,
        petName: selectedPets[0].name,
        petNumber: selectedPets[0].id,
        species: selectedPets[0].species,
        weight: `${selectedPets[0].weight}kg`,
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
            const unitPrice = (field === "unitPrice" ? value : plan.unitPrice) as number;
            const quantity = (field === "quantity" ? value : plan.quantity) as number;
            const discount = (field === "discount" ? value : plan.discount) as number;
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
      const today = new Date().toISOString().split('T')[0];
      const endDate = new Date(Date.now() + 7 * 86400000).toISOString().split('T')[0];

      const baseData = {
          hospitalizationType: formData.hospitalizationType as "入院" | "ホテル",
          ownerName: formData.ownerName,
          petName: formData.petName,
          species: formData.species,
          cageId: formData.cageId,
          startDate: formData.displayDate || today,
          endDate: endDate,
      };

      try {
          if (isEdit && id) {
              await updateHospitalization(id, baseData);
              toast.success("入院情報を更新しました");
          } else {
              await createHospitalization(baseData);
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
    handleFormDataChange
  };
}
