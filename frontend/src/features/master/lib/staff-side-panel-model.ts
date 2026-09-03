import type { Staff } from "../api/staffs";

export interface StaffFormData {
  name: string;
  jobTitleId: string | null;
  licenseNumber: string;
  isActive: boolean;
  email: string;
  password: string;
  staffType: string;
  reservationDisplayName: string;
  reservationVisible: boolean;
  reservationComment: string;
  reservationImageUrl: string;
}

export function staffToFormData(item: Staff | null): StaffFormData {
  return {
    name: item?.name ?? "",
    jobTitleId: item?.occupationId ?? null,
    licenseNumber: item?.licenseNumber ?? "",
    isActive: item?.isActive ?? true,
    email: item?.email ?? "",
    password: "",
    staffType: item?.staffType ?? "doctor",
    reservationDisplayName: item?.reservationDisplayName ?? "",
    reservationVisible: item?.reservationVisible ?? true,
    reservationComment: item?.reservationComment ?? "",
    reservationImageUrl: item?.reservationImageUrl ?? "",
  };
}
