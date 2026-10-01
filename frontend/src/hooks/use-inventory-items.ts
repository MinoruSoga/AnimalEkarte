import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

// バックエンド ListInventory の limit 上限（persistence.MaxMasterListRows）と同値。
const INVENTORY_LIST_ALL_LIMIT = 10000;

// フィルタの「全件から選択」に必要な最小フィールドのみの専用 DTO
// （generated/models は TASK-444-S1 の凍結境界のため新規 import しない）。
interface InventoryItemOptionRow {
  id: number;
  name: string;
}

interface InventoryListResponse {
  data: InventoryItemOptionRow[];
  total: number;
  page: number;
  limit: number;
}

export interface InventoryOption {
  id: string;
  name: string;
}

/**
 * 在庫品の全件一覧（カルテ一覧フィルタ等の選択肢用）。
 * query key は features/inventory と共有してキャッシュを一本化する。
 */
export function useGetAllInventoryItems() {
  return useQuery({
    queryKey: queryKeys.inventoryItems.list({
      page: 1,
      limit: INVENTORY_LIST_ALL_LIMIT,
    }),
    queryFn: async (): Promise<InventoryOption[]> => {
      const { data } = await axios.get<InventoryListResponse>("/v1/inventory", {
        params: { page: 1, limit: INVENTORY_LIST_ALL_LIMIT },
      });
      return data.data.map((item) => ({ id: String(item.id), name: item.name }));
    },
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
