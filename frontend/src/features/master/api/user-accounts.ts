import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { ME_QUERY_KEY } from "@/features/auth";

// ─────────────────────────────────────────────────
// Types（handler の userResponse に対応。string id）
// ─────────────────────────────────────────────────

export interface UserAccountItem {
  id: string;
  email: string;
  display_name: string;
  user_type: string;
  staff_id?: string;
  avatar_url: string;
  status: string;
  permissionGroupIds?: number[];
  created_at: string;
  updated_at: string;
}

export interface UserAccountDetail extends UserAccountItem {
  memberships: { clinic_id: string; is_main: boolean }[];
  permissionGroupIds: number[];
}

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

export const USER_ACCOUNTS_QUERY_KEY = ["user-accounts"] as const;

const listQueryKey = (clinicId: string) =>
  [...USER_ACCOUNTS_QUERY_KEY, clinicId] as const;

const detailQueryKey = (userId: string) =>
  [...USER_ACCOUNTS_QUERY_KEY, "detail", userId] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function getUsers(clinicId: string): Promise<UserAccountItem[]> {
  const { data } = await axios.get<UserAccountItem[]>("/v1/users", {
    params: { clinic_id: clinicId },
  });
  return data;
}

export async function getUser(userId: string): Promise<UserAccountDetail> {
  const { data } = await axios.get<UserAccountDetail>(`/v1/users/${userId}`);
  return data;
}

export async function setUserPermissionGroups(
  userId: string,
  groupIds: number[],
): Promise<void> {
  await axios.put(`/v1/users/${userId}/permission-groups`, { group_ids: groupIds });
}

// ─────────────────────────────────────────────────
// React Query hooks
// ─────────────────────────────────────────────────

export function useGetUsers(clinicId: string) {
  return useQuery({
    queryKey: listQueryKey(clinicId),
    queryFn: () => getUsers(clinicId),
    enabled: !!clinicId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}

export function useGetUser(userId: string | null) {
  return useQuery({
    queryKey: detailQueryKey(userId ?? ""),
    queryFn: () => getUser(userId!),
    enabled: !!userId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}

export function useSetUserPermissionGroups(clinicId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ userId, groupIds }: { userId: string; groupIds: number[] }) =>
      setUserPermissionGroups(userId, groupIds),
    onSuccess: (_data, { userId }) => {
      void queryClient.invalidateQueries({ queryKey: listQueryKey(clinicId) });
      void queryClient.invalidateQueries({ queryKey: detailQueryKey(userId) });
      void queryClient.invalidateQueries({ queryKey: ME_QUERY_KEY });
    },
  });
}
