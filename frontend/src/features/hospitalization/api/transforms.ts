import type { BackendHospitalization } from "./types";

type HospitalizationStatus = "入院中" | "退院済" | "予約" | "一時帰宅";
type HospitalizationType = "入院" | "ホテル";

export const transformHospitalization = (
  hosp: BackendHospitalization
) => {
  const statusMap: Record<string, HospitalizationStatus> = {
    admitted: "入院中",
    discharged: "退院済",
    reserved: "予約",
  };
  const typeMap: Record<string, HospitalizationType> = {
    hospitalization: "入院",
    hotel: "ホテル",
  };

  return {
    id: String(hosp.id ?? 0),
    hospitalizationNo: String(hosp.id ?? ""),
    ownerName: hosp.owner?.name ?? "",
    petName: hosp.pet?.name ?? "",
    species: hosp.pet?.animal_species?.name ?? "",
    hospitalizationType: (typeMap[hosp.hospitalization_type] ?? "入院") as HospitalizationType,
    startDate: hosp.start_date ? hosp.start_date.split("T")[0] : "",
    endDate: hosp.end_date ? hosp.end_date.split("T")[0] : "",
    status: (statusMap[hosp.status] ?? "予約") as HospitalizationStatus,
    cageId: hosp.cage_id ? String(hosp.cage_id) : undefined,
    petId: hosp.pet?.id ? String(hosp.pet.id) : undefined,
    doctorId: hosp.doctor_id ? String(hosp.doctor_id) : undefined,
    doctorName: hosp.doctor?.name ?? undefined,
    petIsDeceased: hosp.pet?.status === "deceased",
    memo: hosp.memo || undefined,
    ownerRequest: hosp.owner_request || undefined,
    staffNotes: hosp.staff_notes || undefined,
  };
};

export type Hospitalization = ReturnType<typeof transformHospitalization>;
