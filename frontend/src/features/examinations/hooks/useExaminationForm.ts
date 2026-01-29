import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { ExaminationRecord } from "@/types";
import { MOCK_EXAMINATION_RECORDS, MOCK_PETS } from "@/config/mock-data";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { findPetByRecord } from "@/utils/pet-matching";

export function useExaminationForm(id?: string) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  // Pet Search State
  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  // Lazy initialization for edit mode
  const [formData, setFormData] = useState<Partial<ExaminationRecord>>(() => {
    if (isEdit && id) {
      const record = MOCK_EXAMINATION_RECORDS.find(r => r.id === id);
      if (record) {
        return record;
      }
      // Fallback default
      return {
        date: "2025/10/10 10:00",
        ownerName: "林 文明",
        petName: "Iris",
        testType: "blood",
        doctor: "dr_a",
        status: "依頼中",
        resultSummary: "",
      };
    }
    return {
      status: "依頼中",
      ownerName: "",
      petName: "",
    };
  });

  // Load initial pet selection for edit mode and handle petId navigation
  useEffect(() => {
    if (isEdit && id) {
        const record = MOCK_EXAMINATION_RECORDS.find(r => r.id === id);
        if (record) {
            const pet = findPetByRecord(record.petName, record.ownerName);
            if (pet) {
                setSelectedPets([pet]);
            }
        }
    } else {
        if (petId) {
            const foundPet = MOCK_PETS.find(p => p.id === petId);
            if (foundPet) {
                setSelectedPets([foundPet]);
            } else {
                navigate("/examinations/select-pet");
            }
        }
    }
  }, [id, isEdit, setSelectedPets, petId, navigate]);

  // Derive form data with pet info at render time (no setState-in-useEffect)
  const formDataWithPet = selectedPets.length > 0
    ? { ...formData, ownerName: selectedPets[0].ownerName, petName: selectedPets[0].name }
    : formData;

  const handleSave = () => {
      // TODO: API call to save examination
      navigate("/examinations");
  };

  return {
    formData: formDataWithPet,
    setFormData,
    petSelection,
    handleSave,
    isEdit
  };
}
