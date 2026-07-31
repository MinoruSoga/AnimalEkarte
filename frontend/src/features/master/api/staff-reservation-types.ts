import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";

// ─────────────────────────────────────────────────
// Staff Capable Reservation Types API (TASK-021 Stage B SoT)
// ─────────────────────────────────────────────────

const STAFF_CAPABLE_ST_KEY = (staffId: string) =>
  queryKeys.staffs.subResource(staffId, "capable-reservation-types");

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
      // In-clinic reservation form candidates (reservation-staffs).
      queryClient.invalidateQueries({
        predicate: (query) =>
          Array.isArray(query.queryKey) &&
          query.queryKey[0] === "clinics" &&
          query.queryKey[2] === "reservation-staffs",
      });
    },
    onError: (error) => handleApiError(error, "設定"),
  });
}
