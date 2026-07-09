import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
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

export type ExamGroupItem = ExamResult;

function transformExamGroup(exam: Examination): ExamGroup {
  return {
    id: exam.id,
    date: exam.date ? exam.date.slice(0, 16).replace("T", " ") : "-",
    machine: exam.machine,
    items: (exam.items ?? []).map(transformExamResult),
  };
}

const getRecordExaminations = async (
  petId: string,
): Promise<ExamGroup[]> => {
  const { data } = await axios.get<{ data: Examination[] }>("/v1/examinations", {
    params: { pet_id: Number(petId), limit: 100 },
  });
  return (data.data ?? []).map(transformExamGroup);
};

export const useGetRecordExaminations = (petId?: string) => {
  return useQuery({
    queryKey: ["examinations", "pet", petId],
    queryFn: () => getRecordExaminations(petId!),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
