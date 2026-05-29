import type {
  CreateStaffRequest,
  UpdateStaffRequest,
} from "../api/staffs";
import type { StaffFormData } from "../components/StaffSidePanelModel";

export function buildStaffCreateRequest(data: StaffFormData): CreateStaffRequest {
  return {
    name: data.name,
    email: data.email,
    password: data.password,
    license_number: data.licenseNumber || undefined,
    occupation_id: data.jobTitleId ?? undefined,
    staff_type: data.staffType,
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
    staff_type: data.staffType,
    reservation_display_name: data.reservationDisplayName || undefined,
    reservation_visible: data.reservationVisible,
    reservation_comment: data.reservationComment || undefined,
    reservation_image_url: data.reservationImageUrl || undefined,
  };
}
