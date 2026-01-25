import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { ExaminationRecord } from "../../../types";
import { MOCK_EXAMINATION_RECORDS, MOCK_PETS } from "../../../lib/constants";
import { usePetSelection } from "../../pets/hooks/usePetSelection";

export function useExaminationForm(id?: string) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  // Pet Search State
  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  const [formData, setFormData] = useState<Partial<ExaminationRecord>>({
    status: "依頼中",
    ownerName: "",
    petName: "",
  });

  // Load initial data
  useEffect(() => {
    if (isEdit && id) {
        // Mock data loading
        const record = MOCK_EXAMINATION_RECORDS.find(r => r.id === id);
        if (record) {
            setFormData(record);
            
            // Normalize strings for comparison (remove spaces, full-width spaces)
            const normalize = (s: string) => s.replace(/[\s\u3000]/g, "");
            
            const pet = MOCK_PETS.find(p => {
                const pName = normalize(p.name);
                const rName = normalize(record.petName);
                const pOwner = normalize(p.ownerName);
                const rOwner = normalize(record.ownerName);
                
                // Check if one includes the other
                return (pName.includes(rName) || rName.includes(pName)) && 
                       (pOwner.includes(rOwner) || rOwner.includes(pOwner));
            });

            if (pet) {
                setSelectedPets([pet]);
            }
        } else {
             // Fallback default
             setFormData({
                date: "2025/10/10 10:00",
                ownerName: "林 文明",
                petName: "Iris",
                testType: "blood",
                doctor: "dr_a",
                status: "依頼中",
                resultSummary: "",
            });
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

  // Sync selected pet to form data
  useEffect(() => {
    if (selectedPets.length > 0) {
      const pet = selectedPets[0];
      setFormData((prev) => ({
        ...prev,
        ownerName: pet.ownerName,
        petName: pet.name,
      }));
    }
  }, [selectedPets]);

  const handleSave = () => {
      // TODO: API call to save examination
      navigate("/examinations");
  };

  return {
    formData,
    setFormData,
    petSelection,
    handleSave,
    isEdit
  };
}
