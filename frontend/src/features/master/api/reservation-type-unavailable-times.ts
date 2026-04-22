import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

// ─────────────────────────────────────────────────
// Raw backend response types
// ─────────────────────────────────────────────────

interface ReservationTypeUnavailableTimeRaw {
  id: number;
  clinic_id: number;
  reservation_type_id: number;
  unavailable_type: string;
  day_of_week?: number;
  specific_date?: string;
  start_time: string;
  end_time: string;
  created_at: string;
  updated_at: string;
}

// ─────────────────────────────────────────────────
// Transform function
// ─────────────────────────────────────────────────

function transformUnavailableTime(
  raw: ReservationTypeUnavailableTimeRaw,
): ReservationTypeUnavailableTime {
  return {
    id: raw.id,
    clinicId: raw.clinic_id,
    reservationTypeId: raw.reservation_type_id,
    unavailableType: raw.unavailable_type,
    dayOfWeek: raw.day_of_week ?? undefined,
    specificDate: raw.specific_date ?? undefined,
    startTime: raw.start_time,
    endTime: raw.end_time,
    createdAt: raw.created_at,
    updatedAt: raw.updated_at,
  };
}

// ─────────────────────────────────────────────────
// Domain type
// ─────────────────────────────────────────────────

export type ReservationTypeUnavailableTime = ReturnType<typeof transformUnavailableTime>;

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export type CreateUnavailableTimeRequest = Omit<
  ReservationTypeUnavailableTime,
  "id" | "clinicId" | "reservationTypeId" | "createdAt" | "updatedAt"
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
  const { data } = await axios.get<{ data: ReservationTypeUnavailableTimeRaw[] }>(
    `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/unavailable-times`,
  );
  return data.data.map(transformUnavailableTime);
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
