import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { TrimmingRecord } from "@/types";
import { transformTrimming } from "./transforms";
import type { BackendTrimming, UpdateTrimmingRequest } from "./types";

export const updateTrimming = async (
  id: string,
  req: UpdateTrimmingRequest
): Promise<TrimmingRecord> => {
  const { data } = await axios.patch<BackendTrimming>(
    `/v1/trimmings/${id}`,
    req
  );
  return transformTrimming(data);
};

export const useUpdateTrimming = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      req,
    }: {
      id: string;
      req: UpdateTrimmingRequest;
    }) => updateTrimming(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["trimmings"] });
    },
  });
};
