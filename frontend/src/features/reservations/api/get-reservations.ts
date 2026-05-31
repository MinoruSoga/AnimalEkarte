import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformReservation } from "./transforms";
import type { Reservation } from "./transforms";
import type { Reservation as BackendReservation } from "@/types/generated/models";

interface ReservationsListResponse {
  data: BackendReservation[];
  total: number;
  page: number;
  limit: number;
}

export interface ReservationFilters {
  date?: string;
  status?: string;
  source?: string;
  petId?: string;
  ownerId?: string;
  enabled?: boolean;
}

function buildReservationParams(filters?: ReservationFilters): Record<string, string | number> {
  const params: Record<string, string | number> = { page: 1, limit: 100 };
  if (filters?.date) params.date = filters.date;
  if (filters?.status) params.status = filters.status;
  if (filters?.source) params.source = filters.source;
  if (filters?.petId) params.pet_id = filters.petId;
  if (filters?.ownerId) params.owner_id = filters.ownerId;
  return params;
}

export const getReservations = async (filters?: ReservationFilters): Promise<Reservation[]> => {
  const { data } = await axios.get<ReservationsListResponse>("/v1/reservations", {
    params: buildReservationParams(filters),
  });
  return data.data.map(transformReservation);
};

export const useGetReservations = (filters?: ReservationFilters) => {
  return useQuery({
    queryKey: ["reservations", filters],
    queryFn: () => getReservations(filters),
    enabled: filters?.enabled ?? true,
    staleTime: QUERY_STALE_TIMES.REALTIME, // 予約一覧は高頻度変更
    gcTime: QUERY_GC_TIMES.SHORT,
  });
};
