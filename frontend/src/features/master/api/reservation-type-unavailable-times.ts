import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationTypeUnavailableTime } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types（models.ts から導出）
// ─────────────────────────────────────────────────

export type CreateUnavailableTimeRequest = Omit<
  ReservationTypeUnavailableTime,
  "id" | "clinic_id" | "reservation_type_id" | "created_at" | "updated_at"
>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const unavailableTimesKey = (reservationTypeId: string) => [
  "masters",
  "reservation-types",
  reservationTypeId,
  "unavailable-times",
] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function getUnavailableTimes(
  clinicId: string,
  reservationTypeId: string,
): Promise<ReservationTypeUnavailableTime[]> {
  const { data } = await axios.get<{ data: ReservationTypeUnavailableTime[] }>(
    `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/unavailable-times`,
  );
  return data.data;
}

async function createUnavailableTime(
  clinicId: string,
  reservationTypeId: string,
  req: CreateUnavailableTimeRequest,
): Promise<void> {
  await axios.post(
    `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/unavailable-times`,
    req,
  );
}

async function deleteUnavailableTime(
  clinicId: string,
  reservationTypeId: string,
  id: number,
): Promise<void> {
  await axios.delete(
    `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/unavailable-times/${id}`,
  );
}

// ─────────────────────────────────────────────────
// Query hooks
// ─────────────────────────────────────────────────

export function useGetUnavailableTimes(clinicId: string, reservationTypeId: string) {
  return useQuery({
    queryKey: unavailableTimesKey(reservationTypeId),
    queryFn: () => getUnavailableTimes(clinicId, reservationTypeId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateUnavailableTime(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateUnavailableTimeRequest) =>
      createUnavailableTime(clinicId, reservationTypeId, req),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: unavailableTimesKey(reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "予約不可時間の作成"),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useDeleteUnavailableTime(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteUnavailableTime(clinicId, reservationTypeId, id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: unavailableTimesKey(reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "予約不可時間の削除"),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
