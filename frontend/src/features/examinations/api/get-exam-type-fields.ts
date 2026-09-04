import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ExaminationType } from "@/types/generated/models";

/**
 * 検査種別の項目テンプレ（exam_type_fields）を表す行。
 * normal_value は表示用として保持し、判定基準値は backend がマスタから解決する。
 */
export interface ExamTypeFieldRow {
  id: number;
  name: string;
  unit: string;
  normalValue: string;
  sortOrder: number;
}

/**
 * GET /v1/masters/examination-types/:id — 検査種別の詳細（items=exam_type_fields 含む）を取得する。
 * 検査フォームで「検査項目テーブルのテンプレ」を組み立てるために使う。
 */
const getExamTypeFields = async (id: string): Promise<ExamTypeFieldRow[]> => {
  const { data } = await axios.get<ExaminationType>(`/v1/masters/examination-types/${id}`);
  const items = data.items ?? [];
  return items
    .slice()
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
    .map((field): ExamTypeFieldRow => ({
      id: Number(field.id),
      name: field.name ?? "",
      unit: field.unit ?? "",
      normalValue: field.normal_value ?? "",
      sortOrder: field.sort_order ?? 0,
    }));
};

export const useGetExamTypeFields = (examTypeId: string) => {
  return useQuery({
    queryKey: queryKeys.examinations.typeFields(examTypeId),
    queryFn: () => getExamTypeFields(examTypeId),
    enabled: !!examTypeId,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
