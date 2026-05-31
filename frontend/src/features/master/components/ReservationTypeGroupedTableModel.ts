import type { ReservationType } from "../api/reservation-types";

export const UNCATEGORIZED_RESERVATION_TYPE_GROUP_ID = "__uncategorized__";

export interface ReservationTypesByGroup {
  map: Map<string, ReservationType[]>;
  uncategorized: ReservationType[];
}

export function groupReservationTypesByGroupId(
  reservationTypes: ReservationType[],
): ReservationTypesByGroup {
  const map = new Map<string, ReservationType[]>();
  const uncategorized: ReservationType[] = [];

  for (const reservationType of reservationTypes) {
    if (!reservationType.groupId) {
      uncategorized.push(reservationType);
      continue;
    }

    const groupReservationTypes = map.get(reservationType.groupId) ?? [];
    groupReservationTypes.push(reservationType);
    map.set(reservationType.groupId, groupReservationTypes);
  }

  return { map, uncategorized };
}
