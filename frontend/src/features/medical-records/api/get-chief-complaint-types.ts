import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ChiefComplaintType as ModelChiefComplaintType } from "@/types/generated/models";

export interface ChiefComplaintType {
  id: number;
  name: string;
}

export const getChiefComplaintTypes = async (): Promise<ChiefComplaintType[]> => {
  const { data } = await axios.get<ModelChiefComplaintType[]>("/v1/masters/chief-complaint-types");
  return data.map((item) => ({
    id: Number(item.id ?? 0),
    name: item.name,
  }));
};

export const useGetChiefComplaintTypes = () =>
  useQuery({
    queryKey: ["masters", "chief-complaints"],
    queryFn: getChiefComplaintTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
