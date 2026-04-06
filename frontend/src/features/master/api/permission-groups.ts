import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { PermissionGroup as ModelPermissionGroup } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export interface CreatePermissionGroupRequest {
  name: string;
  description?: string;
  color: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdatePermissionGroupRequest {
  name?: string;
  description?: string;
  color?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface SetPermissionGroupRulesRequest {
  rules: Array<{
    resource: string;
    can_view: boolean;
    can_create: boolean;
    can_edit: boolean;
    can_delete: boolean;
  }>;
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformPermissionGroup(data: ModelPermissionGroup) {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    description: data.description,
    color: data.color,
    isActive: data.is_active,
    sortOrder: data.sort_order,
    rules: data.rules?.map((rule) => ({
      id: String(rule.id ?? 0),
      groupId: String(rule.group_id ?? 0),
      resource: rule.resource,
      canView: rule.can_view,
      canCreate: rule.can_create,
      canEdit: rule.can_edit,
      canDelete: rule.can_delete,
      createdAt: rule.created_at,
      updatedAt: rule.updated_at,
    })) ?? [],
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type PermissionGroup = ReturnType<typeof transformPermissionGroup>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const PERMISSION_GROUPS_QUERY_KEY = ["masters", "permission-groups"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listPermissionGroups(): Promise<PermissionGroup[]> {
  const { data } = await axios.get<ModelPermissionGroup[]>("/v1/masters/permission-groups");
  return data.map(transformPermissionGroup);
}

export async function getPermissionGroup(id: string): Promise<PermissionGroup> {
  const { data } = await axios.get<ModelPermissionGroup>(`/v1/masters/permission-groups/${id}`);
  return transformPermissionGroup(data);
}

export async function createPermissionGroup(req: CreatePermissionGroupRequest): Promise<PermissionGroup> {
  const { data } = await axios.post<ModelPermissionGroup>("/v1/masters/permission-groups", req);
  return transformPermissionGroup(data);
}

export async function updatePermissionGroup(
  id: string,
  req: UpdatePermissionGroupRequest,
): Promise<PermissionGroup> {
  const { data } = await axios.patch<ModelPermissionGroup>(
    `/v1/masters/permission-groups/${id}`,
    req,
  );
  return transformPermissionGroup(data);
}

export async function deletePermissionGroup(id: string): Promise<void> {
  await axios.delete(`/v1/masters/permission-groups/${id}`);
}

export async function setPermissionGroupRules(
  id: string,
  req: SetPermissionGroupRulesRequest,
): Promise<PermissionGroup> {
  const { data } = await axios.put<ModelPermissionGroup>(
    `/v1/masters/permission-groups/${id}/rules`,
    req,
  );
  return transformPermissionGroup(data);
}

export async function reorderPermissionGroups(ids: string[]): Promise<void> {
  await axios.patch("/v1/masters/permission-groups/reorder", {
    ids: ids.map((id) => parseInt(id, 10)),
  });
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetPermissionGroups() {
  return useQuery({
    queryKey: PERMISSION_GROUPS_QUERY_KEY,
    queryFn: listPermissionGroups,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useGetPermissionGroup(id: string) {
  return useQuery({
    queryKey: [...PERMISSION_GROUPS_QUERY_KEY, id],
    queryFn: () => getPermissionGroup(id),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreatePermissionGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createPermissionGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: PERMISSION_GROUPS_QUERY_KEY });
    },
  });
}

export function useUpdatePermissionGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdatePermissionGroupRequest }) =>
      updatePermissionGroup(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: PERMISSION_GROUPS_QUERY_KEY });
    },
  });
}

export function useDeletePermissionGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deletePermissionGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: PERMISSION_GROUPS_QUERY_KEY });
    },
  });
}

export function useSetPermissionGroupRules() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: SetPermissionGroupRulesRequest }) =>
      setPermissionGroupRules(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: PERMISSION_GROUPS_QUERY_KEY });
    },
  });
}

export function useReorderPermissionGroups() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderPermissionGroups,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: PERMISSION_GROUPS_QUERY_KEY });
    },
  });
}
