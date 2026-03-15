import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Shift } from "../types";
import { transformShift } from "./transforms";
import type { BackendShift } from "./types";

export interface GetShiftsParams {
  date?: string;      // YYYY-MM
  staff_id?: string;
}

export async function getShifts(params: GetShiftsParams): Promise<Shift[]> {
  const query = new URLSearchParams();
  if (params.date) query.set("date", params.date);
  if (params.staff_id) query.set("staff_id", params.staff_id);

  const { data } = await axios.get<BackendShift[]>(`/v1/shifts?${query.toString()}`);
  return (data ?? []).map(transformShift);
}

export function useGetShifts(params: GetShiftsParams) {
  return useQuery({
    queryKey: ["shifts", params.date, params.staff_id],
    queryFn: () => getShifts(params),
  });
}
