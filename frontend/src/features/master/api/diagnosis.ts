import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type {
  DiagnosisCategory as ModelDiagnosisCategory,
  DiagnosisName as ModelDiagnosisName,
} from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Transform functions → domain types (camelCase)
// ─────────────────────────────────────────────────

function transformDiagnosisCategory(data: ModelDiagnosisCategory) {
  return {
    id: data.id,                        // number (uint64)
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
    id: data.id,                                // number (uint64)
    clinicId: data.clinic_id,
    name: data.name,
    isActive: data.is_active,
    description: data.description,
    diagnosisCategoryId: data.diagnosis_category_id, // number (uint64)
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

// Frontend domain types derived from transforms (編集禁止 - ReturnType から自動導出)
export type DiagnosisCategory = ReturnType<typeof transformDiagnosisCategory>;
export type DiagnosisName = ReturnType<typeof transformDiagnosisName>;

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export interface CreateDiagnosisCategoryRequest {
  name: string;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateDiagnosisCategoryRequest {
  name?: string;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface CreateDiagnosisNameRequest {
  name: string;
  diagnosis_category_id: number; // uint64 — string で送ると 400 になるため number 必須
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateDiagnosisNameRequest {
  name?: string;
  diagnosis_category_id?: number; // uint64
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface ReorderDiagnosisCategoryRequest {
  ids: number[];
}

export interface ReorderDiagnosisNameRequest {
  ids: number[];
}

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const DIAGNOSIS_CATEGORIES_KEY = ["masters", "diagnosis-categories"] as const;
const DIAGNOSIS_NAMES_KEY = ["masters", "diagnosis-names"] as const;

// ─────────────────────────────────────────────────
// API functions - DiagnosisCategory
// ─────────────────────────────────────────────────

export async function listDiagnosisCategories(): Promise<DiagnosisCategory[]> {
  const { data } = await axios.get<ModelDiagnosisCategory[]>(
    "/v1/masters/diagnosis-categories",
  );
  return data.map(transformDiagnosisCategory);
}

export async function createDiagnosisCategory(
  req: CreateDiagnosisCategoryRequest,
): Promise<DiagnosisCategory> {
  const { data } = await axios.post<ModelDiagnosisCategory>(
    "/v1/masters/diagnosis-categories",
    req,
  );
  return transformDiagnosisCategory(data);
}

export async function updateDiagnosisCategory(
  id: number,
  req: UpdateDiagnosisCategoryRequest,
): Promise<DiagnosisCategory> {
  const { data } = await axios.patch<ModelDiagnosisCategory>(
    `/v1/masters/diagnosis-categories/${id}`,
    req,
  );
  return transformDiagnosisCategory(data);
}

export async function deleteDiagnosisCategory(id: number): Promise<void> {
  await axios.delete(`/v1/masters/diagnosis-categories/${id}`);
}

export async function reorderDiagnosisCategories(
  req: ReorderDiagnosisCategoryRequest,
): Promise<void> {
  await axios.patch("/v1/masters/diagnosis-categories/reorder", req);
}

// ─────────────────────────────────────────────────
// API functions - DiagnosisName
// ─────────────────────────────────────────────────

export async function listDiagnosisNames(): Promise<DiagnosisName[]> {
  const { data } = await axios.get<ModelDiagnosisName[]>(
    "/v1/masters/diagnosis-names",
  );
  return data.map(transformDiagnosisName);
}

export async function createDiagnosisName(
  req: CreateDiagnosisNameRequest,
): Promise<DiagnosisName> {
  const { data } = await axios.post<ModelDiagnosisName>(
    "/v1/masters/diagnosis-names",
    req,
  );
  return transformDiagnosisName(data);
}

export async function updateDiagnosisName(
  id: number,
  req: UpdateDiagnosisNameRequest,
): Promise<DiagnosisName> {
  const { data } = await axios.patch<ModelDiagnosisName>(
    `/v1/masters/diagnosis-names/${id}`,
    req,
  );
  return transformDiagnosisName(data);
}

export async function deleteDiagnosisName(id: number): Promise<void> {
  await axios.delete(`/v1/masters/diagnosis-names/${id}`);
}

export async function reorderDiagnosisNames(
  req: ReorderDiagnosisNameRequest,
): Promise<void> {
  await axios.patch("/v1/masters/diagnosis-names/reorder", req);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks - DiagnosisCategory
// ─────────────────────────────────────────────────

export function useListDiagnosisCategories() {
  return useQuery({
    queryKey: DIAGNOSIS_CATEGORIES_KEY,
    queryFn: listDiagnosisCategories,
    staleTime: QUERY_STALE_TIMES.STATIC, // マスタデータ: 30分キャッシュ
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateDiagnosisCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createDiagnosisCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DIAGNOSIS_CATEGORIES_KEY });
    },
  });
}

export function useUpdateDiagnosisCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: number; req: UpdateDiagnosisCategoryRequest }) =>
      updateDiagnosisCategory(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DIAGNOSIS_CATEGORIES_KEY });
    },
  });
}

export function useDeleteDiagnosisCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteDiagnosisCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DIAGNOSIS_CATEGORIES_KEY });
    },
  });
}

export function useReorderDiagnosisCategories() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderDiagnosisCategories,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DIAGNOSIS_CATEGORIES_KEY });
    },
  });
}

// ─────────────────────────────────────────────────
// TanStack Query hooks - DiagnosisName
// ─────────────────────────────────────────────────

export function useListDiagnosisNames() {
  return useQuery({
    queryKey: DIAGNOSIS_NAMES_KEY,
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
      queryClient.invalidateQueries({ queryKey: DIAGNOSIS_NAMES_KEY });
    },
  });
}

export function useUpdateDiagnosisName() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: number; req: UpdateDiagnosisNameRequest }) =>
      updateDiagnosisName(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DIAGNOSIS_NAMES_KEY });
    },
  });
}

export function useDeleteDiagnosisName() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteDiagnosisName,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DIAGNOSIS_NAMES_KEY });
    },
  });
}

export function useReorderDiagnosisNames() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderDiagnosisNames,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DIAGNOSIS_NAMES_KEY });
    },
  });
}
