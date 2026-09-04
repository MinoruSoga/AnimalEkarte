import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { formatJSTDate } from "@/lib/jst-date";
import { queryKeys } from "@/lib/query-keys";
import type { VaccinationRecord } from "@/types";
import type { Vaccination } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export interface CreateVaccinationRequest {
  medical_record_id?: number | null;
  pet_id?: number | null;
  vaccine_id: number;
  date: string;
  doctor_id?: number | null;
  next_date?: string | null;
  next_schedule_type?: string;
  lot1?: string;
  lot2?: string;
  lot3?: string;
  lot4?: string;
  supplemental?: string;
  remarks?: string;
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

/** SD-19: 絶対時刻の UTC 切り出しではなく JST 壁日付 YYYY-MM-DD にする。 */
function toJSTDateOnly(iso?: string | null): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (isNaN(d.getTime())) return "";
  return formatJSTDate(d);
}

function transformVaccination(data: Vaccination): VaccinationRecord {
  return {
    id: String(data.id ?? 0),
    petId: data.pet_id ? String(data.pet_id) : undefined,
    medicalRecordId: data.medical_record_id != null ? String(data.medical_record_id) : undefined,
    ownerName: data.pet?.owner?.name ?? "",
    petName: data.pet?.name ?? "",
    vaccineId: String(data.vaccine_id ?? 0),
    vaccineName: data.vaccine?.name ?? "",
    doctor: data.doctor?.name ?? "",
    date: toJSTDateOnly(data.date),
    nextDate: toJSTDateOnly(data.next_date),
    nextScheduleType: data.next_schedule_type || undefined,
    lot1: data.lot1 || undefined,
    lot2: data.lot2 || undefined,
    lot3: data.lot3 || undefined,
    lot4: data.lot4 || undefined,
    supplemental: data.supplemental || undefined,
    remarks: data.remarks || undefined,
  };
}

// ─────────────────────────────────────────────────
// Hook
// ─────────────────────────────────────────────────

/**
 * Mutation hook to create a vaccination record.
 * Used cross-feature from medical-records to register vaccinations.
 * Query key matches features/vaccinations for cache sharing.
 */
export function useCreateVaccination() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateVaccinationRequest): Promise<VaccinationRecord> => {
      const { data } = await axios.post<Vaccination>("/v1/vaccinations", req);
      return transformVaccination(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.vaccinations.all() });
    },
    onError: (error) => handleApiError(error, "ワクチン接種登録"),
  });
}
