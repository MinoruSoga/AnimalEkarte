import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationTypeOccupation as ReservationTypeOccupationRaw } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Transform function
// ─────────────────────────────────────────────────

function transformReservationTypeOccupation(
  raw: ReservationTypeOccupationRaw,
): ReservationTypeOccupation {
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
// Request types
// ─────────────────────────────────────────────────

export type LinkOccupationRequest = { occupation_id: number };
export type UnlinkOccupationRequest = { id: number };

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const occupationsKey = (reservationTypeId: string) => [
  "masters",
  "reservation-types",
  reservationTypeId,
  "occupations",
] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function getReservationTypeOccupations(
  clinicId: string,
  reservationTypeId: string,
): Promise<ReservationTypeOccupation[]> {
  const { data } = await axios.get<{ data: ReservationTypeOccupationRaw[] }>(
    `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/occupations`,
  );
  return data.data.map(transformReservationTypeOccupation);
}

async function linkOccupation(
  clinicId: string,
  reservationTypeId: string,
  occupationId: number,
): Promise<void> {
  await axios.post(
    `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/occupations`,
    { occupation_id: occupationId },
  );
}

async function unlinkOccupation(
  clinicId: string,
  reservationTypeId: string,
  id: number,
): Promise<void> {
  await axios.delete(
    `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/occupations/${id}`,
  );
}

// ─────────────────────────────────────────────────
// Query hooks
// ─────────────────────────────────────────────────

export function useGetReservationTypeOccupations(clinicId: string, reservationTypeId: string) {
  return useQuery({
    queryKey: occupationsKey(reservationTypeId),
    queryFn: () => getReservationTypeOccupations(clinicId, reservationTypeId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useLinkOccupation(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (occupationId: number) => linkOccupation(clinicId, reservationTypeId, occupationId),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: occupationsKey(reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "職種の紐付け"),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useUnlinkOccupation(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unlinkOccupation(clinicId, reservationTypeId, id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: occupationsKey(reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "職種の紐付け解除"),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
