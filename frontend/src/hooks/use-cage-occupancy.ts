import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { useCurrentClinicId } from "@/hooks/use-current-clinic-id";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_GC_TIMES, QUERY_STALE_TIMES } from "@/lib/react-query";

/**
 * EMR-306: ケージ占有状況の共有フック（master 連結ピッカーと入院 board の
 * merge-mode が共用）。
 *
 * src/hooks 配置の理由: 層逆転ルール（src/hooks から @/features への import 禁止）
 * により features/hospitalization/api を再利用できないため、wire 型はローカル定義とし
 * axios を直接呼ぶ（先例: use-cage-merge.ts）。
 *
 * Codex review 対応: GET /v1/hospitalizations は hospitalization:view を要求するため、
 * master-hospitalization:edit のみ持つケージ編集者は占有状況を取得できず fail-closed で
 * 全ケージ選択不可に陥る。占有判定は専用エンドポイント GET /v1/masters/cages/occupancy
 * （merge と同権限・ページングなしの全件返却）を使う。
 * クエリキーは queryKeys.hospitalizations.cageOccupancy で ["hospitalizations"] prefix
 * 配下 — 入院の作成/退院による all() invalidate で占有状況も追従する。
 */

/** GET /v1/masters/cages/occupancy の wire 形。 */
interface CageOccupancyResponse {
  occupied_cage_ids?: number[];
}

const fetchOccupiedCageIds = async (clinicId: string | null): Promise<ReadonlySet<string>> => {
  if (!clinicId) throw new Error("医院が選択されていません。");
  const { data } = await axios.get<CageOccupancyResponse>("/v1/masters/cages/occupancy", {
    headers: { "X-Clinic-ID": clinicId },
  });
  // fail-closed: 契約フィールドを欠く応答は信頼しない（空室と断定しない）。
  // React Query v5 は queryFn の undefined 返却を禁止するため明示的に throw する。
  if (!Array.isArray(data?.occupied_cage_ids)) {
    throw new Error("cage occupancy response missing occupied_cage_ids");
  }
  return new Set(data.occupied_cage_ids.map(String));
};

export interface UseCageOccupancyResult {
  /**
   * admitted/reserved 入院が載っているケージ ID の集合。
   * undefined = 占有状況が不明（読込中・エラー・契約外応答・disabled）。
   * 呼び出し側は fail-closed に倒すこと（占有不明のケージを「空き」と断定しない）。
   */
  occupiedCageIds: ReadonlySet<string> | undefined;
}

export function useCageOccupancy(options?: { enabled?: boolean }): UseCageOccupancyResult {
  const clinicId = useCurrentClinicId();
  const enabled = (options?.enabled ?? true) && !!clinicId;

  const query = useQuery({
    queryKey: queryKeys.hospitalizations.cageOccupancy(clinicId),
    // Capture the same clinic for the key and request, including deferred Axios retries.
    queryFn: () => fetchOccupiedCageIds(clinicId),
    enabled,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });

  const occupiedCageIds = useMemo(() => {
    if (!enabled || query.isPending || query.isError) return undefined;
    return query.data;
  }, [enabled, query.data, query.isPending, query.isError]);

  return { occupiedCageIds };
}
