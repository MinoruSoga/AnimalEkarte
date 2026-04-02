import type { Hospitalization as FrontendHospitalization } from "@/types";
import type { BackendHospitalization } from "./types";
import { formatDate } from "@/utils/format/date";

export const transformHospitalization = (
  hosp: BackendHospitalization
): FrontendHospitalization => {
  const statusMap: Record<string, FrontendHospitalization["status"]> = {
    admitted: "入院中",
    discharged: "退院済",
    reserved: "予約",
  };
  const typeMap: Record<string, FrontendHospitalization["hospitalizationType"]> = {
    hospitalization: "入院",
    hotel: "ホテル",
  };

  return {
    id: String(hosp.id ?? 0),
    hospitalizationNo: String(hosp.id ?? ""),
    ownerName: hosp.owner?.owner_name ?? "",
    petName: hosp.pet?.name ?? "",
    species: hosp.pet?.animal_species?.name ?? "",
    hospitalizationType: typeMap[hosp.hospitalization_type] ?? "入院",
    startDate: formatDate(hosp.start_date),
    endDate: formatDate(hosp.end_date),
    status: statusMap[hosp.status] ?? "予約",
    cageId: hosp.cage_id ? String(hosp.cage_id) : undefined,
    petId: hosp.pet?.id ? String(hosp.pet.id) : undefined,
    petIsDeceased: hosp.pet?.status === "deceased",
  };
};
