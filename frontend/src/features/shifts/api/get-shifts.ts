import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { transformShift } from "./transforms";
import type { Shift } from "./transforms";
import type { ShiftEntry } from "./contract-types";

export interface GetShiftsParams {
  date?: string; // YYYY-MM
  staff_id?: string;
}

async function getShifts(params: GetShiftsParams): Promise<Shift[]> {
  const query = new URLSearchParams();
  if (params.date) query.set("date", params.date);
  if (params.staff_id) query.set("staff_id", params.staff_id);

  const { data } = await axios.get<ShiftEntry[]>(`/v1/shifts?${query.toString()}`);
  return (data ?? []).map(transformShift);
}

/**
 * GET /v1/shifts/{id}（EMR-514 固定API契約）。
 * 編集案を上書きせず、409 version_conflict 時の比較/再確認用に最新状態を取得する。
 */
export async function getShiftById(id: string): Promise<ShiftEntry> {
  const { data } = await axios.get<ShiftEntry>(`/v1/shifts/${id}`);
  return data;
}

export function useGetShifts(params: GetShiftsParams) {
  return useQuery({
    queryKey: queryKeys.shifts.list(params.date, params.staff_id),
    queryFn: () => getShifts(params),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
