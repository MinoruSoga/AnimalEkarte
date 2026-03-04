import type { Hospitalization } from "@/types";
import type { BackendHospitalization } from "./types";

export const transformHospitalization = (
  hosp: BackendHospitalization
): Hospitalization => {
  return {
    id: hosp.id,
    hospitalizationNo: hosp.hospitalization_no,
    ownerName: hosp.owner?.name ?? "",
    petName: hosp.pet?.name ?? "",
    species: hosp.pet?.species ?? "",
    hospitalizationType: (hosp.type as "入院" | "ホテル") || "入院",
    startDate: hosp.start_date,
    endDate: hosp.end_date,
    status: (hosp.status as "入院中" | "退院済" | "予約" | "一時帰宅") || "予約",
    cageId: hosp.cage_id,
  };
};
