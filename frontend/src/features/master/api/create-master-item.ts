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
  // inquiry_template uses title/content instead of name/description
  const isInquiry = category === "inquiry_template";
  const isConsultation = category === "consultation";
  const { data } = await axios.post<GenericMasterBackendItem>(endpoint, {
    ...(isInquiry
      ? { title: req.name, content: req.description ?? "" }
      : { name: req.name, ...(req.code !== undefined && { code: req.code }), price: req.price, description: req.description ?? "" }),
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
      queryClient.invalidateQueries({ queryKey: ["masterItems", category] });
    },
  });
};
