import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

export const deleteAccounting = async (id: string): Promise<void> => {
  await axios.delete(`/v1/accountings/${id}`);
};

export const useDeleteAccounting = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteAccounting,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["accountings"] });
    },
    onError: (error) => {
      handleApiError(error, "削除");
    },
  });
};
