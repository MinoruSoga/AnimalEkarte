import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

const deleteMedicalRecord = async (id: string): Promise<void> => {
  await axios.delete(`/v1/medical-records/${id}`);
};

export const useDeleteMedicalRecord = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteMedicalRecord,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.all() });
    },
    onError: (error) => handleApiError(error, "カルテ削除"),
  });
};
