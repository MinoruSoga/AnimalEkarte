import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { Examination } from "@/types/generated/models";
import { transformExamResult, type ExamResult } from "@/lib/transforms/examination";

/**
 * 医療記録（カルテ）画面用の検査グループ ViewModel。
 *
 * items は examinations feature の shared {@link transformExamResult} に統一する。
 * これにより `referenceValue` (camelCase) / `id: string` / `status: "normal" | "high" | "low"`
 * といった ExamItemsTable と同じ意味のフィールドを共有する。
 *
 * 日付表示だけはカルテ画面固有の `"YYYY-MM-DD HH:mm"` フォーマットを使うため、
 * shared {@link transformExamination} ではなくこちらでラップする。
 */
export interface ExamGroup {
  id: number;
  date: string;
  machine: string;
  items: ExamResult[];
}

export interface RecordExaminationsResult {
  items: ExamGroup[];
  isTruncated: boolean;
}

function transformExamGroup(exam: Examination): ExamGroup {
  return {
    id: exam.id,
    date: exam.date ? exam.date.slice(0, 16).replace("T", " ") : "-",
    machine: exam.machine,
    items: (exam.items ?? []).map(transformExamResult),
  };
}

const getRecordExaminations = async (input: {
  petId?: string;
  medicalRecordId?: string;
}): Promise<RecordExaminationsResult> => {
  const params: Record<string, string | number | boolean> = {
    limit: HISTORY_FETCH_LIMIT,
    include_items: true,
  };
  // 両方渡すと BE は「そのカルテの検査 + 同じペットの未取り込み検査」を返す。
  // record 単独では機器受信（attach 直後・medical_record_id NULL）が消える。
  if (input.medicalRecordId) {
    params.medical_record_id = Number(input.medicalRecordId);
  }
  if (input.petId) {
    params.pet_id = Number(input.petId);
  }
  const { data } = await axios.get<{ data: Examination[]; total?: number }>("/v1/examinations", {
    params,
  });
  const rows = data.data ?? [];
  return {
    items: rows.map(transformExamGroup),
    isTruncated: typeof data.total === "number" && data.total > rows.length,
  };
};

export const useGetRecordExaminations = (petId?: string, medicalRecordId?: string) => {
  return useQuery({
    queryKey: medicalRecordId
      ? queryKeys.examinations.byRecord(medicalRecordId)
      : queryKeys.examinations.byPet(petId!),
    queryFn: () => getRecordExaminations({ petId, medicalRecordId }),
    enabled: Boolean(medicalRecordId || petId),
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
