import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type {
  DiagnosisCategory as ModelDiagnosisCategory,
  DiagnosisName as ModelDiagnosisName,
} from "@/types/generated/models";
import type {
  CreateDiagnosisCategoryRequest,
  UpdateDiagnosisCategoryRequest,
  CreateDiagnosisNameRequest,
  UpdateDiagnosisNameRequest,
  ReorderDiagnosisCategoryRequest,
  ReorderDiagnosisNameRequest,
} from "@/types/diagnosis";

// ─────────────────────────────────────────────────
// Transform functions → domain types (camelCase)
// ─────────────────────────────────────────────────

function transformDiagnosisCategory(data: ModelDiagnosisCategory) {
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
    diagnosisCategoryId: String(data.diagnosis_category_id ?? 0),
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

// Frontend domain types derived from transforms (編集禁止 - ReturnType から自動導出)
export type DiagnosisCategory = ReturnType<typeof transformDiagnosisCategory>;
export type DiagnosisName = ReturnType<typeof transformDiagnosisName>;

// ─────────────────────────────────────────────────
// Request types re-export (@/types/diagnosis から導出済み)
// ─────────────────────────────────────────────────

export type {
  CreateDiagnosisCategoryRequest,
  UpdateDiagnosisCategoryRequest,
  CreateDiagnosisNameRequest,
  UpdateDiagnosisNameRequest,
  ReorderDiagnosisCategoryRequest,
  ReorderDiagnosisNameRequest,
} from "@/types/diagnosis";

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
  id: string,
  req: UpdateDiagnosisCategoryRequest,
): Promise<DiagnosisCategory> {
  const { data } = await axios.patch<ModelDiagnosisCategory>(
    `/v1/masters/diagnosis-categories/${id}`,
    req,
  );
  return transformDiagnosisCategory(data);
}

export async function deleteDiagnosisCategory(id: string): Promise<void> {
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
  id: string,
  req: UpdateDiagnosisNameRequest,
): Promise<DiagnosisName> {
  const { data } = await axios.patch<ModelDiagnosisName>(
    `/v1/masters/diagnosis-names/${id}`,
    req,
  );
  return transformDiagnosisName(data);
}

export async function deleteDiagnosisName(id: string): Promise<void> {
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
    mutationFn: ({ id, req }: { id: string; req: UpdateDiagnosisCategoryRequest }) =>
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
    mutationFn: ({ id, req }: { id: string; req: UpdateDiagnosisNameRequest }) =>
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
