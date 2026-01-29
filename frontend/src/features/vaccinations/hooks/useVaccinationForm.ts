import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { MOCK_PETS } from "@/config/mock-data";
import { MOCK_VACCINATION_RECORDS } from "../api";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { findPetByRecord } from "@/utils/pet-matching";

export function useVaccinationForm(id?: string) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  // Pet Selection
  const petSelection = usePetSelection();
  const { setSelectedPets } = petSelection;

  // Lazy initialization for edit mode
  const getInitialFormData = () => {
    if (isEdit && id) {
      const record = MOCK_VACCINATION_RECORDS.find(r => r.id === id);
      if (record) {
        // Simple mapping for demo
        const mappedVaccineName = record.vaccineName === "狂犬病ワクチン" ? "rabies" :
                                  record.vaccineName === "3種混合ワクチン" ? "mixed5" :
                                  record.vaccineName === "フィラリア予防" ? "filaria" :
                                  record.vaccineName === "ノミダニ予防" ? "flea" : "mixed8";
        return {
          vaccineName: mappedVaccineName,
          date: record.date.replace(/\//g, "-"),
          nextDate: record.nextDate.replace(/\//g, "-"),
        };
      }
    }
    return {
      vaccineName: "",
      date: "",
      nextDate: "",
    };
  };

  const initialFormData = getInitialFormData();

  // Form State
  const [vaccineName, setVaccineName] = useState(initialFormData.vaccineName);
  const [date, setDate] = useState(initialFormData.date);
  const [supplemental, setSupplemental] = useState("");
  const [lot1, setLot1] = useState("");
  const [lot2, setLot2] = useState("");
  const [lot3, setLot3] = useState("");
  const [lot4, setLot4] = useState("");
  const [nextScheduleType, setNextScheduleType] = useState("4weeks");
  const [nextDate, setNextDate] = useState(initialFormData.nextDate);
  const [remarks, setRemarks] = useState("");

  // History Filter State
  const [filterStartDate, setFilterStartDate] = useState("");
  const [filterEndDate, setFilterEndDate] = useState("");
  const [historySearchTerm, setHistorySearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState("desc");

  // Handle pet selection initialization
  useEffect(() => {
    if (isEdit && id) {
      const record = MOCK_VACCINATION_RECORDS.find(r => r.id === id);
      if (record) {
        const pet = findPetByRecord(record.petName, record.ownerName);
        if (pet) {
          setSelectedPets([pet]);
        } else {
          // Fallback for demo
          setSelectedPets([{
            id: "temp",
            name: record.petName,
            species: "犬",
            breed: "Mix",
            gender: "female",
            birthDate: "2015/01/01",
            weight: "10",
            ownerId: "temp_owner",
            ownerName: record.ownerName
          }]);
        }
      }
    } else {
      if (petId) {
        const foundPet = MOCK_PETS.find(p => p.id === petId);
        if (foundPet) {
          setSelectedPets([foundPet]);
        } else {
          navigate("/vaccinations/select-pet");
        }
      }
    }
  }, [id, isEdit, setSelectedPets, petId, navigate]);

  const handleSave = () => {
      // TODO: API call to save vaccination
      navigate("/vaccinations");
  };

  const handleClearHistoryFilter = () => {
      setHistorySearchTerm("");
  };

  return {
    isEdit,
    petSelection,
    form: {
        vaccineName, setVaccineName,
        date, setDate,
        supplemental, setSupplemental,
        lot1, setLot1,
        lot2, setLot2,
        lot3, setLot3,
        lot4, setLot4,
        nextScheduleType, setNextScheduleType,
        nextDate, setNextDate,
        remarks, setRemarks
    },
    historyFilter: {
        filterStartDate, setFilterStartDate,
        filterEndDate, setFilterEndDate,
        historySearchTerm, setHistorySearchTerm,
        sortOrder, setSortOrder,
        handleClearHistoryFilter
    },
    handleSave
  };
}
