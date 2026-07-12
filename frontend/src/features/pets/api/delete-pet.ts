import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

const deletePet = async (id: string): Promise<void> => {
  await axios.delete(`/v1/pets/${id}`);
};

export const useDeletePet = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deletePet,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.pets.list() });
    },
    onError: (error) => handleApiError(error, "ペット削除"),
  });
};
