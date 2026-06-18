import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

/** GET /v1/pets/:id/first-visit のレスポンス（バックエンド petFirstVisitResponse に対応）。 */
interface PetFirstVisitResponse {
  /** 初診日 = 最古の有効カルテ date 由来の派生値。受診歴が無ければ null（捏造しない）。 */
  first_visit_date: string | null;
}

/**
 * #158 飼主レポート: ペットの初診日（最古カルテ date 由来）を取得する。
 * 値は medical_records から導出する派生値で、clinic 隔離・論理削除除外はバックエンドが担保する。
 * 受診歴が無い場合は null を返す。
 */
export const useGetPetFirstVisit = (petId: string | undefined) => {
  return useQuery({
    queryKey: ["pet-first-visit", petId],
    queryFn: async (): Promise<string | null> => {
      const { data } = await axios.get<PetFirstVisitResponse>(`/v1/pets/${petId}/first-visit`);
      return data.first_visit_date;
    },
    enabled: !!petId,
    // 初診日は最古カルテ由来でほぼ変化しない。同レポートの useGetPets / useGetOwner と揃えて STATIC。
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};
