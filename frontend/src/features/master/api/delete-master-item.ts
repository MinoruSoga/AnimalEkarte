import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export const deleteMasterItem = async (id: string): Promise<void> => {
  await axios.delete(`/v1/master-items/${id}`);
};

export const useDeleteMasterItem = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteMasterItem,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masterItems"] });
    },
  });
};
