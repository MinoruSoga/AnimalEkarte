import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

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
  });
};
