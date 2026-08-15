import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { ME_QUERY_KEY, queryKeys } from "@/lib/query-keys";
import type { PermissionGroup as ModelPermissionGroup } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types (derived from models.ts)
// ─────────────────────────────────────────────────

type PermissionGroupRequestBase = Omit<
  ModelPermissionGroup,
  "id" | "clinic_id" | "created_at" | "updated_at" | "rules" | "staffs"
>;

type PermissionGroupRuleRequest = {
  resource: string;
  can_view: boolean;
  can_create: boolean;
  can_edit: boolean;
  can_delete: boolean;
};

export type CreatePermissionGroupRequest = Pick<PermissionGroupRequestBase, "name" | "color"> &
  Partial<Omit<PermissionGroupRequestBase, "name" | "color">> & {
    rules?: PermissionGroupRuleRequest[];
  };

export type UpdatePermissionGroupRequest = Partial<PermissionGroupRequestBase> & {
  rules?: PermissionGroupRuleRequest[];
};

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

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function listPermissionGroups(): Promise<PermissionGroup[]> {
  const { data } = await axios.get<ModelPermissionGroup[]>("/v1/masters/permission-groups");
  return data.map(transformPermissionGroup);
}

async function createPermissionGroup(req: CreatePermissionGroupRequest): Promise<PermissionGroup> {
  const { data } = await axios.post<ModelPermissionGroup>("/v1/masters/permission-groups", req);
  return transformPermissionGroup(data);
}

async function updatePermissionGroup(
  id: string,
  req: UpdatePermissionGroupRequest,
): Promise<PermissionGroup> {
  const { data } = await axios.patch<ModelPermissionGroup>(
    `/v1/masters/permission-groups/${id}`,
    req,
  );
  return transformPermissionGroup(data);
}

async function deletePermissionGroup(id: string): Promise<void> {
  await axios.delete(`/v1/masters/permission-groups/${id}`);
}

async function reorderPermissionGroups(ids: string[]): Promise<void> {
  await axios.patch("/v1/masters/permission-groups/reorder", {
    ids: ids.map((id) => parseInt(id, 10)),
  });
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetPermissionGroups() {
  return useQuery({
    queryKey: queryKeys.masters.category("permission-groups"),
    queryFn: listPermissionGroups,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreatePermissionGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createPermissionGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("permission-groups") });
      queryClient.invalidateQueries({ queryKey: ME_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
}

export function useUpdatePermissionGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdatePermissionGroupRequest }) =>
      updatePermissionGroup(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("permission-groups") });
      queryClient.invalidateQueries({ queryKey: ME_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeletePermissionGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deletePermissionGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("permission-groups") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}

export function useReorderPermissionGroups() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderPermissionGroups,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("permission-groups") });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
}
