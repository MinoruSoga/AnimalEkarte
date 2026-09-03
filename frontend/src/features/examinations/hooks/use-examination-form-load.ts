import { useGetExamination } from "../api/get-examination";
import { useGetExaminationItems } from "../api/get-examination-items";
import { useGetPet } from "@/hooks/use-pet";
import { resolveEntityReadResult, type EntityReadResult } from "@/lib/entity-read-result";
import type { ExaminationRecord } from "../api/transforms";

export function useExaminationFormLoad(id: string | undefined, isEdit: boolean, petId: string | null) {
  const {
    data: examinationData,
    isLoading: isExaminationLoading,
    isError: isExaminationError,
    error: examinationError,
    refetch: refetchExamination,
  } = useGetExamination(id ?? "");
  const entityRead: EntityReadResult<ExaminationRecord> = resolveEntityReadResult({
    id: isEdit ? id : undefined,
    data: examinationData,
    isLoading: isExaminationLoading,
    isError: isExaminationError,
    error: examinationError,
    refetch: refetchExamination,
  });
  const existingExam =
    entityRead.status === "found" ? entityRead.data : undefined;
  const mutationPetId = isEdit ? (existingExam?.petId ?? "") : (petId ?? "");
  const { data: mutationPet, isLoading: isPetLoading } = useGetPet(mutationPetId);
  const { data: existingItems, isSuccess: existingItemsQuerySucceeded } =
    useGetExaminationItems(id ?? "");
  return {
    entityRead,
    existingExam,
    mutationPet,
    isPetLoading,
    existingItems,
    existingItemsQuerySucceeded,
  };
}
