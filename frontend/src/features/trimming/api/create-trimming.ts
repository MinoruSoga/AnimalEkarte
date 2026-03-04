import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { TrimmingRecord } from "@/types";
import { transformTrimming } from "./transforms";
import type { BackendTrimming, CreateTrimmingRequest } from "./types";

export const createTrimming = async (
  req: CreateTrimmingRequest
): Promise<TrimmingRecord> => {
  const { data } = await axios.post<BackendTrimming>("/v1/trimmings", req);
  return transformTrimming(data);
};

export const useCreateTrimming = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createTrimming,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["trimmings"] });
    },
  });
};
