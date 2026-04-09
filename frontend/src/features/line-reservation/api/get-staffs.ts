import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Staff } from "./types";

export async function getReservationStaffs(clinicId: string): Promise<Staff[]> {
  const { data } = await axios.get<Staff[]>(
    `/v1/clinics/${clinicId}/reservation-staffs`
  );
  return data ?? [];
}

export function useGetReservationStaffs(clinicId: string | null) {
  return useQuery({
    queryKey: ["reservation-staffs", clinicId],
    queryFn: () => getReservationStaffs(clinicId!),
    enabled: clinicId !== null,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
