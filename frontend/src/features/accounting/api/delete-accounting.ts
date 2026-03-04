import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

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
  });
};
