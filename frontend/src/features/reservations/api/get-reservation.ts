import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { ReservationAppointment } from "@/types";
import { transformReservation } from "./transforms";
import type { BackendReservation } from "./types";

interface ReservationListResponse {
  data: BackendReservation[];
  total: number;
  page: number;
  limit: number;
}

export const getReservation = async (
  id: string
): Promise<ReservationAppointment> => {
  const { data } = await axios.get<BackendReservation>(
    `/v1/reservations/${id}`
  );
  return transformReservation(data);
};

export const useGetReservation = (id: string) => {
  return useQuery({
    queryKey: ["reservation", id],
    queryFn: () => getReservation(id),
    enabled: !!id,
  });
};

// Fetch reservations by pet ID
export const getReservationsByPetId = async (
  petId: string
): Promise<ReservationAppointment[]> => {
  const { data } = await axios.get<ReservationListResponse>("/v1/reservations", {
    params: { pet_id: petId },
  });
  return data.data.map(transformReservation);
};

export const useGetReservationsByPetId = (petId: string) => {
  return useQuery({
    queryKey: ["reservations", "pet", petId],
    queryFn: () => getReservationsByPetId(petId),
    enabled: !!petId,
  });
};

// Fetch reservations by owner ID
export const getReservationsByOwnerId = async (
  ownerId: string
): Promise<ReservationAppointment[]> => {
  const { data } = await axios.get<ReservationListResponse>("/v1/reservations", {
    params: { owner_id: ownerId },
  });
  return data.data.map(transformReservation);
};

export const useGetReservationsByOwnerId = (ownerId: string) => {
  return useQuery({
    queryKey: ["reservations", "owner", ownerId],
    queryFn: () => getReservationsByOwnerId(ownerId),
    enabled: !!ownerId,
  });
};
