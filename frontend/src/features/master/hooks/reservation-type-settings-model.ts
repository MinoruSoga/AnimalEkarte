import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import type {
  CreateReservationTypeGroupRequest,
  UpdateReservationTypeGroupRequest,
} from "../api/reservation-type-groups";
import type {
  CreateReservationTypeRequest,
  UpdateReservationTypeRequest,
  ReservationType,
} from "../api/reservation-types";
import type { ReservationTypeGroup } from "../api/reservation-type-groups";
import type { GroupFormData } from "../components/ReservationTypeGroupSidePanel";
import type { CategoryFormData } from "../components/ReservationTypeSidePanel";
import { normalizeKana } from "@/lib/normalize-kana";

export function getActiveReservationTypeGroupOptions(groups: ReservationTypeGroup[]) {
  return groups
    .filter((group) => group.isActive)
    .map((group) => ({ id: group.id, name: group.name, color: group.color }));
}

export function matchesReservationTypeSearch(item: ReservationType, term: string): boolean {
  return normalizeKana(item.name).toLowerCase().includes(normalizeKana(term).toLowerCase());
}

export function matchesReservationTypeStatusFilter(
  item: ReservationType,
  filters: ActiveFilter[],
): boolean {
  for (const f of filters) {
    if (f.key === "status" && typeof f.value === "string") {
      const want = f.value === "active";
      if (f.condition === "is" ? item.isActive !== want : item.isActive === want) return false;
    }
  }
  return true;
}

export function buildReservationTypeGroupCreateRequest(
  data: GroupFormData,
): CreateReservationTypeGroupRequest {
  return {
    name: data.name,
    color: data.color || undefined,
    is_active: data.isActive,
  };
}

export function buildReservationTypeGroupUpdateRequest(
  data: GroupFormData,
): UpdateReservationTypeGroupRequest {
  return buildReservationTypeGroupCreateRequest(data);
}

export function buildReservationTypeCreateRequest(
  data: CategoryFormData,
): CreateReservationTypeRequest {
  return {
    name: data.name,
    description: data.description || undefined,
    is_active: true,
    group_id: data.groupId ? Number(data.groupId) : undefined,
    reservation_display_name: data.reservationDisplayName || undefined,
    duration_minutes: data.durationMinutes,
    short_name: data.shortName || undefined,
    reservation_visible: data.reservationVisible,
    reservation_comment: data.reservationComment || undefined,
    reservation_image_url: data.reservationImageUrl || undefined,
    show_short_name: data.showShortName,
    reservation_day_option: data.reservationDayOption as "none" | "weekday" | "saturday" | "anyday",
    is_internal: data.isInternal,
  };
}

export function buildReservationTypeUpdateRequest(
  data: CategoryFormData,
): UpdateReservationTypeRequest {
  return {
    ...buildReservationTypeCreateRequest(data),
    is_active: data.isActive,
  };
}
