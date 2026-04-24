import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

export const deleteTrimming = async (id: string): Promise<void> => {
  await axios.delete(`/v1/trimmings/${id}`);
};

export const useDeleteTrimming = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteTrimming,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["trimmings"] });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
};
