import { jstWallDateToISOString } from "@/lib/jst-date";
import { transformReservation } from "@/lib/transforms/reservation";
import type { Reservation } from "@/lib/transforms/reservation";
import type { CreateReservationRequest } from "./types";

// R-F2-S12: transformReservation / Reservation は lib/transforms/reservation.ts へ移設。
// ここは feature 内既存 import (`./transforms`) を壊さないための re-export。
export { transformReservation };
export type { Reservation };

/** Create-only: empty/"0" omit; positive decimal IDs kept; invalid strings fail closed. */
export const normalizeCreateDoctorID = (doctor: string | undefined): number | undefined => {
  if (doctor == null) return undefined;
  const trimmed = doctor.trim();
  if (trimmed === "" || trimmed === "0") return undefined;
  if (!/^[1-9]\d*$/.test(trimmed)) {
    throw new Error(`Invalid doctor id: ${doctor}`);
  }
  const parsed = Number(trimmed);
  if (!Number.isSafeInteger(parsed)) {
    throw new Error(`Invalid doctor id: ${doctor}`);
  }
  return parsed;
};

export const transformToCreateRequest = (
  data: Partial<Reservation>,
  petId: string,
  ownerId: string,
): CreateReservationRequest => {
  const doctorID = normalizeCreateDoctorID(data.doctor);
  return {
    pet_id: Number(petId),
    owner_id: Number(ownerId),
    start_time: data.start ? jstWallDateToISOString(data.start) : "",
    end_time: data.end ? jstWallDateToISOString(data.end) : "",
    visit_type: data.visitType ?? "first",
    reservation_type_id: Number(data.type ?? 0),
    ...(doctorID !== undefined ? { doctor_id: doctorID } : {}),
    is_designated: data.isDesignated ?? false,
    status: data.status,
    notes: data.notes,
    source: data.source,
    reservation_route: data.reservationRoute ?? undefined,
  };
};
