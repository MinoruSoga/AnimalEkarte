import type { VisitType, ReservationStatus } from "@/types";
import type { ReservationRoute } from "../constants/reservation-route";
export type {
  Reservation,
  ReservationStatus,
  CalendarView,
  VisitType,
  NavigationState,
  Pet,
} from "@/types";
export {
  CALENDAR_VIEW_VALUES,
  RESERVATION_STATUS_VALUES,
  RESERVATION_STATUS_LABELS,
} from "@/types";

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

export interface NewOwnerFormData {
  ownerName: string;
  phone: string;
  petName: string;
  chiefComplaint: string;
  animalSpeciesId: number;
}
