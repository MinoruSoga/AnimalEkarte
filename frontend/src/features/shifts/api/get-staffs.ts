import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ShiftStaff } from "../types";

interface BackendStaff {
  id: string;
  name: string;
  is_active: boolean;
}

export async function getStaffsForShift(): Promise<ShiftStaff[]> {
  const { data } = await axios.get<BackendStaff[]>("/v1/masters/staffs");
  return (data ?? [])
    .filter((s) => s.is_active)
    .map((s) => ({ id: String(s.id), name: s.name }));
}

export function useStaffsForShift() {
  return useQuery({
    queryKey: ["staffs-for-shift"],
    queryFn: getStaffsForShift,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
