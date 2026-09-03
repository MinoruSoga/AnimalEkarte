import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { transformShift } from "./transforms";
import type { Shift } from "./transforms";
import type { BackendShift } from "./types";

export interface GetShiftsParams {
  date?: string; // YYYY-MM
  staff_id?: string;
}

async function getShifts(params: GetShiftsParams): Promise<Shift[]> {
  const query = new URLSearchParams();
  if (params.date) query.set("date", params.date);
  if (params.staff_id) query.set("staff_id", params.staff_id);

  const { data } = await axios.get<BackendShift[]>(`/v1/shifts?${query.toString()}`);
  return (data ?? []).map(transformShift);
}

export function useGetShifts(params: GetShiftsParams) {
  return useQuery({
    queryKey: queryKeys.shifts.list(params.date, params.staff_id),
    queryFn: () => getShifts(params),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
