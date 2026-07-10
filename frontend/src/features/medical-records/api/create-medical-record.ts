import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import type { MedicalRecord } from "@/types";
import { transformMedicalRecord } from "./transforms";
import type { BackendMedicalRecord, CreateMedicalRecordRequest } from "./types";

const createMedicalRecord = async (
  req: CreateMedicalRecordRequest
): Promise<MedicalRecord> => {
  const { data } = await axios.post<BackendMedicalRecord>(
    "/v1/medical-records",
    req
  );
  return transformMedicalRecord(data);
};

export const useCreateMedicalRecord = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createMedicalRecord,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["medical-records"] });
      queryClient.invalidateQueries({ queryKey: ["reception"] });
    },
    onError: (error) => handleApiError(error, "カルテ作成"),
  });
};
