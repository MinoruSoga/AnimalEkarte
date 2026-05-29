import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
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

const availableSlotsKey = (clinicId: string, reservationTypeId: string) => [
  "masters",
  "clinics",
  clinicId,
  "reservation-types",
  reservationTypeId,
  "available-slots",
] as const;

async function getAvailableSlots(
  clinicId: string,
  reservationTypeId: string,
): Promise<ReservationTypeAvailableSlot[]> {
  const { data } = await axios.get<ReservationTypeAvailableSlotRaw[] | { data: ReservationTypeAvailableSlotRaw[] }>(
    `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/available-slots`,
  );
  const items = Array.isArray(data) ? data : data.data;
  return items.map(transformAvailableSlot);
}

async function createAvailableSlot(
  clinicId: string,
  reservationTypeId: string,
  req: CreateAvailableSlotRequest,
): Promise<void> {
  await axios.post(
    `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/available-slots`,
    req,
  );
}

async function deleteAvailableSlot(
  clinicId: string,
  reservationTypeId: string,
  id: number,
): Promise<void> {
  await axios.delete(
    `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/available-slots/${id}`,
  );
}

export function useGetAvailableSlots(clinicId: string, reservationTypeId: string) {
  return useQuery({
    queryKey: availableSlotsKey(clinicId, reservationTypeId),
    queryFn: () => getAvailableSlots(clinicId, reservationTypeId),
    enabled: clinicId !== "" && reservationTypeId !== "",
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateAvailableSlot(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateAvailableSlotRequest) =>
      createAvailableSlot(clinicId, reservationTypeId, req),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: availableSlotsKey(clinicId, reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "予約可能枠の作成"),
  });
}

export function useDeleteAvailableSlot(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteAvailableSlot(clinicId, reservationTypeId, id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: availableSlotsKey(clinicId, reservationTypeId) }),
    onError: (error: unknown) => handleApiError(error, "予約可能枠の削除"),
  });
}
