import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

import { STAFFS_QUERY_KEY } from "./staffs";

// ─────────────────────────────────────────────────
// Staff Excluded Service Types API
// ─────────────────────────────────────────────────

const STAFF_EXCLUDED_ST_KEY = (staffId: string) =>
  [...STAFFS_QUERY_KEY, staffId, "excluded-reservation-types"] as const;

const STAFF_CAPABLE_ST_KEY = (staffId: string) =>
  [...STAFFS_QUERY_KEY, staffId, "capable-reservation-types"] as const;

export function useGetStaffCapableReservationTypes(staffId: string | null) {
  return useQuery({
    queryKey: STAFF_CAPABLE_ST_KEY(staffId ?? ""),
    queryFn: async (): Promise<string[]> => {
      const { data } = await axios.get<{ reservation_type_ids: number[] }>(
        `/v1/masters/staffs/${staffId}/capable-reservation-types`,
      );
      return (data.reservation_type_ids ?? []).map(String);
    },
    enabled: staffId !== null,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useUpdateStaffCapableReservationTypes() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      staffId,
      reservationTypeIds,
    }: {
      staffId: string;
      reservationTypeIds: string[];
    }) => {
      await axios.put(`/v1/masters/staffs/${staffId}/capable-reservation-types`, {
        reservation_type_ids: reservationTypeIds.map((id) => parseInt(id, 10)),
      });
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: STAFF_CAPABLE_ST_KEY(variables.staffId),
      });
      queryClient.invalidateQueries({
        queryKey: STAFF_EXCLUDED_ST_KEY(variables.staffId),
      });
    },
    onError: (error) => handleApiError(error, "設定"),
  });
}
