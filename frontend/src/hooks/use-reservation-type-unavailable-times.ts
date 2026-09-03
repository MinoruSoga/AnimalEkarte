import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { ReservationTypeUnavailableTime as ReservationTypeUnavailableTimeRaw } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Transform function
// ─────────────────────────────────────────────────

function transformUnavailableTime(raw: ReservationTypeUnavailableTimeRaw) {
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
// Query keys
// ─────────────────────────────────────────────────

// master/api/reservation-type-unavailable-times.ts の作成・削除 mutation も同じキーで
// invalidate する必要があるため export する（FE6-16: cross-feature 化に伴う移設）。
export const unavailableTimesKey = (clinicId: string | null, reservationTypeId: string) =>
  queryKeys.masters.reservationTypeSubResource(clinicId, reservationTypeId, "unavailable-times");

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function getUnavailableTimes(
  reservationTypeId: string,
): Promise<ReservationTypeUnavailableTime[]> {
  const { data } = await axios.get<
    ReservationTypeUnavailableTimeRaw[] | { data: ReservationTypeUnavailableTimeRaw[] }
  >(`/v1/masters/reservation-types/${reservationTypeId}/unavailable-times`);
  const items = Array.isArray(data) ? data : data.data;
  return items.map(transformUnavailableTime);
}

// ─────────────────────────────────────────────────
// Query hooks
// ─────────────────────────────────────────────────

export function useGetUnavailableTimes(clinicId: string | null, reservationTypeId: string) {
  return useQuery({
    queryKey: unavailableTimesKey(clinicId, reservationTypeId),
    queryFn: () => getUnavailableTimes(reservationTypeId),
    enabled: clinicId !== null && reservationTypeId !== "",
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
