import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { TrimmingRecord } from "@/types";
import { transformTrimming } from "./transforms";
import type { BackendTrimming } from "./types";

interface TrimmingsListResponse {
  data: BackendTrimming[];
  total: number;
  page: number;
  limit: number;
}

export const getTrimmings = async (): Promise<TrimmingRecord[]> => {
  const { data } = await axios.get<TrimmingsListResponse>("/v1/trimmings");
  return data.data.map(transformTrimming);
};

export const useGetTrimmings = () => {
  return useQuery({
    queryKey: ["trimmings"],
    queryFn: getTrimmings,
  });
};
