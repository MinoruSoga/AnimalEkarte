import { useState, useEffect, useTransition, useCallback } from "react";
import { toast } from "sonner";
import { useNavigate, useSearchParams } from "react-router";
import type { ExaminationRecord } from "@/types";
import { paths } from "@/config/paths";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { useGetPet } from "@/hooks/use-pet";
import { useGetExamination } from "../api/get-examination";
import { useCreateExamination } from "../api/create-examination";
import { useUpdateExamination } from "../api/update-examination";
import { useDeleteExamination } from "../api/delete-examination";
import type { CreateExaminationRequest, UpdateExaminationRequest } from "../api/types";

const EXAM_STATUS_JA_TO_EN: Record<string, "pending" | "in_progress" | "completed"> = {
  "依頼中": "pending",
  "検査中": "in_progress",
  "完了": "completed",
};

// v2: added handleDelete, isDeleting
export function useExaminationForm(id?: string, medicalRecordIdParam?: string) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const petId = searchParams.get("petId");
  const medicalRecordId = medicalRecordIdParam ?? searchParams.get("medicalRecordId") ?? "";
  const isEdit = !!id;

  // Pet Search State
  const petSelection = usePetSelection();
  const { setSelectedPets, selectedPets } = petSelection;

  // API hooks
  const { data: existingExam } = useGetExamination(id ?? "");
  const { data: petFromQuery, isLoading: isPetLoading } = useGetPet(petId ?? "");
  const createMutation = useCreateExamination();
  const updateMutation = useUpdateExamination();
  const deleteMutation = useDeleteExamination();

  // useTransition: save/delete の pending 管理 (rerender-transitions)
  const [isSaveTransitionPending, startSaveTransition] = useTransition();
  const [isDeleteTransitionPending, startDeleteTransition] = useTransition();

  // Local overrides applied on top of server data (only tracks user edits in edit mode)
  const [localOverrides, setLocalOverrides] = useState<Partial<ExaminationRecord>>({});

  // Merge: server data as base + user edits on top
  const formData: Partial<ExaminationRecord> =
    isEdit && existingExam
      ? { ...existingExam, ...localOverrides }
      : { status: "依頼中" as const, ownerName: "", petName: "", ...localOverrides };

  const setFormData = (next: Partial<ExaminationRecord>) => {
    setLocalOverrides((prev) => ({ ...prev, ...next }));
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
    startSaveTransition(() => {
      if (isEdit && id) {
        const req: UpdateExaminationRequest = {
          status: formDataWithPet.status ? EXAM_STATUS_JA_TO_EN[formDataWithPet.status] : undefined,
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
          medical_record_id: medicalRecordId ? Number(medicalRecordId) : null,
          pet_id: Number(pet.id) || null,
          exam_type_id: Number(formDataWithPet.testTypeId) || 0,
          doctor_id: formDataWithPet.doctorId ? Number(formDataWithPet.doctorId) : null,
          date: formDataWithPet.date ?? new Date().toISOString(),
          result_summary: formDataWithPet.resultSummary,
          machine: formDataWithPet.machine,
        };
        createMutation.mutate(req, {
          onSuccess: () => navigate(paths.examinations.getHref()),
        });
      }
    });
  };

  const handleDelete = useCallback((onSuccess?: () => void) => {
    if (!isEdit || !id) return;
    startDeleteTransition(() => {
      deleteMutation.mutate(id, {
        onSuccess: () => {
          toast.success("検査記録を削除しました");
          onSuccess?.();
        },
      });
    });
  }, [isEdit, id, deleteMutation, startDeleteTransition]);

  const isSaving = createMutation.isPending || updateMutation.isPending || isSaveTransitionPending;
  const isDeleting = deleteMutation.isPending || isDeleteTransitionPending;

  return {
    formData: formDataWithPet,
    setFormData,
    petSelection,
    handleSave,
    handleDelete,
    isEdit,
    isSaving,
    isDeleting,
  };
}
