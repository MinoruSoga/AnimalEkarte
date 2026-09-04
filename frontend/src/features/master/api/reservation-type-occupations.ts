import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { ReservationTypeOccupation as ReservationTypeOccupationRaw } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Transform function
// ─────────────────────────────────────────────────

function transformReservationTypeOccupation(raw: ReservationTypeOccupationRaw) {
  return {
    id: raw.id,
    clinicId: raw.clinic_id,
    reservationTypeId: raw.reservation_type_id,
    occupationId: raw.occupation_id,
    occupation: raw.occupation ? { id: raw.occupation.id, name: raw.occupation.name } : undefined,
    createdAt: raw.created_at,
  };
}

// ─────────────────────────────────────────────────
// Domain type
// ─────────────────────────────────────────────────

export type ReservationTypeOccupation = ReturnType<typeof transformReservationTypeOccupation>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const occupationsKey = (clinicId: string, reservationTypeId: string) =>
  queryKeys.masters.reservationTypeSubResource(clinicId, reservationTypeId, "occupations");

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function getReservationTypeOccupations(
  reservationTypeId: string,
): Promise<ReservationTypeOccupation[]> {
  const { data } = await axios.get<{ data: ReservationTypeOccupationRaw[] }>(
    `/v1/masters/reservation-types/${reservationTypeId}/occupations`,
  );
  return data.data.map(transformReservationTypeOccupation);
}

async function linkOccupation(reservationTypeId: string, occupationId: number): Promise<void> {
  await axios.post(`/v1/masters/reservation-types/${reservationTypeId}/occupations`, {
    occupation_id: occupationId,
  });
}

async function unlinkOccupation(reservationTypeId: string, id: number): Promise<void> {
  await axios.delete(`/v1/masters/reservation-types/${reservationTypeId}/occupations/${id}`);
}

// ─────────────────────────────────────────────────
// Query hooks
// ─────────────────────────────────────────────────

export function useGetReservationTypeOccupations(clinicId: string, reservationTypeId: string) {
  return useQuery({
    queryKey: occupationsKey(clinicId, reservationTypeId),
    queryFn: () => getReservationTypeOccupations(reservationTypeId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useLinkOccupation(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (occupationId: number) => linkOccupation(reservationTypeId, occupationId),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: occupationsKey(clinicId, reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "職種の紐付け"),
  });
}

export function useUnlinkOccupation(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unlinkOccupation(reservationTypeId, id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: occupationsKey(clinicId, reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "職種の紐付け解除"),
  });
}
