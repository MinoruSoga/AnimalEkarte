/**
 * TASK-021 Stage B: affirmative capability filter for in-clinic reservation staff candidates.
 * Missing / pending capability metadata is fail-closed (never assume all staff are capable).
 */

export interface StaffCandidateLike {
  id: string | number;
}

export interface ReservationStaffCapabilityLike {
  id: number;
  name?: string;
  /** Affirmative capability surface (preferred). */
  capable_courses?: ReadonlyArray<{ id: number; name?: string }>;
}

/**
 * Filter staff candidates for a selected reservation type using positive capabilities.
 *
 * @returns filtered candidates, or empty when metadata is pending and a type is selected (fail-closed).
 */
export function filterStaffCandidatesByCapability<T extends StaffCandidateLike>(
  candidates: readonly T[],
  selectedReservationTypeId: string | null,
  reservationStaffMap: Map<string, ReservationStaffCapabilityLike> | undefined,
): T[] {
  if (selectedReservationTypeId === null) {
    return [...candidates];
  }
  // Pending metadata: do not offer any staff for the type (fail-closed).
  if (reservationStaffMap === undefined) {
    return [];
  }
  return candidates.filter((staff) => {
    const reservationStaff = reservationStaffMap.get(String(staff.id));
    if (reservationStaff === undefined) {
      // Staff missing from reservation-staffs payload: not capable (fail-closed).
      return false;
    }
    const capable = reservationStaff.capable_courses ?? [];
    return capable.some((course) => String(course.id) === selectedReservationTypeId);
  });
}
