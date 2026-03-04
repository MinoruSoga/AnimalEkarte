import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MasterItem } from "@/types";
import { transformMasterItem } from "./transforms";
import type { BackendMasterItem, CreateMasterItemRequest } from "./types";

export const createMasterItem = async (
  req: CreateMasterItemRequest
): Promise<MasterItem> => {
  const { data } = await axios.post<BackendMasterItem>(
    "/v1/master-items",
    req
  );
  return transformMasterItem(data);
};

export const useCreateMasterItem = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createMasterItem,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masterItems"] });
    },
  });
};
