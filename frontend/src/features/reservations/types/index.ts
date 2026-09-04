import type { VisitType, ReservationStatus } from "@/types";
import type { ReservationRoute } from "../constants/reservation-route";
export type { Reservation, ReservationStatus, CalendarView, NavigationState, Pet } from "@/types";
export { CALENDAR_VIEW_VALUES, RESERVATION_STATUS_VALUES } from "@/types";

/**
 * Form state for creating/editing a reservation.
 * All fields are optional because:
 *  - New forms start with partial defaults
 *  - Edit forms are populated from existing data
 *  - Individual fields are updated independently during editing
 */
export interface ReservationFormData {
  id?: string;
  start?: Date;
  end?: Date;
  ownerName?: string;
  petName?: string;
  visitType?: VisitType;
  type?: string;
  doctor?: string;
  isDesignated?: boolean;
  status?: ReservationStatus;
  notes?: string;
  petId?: string;
  source?: "manual" | "line";
  reservationRoute?: ReservationRoute | null;
}

// FE6-17: 正本は src/types/reservation-form.ts へ移動。ここは re-export のみ。
export type { NewOwnerFormData } from "@/types/reservation-form";
