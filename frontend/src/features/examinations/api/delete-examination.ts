import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

const deleteExamination = async (id: string): Promise<void> => {
  await axios.delete(`/v1/examinations/${id}`);
};

export const useDeleteExamination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteExamination,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.examinations.all() });
    },
    onError: (error) => {
      handleApiError(error, "削除");
    },
  });
};
