import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformExamination, type ExaminationRecord } from "@/lib/transforms/examination";
import type { Examination } from "@/types/generated/models";

// #158 §計画補足: 履歴表示は下書き（依頼中/検査中）を除外する。
// 既存 transformExamination は status を日本語ラベルに変換するため、そのラベルで除外する。
const DRAFT_STATUS_LABELS = new Set(["依頼中", "検査中"]);

const getPetExaminations = async (petId: string): Promise<ExaminationRecord[]> => {
  const { data } = await axios.get<{ data: Examination[] }>("/v1/examinations", {
    params: { pet_id: Number(petId), limit: 100 },
  });
  return (data.data ?? [])
    .map(transformExamination)
    .filter((e) => !DRAFT_STATUS_LABELS.has(e.status));
};

export const useGetPetExaminations = (petId?: string) => {
  return useQuery({
    queryKey: queryKeys.petExaminationsReport(petId!),
    queryFn: () => getPetExaminations(petId!),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
