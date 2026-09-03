import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { formatDate } from "@/lib/format/date";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import {
  transformMedicalRecord,
  type MedicalRecord,
} from "@/lib/transforms/medical-record";
import type { MedicalRecordResponse } from "@/types/generated/medicalrecord-responses";

/**
 * FE-RC-015: list / history query の実体。
 * wire→UI 変換は @/lib/transforms/medical-record が正本（followup3）。
 */

export type MedicalRecordSortKey = "date" | "owner_name" | "pet_name" | "status";

export interface MedicalRecordFilters {
  startDate?: string;
  endDate?: string;
  petId?: string;
  ownerId?: string;
  clinicIds?: string[];
  search?: string;
  status?: string;
  doctorId?: string;
  animalSpeciesId?: string;
  page?: number;
  limit?: number;
  sort?: MedicalRecordSortKey;
  order?: "asc" | "desc";
}

export type MedicalRecordsResult = {
  data: MedicalRecord[];
  total: number;
  page: number;
  limit: number;
};

interface MedicalRecordsListResponse {
  data: MedicalRecordResponse[];
  total: number;
  page: number;
  limit: number;
}

const DEFAULT_PAGE = 1;
const DEFAULT_LIMIT = 20;

export async function getMedicalRecords(
  filters?: MedicalRecordFilters,
): Promise<MedicalRecordsResult> {
  const params: Record<string, string | number> = {
    page: filters?.page ?? DEFAULT_PAGE,
    limit: filters?.limit ?? DEFAULT_LIMIT,
  };
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.petId) params.pet_id = filters.petId;
  if (filters?.ownerId) params.owner_id = filters.ownerId;
  if (filters?.clinicIds && filters.clinicIds.length > 1) {
    params.clinic_ids = filters.clinicIds.join(",");
  }
  if (filters?.search) params.search = filters.search;
  if (filters?.status) params.status = filters.status;
  if (filters?.doctorId) params.doctor_id = filters.doctorId;
  if (filters?.animalSpeciesId) params.animal_species_id = filters.animalSpeciesId;
  if (filters?.sort) params.sort = filters.sort;
  if (filters?.order) params.order = filters.order;

  const { data } = await axios.get<MedicalRecordsListResponse>("/v1/medical-records", {
    params,
  });
  return {
    data: data.data.map(transformMedicalRecord),
    total: data.total,
    page: data.page,
    limit: data.limit,
  };
}

export function useGetMedicalRecords(filters?: MedicalRecordFilters) {
  return useQuery({
    queryKey: queryKeys.medicalRecords.list(filters),
    queryFn: () => getMedicalRecords(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}

/** FEAT-003: 問診履歴 UI 行（feature InterviewHistoryItem と同形）。 */
export interface MedicalRecordInterviewHistoryItem {
  id: string;
  date: string;
  author: string;
  type: string;
  title: string;
  content: string;
}

function transformToHistoryItem(
  record: MedicalRecordResponse,
): MedicalRecordInterviewHistoryItem {
  const chiefComplaint = record.inquiry?.chief_complaint ?? "";
  const content = chiefComplaint || "（記録なし）";
  return {
    id: String(record.id ?? 0),
    date: formatDate(record.date),
    author: record.doctor?.name ?? "-",
    type: record.status === "finalized" ? "確定済" : "作成中",
    title: chiefComplaint || record.record_no,
    content,
  };
}

/** FEAT-003: ペットの問診履歴（InterviewHistoryItem[] 互換）。 */
export function useGetPetMedicalHistory(
  petId: string | undefined,
  excludeRecordId?: string,
): { historyItems: MedicalRecordInterviewHistoryItem[]; isLoading: boolean } {
  const { data, isLoading } = useQuery({
    queryKey: queryKeys.medicalRecords.history(petId!),
    queryFn: async () => {
      const params: Record<string, string> = { limit: "50", page: "1" };
      if (petId) params.pet_id = petId;
      const { data: res } = await axios.get<MedicalRecordsListResponse>(
        "/v1/medical-records",
        { params },
      );
      return res.data;
    },
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });

  const historyItems: MedicalRecordInterviewHistoryItem[] = (data ?? [])
    .filter((r) => !excludeRecordId || String(r.id ?? 0) !== excludeRecordId)
    .map(transformToHistoryItem);

  return { historyItems, isLoading };
}
