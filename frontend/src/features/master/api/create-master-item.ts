import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { MasterItem } from "@/types";
import { MASTER_CATEGORY_ENDPOINT, transformGenericMasterItem, type GenericMasterBackendItem } from "./get-master-items";
import type { CreateMasterItemRequest } from "./types";

export const createMasterItem = async (
  category: string,
  req: CreateMasterItemRequest
): Promise<MasterItem> => {
  const endpoint = MASTER_CATEGORY_ENDPOINT[category];
  if (!endpoint) throw new Error(`No endpoint for category: ${category}`);
  const isConsultation = category === "consultation";
  const { data } = await axios.post<GenericMasterBackendItem>(endpoint, {
    name: req.name,
    ...(req.code !== undefined && { code: req.code }),
    price: req.price,
    description: req.description ?? "",
    is_active: true,
    ...(isConsultation && {
      time_condition: req.timeCondition ?? "anytime",
      ...(req.duration != null && { duration: req.duration }),
    }),
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
      void queryClient.invalidateQueries({ queryKey: queryKeys.masters.category(category) });
    },
    onError: (error) => handleApiError(error, "マスタの作成"),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
