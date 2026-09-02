import type { StaffItem } from "@/hooks/use-staffs";
import type { Hospitalization } from "../api/transforms";

export type HospitalizationFormGate =
  | { kind: "new-pet-loading" }
  | { kind: "new-no-pet" }
  | { kind: "edit-loading" }
  | { kind: "edit-not-found" }
  | { kind: "edit-error"; retryRead?: () => void };

export function resolveHospitalizationFormGate(input: {
  hasSelectedPet: boolean;
  isEdit: boolean;
  petId: string | null;
  isReadLoading: boolean;
  isReadNotFound: boolean;
  isReadError: boolean;
  retryRead?: () => void;
}): HospitalizationFormGate | null {
  if (!input.hasSelectedPet && !input.isEdit && input.petId) return { kind: "new-pet-loading" };
  if (!input.hasSelectedPet && !input.isEdit) return { kind: "new-no-pet" };
  if (input.isEdit && input.isReadLoading) return { kind: "edit-loading" };
  if (input.isEdit && input.isReadNotFound) return { kind: "edit-not-found" };
  if (input.isEdit && input.isReadError) return { kind: "edit-error", retryRead: input.retryRead };
  return null;
}

export function toHospitalizationHistoryItems(records: readonly Hospitalization[]) {
  return records.map((record) => ({
    id: record.id,
    date: record.startDate,
    title: `${record.hospitalizationType}（${record.status}）`,
    subtitle: record.doctorName || record.memo,
  }));
}

export function selectHospitalizationDoctorStaffs(
  staffs: readonly StaffItem[],
  doctorId: string,
): StaffItem[] {
  return staffs.filter((staff) =>
    (staff.isActive && staff.staffType === "doctor") || staff.id === doctorId,
  );
}
