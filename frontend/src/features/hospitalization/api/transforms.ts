import type { BackendHospitalization } from "./types";

/** UI 表示用 status。未知 wire は「不明」（BUG-009 fail-closed。旧 fail-open は「予約」推測）。 */
type HospitalizationStatus = "入院中" | "退院済" | "予約" | "一時帰宅" | "不明";
type HospitalizationType = "入院" | "ホテル";

const KNOWN_STATUS_MAP: Record<string, Exclude<HospitalizationStatus, "不明">> = {
  admitted: "入院中",
  discharged: "退院済",
  reserved: "予約",
};

/**
 * HospitalizationResponse wire → UI list/detail view model.
 * treatment_plans / care_plan_items / daily_records は detail wire に無い
 * （専用 nested endpoint が正本）。
 */
export const transformHospitalization = (hosp: BackendHospitalization) => {
  const typeMap: Record<string, HospitalizationType> = {
    hospitalization: "入院",
    hotel: "ホテル",
  };

  const mappedStatus = KNOWN_STATUS_MAP[hosp.status];
  const status: HospitalizationStatus = mappedStatus ?? "不明";

  return {
    id: String(hosp.id ?? 0),
    hospitalizationNo: String(hosp.id ?? ""),
    ownerName: hosp.owner?.name ?? "",
    petName: hosp.pet?.name ?? "",
    species: hosp.pet?.animal_species?.name ?? "",
    hospitalizationType: (typeMap[hosp.hospitalization_type] ?? "入院") as HospitalizationType,
    startDate: hosp.start_date ? hosp.start_date.split("T")[0] : "",
    endDate: hosp.end_date ? hosp.end_date.split("T")[0] : "",
    status,
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
