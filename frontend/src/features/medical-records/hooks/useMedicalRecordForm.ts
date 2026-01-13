import { useState, useEffect } from "react";
import { useNavigate, useSearchParams, useLocation } from "react-router-dom";
import { toast } from "sonner";
import { Pet } from "../../../types";
import { MOCK_PETS, MOCK_MEDICAL_RECORDS } from "../../../lib/constants";
import { TreatmentItem } from "../components/TreatmentTable";
import { useInventory } from "../../inventory/hooks/useInventory";

export function useMedicalRecordForm(recordId?: string) {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isNewRecord = !recordId;
  const { consumeStock } = useInventory(); // For stock consumption logic
  
  const [activeTab, setActiveTab] = useState("問診");
  const [selectedPet, setSelectedPet] = useState<Pet | null>(null);

  // Treatment Data States
  const [treatmentPlanItems, setTreatmentPlanItems] = useState<TreatmentItem[]>([]);
  const [treatmentCompletedItems, setTreatmentCompletedItems] = useState<TreatmentItem[]>([]);

  // Load initial data
  useEffect(() => {
    if (recordId) {
        // Find record and associated pet
        const record = MOCK_MEDICAL_RECORDS.find(r => r.id === recordId);
        if (record?.petId) {
            const pet = MOCK_PETS.find(p => p.id === record.petId);
            if (pet) {
                setSelectedPet(pet);
            }
        } else {
            // Fallback for mock if link fails
            const mockPet = MOCK_PETS.find(p => p.id === "1");
            if (mockPet) setSelectedPet(mockPet);
        }

        // Mock: Load treatment data
        setTreatmentPlanItems([
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
        ]);
        setTreatmentCompletedItems([
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
        ]);
    } else {
        // New record - check for petId in query params
        if (petId) {
            const pet = MOCK_PETS.find(p => p.id === petId);
            if (pet) {
                setSelectedPet(pet);
            } else {
                navigate("/medical-records/select-pet");
            }
        } else {
            navigate("/medical-records/select-pet");
        }
    }
  }, [recordId, petId, navigate]);

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
      // Inventory consumption logic
      const itemsToConsume = treatmentCompletedItems
        .filter(item => item.inventoryId && item.quantity > 0)
        .map(item => ({
            inventoryId: item.inventoryId!,
            quantity: item.quantity
        }));
      
      if (itemsToConsume.length > 0) {
          consumeStock(itemsToConsume);
          // console.log("Consumed Stock:", itemsToConsume);
      }

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
