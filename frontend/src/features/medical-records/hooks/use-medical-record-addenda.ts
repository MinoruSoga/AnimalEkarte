import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { getMedicalRecordAddenda } from "../api/get-medical-record-addenda";
import { createMedicalRecordAddendum } from "../api/create-medical-record-addendum";
import type { CreateMedicalRecordAddendumInput } from "../api/create-medical-record-addendum";

export function useMedicalRecordAddenda(medicalRecordId: string) {
  return useQuery({
    queryKey: queryKeys.medicalRecords.addenda(medicalRecordId),
    queryFn: () => getMedicalRecordAddenda(medicalRecordId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}

export function useCreateMedicalRecordAddendum(medicalRecordId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateMedicalRecordAddendumInput) =>
      createMedicalRecordAddendum(medicalRecordId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.addenda(medicalRecordId),
      });
    },
  });
}
