import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationSetting } from "./types";

async function getLineReservationSetting(clinicId: string): Promise<ReservationSetting> {
  const { data } = await axios.get<ReservationSetting>(
    `/v1/clinics/${clinicId}/line-reservation-settings`,
  );
  return data;
}

export function useGetLineReservationSetting(clinicId: string | null) {
  return useQuery({
    queryKey: queryKeys.lineReservationSettings(clinicId!),
    queryFn: () => getLineReservationSetting(clinicId!),
    enabled: clinicId !== null,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
