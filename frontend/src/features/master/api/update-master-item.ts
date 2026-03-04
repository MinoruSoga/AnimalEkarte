import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MasterItem } from "@/types";
import { transformMasterItem } from "./transforms";
import type { BackendMasterItem, UpdateMasterItemRequest } from "./types";

export const updateMasterItem = async (
  id: string,
  req: UpdateMasterItemRequest
): Promise<MasterItem> => {
  const { data } = await axios.put<BackendMasterItem>(
    `/v1/master-items/${id}`,
    req
  );
  return transformMasterItem(data);
};

export const useUpdateMasterItem = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      req,
    }: {
      id: string;
      req: UpdateMasterItemRequest;
    }) => updateMasterItem(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masterItems"] });
    },
  });
};
