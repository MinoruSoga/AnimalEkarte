import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { handleApiError } from "@/lib/handle-api-error";
import { MASTER_CATEGORY_ENDPOINT } from "./get-master-items";

export const deleteMasterItem = async (category: string, id: string): Promise<void> => {
  const endpoint = MASTER_CATEGORY_ENDPOINT[category];
  if (!endpoint) throw new Error(`No endpoint for category: ${category}`);
  await axios.delete(`${endpoint}/${id}`);
};

interface UseDeleteMasterItemParams {
  category: string;
}

export const useDeleteMasterItem = ({ category }: UseDeleteMasterItemParams) => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteMasterItem(category, id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.masters.category(category) });
    },
    onError: (error) => handleApiError(error, "マスタの削除"),
  });
};
