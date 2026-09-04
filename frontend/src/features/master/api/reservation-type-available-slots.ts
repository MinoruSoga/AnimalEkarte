import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { ReservationTypeAvailableSlot as ReservationTypeAvailableSlotRaw } from "@/types/generated/models";

function transformAvailableSlot(raw: ReservationTypeAvailableSlotRaw) {
  return {
    id: raw.id,
    clinicId: raw.clinic_id,
    reservationTypeId: raw.reservation_type_id,
    availableType: raw.available_type,
    dayOfWeek: raw.day_of_week ?? undefined,
    specificDate: raw.specific_date ?? undefined,
    startTime: raw.start_time,
    isActive: raw.is_active,
    createdAt: raw.created_at,
    updatedAt: raw.updated_at,
  };
}

export type ReservationTypeAvailableSlot = ReturnType<typeof transformAvailableSlot>;

export type CreateAvailableSlotRequest = {
  available_type: string;
  start_time: string;
  is_active?: boolean;
  day_of_week?: number;
  specific_date?: string;
};

const availableSlotsKey = (clinicId: string, reservationTypeId: string) =>
  queryKeys.masters.reservationTypeSubResource(clinicId, reservationTypeId, "available-slots");

async function getAvailableSlots(
  reservationTypeId: string,
): Promise<ReservationTypeAvailableSlot[]> {
  const { data } = await axios.get<
    ReservationTypeAvailableSlotRaw[] | { data: ReservationTypeAvailableSlotRaw[] }
  >(`/v1/masters/reservation-types/${reservationTypeId}/available-slots`);
  const items = Array.isArray(data) ? data : data.data;
  return items.map(transformAvailableSlot);
}

async function createAvailableSlot(
  reservationTypeId: string,
  req: CreateAvailableSlotRequest,
): Promise<void> {
  await axios.post(`/v1/masters/reservation-types/${reservationTypeId}/available-slots`, req);
}

async function deleteAvailableSlot(reservationTypeId: string, id: number): Promise<void> {
  await axios.delete(`/v1/masters/reservation-types/${reservationTypeId}/available-slots/${id}`);
}

export function useGetAvailableSlots(clinicId: string, reservationTypeId: string) {
  return useQuery({
    queryKey: availableSlotsKey(clinicId, reservationTypeId),
    queryFn: () => getAvailableSlots(reservationTypeId),
    enabled: clinicId !== "" && reservationTypeId !== "",
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateAvailableSlot(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateAvailableSlotRequest) => createAvailableSlot(reservationTypeId, req),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: availableSlotsKey(clinicId, reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "予約可能枠の作成"),
  });
}

export function useDeleteAvailableSlot(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteAvailableSlot(reservationTypeId, id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: availableSlotsKey(clinicId, reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "予約可能枠の削除"),
  });
}
