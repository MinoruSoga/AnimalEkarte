import UserRound from "lucide-react/dist/esm/icons/user-round";
import {
  CONDITIONS_NO_EMPTY,
  type ActiveFilter,
  type FilterProperty,
} from "@/components/shared/PropertyFilter/types";
import { MASTER_STATUS_FILTER } from "../constants/styles";
import type { Occupation } from "../api/occupations";
import type { PermissionGroup } from "../api/permission-groups";
import type { CreateStaffRequest, Staff, UpdateStaffRequest } from "../api/staffs";
import type { StaffFormData } from "../lib/staff-side-panel-model";
import type { StaffType } from "@/types/generated/models";
import { normalizeKana } from "@/lib/normalize-kana";

export function buildStaffIds(staffs: Staff[] | undefined): string[] {
  return (staffs ?? []).map((staff) => staff.id);
}

export function buildGroupsByStaffId({
  staffGroupMap,
  groups,
}: {
  staffGroupMap: Map<string, string[]> | undefined;
  groups: PermissionGroup[];
}): Map<string, PermissionGroup[]> {
  const map = new Map<string, PermissionGroup[]>();
  if (!staffGroupMap) return map;

  for (const [staffId, groupIds] of staffGroupMap.entries()) {
    const groupIdSet = new Set(groupIds);
    map.set(
      staffId,
      groups.filter((group) => groupIdSet.has(group.id)),
    );
  }
  return map;
}

export function buildStaffFilterProperties(occupations: Occupation[]): FilterProperty[] {
  return [
    MASTER_STATUS_FILTER,
    {
      key: "occupationId",
      label: "職種",
      type: "select",
      icon: UserRound,
      conditions: CONDITIONS_NO_EMPTY,
      options: occupations
        .filter((occupation) => occupation.isActive)
        .map((occupation) => ({ value: occupation.id, label: occupation.name })),
    },
  ];
}

export function searchStaff(staff: Staff, lowerSearchTerm: string): boolean {
  return (
    normalizeKana(staff.name).toLowerCase().includes(lowerSearchTerm) ||
    normalizeKana(staff.occupationName ?? "")
      .toLowerCase()
      .includes(lowerSearchTerm)
  );
}

export function filterStaffByMasterFilters(staff: Staff, filters: ActiveFilter[]): boolean {
  for (const filter of filters) {
    if (filter.key === "status" && typeof filter.value === "string") {
      const want = filter.value === "active";
      if (filter.condition === "is" && staff.isActive !== want) return false;
      if (filter.condition === "is_not" && staff.isActive === want) return false;
    }

    if (filter.key === "occupationId" && typeof filter.value === "string") {
      if (filter.condition === "is" && staff.occupationId !== filter.value) return false;
      if (filter.condition === "is_not" && staff.occupationId === filter.value) return false;
    }
  }
  return true;
}

export function buildStaffCreateRequest(data: StaffFormData): CreateStaffRequest {
  return {
    name: data.name,
    email: data.email,
    password: data.password,
    license_number: data.licenseNumber || undefined,
    occupation_id: data.jobTitleId ?? undefined,
    // FE6-2: staffType は Select（StaffType の値域のみ）で選択されるため実行時は常に安全。
    staff_type: data.staffType as StaffType,
    reservation_display_name: data.reservationDisplayName || undefined,
    reservation_visible: data.reservationVisible,
    reservation_comment: data.reservationComment || undefined,
    reservation_image_url: data.reservationImageUrl || undefined,
  };
}

export function buildStaffUpdateRequest(data: StaffFormData): UpdateStaffRequest {
  return {
    name: data.name,
    license_number: data.licenseNumber || undefined,
    is_active: data.isActive,
    occupation_id: data.jobTitleId ?? undefined,
    password: data.password || undefined,
    // FE6-2: staffType は Select（StaffType の値域のみ）で選択されるため実行時は常に安全。
    staff_type: data.staffType as StaffType,
    reservation_display_name: data.reservationDisplayName || undefined,
    reservation_visible: data.reservationVisible,
    reservation_comment: data.reservationComment || undefined,
    reservation_image_url: data.reservationImageUrl || undefined,
  };
}
