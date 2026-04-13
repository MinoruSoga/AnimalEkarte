import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Owner } from "@/types/owner";
import { transformOwner, type OwnerApiResponse } from "./transforms";

interface OwnersResponse {
  data: OwnerApiResponse[];
  total: number;
  page: number;
  limit: number;
}

export const getOwners = async (): Promise<Owner[]> => {
  const { data } = await axios.get<OwnersResponse>("/v1/owners");
  return data.data.map(transformOwner);
};

export const useGetOwners = () => {
  return useQuery({
    queryKey: ["owners"],
    queryFn: getOwners,
    staleTime: QUERY_STALE_TIMES.STATIC, // 飼主マスタは変更頻度が低い
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
