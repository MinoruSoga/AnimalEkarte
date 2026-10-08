import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_GC_TIMES, QUERY_STALE_TIMES } from "@/lib/react-query";

/**
 * EMR-306: ケージ占有状況の共有フック（master 連結 picker と入院 board の
 * merge-mode が共用）。
 *
 * src/hooks 配置の理由: 層逆転ルール（src/hooks から @/features への import 禁止）
 * により features/hospitalization/api/get-hospitalizations.ts を再利用できないため、
 * wire 型はローカル定義とし axios を直接呼ぶ（先例: use-cage-merge.ts）。
 * クエリキーは queryKeys.hospitalizations.cageOccupancy で ["hospitalizations"] prefix
 * 配下 — all() による invalidate に含まれる。
 */

// BE ParsePagination の上限（DefaultMaxPaginationLimit=100）に合わせる。
// features/hospitalization/constants の HOSPITALIZATION_OCCUPANCY_LIMIT と同値だが、
// 層ルール上 feature 定数は import できないためローカル定数とする。
const CAGE_OCCUPANCY_PAGE_LIMIT = 100;

/** GET /v1/hospitalizations の占有判定に必要な最小 wire 形（cage_id のみ使用）。 */
interface CageOccupancyWireRow {
  cage_id?: number | null;
}

interface CageOccupancyWirePage {
  data?: CageOccupancyWireRow[];
  total?: number;
}

type OccupyingWireStatus = "admitted" | "reserved";

const fetchOccupancyPage = async (status: OccupyingWireStatus): Promise<CageOccupancyWirePage> => {
  const { data } = await axios.get<CageOccupancyWirePage>("/v1/hospitalizations", {
    params: { page: 1, limit: CAGE_OCCUPANCY_PAGE_LIMIT, status },
  });
  return data;
};

export interface UseCageOccupancyResult {
  /**
   * admitted/reserved 入院が載っているケージ ID の集合。
   * undefined = 占有状況が不明（読込中・エラー・先頭ページ打ち切り・disabled）。
   * 呼び出し側は fail-closed に倒すこと（占有不明のケージを「空き」と断定しない）。
   */
  occupiedCageIds: ReadonlySet<string> | undefined;
}

export function useCageOccupancy(options?: { enabled?: boolean }): UseCageOccupancyResult {
  const enabled = options?.enabled ?? true;
  const queryOpts = {
    enabled,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  };

  const admittedQuery = useQuery({
    queryKey: queryKeys.hospitalizations.cageOccupancy("admitted"),
    queryFn: () => fetchOccupancyPage("admitted"),
    ...queryOpts,
  });
  const reservedQuery = useQuery({
    queryKey: queryKeys.hospitalizations.cageOccupancy("reserved"),
    queryFn: () => fetchOccupancyPage("reserved"),
    ...queryOpts,
  });

  const occupiedCageIds = useMemo(() => {
    // fail-closed: 読込中（disabled 含む）・エラー・total が先頭ページを超える
    // 打ち切りでは「空室」を断定できないため undefined を返す。
    if (admittedQuery.isPending || reservedQuery.isPending) return undefined;
    if (admittedQuery.isError || reservedQuery.isError) return undefined;
    const admitted = admittedQuery.data;
    const reserved = reservedQuery.data;
    if (admitted === undefined || reserved === undefined) return undefined;
    // total が無い/取得数を超える応答は先頭ページ打ち切りの可能性 — fail-closed。
    if (
      typeof admitted.total !== "number" ||
      typeof reserved.total !== "number" ||
      admitted.total > (admitted.data?.length ?? 0) ||
      reserved.total > (reserved.data?.length ?? 0)
    ) {
      return undefined;
    }
    const rows = [...(admitted.data ?? []), ...(reserved.data ?? [])];
    const set = new Set<string>();
    for (const row of rows) {
      if (row.cage_id != null) set.add(String(row.cage_id));
    }
    return set;
  }, [
    admittedQuery.data,
    admittedQuery.isPending,
    admittedQuery.isError,
    reservedQuery.data,
    reservedQuery.isPending,
    reservedQuery.isError,
  ]);

  return { occupiedCageIds };
}
