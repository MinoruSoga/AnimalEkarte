import { useState, useEffect, useTransition, useCallback, useMemo } from "react";
import { useNavigate, useSearchParams } from "react-router";
import { paths } from "@/config/paths";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
import { useGetVaccination } from "../api/get-vaccination";
import { useCreateVaccination } from "../api/create-vaccination";
import { useUpdateVaccination } from "../api/update-vaccination";
import type { CreateVaccinationRequest, UpdateVaccinationRequest } from "../api/types";

interface VaccinationFormState {
  vaccineId: string;
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
  vaccineId: "",
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
        vaccineId: existingVaccination.vaccineId,
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

  const setField = useCallback(<K extends keyof VaccinationFormState>(key: K, value: VaccinationFormState[K]) => {
    setLocalOverrides((prev) => ({ ...prev, [key]: value }));
  }, []);

  const [isSaveTransitionPending, startSaveTransition] = useTransition();

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
        navigate(paths.vaccinations.selectPet.getHref());
      }
    }
  }, [isEdit, petId, petFromQuery, isPetLoading, setSelectedPets, navigate]);

  const handleSave = () => {
    startSaveTransition(() => {
      if (isEdit && id) {
        const toRFC3339 = (d: string) => d ? `${d}T00:00:00Z` : undefined;
        const req: UpdateVaccinationRequest = {
          date: toRFC3339(formData.date),
          next_date: formData.nextDate ? `${formData.nextDate}T00:00:00Z` : null,
          lot1: formData.lot1 || undefined,
          remarks: formData.remarks || undefined,
        };
        updateMutation.mutate(
          { id, req },
          { onSuccess: () => navigate(paths.vaccinations.getHref()) }
        );
      } else {
        const pet = selectedPets[0];
        if (!pet) return;
        const req: CreateVaccinationRequest = {
          medical_record_id: null,
          pet_id: Number(pet.id),
          vaccine_id: Number(formData.vaccineId),
          date: formData.date ? `${formData.date}T00:00:00Z` : new Date().toISOString(),
          next_date: formData.nextDate ? `${formData.nextDate}T00:00:00Z` : null,
          lot1: formData.lot1 || undefined,
          remarks: formData.remarks || undefined,
        };
        createMutation.mutate(req, {
          onSuccess: () => navigate(paths.vaccinations.getHref()),
        });
      }
    });
  };

  const handleClearHistoryFilter = () => {
    setHistorySearchTerm("");
  };

  const isSaving = createMutation.isPending || updateMutation.isPending || isSaveTransitionPending;

  const setVaccineId = useCallback((v: string) => setField("vaccineId", v), [setField]);
  const setDate = useCallback((v: string) => setField("date", v), [setField]);
  const setSupplemental = useCallback((v: string) => setField("supplemental", v), [setField]);
  const setLot1 = useCallback((v: string) => setField("lot1", v), [setField]);
  const setLot2 = useCallback((v: string) => setField("lot2", v), [setField]);
  const setLot3 = useCallback((v: string) => setField("lot3", v), [setField]);
  const setLot4 = useCallback((v: string) => setField("lot4", v), [setField]);
  const setNextScheduleType = useCallback((v: string) => setField("nextScheduleType", v), [setField]);
  const setNextDate = useCallback((v: string) => setField("nextDate", v), [setField]);
  const setRemarks = useCallback((v: string) => setField("remarks", v), [setField]);

  const form = useMemo(() => ({
    vaccineId: formData.vaccineId,
    setVaccineId,
    date: formData.date,
    setDate,
    supplemental: formData.supplemental,
    setSupplemental,
    lot1: formData.lot1,
    setLot1,
    lot2: formData.lot2,
    setLot2,
    lot3: formData.lot3,
    setLot3,
    lot4: formData.lot4,
    setLot4,
    nextScheduleType: formData.nextScheduleType,
    setNextScheduleType,
    nextDate: formData.nextDate,
    setNextDate,
    remarks: formData.remarks,
    setRemarks,
  }), [
    formData.vaccineId, setVaccineId,
    formData.date, setDate,
    formData.supplemental, setSupplemental,
    formData.lot1, setLot1,
    formData.lot2, setLot2,
    formData.lot3, setLot3,
    formData.lot4, setLot4,
    formData.nextScheduleType, setNextScheduleType,
    formData.nextDate, setNextDate,
    formData.remarks, setRemarks,
  ]);

  return {
    isEdit,
    petSelection,
    form,
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
