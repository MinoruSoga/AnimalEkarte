import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ServiceType } from "./types";

export async function getCourses(clinicId: string): Promise<ServiceType[]> {
  const { data } = await axios.get<ServiceType[]>(
    `/v1/clinics/${clinicId}/reservation-courses`
  );
  return data ?? [];
}

export function useGetCourses(clinicId: string | null) {
  return useQuery({
    queryKey: ["reservation-courses", clinicId],
    queryFn: () => getCourses(clinicId!),
    enabled: clinicId !== null,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
