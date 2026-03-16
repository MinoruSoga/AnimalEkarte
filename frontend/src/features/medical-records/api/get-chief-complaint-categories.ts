import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ChiefComplaintCategory as ModelChiefComplaintCategory } from "@/types/generated/models";

export interface ChiefComplaintCategory {
  id: number;
  name: string;
}

export const getChiefComplaintCategories = async (): Promise<ChiefComplaintCategory[]> => {
  const { data } = await axios.get<ModelChiefComplaintCategory[]>("/v1/masters/chief-complaint-categories");
  return data.map((item) => ({
    id: Number(item.id ?? 0),
    name: item.name,
  }));
};

export const useGetChiefComplaintCategories = () =>
  useQuery({
    queryKey: ["masters", "chief-complaints"],
    queryFn: getChiefComplaintCategories,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
