import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformMedicalRecord, transformToHistoryItem } from "./transforms";
import type { MedicalRecord } from "./transforms";
import type { BackendMedicalRecord } from "./types";
import type { InterviewHistoryItem } from "../types";

interface MedicalRecordsListResponse {
  data: BackendMedicalRecord[];
  total: number;
  page: number;
  limit: number;
}

// BUG-B1: 一覧の server-side pagination/search/filter 対応。
// status は BE enum（draft/finalized）、doctorId/animalSpeciesId は FE で master から解決した ID を渡す。
/** B-1 follow-up: 列ソート server 化で BE が受理する許可キー（medical_record_repository.go の
 * medicalRecordSortColumns と1対1対応）。 */
export type MedicalRecordSortKey = "date" | "owner_name" | "pet_name" | "status";

export interface MedicalRecordFilters {
  startDate?: string; // YYYY-MM-DD
  endDate?: string;   // YYYY-MM-DD
  petId?: string;
  ownerId?: string;
  /** #86: 拠点横断表示。複数医院IDをカンマ区切りで送信。未指定=現在の医院のみ */
  clinicIds?: string[];
  search?: string;
  status?: string; // "draft" | "finalized"
  doctorId?: string;
  animalSpeciesId?: string;
  page?: number;
  limit?: number;
  /** B-1 follow-up: 列ソート server 化。未指定時は BE 既定順（date DESC, created_at DESC）。 */
  sort?: MedicalRecordSortKey;
  order?: "asc" | "desc";
}

export interface MedicalRecordsResult {
  data: MedicalRecord[];
  total: number;
  page: number;
  limit: number;
}

const DEFAULT_PAGE = 1;
const DEFAULT_LIMIT = 20;

const getMedicalRecords = async (
  filters?: MedicalRecordFilters,
): Promise<MedicalRecordsResult> => {
  const params: Record<string, string | number> = {
    page: filters?.page ?? DEFAULT_PAGE,
    limit: filters?.limit ?? DEFAULT_LIMIT,
  };
  if (filters?.startDate) params.start_date = filters.startDate;
  if (filters?.endDate) params.end_date = filters.endDate;
  if (filters?.petId) params.pet_id = filters.petId;
  if (filters?.ownerId) params.owner_id = filters.ownerId;
  if (filters?.clinicIds && filters.clinicIds.length > 1) params.clinic_ids = filters.clinicIds.join(",");
  if (filters?.search) params.search = filters.search;
  if (filters?.status) params.status = filters.status;
  if (filters?.doctorId) params.doctor_id = filters.doctorId;
  if (filters?.animalSpeciesId) params.animal_species_id = filters.animalSpeciesId;
  if (filters?.sort) params.sort = filters.sort;
  if (filters?.order) params.order = filters.order;

  const { data } = await axios.get<MedicalRecordsListResponse>("/v1/medical-records", { params });
  return {
    data: data.data.map(transformMedicalRecord),
    total: data.total,
    page: data.page,
    limit: data.limit,
  };
};

export const useGetMedicalRecords = (filters?: MedicalRecordFilters) => {
  return useQuery({
    queryKey: ["medical-records", filters],
    queryFn: () => getMedicalRecords(filters),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

/** FEAT-003: ペットの問診履歴を取得して InterviewHistoryItem[] に変換する */
export const useGetPetMedicalHistory = (
  petId: string | undefined,
  excludeRecordId?: string,
): { historyItems: InterviewHistoryItem[]; isLoading: boolean } => {
  const { data, isLoading } = useQuery({
    queryKey: ["medical-records", "history", petId],
    queryFn: async () => {
      const params: Record<string, string> = { limit: "50", page: "1" };
      if (petId) params.pet_id = petId;
      const { data: res } = await axios.get<MedicalRecordsListResponse>("/v1/medical-records", { params });
      return res.data;
    },
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });

  const historyItems: InterviewHistoryItem[] = (data ?? [])
    .filter((r) => !excludeRecordId || String(r.id ?? 0) !== excludeRecordId)
    .map(transformToHistoryItem);

  return { historyItems, isLoading };
};
