import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { PermissionGroup } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

export interface RuleInput {
  resource: string;
  can_view: boolean;
  can_create: boolean;
  can_edit: boolean;
  can_delete: boolean;
}

export interface CreatePermissionGroupInput {
  clinicId: string;
  name: string;
  description?: string;
  color?: string;
}

export interface UpdatePermissionGroupInput {
  name?: string;
  description?: string;
  color?: string;
}

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

export const PERMISSION_GROUPS_QUERY_KEY = ["permission-groups"] as const;

const queryKey = (clinicId: string) => [...PERMISSION_GROUPS_QUERY_KEY, clinicId] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function getPermissionGroups(clinicId: string): Promise<PermissionGroup[]> {
  const { data } = await axios.get<PermissionGroup[]>("/v1/permission-groups", {
    params: { clinic_id: clinicId },
  });
  return data;
}

export async function createPermissionGroup(
  input: CreatePermissionGroupInput,
): Promise<PermissionGroup> {
  const { data } = await axios.post<PermissionGroup>(
    "/v1/permission-groups",
    {
      name: input.name,
      description: input.description ?? "",
      color: input.color ?? "#6B7280",
    },
    { params: { clinic_id: input.clinicId } },
  );
  return data;
}

export async function updatePermissionGroup(
  groupId: number,
  input: UpdatePermissionGroupInput,
): Promise<PermissionGroup> {
  const { data } = await axios.patch<PermissionGroup>(
    `/v1/permission-groups/${groupId}`,
    input,
  );
  return data;
}

export async function deletePermissionGroup(groupId: number): Promise<void> {
  await axios.delete(`/v1/permission-groups/${groupId}`);
}

export async function setPermissionGroupRules(
  groupId: number,
  rules: RuleInput[],
): Promise<void> {
  await axios.put(`/v1/permission-groups/${groupId}/rules`, { rules });
}

// ─────────────────────────────────────────────────
// React Query hooks
// ─────────────────────────────────────────────────

export function useGetPermissionGroups(clinicId: string) {
  return useQuery({
    queryKey: queryKey(clinicId),
    queryFn: () => getPermissionGroups(clinicId),
    enabled: !!clinicId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}

export function useCreatePermissionGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createPermissionGroup,
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({ queryKey: queryKey(variables.clinicId) });
    },
  });
}

export function useUpdatePermissionGroup(clinicId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ groupId, input }: { groupId: number; input: UpdatePermissionGroupInput }) =>
      updatePermissionGroup(groupId, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKey(clinicId) });
    },
  });
}

export function useDeletePermissionGroup(clinicId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (groupId: number) => deletePermissionGroup(groupId),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKey(clinicId) });
    },
  });
}

export function useSetPermissionGroupRules(clinicId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ groupId, rules }: { groupId: number; rules: RuleInput[] }) =>
      setPermissionGroupRules(groupId, rules),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKey(clinicId) });
    },
  });
}
