import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { formatDate } from "@/lib/format/date";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { MedicalRecordResponse } from "@/types/generated/medicalrecord-responses";

/**
 * FE-RC-015 (followup2): list query の実体を hooks に置く。
 * features → hooks は許可方向。hooks → features は禁止。
 *
 * recommendation_reason の狭義型は feature constants 側の正本のまま。
 * ここでは wire 文字列を保持し、消費者は既存 MedicalRecord 型と構造的に互換。
 */
type MedicalRecordStatus = "作成中" | "確定済";

const STATUS_MAP: Record<string, MedicalRecordStatus> = {
  draft: "作成中",
  finalized: "確定済",
};

function toVisitTypeLabel(visitType?: string | null): string | undefined {
  if (!visitType) return undefined;
  if (visitType === "first" || visitType === "初診") return "初診";
  if (visitType === "revisit" || visitType === "再診") return "再診";
  return visitType;
}

/** List/detail wire → UI row（feature api/transforms.transformMedicalRecord と同形）。 */
function transformMedicalRecordForList(record: MedicalRecordResponse) {
  return {
    id: String(record.id ?? 0),
    recordNo: record.record_no,
    date: formatDate(record.date),
    ownerId: record.owner_id ? String(record.owner_id) : undefined,
    ownerName: record.owner?.name ?? "",
    petId: record.pet_id ? String(record.pet_id) : undefined,
    petName: record.pet?.name ?? "",
    // TASK-444-S1: hooks は @/types/generated/models 禁止のため wire リテラルを直書き。
    petIsDeceased: record.pet?.status === "deceased",
    species: record.pet?.animal_species?.name ?? "",
    chiefComplaint: record.inquiry?.chief_complaint ?? "",
    chiefComplaintTypeId: record.inquiry?.chief_complaint_type_id ?? null,
    doctor: record.doctor?.name ?? String(record.doctor_id ?? ""),
    visitType: toVisitTypeLabel(record.visit_type),
    nextVisitRecommendedDate: record.next_visit_recommended_date ?? "",
    subjective: undefined as string | undefined,
    objective: undefined as string | undefined,
    assessment: undefined as string | undefined,
    plan: undefined as string | undefined,
    surgeryNotes: undefined as string | undefined,
    diagnosis: undefined as string | undefined,
    treatment: undefined as string | undefined,
    prescription: undefined as string | undefined,
    diagnosis1CategoryId: null as number | null,
    diagnosis1NameId: null as number | null,
    diagnosis2CategoryId: null as number | null,
    diagnosis2NameId: null as number | null,
    notes: record.inquiry?.notes || undefined,
    accountingId: record.accounting_id ? String(record.accounting_id) : undefined,
    visitCount: record.visit_count,
    version: record.version,
    recommendationReason: (record.recommendation_reason ?? null) as
      | "revisit"
      | "checkup"
      | "prevention"
      | "exam"
      | null,
    clinicId: record.clinic_id ? String(record.clinic_id) : undefined,
    status: (STATUS_MAP[record.status] ?? "作成中") as MedicalRecordStatus,
  };
}

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
  data: ReturnType<typeof transformMedicalRecordForList>[];
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
    data: data.data.map(transformMedicalRecordForList),
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
