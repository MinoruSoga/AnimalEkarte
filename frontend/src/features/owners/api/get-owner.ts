import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Owner } from "@/types/owner";
import { transformOwner, type OwnerApiResponse } from "./transforms";

export const getOwner = async (id: string): Promise<Owner> => {
  const { data } = await axios.get<OwnerApiResponse>(`/v1/owners/${id}`);
  return transformOwner(data);
};

export const useGetOwner = (id: string) => {
  return useQuery({
    queryKey: ["owners", id],
    queryFn: () => getOwner(id),
    staleTime: QUERY_STALE_TIMES.STATIC, // 単一飼主データも変更頻度は低い
    gcTime: QUERY_GC_TIMES.LONG,
    enabled: !!id,
  });
};
