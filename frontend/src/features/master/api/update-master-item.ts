import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MasterItem } from "@/types";
import { MASTER_CATEGORY_ENDPOINT, transformGenericMasterItem, type GenericMasterBackendItem } from "./get-master-items";
import type { UpdateMasterItemRequest } from "./types";

export const updateMasterItem = async (
  category: string,
  id: string,
  req: UpdateMasterItemRequest
): Promise<MasterItem> => {
  const endpoint = MASTER_CATEGORY_ENDPOINT[category];
  if (!endpoint) throw new Error(`No endpoint for category: ${category}`);
  const payload: Record<string, unknown> = {};
  const isInquiry = category === "inquiry_template";
  if (req.name !== undefined) payload[isInquiry ? "title" : "name"] = req.name;
  if (req.code !== undefined) payload.code = req.code;
  if (req.price !== undefined && !isInquiry) payload.price = req.price;
  if (req.description !== undefined) payload[isInquiry ? "content" : "description"] = req.description;
  if (req.status !== undefined) payload.is_active = req.status !== "inactive";
  if (category === "consultation") {
    if (req.timeCondition !== undefined) payload.time_condition = req.timeCondition;
    if (req.duration !== undefined) payload.duration = req.duration;
  }
  const { data } = await axios.patch<GenericMasterBackendItem>(`${endpoint}/${id}`, payload);
  return transformGenericMasterItem(data);
};

interface UseUpdateMasterItemParams {
  category: string;
}

export const useUpdateMasterItem = ({ category }: UseUpdateMasterItemParams) => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateMasterItemRequest }) =>
      updateMasterItem(category, id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masterItems", category] });
    },
  });
};
