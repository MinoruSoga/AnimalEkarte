import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import type { ExaminationRecord, ExamResult } from "@/types";
import type {
  Examination,
  ExamResult as ModelExamResult,
} from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export interface ExaminationFilters {
  startDate?: string;
  endDate?: string;
  petId?: string;
}

export interface UpdateExaminationRequest {
  medical_record_id?: number | null;
  status?: "pending" | "in_progress" | "result_entered" | "completed" | "confirmed";
  result_summary?: string;
  machine?: string;
  date?: string;
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

const EXAM_STATUS_EN_TO_JA: Record<
  string,
  "依頼中" | "検査中" | "結果入力済み" | "完了" | "確定"
> = {
  pending: "依頼中",
  in_progress: "検査中",
  result_entered: "結果入力済み",
  completed: "完了",
  confirmed: "確定",
};

function transformExamResult(item: ModelExamResult): ExamResult {
  return {
    id: String(item.id ?? 0),
    name: item.name ?? "",
    result: item.result ?? "",
    unit: item.unit ?? "",
    referenceRange: item.reference_value ?? "",
    isAbnormal: item.is_abnormal ?? false,
  };
}

function transformExamination(data: Examination): ExaminationRecord {
  return {
    id: String(data.id ?? 0),
    date: data.date ? data.date.split("T")[0] : "",
    ownerName: data.pet?.owner?.name ?? "",
    petName: data.pet?.name ?? "",
    petId: data.pet_id ? String(data.pet_id) : undefined,
    medicalRecordId: data.medical_record_id ? String(data.medical_record_id) : undefined,
    testType: data.exam_type?.name ?? "",
    testTypeId: String(data.exam_type_id ?? ""),
    doctor: data.doctor?.name ?? String(data.doctor_id ?? ""),
    doctorId: String(data.doctor_id ?? ""),
    status: EXAM_STATUS_EN_TO_JA[data.status ?? ""] ?? "依頼中",
    resultSummary: data.result_summary ?? undefined,
    machine: data.machine ?? undefined,
    items: data.items?.map(transformExamResult),
  };
}

// ─────────────────────────────────────────────────
// Hooks
// ─────────────────────────────────────────────────

interface ExaminationsListResponse {
  data: Examination[];
  total: number;
  page: number;
  limit: number;
}

/**
 * Read-only hook to fetch examinations with optional filters.
 * Query key matches features/examinations for cache sharing.
 */
export function useGetExaminations(filters?: ExaminationFilters) {
  return useQuery({
    queryKey: ["examinations", filters],
    queryFn: async (): Promise<ExaminationRecord[]> => {
      const params: Record<string, string> = {};
      if (filters?.startDate) params.start_date = filters.startDate;
      if (filters?.endDate) params.end_date = filters.endDate;
      if (filters?.petId) params.pet_id = filters.petId;
      const { data } = await axios.get<ExaminationsListResponse>("/v1/examinations", { params });
      return data.data.map(transformExamination);
    },
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}

/**
 * Mutation hook to update an examination record.
 * Used cross-feature to link examinations to medical records.
 */
export function useUpdateExamination() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      req,
    }: {
      id: string;
      req: UpdateExaminationRequest;
    }): Promise<ExaminationRecord> => {
      const { data } = await axios.patch<Examination>(`/v1/examinations/${id}`, req);
      return transformExamination(data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["examinations"] });
    },
    onError: (error) => handleApiError(error, "検査更新"),
  });
}
