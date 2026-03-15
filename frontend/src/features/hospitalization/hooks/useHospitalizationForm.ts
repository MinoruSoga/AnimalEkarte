import { useState, useEffect } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { toast } from "sonner";
import { handleApiError } from "@/lib/handle-api-error";
import { paths } from "@/config/paths";
import type { TreatmentPlan } from "@/types";
import type { Pet } from "@/types";
import type { HospitalizationFormData } from "../types";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { usePetInfo } from "@/hooks/use-pet";
import {
  createHospitalization,
  updateHospitalization,
  useGetHospitalizationRaw,
} from "../api";

export function useHospitalizationForm(id?: string, onSuccess?: () => void) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  const petSelection = usePetSelection();
  const { selectedPets, setSelectedPets } = petSelection;

  // petId が URL にある場合は React Query でフェッチ（cross-feature 直接呼び出しを回避）
  const { pet: petFromQuery, isLoading: isPetLoading } = usePetInfo(petId ?? "");

  // rerender-lazy-state-init: Date.now() は impure。lazy init で初回レンダーのみ実行
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
    endDate: new Date(Date.now() + 7 * 86400000).toISOString().split("T")[0],
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

  const {
    data: hospitalizationData,
    isLoading,
    isError,
  } = useGetHospitalizationRaw(id);

  useEffect(() => {
    if (!hospitalizationData) return;
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 非同期サーバーデータでフォームを初期化するパターン。React 18 が自動バッチするため実害なし
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
        : new Date(Date.now() + 7 * 86400000).toISOString().split("T")[0],
      memo: hospitalizationData.memo ?? "",
      ownerRequest: hospitalizationData.owner_request ?? "",
      staffNotes: hospitalizationData.staff_notes ?? "",
    }));
    // ペット情報を復元
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
      toast.error("入院情報の取得に失敗しました");
    }
  }, [isError]);

  // petId が URL から来た場合: usePetInfo の結果を selectedPets に反映
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

    try {
      if (isEdit && id) {
        await updateHospitalization(id, {
          type: formData.hospitalizationType === "入院" ? "hospitalization" : "hotel",
          owner_request: formData.ownerRequest,
          staff_notes: formData.staffNotes,
          memo: formData.memo,
          cage_id: formData.cageId || undefined,
        });
        toast.success("入院情報を更新しました");
      } else {
        const startISO = (formData.displayDate || today) + "T00:00:00Z";
        const endISO = (formData.endDate || new Date(Date.now() + 7 * 86400000).toISOString().split("T")[0]) + "T00:00:00Z";
        await createHospitalization({
          pet_id: pet.id,
          owner_id: pet.ownerId || "",
          type: formData.hospitalizationType === "入院" ? "hospitalization" : "hotel",
          start_date: startISO,
          end_date: endISO,
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
            navigate(paths.hospitalization.getHref());
          }
        }, 500);
      }
    } catch (error) {
      handleApiError(error, "保存");
    }
  };

  return {
    isEdit,
    isLoading,
    isError,
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
