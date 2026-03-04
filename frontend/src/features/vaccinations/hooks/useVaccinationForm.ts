import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/features/pets/api/get-pet";
import {
  useGetVaccination,
  useCreateVaccination,
  useUpdateVaccination,
} from "../api";
import type { CreateVaccinationRequest, UpdateVaccinationRequest } from "../api";

interface VaccinationFormState {
  vaccineName: string;
  date: string;
  supplemental: string;
  lot1: string;
  lot2: string;
  lot3: string;
  lot4: string;
  nextScheduleType: string;
  nextDate: string;
  remarks: string;
}

const DEFAULT_FORM: VaccinationFormState = {
  vaccineName: "",
  date: "",
  supplemental: "",
  lot1: "",
  lot2: "",
  lot3: "",
  lot4: "",
  nextScheduleType: "4weeks",
  nextDate: "",
  remarks: "",
};

export function useVaccinationForm(id?: string) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  // Pet Selection
  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  // API hooks
  const { data: existingVaccination } = useGetVaccination(id ?? "");
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");
  const createMutation = useCreateVaccination();
  const updateMutation = useUpdateVaccination();

  // Local overrides: tracks user edits on top of server data
  const [localOverrides, setLocalOverrides] = useState<Partial<VaccinationFormState>>({});

  // Merge: server data as base + user edits on top
  const formData: VaccinationFormState = isEdit && existingVaccination
    ? {
        vaccineName: existingVaccination.vaccineName,
        date: existingVaccination.date ? existingVaccination.date.slice(0, 10) : "",
        supplemental: "",
        lot1: "",
        lot2: "",
        lot3: "",
        lot4: "",
        nextScheduleType: "4weeks",
        nextDate: existingVaccination.nextDate ? existingVaccination.nextDate.slice(0, 10) : "",
        remarks: "",
        ...localOverrides,
      }
    : { ...DEFAULT_FORM, ...localOverrides };

  const setField = <K extends keyof VaccinationFormState>(key: K, value: VaccinationFormState[K]) => {
    setLocalOverrides((prev) => ({ ...prev, [key]: value }));
  };

  // History Filter State
  const [filterStartDate, setFilterStartDate] = useState("");
  const [filterEndDate, setFilterEndDate] = useState("");
  const [historySearchTerm, setHistorySearchTerm] = useState("");
  const [sortOrder, setSortOrder] = useState("desc");

  // New mode: populate pet selection from petId query param
  useEffect(() => {
    if (!isEdit) {
      if (petFromQuery) {
        setSelectedPets([petFromQuery]);
      } else if (!petId && !isPetLoading) {
        navigate("/vaccinations/select-pet");
      }
    }
  }, [isEdit, petId, petFromQuery, isPetLoading, setSelectedPets, navigate]);

  const handleSave = () => {
    if (isEdit && id) {
      const req: UpdateVaccinationRequest = {
        vaccine_name: formData.vaccineName || undefined,
        vaccination_date: formData.date || undefined,
        next_date: formData.nextDate || null,
        lot_number: formData.lot1 || undefined,
        notes: formData.remarks || undefined,
      };
      updateMutation.mutate(
        { id, req },
        { onSuccess: () => navigate("/vaccinations") }
      );
    } else {
      const pet = selectedPets[0];
      if (!pet) return;
      const req: CreateVaccinationRequest = {
        pet_id: pet.id,
        owner_id: pet.ownerId,
        vaccine_name: formData.vaccineName,
        vaccination_date: formData.date || new Date().toISOString(),
        next_date: formData.nextDate || null,
        lot_number: formData.lot1 || undefined,
        notes: formData.remarks || undefined,
      };
      createMutation.mutate(req, {
        onSuccess: () => navigate("/vaccinations"),
      });
    }
  };

  const handleClearHistoryFilter = () => {
    setHistorySearchTerm("");
  };

  const isSaving = createMutation.isPending || updateMutation.isPending;

  return {
    isEdit,
    petSelection,
    form: {
      vaccineName: formData.vaccineName,
      setVaccineName: (v: string) => setField("vaccineName", v),
      date: formData.date,
      setDate: (v: string) => setField("date", v),
      supplemental: formData.supplemental,
      setSupplemental: (v: string) => setField("supplemental", v),
      lot1: formData.lot1,
      setLot1: (v: string) => setField("lot1", v),
      lot2: formData.lot2,
      setLot2: (v: string) => setField("lot2", v),
      lot3: formData.lot3,
      setLot3: (v: string) => setField("lot3", v),
      lot4: formData.lot4,
      setLot4: (v: string) => setField("lot4", v),
      nextScheduleType: formData.nextScheduleType,
      setNextScheduleType: (v: string) => setField("nextScheduleType", v),
      nextDate: formData.nextDate,
      setNextDate: (v: string) => setField("nextDate", v),
      remarks: formData.remarks,
      setRemarks: (v: string) => setField("remarks", v),
    },
    historyFilter: {
      filterStartDate, setFilterStartDate,
      filterEndDate, setFilterEndDate,
      historySearchTerm, setHistorySearchTerm,
      sortOrder, setSortOrder,
      handleClearHistoryFilter,
    },
    handleSave,
    isSaving,
  };
}
