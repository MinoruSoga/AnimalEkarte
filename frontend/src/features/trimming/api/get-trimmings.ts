import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { TrimmingUI } from "@/types";
import { transformTrimming } from "./transforms";
import type { TrimmingListResponse } from "@/types/trimming";

export const getTrimmings = async (): Promise<TrimmingUI[]> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings");
  return data.data.map(transformTrimming);
};

export const useGetTrimmings = () => {
  return useQuery({
    queryKey: ["trimmings"],
    queryFn: getTrimmings,
  });
};
