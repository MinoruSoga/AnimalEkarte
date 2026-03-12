import type { Pet, VisitType, ReservationStatus } from "@/types";
export type {
  ReservationAppointment,
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
}

/** Default values for a new reservation form */
export function createDefaultReservationFormData(): ReservationFormData {
  const defaultStart = new Date();
  defaultStart.setHours(10, 0, 0, 0);
  const defaultEnd = new Date(defaultStart);
  defaultEnd.setHours(11, 0, 0, 0);

  return {
    start: defaultStart,
    end: defaultEnd,
    visitType: "first",
    type: "診療",
    doctor: "医師A",
    isDesignated: false,
    status: "confirmed",
  };
}

export type ReservationFormSaveHandler = (
  data: ReservationFormData,
  selectedPets: Pet[]
) => void;
