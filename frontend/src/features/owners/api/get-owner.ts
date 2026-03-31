import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Owner } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { Owner as BackendOwner } from "@/types/generated/models";

export const getOwner = async (id: string): Promise<Owner> => {
  const { data } = await axios.get<BackendOwner>(`/v1/owners/${id}`);
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
