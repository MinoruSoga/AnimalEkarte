import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { ShiftStaff } from "../types";

/** GET /v1/masters/staffs の最小フィールド（職種は occupations マスタ） */
interface BackendStaffOccupation {
  id: number | string;
  name: string;
}

interface BackendStaff {
  id: number | string;
  name: string;
  is_active: boolean;
  occupation_id?: number | string | null;
  occupation?: BackendStaffOccupation | null;
}

export const OCCUPATION_FILTER_ALL = "all";
export const OCCUPATION_FILTER_UNSET = "unset";

async function getStaffsForShift(): Promise<ShiftStaff[]> {
  const { data } = await axios.get<BackendStaff[]>("/v1/masters/staffs");
  return (data ?? [])
    .filter((s) => s.is_active)
    .map((s) => {
      const occupationId =
        s.occupation?.id != null
          ? String(s.occupation.id)
          : s.occupation_id != null
            ? String(s.occupation_id)
            : null;
      const occupationName = s.occupation?.name?.trim() ? s.occupation.name.trim() : null;
      return {
        id: String(s.id),
        name: s.name,
        occupationId,
        occupationName,
      };
    });
}

export function useGetStaffsForShift() {
  return useQuery({
    queryKey: queryKeys.staffsForShift(),
    queryFn: getStaffsForShift,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

/** 職種フィルタ値（all / unset / occupation id）でスタッフを絞る。表示のみ。 */
export function filterStaffsByOccupation(
  staffs: ShiftStaff[],
  occupationFilter: string,
): ShiftStaff[] {
  if (occupationFilter === OCCUPATION_FILTER_ALL) return staffs;
  if (occupationFilter === OCCUPATION_FILTER_UNSET) {
    return staffs.filter((s) => s.occupationId == null);
  }
  return staffs.filter((s) => s.occupationId === occupationFilter);
}
