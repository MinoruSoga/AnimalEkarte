import type { VisitType, ReservationStatus } from "@/types";
export type {
  Appointment,
  ReservationStatus,
  CalendarView,
  VisitType,
  ReservationType,
  NavigationState,
  Pet,
} from "@/types";
export {
  CALENDAR_VIEW_VALUES,
  RESERVATION_STATUS_VALUES,
  RESERVATION_STATUS_LABELS,
  RESERVATION_TYPE_VALUES,
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
}

