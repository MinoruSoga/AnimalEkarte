import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { StaffSchedule } from "./types";

export async function getStaffSchedules(
  clinicId: string,
  staffId: number,
  month: string
): Promise<StaffSchedule[]> {
  const { data } = await axios.get<StaffSchedule[]>(
    `/v1/clinics/${clinicId}/reservation-staffs/${staffId}/schedules`,
    { params: { month } }
  );
  return data ?? [];
}

export function useGetStaffSchedules(
  clinicId: string | null,
  staffId: number | null,
  month: string
) {
  return useQuery({
    queryKey: ["staff-schedules", clinicId, staffId, month],
    queryFn: () => getStaffSchedules(clinicId!, staffId!, month),
    enabled: clinicId !== null && staffId !== null,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
