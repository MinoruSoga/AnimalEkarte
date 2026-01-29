import { useState, useEffect } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router";
import { toast } from "sonner";
import { Pet } from "@/types";
import { MOCK_PETS, MOCK_MEDICAL_RECORDS } from "@/config/mock-data";
import { TreatmentItem } from "../components/TreatmentTable";

export function useMedicalRecordForm(recordId?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isNewRecord = !recordId;

  const [activeTab, setActiveTab] = useState("問診");
  // Lazy initialization for both edit and new modes
  const [selectedPet, _setSelectedPet] = useState<Pet | null>(() => {
    if (recordId) {
      const record = MOCK_MEDICAL_RECORDS.find(r => r.id === recordId);
      if (record?.petId) {
        const pet = MOCK_PETS.find(p => p.id === record.petId);
        if (pet) return pet;
      }
      // Fallback for mock if link fails
      const mockPet = MOCK_PETS.find(p => p.id === "1");
      if (mockPet) return mockPet;
    } else if (petId) {
      // Initialize with petId for new records
      const pet = MOCK_PETS.find(p => p.id === petId);
      if (pet) return pet;
    }
    return null;
  });

  // Treatment Data States (lazy init for edit mode)
  const [treatmentPlanItems, setTreatmentPlanItems] = useState<TreatmentItem[]>(() => {
    if (recordId) {
      return [
        {
          id: 1,
          selected: false,
          status: "完了",
          content: "recheck(新料金)1",
          memo: "再診料099",
          insurance: true,
          unitPrice: 990,
          quantity: 1,
          discountRate: 0,
          discountAmount: 0,
        },
        // ... other mock data can be added here
      ];
    }
    return [];
  });

  const [treatmentCompletedItems, setTreatmentCompletedItems] = useState<TreatmentItem[]>(() => {
    if (recordId) {
      return [
        {
          id: 101,
          selected: false,
          content: "recheck(新料金)1",
          memo: "再診料099",
          insurance: true,
          unitPrice: 990,
          quantity: 1,
          discountRate: 0,
          discountAmount: 0,
        },
      ];
    }
    return [];
  });

  // Handle navigation for new records when petId is missing or invalid
  useEffect(() => {
    if (!recordId && !selectedPet) {
      navigate("/medical-records/select-pet");
    }
  }, [recordId, selectedPet, navigate]);

  const handleBack = () => {
      if (location.state?.from) {
          navigate(location.state.from);
          return;
      }

      if (!recordId) {
          navigate("/medical-records/select-pet");
      } else {
          navigate("/medical-records");
      }
  };

  const handleSave = () => {
      // TODO: Implement inventory consumption when inventory feature is ready
      toast.success(isNewRecord ? "カルテを作成しました" : "カルテを更新しました");
      setTimeout(() => {
          if (location.state?.from) {
              navigate(location.state.from);
          } else {
              navigate("/medical-records");
          }
      }, 800);
  };

  return {
    isNewRecord,
    activeTab,
    setActiveTab,
    selectedPet,
    handleBack,
    handleSave,
    treatmentPlanItems,
    setTreatmentPlanItems,
    treatmentCompletedItems,
    setTreatmentCompletedItems
  };
}
