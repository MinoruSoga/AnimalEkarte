import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router";
import type { ExaminationRecord } from "@/types";
import { paths } from "@/config/paths";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
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
        navigate(paths.examinations.selectPet.getHref());
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
        machine: formDataWithPet.machine,
        date: formDataWithPet.date,
      };
      updateMutation.mutate(
        { id, req },
        { onSuccess: () => navigate(paths.examinations.getHref()) }
      );
    } else {
      const pet = selectedPets[0];
      if (!pet) return;
      const req: CreateExaminationRequest = {
        medical_record_id: "",
        pet_id: pet.id,
        exam_type_id: formDataWithPet.testType ?? "",
        date: formDataWithPet.date ?? new Date().toISOString(),
        result_summary: formDataWithPet.resultSummary,
        machine: formDataWithPet.machine,
      };
      createMutation.mutate(req, {
        onSuccess: () => navigate(paths.examinations.getHref()),
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
