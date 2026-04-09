import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationCustomer } from "./types";

export async function getReservationCustomers(clinicId: string): Promise<ReservationCustomer[]> {
  const { data } = await axios.get<ReservationCustomer[]>(
    `/v1/clinics/${clinicId}/reservation-customers`
  );
  return data ?? [];
}

export function useGetReservationCustomers(clinicId: string | null) {
  return useQuery({
    queryKey: ["reservation-customers", clinicId],
    queryFn: () => getReservationCustomers(clinicId!),
    enabled: clinicId !== null,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
