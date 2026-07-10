import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

import { STAFFS_QUERY_KEY } from "./staffs";

// ─────────────────────────────────────────────────
// Staff Permission Groups API
// ─────────────────────────────────────────────────

const getAllPermissionGroupMapKey = (staffIds: string[]) =>
  [...STAFFS_QUERY_KEY, "all-permission-group-map", ...staffIds] as const;

const STAFF_PERM_GROUPS_KEY = (staffId: string) =>
  [...STAFFS_QUERY_KEY, staffId, "permission-groups"] as const;

/**
 * 全スタッフの権限グループIDマップを一括取得する。
 * staffId → groupId[] の Map を返す。
 */
export function useGetAllStaffPermissionGroupMap(staffIds: string[]) {
  return useQuery({
    queryKey: getAllPermissionGroupMapKey(staffIds),
    queryFn: async (): Promise<Map<string, string[]>> => {
      const map = new Map<string, string[]>();
      await Promise.all(
        staffIds.map(async (id) => {
          try {
            const { data } = await axios.get<{ group_ids: number[] }>(
              `/v1/masters/staffs/${id}/permission-groups`,
            );
            map.set(id, (data.group_ids ?? []).map(String));
          } catch {
            // バッチ取得: 個別スタッフの失敗（404含む）はスキップして継続
            map.set(id, []);
          }
        }),
      );
      return map;
    },
    enabled: staffIds.length > 0,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useGetStaffPermissionGroups(staffId: string | null) {
  return useQuery({
    queryKey: STAFF_PERM_GROUPS_KEY(staffId ?? ""),
    queryFn: async (): Promise<string[]> => {
      const { data } = await axios.get<{ group_ids: number[] }>(
        `/v1/masters/staffs/${staffId}/permission-groups`,
      );
      return (data.group_ids ?? []).map(String);
    },
    enabled: staffId !== null,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useUpdateStaffPermissionGroups() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      staffId,
      groupIds,
    }: {
      staffId: string;
      groupIds: string[];
    }) => {
      await axios.put(`/v1/masters/staffs/${staffId}/permission-groups`, {
        group_ids: groupIds.map((id) => parseInt(id, 10)),
      });
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: STAFF_PERM_GROUPS_KEY(variables.staffId),
      });
    },
    onError: (error) => handleApiError(error, "設定"),
  });
}
