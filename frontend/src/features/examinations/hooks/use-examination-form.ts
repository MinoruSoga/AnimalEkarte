import { useState, useEffect, useTransition, useCallback, useActionState, useRef } from "react";
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
import type { ActionState } from "@/types/form";
import { INITIAL_ACTION_STATE } from "@/types/form";

const EXAM_STATUS_JA_TO_EN: Record<string, "pending" | "in_progress" | "result_entered" | "completed" | "confirmed"> = {
  "依頼中": "pending",
  "検査中": "in_progress",
  "結果入力済み": "result_entered",
  "完了": "completed",
  "確定": "confirmed",
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
  const [isDeleteTransitionPending, startDeleteTransition] = useTransition();

  // Local overrides applied on top of server data (only tracks user edits in edit mode)
  // useActionState の前に宣言: callback 内で formDataWithPet を参照するため
  const [localOverrides, setLocalOverrides] = useState<Partial<ExaminationRecord>>({});

  // Merge: server data as base + user edits on top
  const formData: Partial<ExaminationRecord> =
    isEdit && existingExam
      ? { ...existingExam, ...localOverrides }
      : { status: "依頼中" as const, ownerName: "", petName: "", ...localOverrides };

  const setFormData = (next: Partial<ExaminationRecord>) => {
    setLocalOverrides((prev) => ({ ...prev, ...next }));
  };

  // Derive form data with pet info at render time (no setState-in-useEffect)
  const formDataWithPet =
    selectedPets.length > 0
      ? {
          ...formData,
          ownerName: selectedPets[0].ownerName,
          petName: selectedPets[0].name,
        }
      : formData;

  // useActionState の stale closure 対策: 最新の formDataWithPet を ref で保持
  // (use-medical-record-form.ts の activeTabRef と同じパターン)
  const formDataWithPetRef = useRef(formDataWithPet);
  useEffect(() => {
    formDataWithPetRef.current = formDataWithPet;
  });

  /**
   * React 19 useActionState を使用したフォームアクション
   */
  const [formState, formAction, isPending] = useActionState(
    async (_prevState: ActionState, _formData: FormData): Promise<ActionState> => {
      const current = formDataWithPetRef.current;
      // フロントエンド・バリデーション
      const errors: Record<string, string> = {};
      if (!current.testTypeId) errors.testTypeId = "検査種別を選択してください";
      if (!current.doctorId) errors.doctorId = "担当医を選択してください";

      if (Object.keys(errors).length > 0) {
        toast.error("未入力の項目があります");
        return { success: false, fieldErrors: errors, timestamp: Date.now() };
      }

      try {
        if (isEdit && id) {
          const req: UpdateExaminationRequest = {
            status: current.status ? EXAM_STATUS_JA_TO_EN[current.status] : undefined,
            result_summary: current.resultSummary,
            machine: current.machine,
            date: current.date
              ? current.date.includes("T")
                ? current.date
                : `${current.date}T00:00:00Z`
              : undefined,
          };
          await updateMutation.mutateAsync({ id, req });
        } else {
          const pet = selectedPets[0];
          if (!pet) return { success: false, timestamp: Date.now() };
          const req: CreateExaminationRequest = {
            medical_record_id: medicalRecordId ? Number(medicalRecordId) : null,
            pet_id: Number(pet.id) || null,
            exam_type_id: Number(current.testTypeId) || 0,
            doctor_id: current.doctorId ? Number(current.doctorId) : null,
            date: current.date ?? new Date().toISOString(),
            result_summary: current.resultSummary,
            machine: current.machine,
          };
          await createMutation.mutateAsync(req);
        }
        return { success: true, timestamp: Date.now() };
      } catch {
        toast.error("保存に失敗しました");
        return { success: false, timestamp: Date.now() };
      }
    },
    INITIAL_ACTION_STATE
  );

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

  const isSaving = isPending;
  const isDeleting = deleteMutation.isPending || isDeleteTransitionPending;

  return {
    formData: formDataWithPet,
    setFormData,
    petSelection,
    formAction,
    formState,
    handleDelete,
    isEdit,
    isSaving,
    isDeleting,
  };
}
