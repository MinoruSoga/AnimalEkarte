import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";
import type { ExaminationRecord } from "@/types";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/features/pets/api/get-pet";
import {
  useGetExamination,
  useCreateExamination,
  useUpdateExamination,
} from "../api";
import type { CreateExaminationRequest, UpdateExaminationRequest } from "../api";

export function useExaminationForm(id?: string) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const isEdit = !!id;

  // Pet Search State
  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  // API hooks
  const { data: existingExam } = useGetExamination(id ?? "");
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");
  const createMutation = useCreateExamination();
  const updateMutation = useUpdateExamination();

  // Local overrides applied on top of server data (only tracks user edits in edit mode)
  const [localOverrides, setLocalOverrides] = useState<Partial<ExaminationRecord>>({});

  // Merge: server data as base + user edits on top
  const formData: Partial<ExaminationRecord> =
    isEdit && existingExam
      ? { ...existingExam, ...localOverrides }
      : { status: "依頼中" as const, ownerName: "", petName: "", ...localOverrides };

  const setFormData = (next: Partial<ExaminationRecord>) => {
    setLocalOverrides(next);
  };

  // New mode: populate pet selection from petId query param
  useEffect(() => {
    if (!isEdit) {
      if (petFromQuery) {
        setSelectedPets([petFromQuery]);
      } else if (!petId && !isPetLoading) {
        // No petId provided and not loading — redirect to pet selection
        navigate("/examinations/select-pet");
      }
      // If petId is provided but petFromQuery is not yet resolved, wait
    }
  }, [isEdit, petId, petFromQuery, isPetLoading, setSelectedPets, navigate]);

  // Derive form data with pet info at render time (no setState-in-useEffect)
  const formDataWithPet =
    selectedPets.length > 0
      ? {
          ...formData,
          ownerName: selectedPets[0].ownerName,
          petName: selectedPets[0].name,
        }
      : formData;

  const handleSave = () => {
    if (isEdit && id) {
      const req: UpdateExaminationRequest = {
        status: formDataWithPet.status,
        result_summary: formDataWithPet.resultSummary,
        test_type: formDataWithPet.testType,
        machine: formDataWithPet.machine,
        examination_date: formDataWithPet.date,
      };
      updateMutation.mutate(
        { id, req },
        { onSuccess: () => navigate("/examinations") }
      );
    } else {
      const pet = selectedPets[0];
      if (!pet) return;
      const req: CreateExaminationRequest = {
        pet_id: pet.id,
        owner_id: pet.ownerId,
        examination_date:
          formDataWithPet.date ?? new Date().toISOString(),
        test_type: formDataWithPet.testType ?? "",
        result_summary: formDataWithPet.resultSummary,
        machine: formDataWithPet.machine,
      };
      createMutation.mutate(req, {
        onSuccess: () => navigate("/examinations"),
      });
    }
  };

  const isSaving = createMutation.isPending || updateMutation.isPending;

  return {
    formData: formDataWithPet,
    setFormData,
    petSelection,
    handleSave,
    isEdit,
    isSaving,
  };
}
