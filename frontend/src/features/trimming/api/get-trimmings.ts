import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { TrimmingRecord } from "@/types";
import { transformTrimming } from "./transforms";
import type { BackendTrimming } from "./types";

export const getTrimmings = async (): Promise<TrimmingRecord[]> => {
  const { data } = await axios.get<BackendTrimming[]>("/v1/trimmings");
  return data.map(transformTrimming);
};

export const useGetTrimmings = () => {
  return useQuery({
    queryKey: ["trimmings"],
    queryFn: getTrimmings,
  });
};
