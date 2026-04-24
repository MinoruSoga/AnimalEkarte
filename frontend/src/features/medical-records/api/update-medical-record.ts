import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { transformMedicalRecord } from "./transforms";
import type { MedicalRecord } from "./transforms";
import type { BackendMedicalRecord, UpdateMedicalRecordRequest } from "./types";

export const updateMedicalRecord = async (
  id: string,
  req: UpdateMedicalRecordRequest
): Promise<MedicalRecord> => {
  const { data } = await axios.patch<BackendMedicalRecord>(
    `/v1/medical-records/${id}`,
    req
  );
  return transformMedicalRecord(data);
};

export const useUpdateMedicalRecord = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateMedicalRecordRequest }) =>
      updateMedicalRecord(id, req),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["medical-records"] });
      queryClient.invalidateQueries({ queryKey: ["medical-record", id] });
    },
    onError: (error) => {
      handleApiError(error, "更新");
    },
  });
};
