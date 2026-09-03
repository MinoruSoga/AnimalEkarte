import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type {
  DiagnosisType as ModelDiagnosisType,
  DiagnosisName as ModelDiagnosisName,
} from "@/types/generated/models";
import type {
  CreateDiagnosisTypeRequest,
  UpdateDiagnosisTypeRequest,
  CreateDiagnosisNameRequest,
  UpdateDiagnosisNameRequest,
  ReorderDiagnosisTypeRequest,
  ReorderDiagnosisNameRequest,
} from "@/types/diagnosis";

// ─────────────────────────────────────────────────
// Transform functions → domain types (camelCase)
// ─────────────────────────────────────────────────

function transformDiagnosisType(data: ModelDiagnosisType) {
  return {
    id: String(data.id ?? 0),
    clinicId: data.clinic_id,
    name: data.name,
    isActive: data.is_active,
    description: data.description,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

function transformDiagnosisName(data: ModelDiagnosisName) {
  return {
    id: String(data.id ?? 0),
    clinicId: data.clinic_id,
    name: data.name,
    isActive: data.is_active,
    description: data.description,
    diagnosisTypeId: String(data.diagnosis_type_id ?? 0),
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

// Frontend domain types derived from transforms (編集禁止 - ReturnType から自動導出)
export type DiagnosisType = ReturnType<typeof transformDiagnosisType>;
export type DiagnosisName = ReturnType<typeof transformDiagnosisName>;

// ─────────────────────────────────────────────────
// Request types re-export (@/types/diagnosis から導出済み)
// ─────────────────────────────────────────────────

export type { UpdateDiagnosisTypeRequest, UpdateDiagnosisNameRequest } from "@/types/diagnosis";

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

// ─────────────────────────────────────────────────
// API functions - DiagnosisType
// ─────────────────────────────────────────────────

async function listDiagnosisTypes(): Promise<DiagnosisType[]> {
  const { data } = await axios.get<{ data: ModelDiagnosisType[] }>("/v1/masters/diagnosis-types");
  return data.data.map(transformDiagnosisType);
}

async function createDiagnosisType(req: CreateDiagnosisTypeRequest): Promise<DiagnosisType> {
  const { data } = await axios.post<ModelDiagnosisType>("/v1/masters/diagnosis-types", req);
  return transformDiagnosisType(data);
}

async function updateDiagnosisType(
  id: string,
  req: UpdateDiagnosisTypeRequest,
): Promise<DiagnosisType> {
  const { data } = await axios.patch<ModelDiagnosisType>(`/v1/masters/diagnosis-types/${id}`, req);
  return transformDiagnosisType(data);
}

async function deleteDiagnosisType(id: string): Promise<void> {
  await axios.delete(`/v1/masters/diagnosis-types/${id}`);
}

async function reorderDiagnosisTypes(req: ReorderDiagnosisTypeRequest): Promise<void> {
  await axios.patch("/v1/masters/diagnosis-types/reorder", req);
}

// ─────────────────────────────────────────────────
// API functions - DiagnosisName
// ─────────────────────────────────────────────────

async function listDiagnosisNames(): Promise<DiagnosisName[]> {
  const { data } = await axios.get<{ data: ModelDiagnosisName[] }>("/v1/masters/diagnosis-names");
  return data.data.map(transformDiagnosisName);
}

async function createDiagnosisName(req: CreateDiagnosisNameRequest): Promise<DiagnosisName> {
  const { data } = await axios.post<ModelDiagnosisName>("/v1/masters/diagnosis-names", req);
  return transformDiagnosisName(data);
}

async function updateDiagnosisName(
  id: string,
  req: UpdateDiagnosisNameRequest,
): Promise<DiagnosisName> {
  const { data } = await axios.patch<ModelDiagnosisName>(`/v1/masters/diagnosis-names/${id}`, req);
  return transformDiagnosisName(data);
}

async function deleteDiagnosisName(id: string): Promise<void> {
  await axios.delete(`/v1/masters/diagnosis-names/${id}`);
}

async function reorderDiagnosisNames(req: ReorderDiagnosisNameRequest): Promise<void> {
  await axios.patch("/v1/masters/diagnosis-names/reorder", req);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks - DiagnosisType
// ─────────────────────────────────────────────────

export function useGetDiagnosisTypes() {
  return useQuery({
    queryKey: queryKeys.masters.category("diagnosis-types"),
    queryFn: listDiagnosisTypes,
    staleTime: QUERY_STALE_TIMES.STATIC, // マスタデータ: 30分キャッシュ
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateDiagnosisType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createDiagnosisType,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("diagnosis-types") });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
}

export function useUpdateDiagnosisType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateDiagnosisTypeRequest }) =>
      updateDiagnosisType(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("diagnosis-types") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteDiagnosisType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteDiagnosisType,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("diagnosis-types") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}

export function useReorderDiagnosisTypes() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderDiagnosisTypes,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("diagnosis-types") });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
}

// ─────────────────────────────────────────────────
// TanStack Query hooks - DiagnosisName
// ─────────────────────────────────────────────────

export function useGetDiagnosisNames() {
  return useQuery({
    queryKey: queryKeys.masters.category("diagnosis-names"),
    queryFn: listDiagnosisNames,
    staleTime: QUERY_STALE_TIMES.STATIC, // マスタデータ: 30分キャッシュ
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateDiagnosisName() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createDiagnosisName,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("diagnosis-names") });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
}

export function useUpdateDiagnosisName() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateDiagnosisNameRequest }) =>
      updateDiagnosisName(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("diagnosis-names") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteDiagnosisName() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteDiagnosisName,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("diagnosis-names") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}

export function useReorderDiagnosisNames() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderDiagnosisNames,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("diagnosis-names") });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
}
