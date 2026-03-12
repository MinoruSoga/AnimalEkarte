import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MasterItem } from "@/types";
import { MASTER_CATEGORY_ENDPOINT, transformGenericMasterItem, type GenericMasterBackendItem } from "./get-master-items";
import type { CreateMasterItemRequest } from "./types";

export const createMasterItem = async (
  category: string,
  req: CreateMasterItemRequest
): Promise<MasterItem> => {
  const endpoint = MASTER_CATEGORY_ENDPOINT[category];
  if (!endpoint) throw new Error(`No endpoint for category: ${category}`);
  const { data } = await axios.post<GenericMasterBackendItem>(endpoint, {
    name: req.name,
    code: req.code ?? "",
    price: req.price,
    description: req.description ?? "",
    is_active: true,
  });
  return transformGenericMasterItem(data);
};

interface UseCreateMasterItemParams {
  category: string;
}

export const useCreateMasterItem = ({ category }: UseCreateMasterItemParams) => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateMasterItemRequest) => createMasterItem(category, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masterItems", category] });
    },
  });
};
