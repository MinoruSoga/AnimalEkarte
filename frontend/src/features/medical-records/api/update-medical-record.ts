import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MedicalRecord } from "@/types";
import { transformMedicalRecord } from "./transforms";
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
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["medical-records"] });
    },
  });
};
