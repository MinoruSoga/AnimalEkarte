import { jstWallDateToISOString } from "@/lib/jst-date";
import { transformReservation } from "@/lib/transforms/reservation";
import type { Reservation } from "@/lib/transforms/reservation";
import type { CreateReservationRequest } from "./types";

// R-F2-S12: transformReservation / Reservation は lib/transforms/reservation.ts へ移設。
// ここは feature 内既存 import (`./transforms`) を壊さないための re-export。
export { transformReservation };
export type { Reservation };

export const transformToCreateRequest = (
  data: Partial<Reservation>,
  petId: string,
  ownerId: string,
): CreateReservationRequest => {
  return {
    pet_id: Number(petId),
    owner_id: Number(ownerId),
    start_time: data.start ? jstWallDateToISOString(data.start) : "",
    end_time: data.end ? jstWallDateToISOString(data.end) : "",
    visit_type: data.visitType ?? "first",
    reservation_type_id: Number(data.type ?? 0),
    doctor_id: data.doctor ? Number(data.doctor) : undefined,
    is_designated: data.isDesignated ?? false,
    status: data.status,
    notes: data.notes,
    source: data.source,
    reservation_route: data.reservationRoute ?? undefined,
  };
};
