import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationTypeOccupation } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const occupationsKey = (reservationTypeId: string) => [
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
  const { data } = await axios.get<{ data: ReservationTypeOccupation[] }>(
    `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/occupations`,
  );
  return data.data;
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
