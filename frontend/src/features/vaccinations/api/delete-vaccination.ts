import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

const deleteVaccination = async (id: string): Promise<void> => {
  await axios.delete(`/v1/vaccinations/${id}`);
};

export const useDeleteVaccination = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteVaccination,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vaccinations.all() });
    },
    onError: (error) => {
      handleApiError(error, "削除");
    },
  });
};
