import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { LineCustomer } from "./types";

async function getLineCustomers(clinicId: string): Promise<LineCustomer[]> {
  const { data } = await axios.get<LineCustomer[]>(`/v1/clinics/${clinicId}/line-customers`);
  return data ?? [];
}

export function useGetLineCustomers(clinicId: string | null) {
  return useQuery({
    queryKey: queryKeys.lineCustomers(clinicId!),
    queryFn: () => getLineCustomers(clinicId!),
    enabled: clinicId !== null,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
