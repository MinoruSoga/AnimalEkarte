import { useMutation, useQueryClient, type QueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

/**
 * EMR-306 ケージ一時連結の mutation hooks（self-contained）。
 *
 * src/hooks 配置の理由: ケージは master feature の一覧 CRUD
 * （features/master/api/cages.ts = queryKeys.masters.category("cages")）だけでなく、
 * 入院系 UI が使う汎用マスタ hook（use-master-items.ts の "cage" =
 * queryKeys.masters.category("cage")）からも参照される。層逆転ルール（src/hooks
 * から @/features への import 禁止）と generated-models freeze ルールにより、
 * features 側の transformCage / ModelCage を再利用できないため wire 型は
 * ローカル定義とし、axios を直接呼ぶ（先例: use-master-items.ts）。
 *
 * invalidate は "cages" と "cage" の両方を別々に行う — masters.category(name) は
 * ["masters", name] であり、片方はもう片方の prefix ではないため 2 回必要
 * （先例: use-update-reservation.ts の list + detail 両 invalidate）。
 */

interface MergeCageRequest {
  /** 代表以外の構成ケージ ID（int64 配列、最低1件） */
  member_cage_ids: number[];
}

export interface MergeCagesVariables {
  /** 代表ケージ ID。POST /v1/masters/cages/:id/merge の :id になる */
  representativeId: string;
  /** 代表以外の構成ケージ ID 配列（member_cage_ids にそのまま送る） */
  memberCageIds: number[];
}

const mergeCages = async ({
  representativeId,
  memberCageIds,
}: MergeCagesVariables): Promise<void> => {
  const body: MergeCageRequest = { member_cage_ids: memberCageIds };
  await axios.post(`/v1/masters/cages/${representativeId}/merge`, body);
};

const unmergeCages = async (representativeId: string): Promise<void> => {
  await axios.post(`/v1/masters/cages/${representativeId}/unmerge`);
};

function invalidateCageMergeQueries(queryClient: QueryClient): void {
  queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("cages") });
  queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("cage") });
}

/**
 * ケージ一時連結（POST /v1/masters/cages/:id/merge）。
 * 代表ケージ id に member_cage_ids のケージを連結する。
 */
export function useMergeCages() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: mergeCages,
    onSuccess: () => invalidateCageMergeQueries(queryClient),
    onError: (error) => handleApiError(error, "ケージ連結"),
  });
}

/**
 * ケージ一時連結の解除（POST /v1/masters/cages/:id/unmerge）。
 * :id はグループの代表ケージ ID（メンバー行からは merged_into_cage_id で解決する）。
 */
export function useUnmergeCages() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: unmergeCages,
    onSuccess: () => invalidateCageMergeQueries(queryClient),
    onError: (error) => handleApiError(error, "ケージ連結解除"),
  });
}
