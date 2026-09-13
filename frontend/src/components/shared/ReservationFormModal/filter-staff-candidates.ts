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

/** Confirmed type/date orphan: show reason and block submit until clear/reselect. */
export const STAFF_ORPHAN_REASON_MESSAGE =
  "この条件では指定できない担当者です。解除するか、対応可能な担当者を選び直してください。";

export interface StaffSelectionEligibilityInput {
  doctorId: string;
  eligibleOptionIds: ReadonlySet<string>;
  nameById: ReadonlyMap<string, string>;
  /** True when filters' candidate queries have defined success data. */
  candidatesSettled: boolean;
  /** True when any candidate query for the current filters failed. */
  hasQueryError: boolean;
}

export interface StaffSelectionEligibility {
  isConfirmedOrphan: boolean;
  /** Name for SearchableSelect fallback when value is options-orphan. */
  displayLabel: string | undefined;
  reasonMessage: string | null;
}

/**
 * Resolve display/eligibility for a retained doctor id under type/date filters.
 * Loading and query error must not be treated as confirmed orphan.
 */
export function resolveStaffSelectionEligibility(
  input: StaffSelectionEligibilityInput,
): StaffSelectionEligibility {
  const doctorId = input.doctorId.trim();
  const hasDoctor = doctorId.length > 0;
  const displayLabel = hasDoctor ? input.nameById.get(doctorId) : undefined;
  const isEligible = hasDoctor && input.eligibleOptionIds.has(doctorId);
  const isConfirmedOrphan =
    hasDoctor && input.candidatesSettled && !input.hasQueryError && !isEligible;

  return {
    isConfirmedOrphan,
    displayLabel,
    reasonMessage: isConfirmedOrphan ? STAFF_ORPHAN_REASON_MESSAGE : null,
  };
}
