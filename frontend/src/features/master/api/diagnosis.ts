import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

// ─────────────────────────────────────────────────
// Backend types (snake_case) - api.yaml 準拠
// ─────────────────────────────────────────────────

export interface BackendDiagnosisCategory {
  id: string;
  clinic_id: string;
  name: string;
  is_active: boolean;
  description: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface BackendDiagnosisName {
  id: string;
  clinic_id: string;
  name: string;
  is_active: boolean;
  description: string;
  diagnosis_category_id: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

// ─────────────────────────────────────────────────
// Frontend display types (camelCase)
// ─────────────────────────────────────────────────

export interface DiagnosisCategory {
  id: string;
  clinicId: string;
  name: string;
  isActive: boolean;
  description: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

export interface DiagnosisName {
  id: string;
  clinicId: string;
  name: string;
  isActive: boolean;
  description: string;
  diagnosisCategoryId: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
}

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
  diagnosis_category_id: string;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateDiagnosisNameRequest {
  name?: string;
  diagnosis_category_id?: string;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

// ─────────────────────────────────────────────────
// Transform functions
// ─────────────────────────────────────────────────

function transformDiagnosisCategory(data: BackendDiagnosisCategory): DiagnosisCategory {
  return {
    id: data.id,
    clinicId: data.clinic_id,
    name: data.name,
    isActive: data.is_active,
    description: data.description,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

function transformDiagnosisName(data: BackendDiagnosisName): DiagnosisName {
  return {
    id: data.id,
    clinicId: data.clinic_id,
    name: data.name,
    isActive: data.is_active,
    description: data.description,
    diagnosisCategoryId: data.diagnosis_category_id,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
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
  const { data } = await axios.get<BackendDiagnosisCategory[]>(
    "/v1/masters/diagnosis-categories",
  );
  return data.map(transformDiagnosisCategory);
}

export async function createDiagnosisCategory(
  req: CreateDiagnosisCategoryRequest,
): Promise<DiagnosisCategory> {
  const { data } = await axios.post<BackendDiagnosisCategory>(
    "/v1/masters/diagnosis-categories",
    req,
  );
  return transformDiagnosisCategory(data);
}

export async function updateDiagnosisCategory(
  id: string,
  req: UpdateDiagnosisCategoryRequest,
): Promise<DiagnosisCategory> {
  const { data } = await axios.patch<BackendDiagnosisCategory>(
    `/v1/masters/diagnosis-categories/${id}`,
    req,
  );
  return transformDiagnosisCategory(data);
}

export async function deleteDiagnosisCategory(id: string): Promise<void> {
  await axios.delete(`/v1/masters/diagnosis-categories/${id}`);
}

// ─────────────────────────────────────────────────
// API functions - DiagnosisName
// ─────────────────────────────────────────────────

export async function listDiagnosisNames(): Promise<DiagnosisName[]> {
  const { data } = await axios.get<BackendDiagnosisName[]>(
    "/v1/masters/diagnosis-names",
  );
  return data.map(transformDiagnosisName);
}

export async function createDiagnosisName(
  req: CreateDiagnosisNameRequest,
): Promise<DiagnosisName> {
  const { data } = await axios.post<BackendDiagnosisName>(
    "/v1/masters/diagnosis-names",
    req,
  );
  return transformDiagnosisName(data);
}

export async function updateDiagnosisName(
  id: string,
  req: UpdateDiagnosisNameRequest,
): Promise<DiagnosisName> {
  const { data } = await axios.patch<BackendDiagnosisName>(
    `/v1/masters/diagnosis-names/${id}`,
    req,
  );
  return transformDiagnosisName(data);
}

export async function deleteDiagnosisName(id: string): Promise<void> {
  await axios.delete(`/v1/masters/diagnosis-names/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks - DiagnosisCategory
// ─────────────────────────────────────────────────

export function useListDiagnosisCategories() {
  return useQuery({
    queryKey: DIAGNOSIS_CATEGORIES_KEY,
    queryFn: listDiagnosisCategories,
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

// ─────────────────────────────────────────────────
// TanStack Query hooks - DiagnosisName
// ─────────────────────────────────────────────────

export function useListDiagnosisNames() {
  return useQuery({
    queryKey: DIAGNOSIS_NAMES_KEY,
    queryFn: listDiagnosisNames,
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
